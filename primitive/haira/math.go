package haira

import (
	"math"
	"math/rand"
	"sort"
)

// Math constants
var MathPI = math.Pi
var MathE = math.E

// MathAbs returns the absolute value.
func MathAbs(x any) float64 {
	return math.Abs(toFloat64(x))
}

// MathMin returns the smaller of two numbers.
func MathMin(a, b any) float64 {
	return math.Min(toFloat64(a), toFloat64(b))
}

// MathMax returns the larger of two numbers.
func MathMax(a, b any) float64 {
	return math.Max(toFloat64(a), toFloat64(b))
}

// MathClamp clamps a value between min and max.
func MathClamp(x, min, max any) float64 {
	v := toFloat64(x)
	lo := toFloat64(min)
	hi := toFloat64(max)
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// MathFloor returns the largest integer less than or equal to x.
func MathFloor(x any) float64 {
	return math.Floor(toFloat64(x))
}

// MathCeil returns the smallest integer greater than or equal to x.
func MathCeil(x any) float64 {
	return math.Ceil(toFloat64(x))
}

// MathRound returns the nearest integer, rounding half away from zero.
func MathRound(x any) float64 {
	return math.Round(toFloat64(x))
}

// MathTrunc returns the integer part of x.
func MathTrunc(x any) float64 {
	return math.Trunc(toFloat64(x))
}

// MathPow returns x raised to the power y.
func MathPow(x, y any) float64 {
	return math.Pow(toFloat64(x), toFloat64(y))
}

// MathSqrt returns the square root of x.
func MathSqrt(x any) float64 {
	return math.Sqrt(toFloat64(x))
}

// MathCbrt returns the cube root of x.
func MathCbrt(x any) float64 {
	return math.Cbrt(toFloat64(x))
}

// MathExp returns e raised to the power x.
func MathExp(x any) float64 {
	return math.Exp(toFloat64(x))
}

// MathLog returns the natural logarithm of x.
func MathLog(x any) float64 {
	return math.Log(toFloat64(x))
}

// MathLog10 returns the base-10 logarithm of x.
func MathLog10(x any) float64 {
	return math.Log10(toFloat64(x))
}

// MathLog2 returns the base-2 logarithm of x.
func MathLog2(x any) float64 {
	return math.Log2(toFloat64(x))
}

// MathSin returns the sine of x (in radians).
func MathSin(x any) float64 {
	return math.Sin(toFloat64(x))
}

// MathCos returns the cosine of x (in radians).
func MathCos(x any) float64 {
	return math.Cos(toFloat64(x))
}

// MathTan returns the tangent of x (in radians).
func MathTan(x any) float64 {
	return math.Tan(toFloat64(x))
}

// MathAsin returns the arcsine of x.
func MathAsin(x any) float64 {
	return math.Asin(toFloat64(x))
}

// MathAcos returns the arccosine of x.
func MathAcos(x any) float64 {
	return math.Acos(toFloat64(x))
}

// MathAtan returns the arctangent of x.
func MathAtan(x any) float64 {
	return math.Atan(toFloat64(x))
}

// MathAtan2 returns the arctangent of y/x.
func MathAtan2(y, x any) float64 {
	return math.Atan2(toFloat64(y), toFloat64(x))
}

// MathRandom returns a random float64 in [0.0, 1.0).
func MathRandom() float64 {
	return rand.Float64()
}

// MathRandomInt returns a random int in [min, max).
func MathRandomInt(min, max any) int {
	lo := int(toFloat64(min))
	hi := int(toFloat64(max))
	if hi <= lo {
		return lo
	}
	return lo + rand.Intn(hi-lo)
}

// MathSum sums all numeric elements in an array.
func MathSum(arr any) float64 {
	s := ToSlice(arr)
	var total float64
	for _, item := range s {
		total += toFloat64(item)
	}
	return total
}

// MathAvg returns the average of all numeric elements in an array.
func MathAvg(arr any) float64 {
	s := ToSlice(arr)
	if len(s) == 0 {
		return 0
	}
	var total float64
	for _, item := range s {
		total += toFloat64(item)
	}
	return total / float64(len(s))
}

// MathMedian returns the median value of a numeric array.
func MathMedian(arr any) float64 {
	s := ToSlice(arr)
	if len(s) == 0 {
		return 0
	}
	floats := make([]float64, len(s))
	for i, item := range s {
		floats[i] = toFloat64(item)
	}
	sort.Float64s(floats)
	n := len(floats)
	if n%2 == 0 {
		return (floats[n/2-1] + floats[n/2]) / 2
	}
	return floats[n/2]
}

// MathSign returns -1, 0, or 1 indicating the sign of x.
func MathSign(x any) float64 {
	v := toFloat64(x)
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}

// MathHypot returns sqrt(x*x + y*y) without overflow.
func MathHypot(x, y any) float64 {
	return math.Hypot(toFloat64(x), toFloat64(y))
}

// MathGcd returns the greatest common divisor of a and b.
func MathGcd(a, b any) int {
	x := int(math.Abs(toFloat64(a)))
	y := int(math.Abs(toFloat64(b)))
	for y != 0 {
		x, y = y, x%y
	}
	return x
}

// MathLcm returns the least common multiple of a and b.
func MathLcm(a, b any) int {
	x := int(math.Abs(toFloat64(a)))
	y := int(math.Abs(toFloat64(b)))
	if x == 0 || y == 0 {
		return 0
	}
	return x / MathGcd(x, y) * y
}

// MathLerp performs linear interpolation between a and b by t.
func MathLerp(a, b, t any) float64 {
	fa := toFloat64(a)
	fb := toFloat64(b)
	ft := toFloat64(t)
	return fa + (fb-fa)*ft
}

// toFloat64 converts any numeric value to float64.
func toFloat64(v any) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case float64:
		return n
	case int64:
		return float64(n)
	case float32:
		return float64(n)
	default:
		return 0
	}
}
