// Package cryptox 提供数据加密原语, 用于保护敏感字段 (Kubeconfig / OAuth Secret 等)。
//
// 设计原则:
//
//   - 加密算法: AES-256-GCM (认证加密, 同时保证机密性 + 完整性)
//   - 密钥长度: 32 字节 (256 位), 由调用方从 envVault / KMS / 环境变量传入
//   - Nonce 随机: 每次加密都用 crypto/rand 生成新 nonce
//   - 密文格式: nonce(12B) || ciphertext || gcm_tag(16B), 全部连在一起存
//   - 不持久化密钥: 进程退出后 DEK 随 GC 释放
//
// 不放这里的东西:
//
//   - 密钥加载 (envVault / KMS 集成) — 那是 loader 子包的事
//   - 字段级 schema / migration — 那是业务 model 的事
package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// ErrInvalidKey 密钥长度不合法 (不是 32 字节)。
var ErrInvalidKey = errors.New("cryptox: key must be 32 bytes (AES-256)")

// nonceSize GCM nonce 长度, 12 字节是 GCM 标准.
const nonceSize = 12

// Service 是加密服务, 持有 DEK, 提供 Encrypt / Decrypt.
type Service struct {
	gcm cipher.AEAD
}

// New 用 32 字节密钥构造加密服务.
//
// TODO 验证密钥长度 (如果不是 32 字节返回 ErrInvalidKey).
// 推荐用 crypto/subtle 做常量时间比较防御侧信道, 这里只校验长度, 业务调一次即可.
func New(key []byte) (*Service, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidKey, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	return &Service{gcm: gcm}, nil
}

// Encrypt 加密明文, 返回 nonce||ciphertext||tag 的拼接.
//
// 输出长度 = len(plaintext) + nonceSize + 16 (gcm tag).
// TODO:
//
//	data := s.Encrypt(plaintext)
//	// 存到 DB 的 bytea 字段
func (s *Service) Encrypt(plaintext []byte) ([]byte, error) {
	// TODO 实现:
	//   1. 分配 nonceSize 字节 buffer
	//   2. io.ReadFull(rand.Reader, nonce)  // 随机 nonce
	//   3. ciphertext := s.gcm.Seal(nil, nonce, plaintext, nil)
	//   4. return append(nonce, ciphertext...), nil
	_ = plaintext
	return nil, fmt.Errorf("not implemented")
}

// Decrypt 解密密文 (Encrypt 的逆操作), 输入格式 = nonce||ciphertext||tag.
//
// TODO:
//
//	plaintext, err := s.Decrypt(ciphertext)
//	if errors.Is(err, ErrAuthFailure) { /* GCM 认证失败, 密钥错或数据被篡改 */ }
func (s *Service) Decrypt(ciphertext []byte) ([]byte, error) {
	// TODO 实现:
	//   1. if len(ciphertext) < nonceSize+16 { return nil, errors.New("ciphertext too short") }
	//   2. nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	//   3. return s.gcm.Open(nil, nonce, ct, nil)
	_ = ciphertext
	return nil, fmt.Errorf("not implemented")
}

// EncryptString / DecryptString 字符串便利函数 (业务层多在用 string).
func (s *Service) EncryptString(plaintext string) ([]byte, error) {
	return s.Encrypt([]byte(plaintext))
}

func (s *Service) DecryptString(ciphertext []byte) (string, error) {
	pt, err := s.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

// 编译期防 io/rand import 未用.
var _ = io.ReadFull
var _ = rand.Read
