package api

import (
	"encoding/json"
	"strings"
)

// redactedValue replaces a sensitive field's value in debug output.
const redactedValue = "***REDACTED***"

// sensitiveKeys are request/response field names whose VALUES must never reach
// verbose/debug logs (ZR_DEBUG=api / -vv). Keys are stored in separator-free
// lowercase form and matched via normalizeMaskKey, so a single entry covers the
// camelCase, snake_case and kebab-case spellings of the same field — e.g.
// "accesstoken" matches accessToken, access_token and access-token (Zuora's
// OAuth endpoint returns the snake_case forms). This is a deliberately
// conservative payment-card + bank + credential set; extend it here as new
// sensitive fields surface — it is the single source of truth for masking.
var sensitiveKeys = map[string]bool{
	"creditcardnumber":  true,
	"cardnumber":        true,
	"cardsecuritycode":  true,
	"securitycode":      true,
	"cvv":               true,
	"cvc":               true,
	"bankaccountnumber": true,
	"password":          true,
	"secret":            true,
	"clientsecret":      true,
	"token":             true,
	"accesstoken":       true,
	"refreshtoken":      true,
	"apikey":            true,
}

// maskSecrets redacts the values of known-sensitive fields in a JSON body so
// they cannot leak into verbose/debug logs. It walks objects and arrays
// recursively. A body that is not valid JSON is returned unchanged (there is no
// field structure to mask) — request bodies are JSON-validated upstream and
// Zuora responses are JSON, so this is the normal path.
func maskSecrets(body []byte) []byte {
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		return body
	}
	out, err := json.Marshal(maskValue(v))
	if err != nil {
		return body
	}
	return out
}

// normalizeMaskKey folds a field name to its separator-free lowercase form so a
// single sensitiveKeys entry matches the camelCase, snake_case and kebab-case
// spellings of the same field (accessToken / access_token / access-token). (#426)
func normalizeMaskKey(k string) string {
	k = strings.ToLower(k)
	k = strings.ReplaceAll(k, "_", "")
	k = strings.ReplaceAll(k, "-", "")
	return k
}

func maskValue(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, val := range t {
			if sensitiveKeys[normalizeMaskKey(k)] {
				t[k] = redactedValue
			} else {
				t[k] = maskValue(val)
			}
		}
		return t
	case []interface{}:
		for i, item := range t {
			t[i] = maskValue(item)
		}
		return t
	default:
		return v
	}
}
