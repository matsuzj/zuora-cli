package api

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

// FuzzReadOnlyAllowlist machine-checks the read-only mode's security
// invariant instead of enumerating near-misses by hand: a POST is allowed
// if and only if its normalized path is in the allowlist (or matches an
// allowlisted pattern), write verbs are never allowed, and extractPath's
// normalization is stable. A divergence here would be a read-only BYPASS —
// a write slipping through --read-only — which is the CLI's clearest
// security property.
func FuzzReadOnlyAllowlist(f *testing.F) {
	for _, p := range readOnlyPOSTAllowList {
		f.Add("POST", p)
	}
	f.Add("POST", "v1/subscriptions/sub-1/preview")
	f.Add("POST", "meters/m1/summary")
	f.Add("POST", "https://rest.test.ap.zuora.com/v1/action/query?x=1")
	f.Add("POST", "/V1/Action/Query")
	f.Add("POST", "v1/action/queryx")
	f.Add("POST", "v1/action/query/../../v1/orders")
	f.Add("POST", "%2e%2e/v1/action/query")
	f.Add("POST", "v1/action/query\x00")
	f.Add("POST", "query/jobs")
	f.Add("PUT", "v1/action/query")
	f.Add("DELETE", "query/jobs/abc")
	f.Add("get", "v1/orders")
	f.Fuzz(func(t *testing.T, method, path string) {
		allowed := isReadOnlyAllowed(method, path) // must never panic

		p := extractPath(path)
		if strings.Contains(p, "?") {
			t.Errorf("extractPath(%q) = %q still contains a query separator", path, p)
		}
		if strings.HasPrefix(p, "/") {
			t.Errorf("extractPath(%q) = %q keeps a leading slash", path, p)
		}
		if p != strings.ToLower(p) {
			t.Errorf("extractPath(%q) = %q is not lowercased", path, p)
		}
		if again := extractPath(p); again != p {
			t.Errorf("extractPath is not idempotent: %q -> %q -> %q", path, p, again)
		}

		switch strings.ToUpper(method) {
		case "GET", "HEAD", "OPTIONS":
			if !allowed {
				t.Errorf("safe method %q must always be allowed", method)
			}
		case "POST":
			inList := false
			for _, a := range readOnlyPOSTAllowList {
				if p == a {
					inList = true
					break
				}
			}
			if !inList {
				for _, re := range readOnlyPOSTPatterns {
					if re.MatchString(p) {
						inList = true
						break
					}
				}
			}
			if allowed != inList {
				t.Errorf("POST %q: isReadOnlyAllowed=%v but allowlist membership of %q=%v", path, allowed, p, inList)
			}
		default:
			if allowed {
				t.Errorf("write method %q must never be allowed in read-only mode (path %q)", method, path)
			}
		}
	})
}

// FuzzParseAPIError feeds the fully attacker/gateway-controlled error-body
// parser arbitrary bytes: it must never panic, must preserve the status,
// must cap non-JSON echo bodies, and successEnvelopeError /
// isRetriableSuccessEnvelope must agree with an independent computation of
// the envelope and transient-code rules.
func FuzzParseAPIError(f *testing.F) {
	seeds := []string{
		`{"reasons":[{"code":53100020,"message":"bad value"}]}`,
		`{"reasons":[{"code":"INVALID","message":"a"},{"code":58730050,"message":"b"}]}`,
		`{"Success":false,"Errors":[{"Code":"DUP","Message":"dup"}]}`,
		`{"error":{"code":"X","message":"y"}}`,
		`{"message":"plain"}`,
		`{"success":false,"reasons":[{"code":58730061,"message":"lock"}]}`,
		`{"success":true}`,
		`{"success":"false"}`,
		`<html>` + strings.Repeat("A", 600) + `%s%d\x1b[2J</html>`,
		"trunc\xe3",
		``,
	}
	for _, s := range seeds {
		f.Add(200, []byte(s))
		f.Add(400, []byte(s))
		f.Add(500, []byte(s))
	}
	f.Fuzz(func(t *testing.T, status int, body []byte) {
		apiErr := parseAPIError(status, body) // must never panic
		if apiErr == nil {
			t.Fatal("parseAPIError returned nil")
		}
		if apiErr.StatusCode != status {
			t.Errorf("StatusCode %d not preserved (got %d)", status, apiErr.StatusCode)
		}
		_ = apiErr.Error() // must never panic (bodies may contain fmt verbs)
		if !json.Valid(body) && len(apiErr.Message) > maxRawErrorBody+3 {
			t.Errorf("non-JSON body echo not capped: len=%d", len(apiErr.Message))
		}

		env := successEnvelopeError(status, body)
		var envelope struct {
			S  *bool `json:"success"`
			SU *bool `json:"Success"`
		}
		if json.Unmarshal(body, &envelope) == nil {
			want := (envelope.S != nil && !*envelope.S) || (envelope.SU != nil && !*envelope.SU)
			if (env != nil) != want {
				t.Errorf("successEnvelopeError=%v disagrees with independent envelope read (want error: %v) for %q", env, want, body)
			}
		} else if env != nil {
			t.Errorf("non-decodable body must not produce an envelope error: %q", body)
		}

		if env != nil {
			ae, ok := env.(*APIError)
			if !ok {
				t.Fatalf("envelope error is %T, want *APIError", env)
			}
			retriable := isRetriableSuccessEnvelope(env, "GET")
			n, convErr := strconv.Atoi(ae.Code)
			want := convErr == nil && (n%100 == 50 || n%100 == 61 || n%100 == 70 || n%100 == 99)
			if retriable != want {
				t.Errorf("isRetriableSuccessEnvelope=%v disagrees with independent transient-code rule=%v (code %q)", retriable, want, ae.Code)
			}
		}
	})
}
