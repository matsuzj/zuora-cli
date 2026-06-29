package api

import (
	"encoding/json"
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
