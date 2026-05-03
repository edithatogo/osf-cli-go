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

const redacted = "[REDACTED]"

var tokenLikePattern = regexp.MustCompile(`(?i)\b[A-Za-z0-9_-]{24,}\b`)
var bearerPattern = regexp.MustCompile(`(?i)\b(Bearer\s+)([A-Za-z0-9._~+/=-]{8,})\b`)
var envAssignmentPattern = regexp.MustCompile(`(?i)\b(OSF_TOKEN\s*=\s*)([^\s]+)\b`)

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
