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
