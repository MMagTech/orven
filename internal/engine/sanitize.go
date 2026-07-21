package engine

import (
	"regexp"
	"strings"
)

// The publication boundary (docs/CONSTRAINTS.md §16): a plugin may use
// its assigned credential, but nothing a plugin returns may carry a
// credential onward into storage, briefings, logs, or the UI.
//
// The sanitizer is deliberately precise and non-destructive: it alters
// text only when an assigned secret value actually appears, or when
// text matches a well-defined credential context (an Authorization or
// API-key header, or a credential-bearing query parameter). It never
// does broad keyword matching — prose like "2 tokens expired" or "the
// API key was rejected" passes through byte-for-byte.
//
// This protects against accidental disclosure. It cannot stop a
// malicious plugin from deliberately transforming (encoding, splitting)
// or exfiltrating a secret it was given — that residual risk is
// governed by the install-time trust decision and plugin review.

const redacted = "[redacted]"

// credentialPatterns are the only pattern-based redactions performed.
// Each match must be a credential *context* carrying a value, never a
// bare keyword.
var credentialPatterns = []*regexp.Regexp{
	// credential-bearing query parameters: ?api_key=..., &token=..., etc.
	regexp.MustCompile(`(?i)([?&](?:api_?key|access_?token|token|password|secret|auth)=)[^&\s"']+`),
	// Authorization header values, with or without a scheme
	regexp.MustCompile(`(?i)\b(authorization\s*:\s*(?:bearer\s+|basic\s+|digest\s+|token\s+)?)[^\s"']+`),
	// API-key style headers: X-Api-Key: ..., Api-Key: ...
	regexp.MustCompile(`(?i)\b((?:x-)?api-?key\s*:\s*)[^\s"']+`),
}

// minSecretLen guards precision: assigned values shorter than this are
// not used for exact matching, because scrubbing very short strings
// would mangle unrelated text (a 2-character "secret" appears in
// ordinary prose constantly).
const minSecretLen = 4

type sanitizer struct{ values []string }

func newSanitizer(secrets map[string]string) *sanitizer {
	s := &sanitizer{}
	for _, v := range secrets {
		if len(v) >= minSecretLen {
			s.values = append(s.values, v)
		}
	}
	return s
}

// clean redacts assigned secret values and credential-shaped fragments.
// Text with neither is returned unchanged.
func (s *sanitizer) clean(text string) string {
	if text == "" {
		return text
	}
	for _, v := range s.values {
		text = strings.ReplaceAll(text, v, redacted)
	}
	for _, re := range credentialPatterns {
		text = re.ReplaceAllString(text, "${1}"+redacted)
	}
	return text
}

// ContainsCredentialPattern reports the first credential-shaped
// fragment in text, or "" if none. Used by `orven validate` so
// contributors are told about credential-shaped output before it ever
// runs against a real system. (Validation inspects raw, unsanitized
// output on purpose: the runtime scrubber must never mask a leak from
// the validator.)
func ContainsCredentialPattern(text string) string {
	for _, re := range credentialPatterns {
		if m := re.FindString(text); m != "" {
			return m
		}
	}
	return ""
}
