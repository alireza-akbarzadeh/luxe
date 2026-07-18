package utils

import (
	"strings"
	"testing"
)

func TestPasswordResetEmail(t *testing.T) {
	subject, body := PasswordResetEmail("https://luxe.example", "tok-abc")
	if subject == "" {
		t.Fatal("expected subject")
	}
	checks := []string{
		"Reset your password",
		"https://luxe.example/reset-password?token=tok-abc",
		"https://luxe.example/apple-touch-icon.png",
		"#c9a96e",
		"1 hour",
	}
	for _, want := range checks {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q", want)
		}
	}
}

func TestVerificationEmail(t *testing.T) {
	subject, body := VerificationEmail("https://luxe.example/", "verify-1")
	if !strings.Contains(subject, "Verify") {
		t.Fatalf("unexpected subject: %s", subject)
	}
	if !strings.Contains(body, "https://luxe.example/verify-email?token=verify-1") {
		t.Fatal("missing verify URL")
	}
}
