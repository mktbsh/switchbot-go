package switchbot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"
	"time"
)

// Helper function to validate UUIDv7 format in tests
var uuidV7Regex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestSetAuthorizationHeader(t *testing.T) {
	// Setup a dummy client
	mockToken := "test-token-123"
	mockSecret := "test-secret-abc"
	client, err := NewClient(mockToken, mockSecret) // Use default options
	if err != nil {
		t.Fatalf("Failed to create client for testing: %v", err)
	}

	// Create a dummy request
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)

	// Call the function to test
	client.setAuthorizationHeader(req)

	// --- Assertions ---
	t.Run("Check Authorization Header", func(t *testing.T) {
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("Authorization header is missing")
		}
		if authHeader != mockToken {
			t.Errorf("Authorization header = %q; want %q", authHeader, mockToken)
		}
	})

	t.Run("Check Content-Type Header", func(t *testing.T) {
		contentTypeHeader := req.Header.Get("Content-Type")
		expectedContentType := "application/json; charset=utf-8"
		if contentTypeHeader == "" {
			t.Error("Content-Type header is missing")
		}
		if contentTypeHeader != expectedContentType {
			t.Errorf("Content-Type header = %q; want %q", contentTypeHeader, expectedContentType)
		}
	})

	// Get the actual t and nonce set by the function
	actualT := req.Header.Get("t")
	actualNonce := req.Header.Get("nonce")
	actualSign := req.Header.Get("sign")

	t.Run("Check Timestamp (t) Header", func(t *testing.T) {
		if actualT == "" {
			t.Fatal("t header is missing") // Fatal because sign check depends on it
		}
		// Check if it's a parseable integer
		ts, err := strconv.ParseInt(actualT, 10, 64)
		if err != nil {
			t.Errorf("t header %q is not a valid integer: %v", actualT, err)
		}
		// Check if it's a reasonable timestamp (e.g., within +/- 1 minute of now)
		nowMillis := time.Now().UnixMilli()
		if ts < nowMillis-60000 || ts > nowMillis+60000 {
			t.Errorf("t header timestamp %d seems unreasonable compared to now %d", ts, nowMillis)
		}
	})

	t.Run("Check Nonce Header", func(t *testing.T) {
		if actualNonce == "" {
			t.Fatal("nonce header is missing") // Fatal because sign check depends on it
		}
		// Check if it matches UUIDv7 format
		if !uuidV7Regex.MatchString(actualNonce) {
			t.Errorf("nonce header %q does not match UUIDv7 format", actualNonce)
		}
	})

	t.Run("Check Sign Header", func(t *testing.T) {
		if actualSign == "" {
			t.Error("sign header is missing")
		}

		// Recalculate the expected signature using the actual t and nonce
		stringToSign := mockToken + actualT + actualNonce
		mac := hmac.New(sha256.New, []byte(mockSecret))
		mac.Write([]byte(stringToSign))
		expectedSign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

		if actualSign != expectedSign {
			t.Errorf("sign header = %q; want %q (calculated from t=%s, nonce=%s)",
				actualSign, expectedSign, actualT, actualNonce)
		}
	})
}

func TestGenerateTimestamp(t *testing.T) {
	before := time.Now().UnixMilli()
	// Allow very brief execution time
	time.Sleep(1 * time.Millisecond)
	tsStr := generateTimestamp()
	time.Sleep(1 * time.Millisecond)
	after := time.Now().UnixMilli()

	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		t.Fatalf("generateTimestamp() returned non-integer string %q: %v", tsStr, err)
	}

	// Check if timestamp falls within the expected range (allow some tolerance)
	tolerance := int64(10) // 10 milliseconds tolerance
	if ts < before-tolerance || ts > after+tolerance {
		t.Errorf("generateTimestamp() timestamp %d is outside expected range [%d, %d]", ts, before, after)
	}
}

func TestGenerateNonce(t *testing.T) {
	t.Run("FormatAndNoError", func(t *testing.T) {
		nonce := generateNonce() // Relies on getUUIDv7String not returning error in practice

		if nonce == "" {
			t.Fatal("generateNonce() returned an empty string")
		}
		if !uuidV7Regex.MatchString(nonce) {
			t.Errorf("generateNonce() returned string %q does not match UUIDv7 format", nonce)
		}
	})

	t.Run("Uniqueness", func(t *testing.T) {
		iterations := 100
		generated := make(map[string]bool)
		for i := 0; i < iterations; i++ {
			nonce := generateNonce()
			if generated[nonce] {
				t.Fatalf("generateNonce() generated duplicate nonce %q on iteration %d", nonce, i)
			}
			generated[nonce] = true
			// Optional short sleep if concerned about extremely rapid generation
			// time.Sleep(time.Microsecond * 10)
		}
		if len(generated) != iterations {
			t.Errorf("generateNonce() did not produce %d unique nonces, got %d", iterations, len(generated))
		}
	})
}
