package authz

import (
	"fmt"
	"testing"
	"time"
)

func TestCode_IsExpired(t *testing.T) {
	tests := map[string]struct {
		code     Code
		expected bool
	}{
		"just same": {
			code: Code{
				createdAt: time.Now(),
			},
			expected: false,
		},
		"decent": {
			code: Code{
				createdAt: time.Now().Add(-5 * time.Minute),
			},
			expected: false,
		},
		"just old": {
			code: Code{
				createdAt: time.Now().Add(-10 * time.Minute),
			},
			expected: true,
		},
		"too old": {
			code: Code{
				createdAt: time.Now().Add(-10 * time.Hour),
			},
			expected: true,
		},
	}

	for n, test := range tests {
		t.Run(n, func(t *testing.T) {
			actual := test.code.IsExpired()
			if actual != test.expected {
				t.Errorf("should have returned the expected value: %v: %s", test.code, reportUnexpected("validity", actual, test.expected))
				return
			}
		})
	}
}

func TestAccessToken_IsExpired(t *testing.T) {
	tests := map[string]struct {
		tok      AccessToken
		expected bool
	}{
		"just same": {
			tok: AccessToken{
				createdAt: time.Now(),
			},
			expected: false,
		},
		"decent": {
			tok: AccessToken{
				createdAt: time.Now().Add(-30 * time.Minute),
			},
			expected: false,
		},
		"just old": {
			tok: AccessToken{
				createdAt: time.Now().Add(-1 * time.Hour),
			},
			expected: true,
		},
		"too old": {
			tok: AccessToken{
				createdAt: time.Now().Add(-2 * time.Hour),
			},
			expected: true,
		},
	}

	for n, test := range tests {
		t.Run(n, func(t *testing.T) {
			actual := test.tok.IsExpired()
			if actual != test.expected {
				t.Errorf("should have returned the expected value: %v: %s", test.tok, reportUnexpected("validity", actual, test.expected))
				return
			}
		})
	}
}

func reportUnexpected(name string, actual, expected interface{}) error {
	return fmt.Errorf("unexpected %s: got %v, but expected %v", name, actual, expected)
}
