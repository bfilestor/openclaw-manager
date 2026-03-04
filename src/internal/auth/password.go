package auth

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordEmpty    = errors.New("password empty")
	ErrPasswordTooShort = errors.New("password too short")
	ErrPasswordWeak     = errors.New("password weak")
)

type PasswordService struct {
	Cost         int
	MinLength    int
	MaxLength    int
}

func NewPasswordService() *PasswordService {
	return &PasswordService{Cost: 12, MinLength: 8, MaxLength: 128}
}

func (s *PasswordService) Hash(plain string) (string, error) {
	if err := s.ValidateStrength(plain); err != nil {
		return "", err
	}
	b, err := bcrypt.GenerateFromPassword([]byte(plain), s.Cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *PasswordService) Verify(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func (s *PasswordService) ValidateStrength(plain string) error {
	if plain == "" {
		return ErrPasswordEmpty
	}
	if len(plain) < s.MinLength {
		return ErrPasswordTooShort
	}
	if len(plain) > s.MaxLength {
		return ErrPasswordWeak
	}

	hasLetter, hasDigit := false, false
	for _, r := range plain {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return ErrPasswordWeak
	}
	return nil
}
