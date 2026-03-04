package security

import (
	"strings"
	"testing"
	"time"

	"openclaw-manager/internal/auth"
	"openclaw-manager/internal/storage"
)

func TestPathTraversalRejected(t *testing.T) {
	v, _ := storage.NewPathValidator([]string{"/tmp/allowed"})
	if _, err := v.Validate("/tmp/allowed/../../etc/passwd"); err == nil {
		t.Fatalf("expected traversal to be rejected")
	}
}

func TestJWTTamperRejected(t *testing.T) {
	svc := &auth.JWTService{Secret: []byte("abcdefghijklmnopqrstuvwxyz123456"), AccessTokenTTL: time.Hour}
	tok, _, err := svc.SignAccessToken("u1", "Viewer")
	if err != nil { t.Fatal(err) }
	parts := strings.Split(tok, ".")
	if len(parts) != 3 { t.Fatalf("invalid token") }
	tampered := parts[0] + "." + parts[1] + ".abcdef"
	if _, err := svc.VerifyAccessToken(tampered); err == nil {
		t.Fatalf("tampered token should fail")
	}
}
