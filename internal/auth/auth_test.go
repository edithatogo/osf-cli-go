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

func TestRedactHandlesEmptyInput(t *testing.T) {
	if got := Redact(""); got != "" {
		t.Fatalf("Redact empty input = %q, want empty string", got)
	}
}

func TestRedactWithEmptySecret(t *testing.T) {
	got := Redact("token=secret", "")
	if !strings.Contains(got, "token=secret") {
		t.Fatalf("Redact removed content despite empty secret: %q", got)
	}
}

func TestRedactEnvAssignment(t *testing.T) {
	token := "osf_live_token_abc123def456ghi789"
	raw := "export OSF_TOKEN=" + token

	got := Redact(raw, token)

	if strings.Contains(got, token) {
		t.Fatalf("Redact leaked token: %q", got)
	}
	if !strings.Contains(got, redacted) {
		t.Fatalf("Redact output = %q, want redaction marker", got)
	}
}

func TestRedactBearerHeader(t *testing.T) {
	got := Redact("Authorization: Bearer osf_live_token_abc123def456ghi789xyz")

	if strings.Contains(got, "osf_live_token_abc123def456ghi789xyz") {
		t.Fatalf("Redact leaked bearer token: %q", got)
	}
	if !strings.Contains(got, "Bearer "+redacted) {
		t.Fatalf("Redact output = %q, want Bearer [REDACTED]", got)
	}
}

func TestRedactErrorNil(t *testing.T) {
	if got := RedactError(nil); got != nil {
		t.Fatalf("RedactError(nil) = %v, want nil", got)
	}
}

func TestMissingTokenErrorEmptyEnv(t *testing.T) {
	err := MissingTokenError{Env: ""}
	msg := err.Error()
	if !strings.Contains(msg, TokenEnv) {
		t.Fatalf("MissingTokenError.Error() = %q, want reference to %s", msg, TokenEnv)
	}
}

func TestEnvSourceLookup(t *testing.T) {
	t.Setenv(TokenEnv, "test-token-value-12345")
	source := EnvSource{}
	token, ok := source.Lookup(TokenEnv)
	if !ok {
		t.Fatal("EnvSource.Lookup returned false, want true")
	}
	if token != "test-token-value-12345" {
		t.Fatalf("EnvSource.Lookup = %q, want %q", token, "test-token-value-12345")
	}
	_, ok = source.Lookup("NONEXISTENT_VAR_12345")
	if ok {
		t.Fatal("EnvSource.Lookup for nonexistent var returned true, want false")
	}
}

func TestRedactRemovesURLEncodedToken(t *testing.T) {
	token := "my_api_token_12345"
	raw := "Bearer " + token + "&next=/v2/nodes/"
	got := Redact(raw, token)
	if strings.Contains(got, token) {
		t.Fatalf("Redact leaked token: %q", got)
	}
	if !strings.Contains(got, redacted) {
		t.Fatalf("Redact output = %q, want redaction marker", got)
	}
}

func TestRedactRemovesMultilineToken(t *testing.T) {
	token := "abc123def456ghi789jkl012"
	raw := "line1\nBearer " + token + "\nline3"
	got := Redact(raw, token)
	if strings.Contains(got, token) {
		t.Fatalf("Redact leaked token across lines: %q", got)
	}
	if !strings.Contains(got, redacted) {
		t.Fatalf("Redact output = %q, want redaction marker", got)
	}
}

func TestRedactWithPartialTokenMatch(t *testing.T) {
	text := "OSF_TOKEN=super_secret_token_12345"
	got := Redact(text)
	if strings.Contains(got, "super_secret_token_12345") {
		t.Fatalf("Redact leaked auto-detected token assignment: %q", got)
	}
	if !strings.Contains(got, redacted) {
		t.Fatalf("Redact output = %q, want redaction marker", got)
	}
}

func TestRedactErrorWithMultipleSecrets(t *testing.T) {
	token1 := "tok_12345"
	token2 := "sec_67890"
	err := RedactError(errors.New(token1+" and "+token2), token1, token2)
	if err == nil {
		t.Fatal("RedactError returned nil")
	}
	msg := err.Error()
	if strings.Contains(msg, token1) || strings.Contains(msg, token2) {
		t.Fatalf("RedactError leaked secrets: %q", msg)
	}
}

func TestLoadTokenEmptyEnvValue(t *testing.T) {
	_, err := LoadToken(FuncSource(func(name string) (string, bool) {
		return "", true
	}))
	if err == nil {
		t.Fatal("LoadToken returned nil error, want missing-token error for empty value")
	}
}

func TestRedactDoesNotCorruptNormalText(t *testing.T) {
	got := Redact("hello world this is normal text")
	if got != "hello world this is normal text" {
		t.Fatalf("Redact corrupted normal text: %q", got)
	}
}

func TestRedactWithShortSecret(t *testing.T) {
	got := Redact("key=abc", "abc")
	if !strings.Contains(got, "key=[REDACTED]") {
		t.Fatalf("Redact = %q, want redacted short secret", got)
	}
}

