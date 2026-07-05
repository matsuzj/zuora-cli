package api

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaskSecrets(t *testing.T) {
	t.Run("redacts nested + array secrets, preserves non-secret identifiers", func(t *testing.T) {
		in := []byte(`{
			"accountNumber": "A-001",
			"creditCard": {"cardNumber": "4111111111111111", "cardSecurityCode": "123"},
			"password": "hunter2",
			"items": [{"token": "tok_abc"}]
		}`)
		out := maskSecrets(in)

		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(out, &m))
		assert.Equal(t, "A-001", m["accountNumber"], "a non-secret identifier must NOT be masked")
		cc := m["creditCard"].(map[string]interface{})
		assert.Equal(t, redactedValue, cc["cardNumber"])
		assert.Equal(t, redactedValue, cc["cardSecurityCode"])
		assert.Equal(t, redactedValue, m["password"])
		arr := m["items"].([]interface{})
		assert.Equal(t, redactedValue, arr[0].(map[string]interface{})["token"], "a secret inside an array must be redacted")

		// The raw secret values must not survive ANYWHERE in the serialized output.
		s := string(out)
		assert.NotContains(t, s, "4111111111111111")
		assert.NotContains(t, s, "hunter2")
		assert.NotContains(t, s, "tok_abc")
	})

	t.Run("key match is case-insensitive", func(t *testing.T) {
		out := string(maskSecrets([]byte(`{"CardNumber":"4111","CVV":"999"}`)))
		assert.NotContains(t, out, "4111")
		assert.NotContains(t, out, "999")
	})

	t.Run("snake_case and kebab-case credential keys are redacted", func(t *testing.T) {
		// Zuora's OAuth endpoint returns snake_case credential fields; these must
		// be masked despite sensitiveKeys storing the fused form. Bites if key
		// matching reverts to plain strings.ToLower (no separator folding). (#426)
		out := string(maskSecrets([]byte(`{
			"access_token": "at_secret",
			"refresh_token": "rt_secret",
			"client_secret": "cs_secret",
			"api-key": "ak_secret"
		}`)))
		assert.NotContains(t, out, "at_secret")
		assert.NotContains(t, out, "rt_secret")
		assert.NotContains(t, out, "cs_secret")
		assert.NotContains(t, out, "ak_secret")
	})

	t.Run("non-JSON body is returned unchanged", func(t *testing.T) {
		in := []byte("not json <html>error</html>")
		assert.Equal(t, in, maskSecrets(in))
	})
}

// TestMaskSecrets_EveryKeyEverySpelling iterates the ACTUAL sensitiveKeys
// table (same package) and, for each entry, derives spellings that must fold
// back to it under normalizeMaskKey — fused, UPPER, mid-word camelCase,
// snake_case and kebab-case — asserting the secret value never survives
// maskSecrets. Because it enumerates the live table, any key whose matching
// breaks (e.g. separator folding removed, case folding removed) fails here
// for every affected entry, not just the hand-picked ones above.
func TestMaskSecrets_EveryKeyEverySpelling(t *testing.T) {
	// Floor pin: shrinking the table below its audited size is a masking
	// regression (the "someone trims the list" class). Adding keys is fine.
	require.GreaterOrEqual(t, len(sensitiveKeys), 14,
		"sensitiveKeys shrank below its audited size — a sensitive field would now leak into verbose logs")

	for key := range sensitiveKeys {
		require.GreaterOrEqual(t, len(key), 2, "sensitiveKeys entry too short to derive spellings: %q", key)
		mid := len(key) / 2
		spellings := map[string]string{
			"fused": key,
			"upper": strings.ToUpper(key),
			"camel": key[:mid] + strings.ToUpper(key[mid:mid+1]) + key[mid+1:],
			"snake": key[:mid] + "_" + key[mid:],
			"kebab": key[:mid] + "-" + key[mid:],
		}
		for style, spelled := range spellings {
			t.Run(key+"/"+style, func(t *testing.T) {
				secret := "SENTINEL-" + key + "-VALUE"
				body, err := json.Marshal(map[string]string{spelled: secret})
				require.NoError(t, err)
				out := string(maskSecrets(body))
				assert.NotContains(t, out, secret,
					"spelling %q of sensitive key %q must be redacted", spelled, key)
				assert.Contains(t, out, redactedValue)
			})
		}
	}
}
