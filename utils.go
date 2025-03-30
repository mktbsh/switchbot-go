package switchbot

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

const uuidV7Format = "%x-%x-%x-%x-%x"

// getUUIDv7 generates a UUIDv7 (time-based) value.
func getUUIDv7() ([16]byte, error) {
	var value [16]byte
	_, err := rand.Read(value[:])
	if err != nil {
		return value, err
	}

	ts := big.NewInt(time.Now().UnixMilli())
	ts.FillBytes(value[0:6])
	value[6] = (value[6] & 0x0F) | 0x70
	value[8] = (value[8] & 0x3F) | 0x80
	return value, nil
}

// getUUIDv7String generates a UUIDv7 (time-based) value and returns it as a string.
func getUUIDv7String() (string, error) {
	value, err := getUUIDv7()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(uuidV7Format, value[:4], value[4:6], value[6:8], value[8:10], value[10:]), nil
}

// isEmptyJSONBody checks if the JSON body is empty or contains only null or empty object.
func isEmptyJSONBody(body json.RawMessage) bool {
	return len(body) == 0 || string(body) == "{}" || string(body) == "null"
}
