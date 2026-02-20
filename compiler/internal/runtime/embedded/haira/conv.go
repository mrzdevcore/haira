package haira

import (
	"fmt"
	"strconv"
	"strings"
)

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

func ConvFloatToString(v any) string {
	return strconv.FormatFloat(toFloat64(v), 'f', -1, 64)
}

func ConvBoolToString(v any) string {
	if b, ok := v.(bool); ok {
		return strconv.FormatBool(b)
	}
	return fmt.Sprintf("%v", v)
}

func ConvStringToInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("cannot convert %q to int", s)
	}
	return n, nil
}

func ConvStringToFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert %q to float", s)
	}
	return f, nil
}

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

func ConvIntToFloat(v any) float64 { return toFloat64(v) }
func ConvFloatToInt(v any) int     { return int(toFloat64(v)) }
func ConvIntToHex(v any) string    { return fmt.Sprintf("%x", int(toFloat64(v))) }
func ConvIntToBinary(v any) string { return strconv.FormatInt(int64(toFloat64(v)), 2) }
func ConvIntToOctal(v any) string  { return strconv.FormatInt(int64(toFloat64(v)), 8) }
func ConvToString(v any) string    { return Str(v) }

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
