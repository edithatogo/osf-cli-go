package auth

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// TokenEnv is the environment variable that stores the OSF personal access token.
const TokenEnv = "OSF_TOKEN"

// UsernameEnv is the environment variable that stores the OSF username/email.
const UsernameEnv = "OSF_USERNAME"

// PasswordEnv is the environment variable that stores the OSF password.
const PasswordEnv = "OSF_PASSWORD"

const redacted = "[REDACTED]"

var tokenLikePattern = regexp.MustCompile(`(?i)\b[A-Za-z0-9_-]{24,}\b`)
var bearerPattern = regexp.MustCompile(`(?i)\b(Bearer\s+)([A-Za-z0-9._~+/=-]{8,})\b`)
var basicPattern = regexp.MustCompile(`(?i)\b(Basic\s+)([A-Za-z0-9+/=]{8,})\b`)
var envAssignmentPattern = regexp.MustCompile(`(?i)\b((?:OSF_TOKEN|OSF_USERNAME|OSF_PASSWORD)\s*=\s*)([^\s]+)\b`)

// Source supplies named values without forcing callers to read the process env.
type Source interface {
	Lookup(name string) (string, bool)
}

// FuncSource adapts a lookup function into a Source.
type FuncSource func(name string) (string, bool)

// Lookup implements Source.
func (f FuncSource) Lookup(name string) (string, bool) {
	return f(name)
}

// EnvSource reads values from the process environment.
type EnvSource struct{}

// Lookup implements Source.
func (EnvSource) Lookup(name string) (string, bool) {
	return os.LookupEnv(name)
}

// Mode identifies the credential mechanism used for OSF requests.
type Mode string

const (
	// ModeAnonymous means no credentials were supplied.
	ModeAnonymous Mode = "anonymous"
	// ModeBearerToken means requests should use Authorization: Bearer.
	ModeBearerToken Mode = "bearer-token"
	// ModeUsernamePassword means requests should use username/password signing.
	ModeUsernamePassword Mode = "username-password"
)

// Credentials contains the selected OSF credential material.
type Credentials struct {
	Mode     Mode
	Token    string
	Username string
	Password string
}

// Authenticated reports whether the credentials can authenticate OSF requests.
func (c Credentials) Authenticated() bool {
	switch c.Mode {
	case ModeBearerToken:
		return strings.TrimSpace(c.Token) != ""
	case ModeUsernamePassword:
		return strings.TrimSpace(c.Username) != "" && strings.TrimSpace(c.Password) != ""
	default:
		return false
	}
}

// Secrets returns concrete secret values that should be redacted from output.
func (c Credentials) Secrets() []string {
	secrets := make([]string, 0, 3)
	if strings.TrimSpace(c.Token) != "" {
		secrets = append(secrets, strings.TrimSpace(c.Token))
	}
	if strings.TrimSpace(c.Username) != "" {
		secrets = append(secrets, strings.TrimSpace(c.Username))
	}
	if strings.TrimSpace(c.Password) != "" {
		secrets = append(secrets, strings.TrimSpace(c.Password))
	}
	return secrets
}

// LoadToken returns the trimmed OSF personal access token from the supplied source.
func LoadToken(source Source) (string, error) {
	if source == nil {
		source = EnvSource{}
	}

	raw, ok := source.Lookup(TokenEnv)
	if !ok {
		return "", MissingTokenError{Env: TokenEnv}
	}

	token := strings.TrimSpace(raw)
	if token == "" {
		return "", MissingTokenError{Env: TokenEnv}
	}

	return token, nil
}

// LoadCredentials selects credentials from the supplied source.
//
// OSF_TOKEN has precedence. OSF_USERNAME and OSF_PASSWORD are used only when no
// token is present.
func LoadCredentials(source Source) (Credentials, error) {
	if source == nil {
		source = EnvSource{}
	}

	if token, err := LoadToken(source); err == nil {
		return Credentials{Mode: ModeBearerToken, Token: token}, nil
	}

	usernameRaw, hasUsername := source.Lookup(UsernameEnv)
	passwordRaw, hasPassword := source.Lookup(PasswordEnv)
	username := strings.TrimSpace(usernameRaw)
	password := strings.TrimSpace(passwordRaw)
	if username == "" && password == "" && !hasUsername && !hasPassword {
		return Credentials{Mode: ModeAnonymous}, MissingCredentialsError{}
	}
	if username == "" || password == "" {
		return Credentials{Mode: ModeAnonymous}, MissingCredentialsError{UsernamePresent: username != "", PasswordPresent: password != ""}
	}
	return Credentials{Mode: ModeUsernamePassword, Username: username, Password: password}, nil
}

// MissingTokenError reports that no usable token was found.
type MissingTokenError struct {
	Env string
}

// Error implements error.
func (e MissingTokenError) Error() string {
	env := e.Env
	if env == "" {
		env = TokenEnv
	}
	return fmt.Sprintf("missing OSF personal access token; set %s", env)
}

// MissingCredentialsError reports that no complete supported credential set was found.
type MissingCredentialsError struct {
	UsernamePresent bool
	PasswordPresent bool
}

// Error implements error.
func (e MissingCredentialsError) Error() string {
	if e.UsernamePresent && !e.PasswordPresent {
		return fmt.Sprintf("missing OSF password; set %s or use %s", PasswordEnv, TokenEnv)
	}
	if !e.UsernamePresent && e.PasswordPresent {
		return fmt.Sprintf("missing OSF username; set %s or use %s", UsernameEnv, TokenEnv)
	}
	return fmt.Sprintf("missing OSF credentials; set %s or %s and %s", TokenEnv, UsernameEnv, PasswordEnv)
}

// Redact removes token-like values from loggable text.
func Redact(text string, secrets ...string) string {
	if text == "" {
		return text
	}

	out := text
	for _, secret := range secrets {
		secret = strings.TrimSpace(secret)
		if secret == "" {
			continue
		}
		out = strings.ReplaceAll(out, secret, redacted)
	}

	out = bearerPattern.ReplaceAllString(out, "${1}"+redacted)
	out = basicPattern.ReplaceAllString(out, "${1}"+redacted)
	out = envAssignmentPattern.ReplaceAllString(out, "${1}"+redacted)
	out = tokenLikePattern.ReplaceAllStringFunc(out, func(match string) string {
		if match == redacted {
			return match
		}
		return redacted
	})

	return out
}

// RedactError returns an error whose message has sensitive values removed.
func RedactError(err error, secrets ...string) error {
	if err == nil {
		return nil
	}
	return errors.New(Redact(err.Error(), secrets...))
}
