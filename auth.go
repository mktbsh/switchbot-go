package switchbot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"
	"time"
)

func (c *Client) setAuthorizationHeader(req *http.Request) {
	t := generateTimestamp()
	n := generateNonce()

	mac := hmac.New(sha256.New, []byte(c.secret))
	mac.Write([]byte(c.token + t + n))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	header := req.Header
	header.Set("Authorization", c.token)
	header.Set("t", t)
	header.Set("sign", signature)
	header.Set("nonce", n)
	header.Set("Content-Type", "application/json; charset=utf-8")
}

// generateTimestamp generates a timestamp in milliseconds since epoch.
func generateTimestamp() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}

// generateNonce generates a nonce using UUIDv7.
func generateNonce() string {
	uuid, _ := getUUIDv7String()
	return uuid
}
