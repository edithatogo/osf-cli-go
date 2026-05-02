package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestLoadTokenReturnsToken(t *testing.T) {
	token, err := LoadToken(FuncSource(func(name string) (string, bool) {
		if name == TokenEnv {
			return "osf_live_token_1234567890abcdef", true
		}
		return "", false
	}))

	if err != nil {
		t.Fatalf("LoadToken returned error: %v", err)
	}
	if got, want := token, "osf_live_token_1234567890abcdef"; got != want {
		t.Fatalf("LoadToken returned %q, want %q", got, want)
	}
}

func TestLoadTokenMissing(t *testing.T) {
	_, err := LoadToken(FuncSource(func(name string) (string, bool) {
		return "", false
	}))

	var missing MissingTokenError
	if !errors.As(err, &missing) {
		t.Fatalf("LoadToken error = %v, want MissingTokenError", err)
	}
	if got := err.Error(); strings.Contains(got, "osf_live_token") {
		t.Fatalf("missing-token error leaked token text: %q", got)
	}
}

func TestLoadTokenTrimsWhitespace(t *testing.T) {
	token, err := LoadToken(FuncSource(func(name string) (string, bool) {
		if name == TokenEnv {
			return " \t  osf_live_token_1234567890abcdef  \n", true
		}
		return "", false
	}))

	if err != nil {
		t.Fatalf("LoadToken returned error: %v", err)
	}
	if got, want := token, "osf_live_token_1234567890abcdef"; got != want {
		t.Fatalf("LoadToken returned %q, want %q", got, want)
	}
}

func TestLoadTokenWhitespaceOnlyIsMissing(t *testing.T) {
	_, err := LoadToken(FuncSource(func(name string) (string, bool) {
		if name == TokenEnv {
			return " \t \n ", true
		}
		return "", false
	}))

	if err == nil {
		t.Fatal("LoadToken returned nil error, want missing-token error")
	}
	var missing MissingTokenError
	if !errors.As(err, &missing) {
		t.Fatalf("LoadToken error = %v, want MissingTokenError", err)
	}
}

func TestRedactRemovesTokenLikeValues(t *testing.T) {
	token := "osf_live_token_1234567890abcdef"
	raw := "request failed for token=" + token + " with Authorization: Bearer " + token

	got := Redact(raw, token)

	if strings.Contains(got, token) {
		t.Fatalf("Redact leaked token: %q", got)
	}
	if !strings.Contains(got, redacted) {
		t.Fatalf("Redact output = %q, want redaction marker", got)
	}
}

func TestRedactError(t *testing.T) {
	token := "osf_live_token_1234567890abcdef"
	err := RedactError(errors.New("bad request: "+token), token)

	if err == nil {
		t.Fatal("RedactError returned nil")
	}
	if strings.Contains(err.Error(), token) {
		t.Fatalf("RedactError leaked token: %q", err.Error())
	}
}
