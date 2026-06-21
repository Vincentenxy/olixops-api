package cryptox

import (
	"bytes"
	"errors"
	"testing"
)

// 测试密钥: 必须正好 32 字节
var testKey = []byte("01234567890123456789012345678901") // 32 字节

func TestService_InvalidKeyLength(t *testing.T) {
	// TODO: 测三种情况
	//   1. 16 字节 → ErrInvalidKey
	//   2. 32 字节 (正确) → nil
	//   3. 64 字节 → ErrInvalidKey
	cases := []struct {
		name    string
		keyLen  int
		wantErr bool
	}{
		{"16 bytes too short", 16, true},
		{"32 bytes correct", 32, false},
		{"64 bytes too long", 64, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := bytes.Repeat([]byte{0xAB}, tc.keyLen)
			_, err := New(key)
			if (err != nil) != tc.wantErr {
				t.Errorf("err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestService_EncryptDecryptRoundTrip(t *testing.T) {
	// TODO 实现完后才能跑通这个测试
	t.Skip("TODO: 等 Encrypt/Decrypt 实现完成")

	s, err := New(testKey)
	if err != nil {
		t.Fatal(err)
	}
	plaintexts := []string{
		"",
		"hello world",
		"kubeconfig 内容: apiVersion: v1\nclusters: ...",
		"中文 + emoji 🦀 测试",
		string(bytes.Repeat([]byte("A"), 4096)), // 4KB
	}
	for _, pt := range plaintexts {
		t.Run(pt, func(t *testing.T) {
			ct, err := s.EncryptString(pt)
			if err != nil {
				t.Fatal(err)
			}
			if pt != "" && bytes.Contains(ct, []byte(pt)) {
				t.Error("ciphertext leaks plaintext!")
			}
			got, err := s.DecryptString(ct)
			if err != nil {
				t.Fatal(err)
			}
			if got != pt {
				t.Errorf("got %q, want %q", got, pt)
			}
		})
	}
}

func TestService_NonceUniqueness(t *testing.T) {
	// TODO: 同一明文 + 同一密钥, 两次加密结果必须不同 (因为 nonce 随机)
	t.Skip("TODO")
	s, _ := New(testKey)
	a, _ := s.EncryptString("same plaintext")
	b, _ := s.EncryptString("same plaintext")
	if bytes.Equal(a, b) {
		t.Error("two encryptions of same plaintext should differ (nonce randomness)")
	}
}

func TestService_DecryptTamperedCiphertext(t *testing.T) {
	// TODO: 改 1 字节密文, 解密应该返回 GCM 认证失败
	t.Skip("TODO")
	s, _ := New(testKey)
	ct, _ := s.EncryptString("hello")
	ct[20] ^= 0x01 // 翻转 1 比特
	_, err := s.DecryptString(ct)
	if !errors.Is(err /* ErrAuthFailure, */, nil) {
		// 期待非 nil
	}
	if err == nil {
		t.Error("decrypt of tampered ciphertext should fail")
	}
}
