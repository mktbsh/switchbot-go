package switchbot

import (
	"encoding/binary"
	"encoding/json"
	"regexp"
	"testing"
	"time"
)

func TestGetUUIDv7(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		_, err := getUUIDv7()
		if err != nil {
			t.Fatalf("getUUIDv7() returned an unexpected error: %v", err)
		}
	})

	t.Run("VersionAndVariantBits", func(t *testing.T) {
		value, err := getUUIDv7()
		if err != nil {
			t.Fatalf("getUUIDv7() failed: %v", err)
		}

		// Check Version (7) -> bits 0111xxxx in value[6]
		if version := value[6] >> 4; version != 0x07 {
			t.Errorf("getUUIDv7() generated incorrect version bits: got %02x, want 7x", value[6])
		}

		// Check Variant (RFC 4122) -> bits 10xxxxxx in value[8]
		// 0b10xxxxxx should be one of 8, 9, a, b in the first nibble
		if variant := value[8] >> 6; variant != 0x02 { // 0b10 -> 2
			t.Errorf("getUUIDv7() generated incorrect variant bits: got %02x, want 10xxxxxx", value[8])
		}
	})

	t.Run("TimestampAccuracy", func(t *testing.T) {
		// Allow a small delta for execution time variance
		before := time.Now().UnixMilli()
		// Short sleep might be needed on fast machines if resolution isn't high enough
		time.Sleep(1 * time.Millisecond)
		value, err := getUUIDv7()
		time.Sleep(1 * time.Millisecond)
		after := time.Now().UnixMilli()

		if err != nil {
			t.Fatalf("getUUIDv7() failed: %v", err)
		}

		// Extract timestamp (first 48 bits)
		var tsBytes [8]byte // Use 8 bytes for uint64
		// Copy the first 6 bytes (48 bits) into the higher bytes of the 8-byte slice
		// This correctly positions the 48 bits when interpreted as BigEndian uint64.
		copy(tsBytes[2:], value[0:6])
		// Convert the 8 bytes to uint64. The first 2 bytes are zero padding.
		extractedTS := binary.BigEndian.Uint64(tsBytes[:]) // <-- Remove the ">> 16" shift

		// Check if the timestamp is within a reasonable range (e.g., +/- 10ms)
		tolerance := int64(10) // milliseconds
		if int64(extractedTS) < before-tolerance || int64(extractedTS) > after+tolerance {
			t.Errorf("getUUIDv7() timestamp %d is outside the expected range [%d, %d] (tolerance %dms)",
				extractedTS, before, after, tolerance)
		}
		// Optional: Check against current time to catch larger drifts during test run
		now := time.Now().UnixMilli()
		toleranceNow := int64(20) // Allow slightly more tolerance against 'now'
		if int64(extractedTS) > now+toleranceNow {
			t.Errorf("getUUIDv7() timestamp %d is unexpectedly far in the future compared to now %d (tolerance %dms)",
				extractedTS, now, toleranceNow)
		}
		// Sanity check: timestamp should be a large number (e.g., after 2024)
		minExpectedTS := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
		if int64(extractedTS) < minExpectedTS {
			t.Errorf("getUUIDv7() timestamp %d seems unreasonably small (expected > %d)", extractedTS, minExpectedTS)
		}
	})

	t.Run("Uniqueness", func(t *testing.T) {
		iterations := 100 // Generate multiple UUIDs to increase chance of collision detection
		generated := make(map[[16]byte]bool)

		for i := 0; i < iterations; i++ {
			uuid, err := getUUIDv7()
			if err != nil {
				t.Fatalf("getUUIDv7() failed during uniqueness test on iteration %d: %v", i, err)
			}
			if generated[uuid] {
				t.Fatalf("getUUIDv7() generated duplicate UUID %x on iteration %d", uuid, i)
			}
			generated[uuid] = true
			// Optional: short sleep if very high speed generation might hit time resolution limits
			// time.Sleep(1 * time.Millisecond)
		}
		if len(generated) != iterations {
			t.Errorf("Expected %d unique UUIDs, but map contains %d", iterations, len(generated))
		}
	})
}

func TestGetUUIDv7String(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		_, err := getUUIDv7String()
		if err != nil {
			t.Fatalf("getUUIDv7String() returned an unexpected error: %v", err)
		}
	})

	t.Run("Format", func(t *testing.T) {
		uuidStr, err := getUUIDv7String()
		if err != nil {
			t.Fatalf("getUUIDv7String() failed: %v", err)
		}

		// Regex to validate UUID format (xxxxxxxx-xxxx-7xxx-[89ab]xxx-xxxxxxxxxxxx)
		// Version 7 is checked by the '7'.
		// Variant RFC 4122 is checked by [89ab].
		uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

		if !uuidRegex.MatchString(uuidStr) {
			t.Errorf("getUUIDv7String() generated string '%s' does not match UUIDv7 format regex", uuidStr)
		}
	})

	t.Run("Uniqueness", func(t *testing.T) {
		iterations := 100
		generated := make(map[string]bool)

		for i := 0; i < iterations; i++ {
			uuidStr, err := getUUIDv7String()
			if err != nil {
				t.Fatalf("getUUIDv7String() failed during uniqueness test on iteration %d: %v", i, err)
			}
			if generated[uuidStr] {
				t.Fatalf("getUUIDv7String() generated duplicate UUID string '%s' on iteration %d", uuidStr, i)
			}
			generated[uuidStr] = true
			// Optional short sleep
			// time.Sleep(1 * time.Millisecond)
		}
		if len(generated) != iterations {
			t.Errorf("Expected %d unique UUID strings, but map contains %d", iterations, len(generated))
		}
	})
}

func TestIsEmptyJSONBody(t *testing.T) {
	testCases := []struct {
		name     string
		input    json.RawMessage
		expected bool
	}{
		{
			name:     "Nil Input",
			input:    nil,
			expected: true,
		},
		{
			name:     "Empty Byte Slice",
			input:    []byte{},
			expected: true,
		},
		{
			name:     "Empty Object String",
			input:    []byte("{}"),
			expected: true,
		},
		{
			name:     "Null String",
			input:    []byte("null"),
			expected: true,
		},
		{
			name:     "Empty Object RawMessage",
			input:    json.RawMessage("{}"),
			expected: true,
		},
		{
			name:     "Null RawMessage",
			input:    json.RawMessage("null"),
			expected: true,
		},
		{
			name:     "Valid JSON Object",
			input:    json.RawMessage(`{"key":"value"}`),
			expected: false,
		},
		{
			name:     "Valid JSON Array",
			input:    json.RawMessage(`[1, 2, 3]`),
			expected: false, // Array is not considered empty by this function
		},
		{
			name:     "Empty JSON Array",
			input:    json.RawMessage(`[]`),
			expected: false, // Empty array is not considered empty by this function
		},
		{
			name:     "Non-empty String (Invalid JSON)",
			input:    json.RawMessage("hello"),
			expected: false,
		},
		{
			name:     "Whitespace only",
			input:    json.RawMessage("   "),
			expected: false,
		},
		{
			name:     "Number",
			input:    json.RawMessage("123"),
			expected: false,
		},
		{
			name:     "Boolean True",
			input:    json.RawMessage("true"),
			expected: false,
		},
		{
			name:     "Boolean False",
			input:    json.RawMessage("false"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isEmptyJSONBody(tc.input)
			if result != tc.expected {
				t.Errorf("isEmptyJSONBody(%q) = %v; want %v", string(tc.input), result, tc.expected)
			}
		})
	}
}
