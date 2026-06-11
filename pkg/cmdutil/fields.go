package cmdutil

import (
	"fmt"
	"strconv"
)

// GetString returns m[key] as a display string, or "" when absent/nil.
//
// JSON numbers decode to float64 and are rendered with fmt's default verb, so
// large values use scientific notation (e.g. 1000000 -> "1e+06"). Use this for
// string-valued fields (ids, statuses, names). For numeric/monetary fields that
// must read as plain decimals, use GetDecimal.
func GetString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// GetDecimal returns m[key] as a display string like GetString, but formats
// JSON numbers (float64) as plain decimals rather than scientific notation
// (e.g. 1000000 -> "1000000", not "1e+06"). Use this for monetary amounts,
// balances, quantities, and other numeric fields.
func GetDecimal(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if f, ok := v.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return fmt.Sprintf("%v", v)
}

// GetMoney returns m[key] formatted with a fixed two decimals ("50.00") — the
// display contract for monetary amounts and balances (E2E and user scripts
// depend on the two-decimal form). GetDecimal strips trailing zeros and is for
// non-monetary numerics.
func GetMoney(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if f, ok := v.(float64); ok {
		return fmt.Sprintf("%.2f", f)
	}
	return fmt.Sprintf("%v", v)
}

// GetBool returns m[key] as "true"/"false", or "" when absent, nil, or not a
// bool.
func GetBool(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if b, ok := v.(bool); ok {
		return strconv.FormatBool(b)
	}
	return ""
}

// GetInt returns m[key] as a decimal integer string (JSON numbers are float64;
// the value is truncated, not rounded), or "" when absent/nil/non-numeric.
func GetInt(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if f, ok := v.(float64); ok {
		return strconv.Itoa(int(f))
	}
	return ""
}
