package haira

import (
	"math"
	"math/rand"
)

var MathPI = math.Pi
var MathE = math.E

func MathAbs(x any) float64      { return math.Abs(toFloat64(x)) }
func MathMin(a, b any) float64   { return math.Min(toFloat64(a), toFloat64(b)) }
func MathMax(a, b any) float64   { return math.Max(toFloat64(a), toFloat64(b)) }
func MathFloor(x any) float64    { return math.Floor(toFloat64(x)) }
func MathCeil(x any) float64     { return math.Ceil(toFloat64(x)) }
func MathRound(x any) float64    { return math.Round(toFloat64(x)) }
func MathTrunc(x any) float64    { return math.Trunc(toFloat64(x)) }
func MathPow(x, y any) float64   { return math.Pow(toFloat64(x), toFloat64(y)) }
func MathSqrt(x any) float64     { return math.Sqrt(toFloat64(x)) }
func MathCbrt(x any) float64     { return math.Cbrt(toFloat64(x)) }
func MathExp(x any) float64      { return math.Exp(toFloat64(x)) }
func MathLog(x any) float64      { return math.Log(toFloat64(x)) }
func MathLog10(x any) float64    { return math.Log10(toFloat64(x)) }
func MathLog2(x any) float64     { return math.Log2(toFloat64(x)) }
func MathSin(x any) float64      { return math.Sin(toFloat64(x)) }
func MathCos(x any) float64      { return math.Cos(toFloat64(x)) }
func MathTan(x any) float64      { return math.Tan(toFloat64(x)) }
func MathAsin(x any) float64     { return math.Asin(toFloat64(x)) }
func MathAcos(x any) float64     { return math.Acos(toFloat64(x)) }
func MathAtan(x any) float64     { return math.Atan(toFloat64(x)) }
func MathAtan2(y, x any) float64 { return math.Atan2(toFloat64(y), toFloat64(x)) }
func MathRandom() float64        { return rand.Float64() }

func MathClamp(x, min, max any) float64 {
	v, lo, hi := toFloat64(x), toFloat64(min), toFloat64(max)
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func MathRandomInt(min, max any) int {
	lo := int(toFloat64(min))
	hi := int(toFloat64(max))
	if hi <= lo {
		return lo
	}
	return lo + rand.Intn(hi-lo)
}
