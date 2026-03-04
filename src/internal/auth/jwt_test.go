package auth

import (
	"testing"
	"time"
)

type fakeBlacklist struct{ m map[string]bool }

func (f *fakeBlacklist) ExistsJTI(jti string) (bool, error) { return f.m[jti], nil }

func TestJWTSignVerify(t *testing.T) {
	s := &JWTService{Secret: []byte("abcdefghijklmnopqrstuvwxyz123456"), AccessTokenTTL: 15 * time.Minute}
	tok, jti, err := s.SignAccessToken("u1", "Admin")
	if err != nil || jti == "" {
		t.Fatalf("sign failed err=%v jti=%s", err, jti)
	}
	c, err := s.VerifyAccessToken(tok)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if c.Subject != "u1" || c.Role != "Admin" || c.ID == "" {
		t.Fatalf("claims mismatch: %+v", c)
	}
}

func TestJWTExpiredAndInvalidAndRevoked(t *testing.T) {
	s := &JWTService{Secret: []byte("abcdefghijklmnopqrstuvwxyz123456"), AccessTokenTTL: time.Nanosecond}
	tok, _, _ := s.SignAccessToken("u1", "Viewer")
	time.Sleep(time.Millisecond)
	if _, err := s.VerifyAccessToken(tok); err != ErrTokenExpired {
		t.Fatalf("expect expired, got %v", err)
	}
	if _, err := s.VerifyAccessToken("bad"); err != ErrTokenInvalid {
		t.Fatalf("expect invalid, got %v", err)
	}

	s2 := &JWTService{Secret: []byte("abcdefghijklmnopqrstuvwxyz123456"), AccessTokenTTL: time.Minute}
	t2, j2, _ := s2.SignAccessToken("u2", "Viewer")
	s2.BlacklistChecker = &fakeBlacklist{m: map[string]bool{j2: true}}
	if _, err := s2.VerifyAccessToken(t2); err != ErrTokenRevoked {
		t.Fatalf("expect revoked, got %v", err)
	}
}
