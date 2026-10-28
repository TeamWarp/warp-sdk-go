// File generated from our OpenAPI spec by Scalar. See README.md for details.

// Package webhooks implements Standard Webhooks signature verification
// (https://github.com/standard-webhooks/standard-webhooks) using only the
// standard library, so generated SDKs verify webhook payloads without any
// third-party module. It is internal because only the SDK's own webhook parse
// methods call it.
package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DefaultTolerance bounds how far a webhook timestamp may drift from "now"
// (in either direction) before verification fails, limiting replay windows.
const DefaultTolerance = 5 * time.Minute

// secretPrefix is the optional Standard Webhooks secret prefix; the remainder
// of the secret is base64.
const secretPrefix = "whsec_"

// signatureVersion marks the HMAC scheme of one entry in the signature header.
const signatureVersion = "v1"

// The standard header names a delivery carries its id, timestamp and
// signatures in.
const (
	idHeader        = "webhook-id"
	timestampHeader = "webhook-timestamp"
	signatureHeader = "webhook-signature"
)

// Verify checks a Standard Webhooks signature: the HMAC-SHA256 of
// "{id}.{timestamp}.{payload}" keyed with the base64-decoded secret must match
// one of the space-separated "v1,<base64>" entries in the "webhook-signature"
// header, and the "webhook-timestamp" header must be within tolerance of now.
// Comparison is constant-time.
func Verify(payload []byte, headers http.Header, secret string, now time.Time, tolerance time.Duration) error {
	// Decoding the secret first means a misconfigured secret is reported as
	// such, whatever the delivery carries.
	key, err := decodeSecret(secret)
	if err != nil {
		return err
	}

	id := headers.Get(idHeader)
	timestamp := headers.Get(timestampHeader)
	signatures := headers.Get(signatureHeader)
	if id == "" || timestamp == "" || signatures == "" {
		return fmt.Errorf("webhooks: delivery is missing the %s, %s or %s header", idHeader, timestampHeader, signatureHeader)
	}

	sentAt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		// The wrapped strconv error already quotes the header's value.
		return fmt.Errorf("webhooks: %s header is not a Unix timestamp: %w", timestampHeader, err)
	}
	if err := checkFreshness(time.Unix(sentAt, 0), now, tolerance); err != nil {
		return err
	}

	// The signed timestamp is the parsed integer written back out, so a header
	// sent as "+1614265330" or "01614265330" is checked against the signature
	// of "1614265330".
	expected := sign(key, id, strconv.FormatInt(sentAt, 10), payload)
	if !anySignatureMatches(signatures, expected) {
		return errors.New("webhooks: no v1 signature matches the payload")
	}
	return nil
}

// checkFreshness rejects a delivery sent more than tolerance before or after
// now, which bounds how long a captured delivery can be replayed. A
// non-positive tolerance means DefaultTolerance.
func checkFreshness(sentAt time.Time, now time.Time, tolerance time.Duration) error {
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	switch {
	case now.Sub(sentAt) > tolerance:
		return fmt.Errorf("webhooks: delivery timestamp is more than %s in the past", tolerance)
	case sentAt.Sub(now) > tolerance:
		return fmt.Errorf("webhooks: delivery timestamp is more than %s in the future", tolerance)
	}
	return nil
}

// anySignatureMatches reports whether one of the space-separated
// "<version>,<signature>" entries is a v1 signature equal to expected. Entries
// with another version, or without a comma, are skipped so a sender can list
// future schemes beside v1.
func anySignatureMatches(signatures string, expected string) bool {
	for _, entry := range strings.Split(signatures, " ") {
		version, rest, found := strings.Cut(entry, ",")
		if !found || version != signatureVersion {
			continue
		}
		// Anything after a second comma is not part of the signature.
		signature, _, _ := strings.Cut(rest, ",")
		// hmac.Equal is constant-time, so a forged signature cannot be probed
		// byte by byte through timing differences.
		if hmac.Equal([]byte(signature), []byte(expected)) {
			return true
		}
	}
	return false
}

// Sign computes the base64 Standard Webhooks v1 signature (without the "v1,"
// prefix) over "{id}.{timestamp}.{payload}". Exposed so tests can craft valid
// signed payloads.
func Sign(id string, timestamp string, payload []byte, secret string) (string, error) {
	key, err := decodeSecret(secret)
	if err != nil {
		return "", err
	}
	return sign(key, id, timestamp, payload), nil
}

// decodeSecret strips the optional "whsec_" prefix and base64-decodes the
// rest into the HMAC key.
func decodeSecret(secret string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(secret, secretPrefix))
	if err != nil {
		return nil, fmt.Errorf("webhooks: signing secret is not valid base64: %w", err)
	}
	// A bare "whsec_" decodes to an empty key, and an HMAC under an empty key
	// is one anyone can compute, so refuse it rather than accept forgeries.
	if len(key) == 0 {
		return nil, errors.New("webhooks: signing secret is empty")
	}
	return key, nil
}

// sign returns the base64 HMAC-SHA256 of "{id}.{timestamp}.{payload}" under key.
func sign(key []byte, id string, timestamp string, payload []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(id))
	mac.Write([]byte("."))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(payload)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
