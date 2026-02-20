package haira

import (
	"fmt"
	"strconv"
	"strings"
)

// ConvIntToString converts an int to a string.
func ConvIntToString(v any) string {
	switch n := v.(type) {
	case int:
		return strconv.Itoa(n)
	case float64:
		return strconv.Itoa(int(n))
	case int64:
		return strconv.FormatInt(n, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ConvFloatToString converts a float to a string.
func ConvFloatToString(v any) string {
	f := toFloat64(v)
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// ConvBoolToString converts a bool to "true" or "false".
func ConvBoolToString(v any) string {
	if b, ok := v.(bool); ok {
		return strconv.FormatBool(b)
	}
	return fmt.Sprintf("%v", v)
}

// ConvStringToInt parses a string as an integer.
func ConvStringToInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("cannot convert %q to int", s)
	}
	return n, nil
}

// ConvStringToFloat parses a string as a float64.
func ConvStringToFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert %q to float", s)
	}
	return f, nil
}

// ConvStringToBool parses a string as a bool.
// Accepts "true", "false", "1", "0", "yes", "no".
func ConvStringToBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("cannot convert %q to bool", s)
	}
}

// ConvIntToFloat converts an int to a float64.
func ConvIntToFloat(v any) float64 {
	return toFloat64(v)
}

// ConvFloatToInt converts a float64 to an int (truncates).
func ConvFloatToInt(v any) int {
	return int(toFloat64(v))
}

// ConvIntToHex converts an int to a hexadecimal string.
func ConvIntToHex(v any) string {
	return fmt.Sprintf("%x", int(toFloat64(v)))
}

// ConvIntToBinary converts an int to a binary string.
func ConvIntToBinary(v any) string {
	return strconv.FormatInt(int64(toFloat64(v)), 2)
}

// ConvIntToOctal converts an int to an octal string.
func ConvIntToOctal(v any) string {
	return strconv.FormatInt(int64(toFloat64(v)), 8)
}

// ConvToString converts any value to its string representation.
func ConvToString(v any) string {
	return Str(v)
}

// ConvHexToInt converts a hexadecimal string to an int.
func ConvHexToInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	n, err := strconv.ParseInt(s, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert %q from hex to int", s)
	}
	return int(n), nil
}