func TestLoadTokenWithNilSource(t *testing.T) {
	t.Setenv("OSF_TOKEN", "nil-source-token-12345")
	token, err := LoadToken(nil)
	if err != nil {
		t.Fatalf("LoadToken(nil) returned error: %v", err)
	}
	if token != "nil-source-token-12345" {
		t.Fatalf("LoadToken(nil) = %q, want %q", token, "nil-source-token-12345")
	}
}

func TestLoadCredentialsPrefersToken(t *testing.T) {
	credentials, err := LoadCredentials(FuncSource(func(name string) (string, bool) {
		switch name {
		case TokenEnv:
			return "token-123", true
		case UsernameEnv:
			return "user@example.org", true
		case PasswordEnv:
			return "password-123", true
		default:
			return "", false
		}
	}))
	if err != nil {
		t.Fatalf("LoadCredentials returned error: %v", err)
	}
	if credentials.Mode != ModeBearerToken || credentials.Token != "token-123" {
		t.Fatalf("credentials = %+v, want bearer token", credentials)
	}
}

func TestLoadCredentialsUsernamePasswordFallback(t *testing.T) {
	credentials, err := LoadCredentials(FuncSource(func(name string) (string, bool) {
		switch name {
		case UsernameEnv:
			return " user@example.org ", true
		case PasswordEnv:
			return " password-123 ", true
		default:
			return "", false
		}
	}))
	if err != nil {
		t.Fatalf("LoadCredentials returned error: %v", err)
	}
	if credentials.Mode != ModeUsernamePassword || credentials.Username != "user@example.org" || credentials.Password != "password-123" {
		t.Fatalf("credentials = %+v, want username/password", credentials)
	}
}

func TestLoadCredentialsPartialUsernamePassword(t *testing.T) {
	_, err := LoadCredentials(FuncSource(func(name string) (string, bool) {
		if name == UsernameEnv {
			return "user@example.org", true
		}
		return "", false
	}))
	if err == nil {
		t.Fatal("LoadCredentials returned nil error, want partial credential error")
	}
	var missing MissingCredentialsError
	if !errors.As(err, &missing) {
		t.Fatalf("error = %T %v, want MissingCredentialsError", err, err)
	}
	if !strings.Contains(err.Error(), PasswordEnv) {
		t.Fatalf("error = %q, want password env mention", err.Error())
	}
}

func TestCredentialsSecrets(t *testing.T) {
	credentials := Credentials{Mode: ModeUsernamePassword, Username: "user@example.org", Password: "password-123"}
	secrets := credentials.Secrets()
	if len(secrets) != 2 || secrets[0] != "user@example.org" || secrets[1] != "password-123" {
		t.Fatalf("Secrets = %#v", secrets)
	}
}

func TestRedactBasicAuthAndPasswordEnv(t *testing.T) {
	raw := "Authorization: Basic dXNlckBleGFtcGxlLm9yZzpwYXNz OSF_PASSWORD=password-123"
	got := Redact(raw, "password-123")
	if strings.Contains(got, "dXNlckBleGFtcGxlLm9yZzpwYXNz") || strings.Contains(got, "password-123") {
		t.Fatalf("Redact leaked basic auth/password: %q", got)
	}
	if !strings.Contains(got, "Basic "+redacted) || !strings.Contains(got, "OSF_PASSWORD="+redacted) {
		t.Fatalf("Redact = %q, want basic and env redaction", got)
	}
}

func TestRedactCallbackWithAlreadyRedacted(t *testing.T) {
	got := Redact(redacted + " and " + redacted)
	if got != redacted+" and "+redacted {
		t.Fatalf("Redact corrupted already redacted text: %q", got)
	}
}

func TestRedactLongTokenLikeValue(t *testing.T) {
	longToken := "abcdefghijklmnopqrstuvwx1234567890"
	got := Redact("token is "+longToken, longToken)
	if strings.Contains(got, longToken) {
		t.Fatalf("Redact leaked long token: %q", got)
	}
}

func TestAPIErrorFmt(t *testing.T) {
	err := MissingTokenError{Env: TokenEnv}
	msg := err.Error()
	if !strings.Contains(msg, TokenEnv) {
		t.Fatalf("MissingTokenError.Error() = %q, want %s", msg, TokenEnv)
	}
}

func TestFuncSourceLookup(t *testing.T) {
	source := FuncSource(func(name string) (string, bool) {
		if name == "MY_KEY" {
			return "my-value", true
		}
		return "", false
	})
	v, ok := source.Lookup("MY_KEY")
	if !ok || v != "my-value" {
		t.Fatalf("FuncSource.Lookup = %q, %v, want %q, true", v, ok, "my-value")
	}
	v, ok = source.Lookup("UNKNOWN")
	if ok {
		t.Fatalf("FuncSource.Lookup for unknown = %q, %v, want \"\", false", v, ok)
	}
}
