package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHashAndVerify(t *testing.T) {
	s := NewPasswordService()
	h1, err := s.Hash("Pass1234")
	if err != nil {
		t.Fatal(err)
	}
	h2, err := s.Hash("Pass1234")
	if err != nil {
		t.Fatal(err)
	}
	if h1 == "Pass1234" || h2 == "Pass1234" {
		t.Fatal("hash must not equal plain")
	}
	if h1 == h2 {
		t.Fatal("same password should generate different hash")
	}
	if !s.Verify("Pass1234", h1) || s.Verify("wrong", h1) {
		t.Fatal("verify mismatch")
	}
	cost, err := bcrypt.Cost([]byte(h1))
	if err != nil || cost != 12 {
		t.Fatalf("expected cost=12 got=%d err=%v", cost, err)
	}
}

func TestPasswordStrength(t *testing.T) {
	s := NewPasswordService()
	cases := []struct {
		p       string
		expects error
	}{
		{"", ErrPasswordEmpty},
		{"abc123", ErrPasswordTooShort},
		{"12345678", ErrPasswordWeak},
		{"abcdefgh", ErrPasswordWeak},
		{"Pass1234", nil},
	}
	for _, c := range cases {
		err := s.ValidateStrength(c.p)
		if err != c.expects {
			t.Fatalf("password=%q expect=%v got=%v", c.p, c.expects, err)
		}
	}
}
