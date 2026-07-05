package auth

import "golang.org/x/crypto/bcrypt"

// BcryptHasher 是 PasswordHasher 的默认实现。
type BcryptHasher struct {
	Cost int
}

// 编译期断言: BcryptHasher 必须实现 PasswordHasher interface.
var _ PasswordHasher = (*BcryptHasher)(nil)

// NewBcryptHasher 构造哈希器。cost <= 0 时使用 bcrypt 默认值 10。
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{Cost: cost}
}

// Hash 生成密码哈希。
func (b *BcryptHasher) Hash(plain string) (string, error) {
	bs, err := bcrypt.GenerateFromPassword([]byte(plain), b.Cost)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

// Verify 校验密码;校验失败时返回 ErrInvalidPassword。
func (b *BcryptHasher) Verify(plain, hashed string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)); err != nil {
		return ErrInvalidPassword
	}
	return nil
}
