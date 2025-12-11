package common

import (
	"math"
	"strings"
)

func Log10_f32(val float32) float32 {
	return float32(math.Log10(float64(val)))
}

func Power_f32(base float32, exp float32) float32 {
	return float32(math.Pow(float64(base), float64(exp)))
}

func Exppower_f32(exp float32) float32 {
	return float32(math.Pow(math.E, float64(exp)))
}

func Sigma_f32(x float32, exp float32) float32 {
	return 1. / (1. + Power_f32(x/(1-x), exp))
}

func CloneMap(src map[string]float32) map[string]float32 {
	if src == nil {
		return nil
	}
	dst := make(map[string]float32, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// AggregateScores combina i punteggi dei nodi secondo il metodo scelto.
// Supporta: "sum", "geometric", "squaredsum"
func AggregateScores(scores []float32, method string) float32 {
	if len(scores) == 0 {
		return 0
	}

	switch strings.ToLower(method) {
	case "geometric":
		prod := float64(1)
		for _, s := range scores {
			// evita 0 assoluti (che azzerano tutto)
			prod *= math.Max(float64(s), 1e-6)
		}
		return float32(math.Pow(prod, 1/float64(len(scores))))

	case "squaredsum":
		var sum float64 = 0.
		for _, s := range scores {
			sum += math.Pow(float64(s), 2)
		}
		return float32(math.Sqrt(sum)) // monotona a media quadratica

	default: // "sum" o sconosciuto
		var sum float32 = 0.
		for _, s := range scores {
			sum += s
		}
		return sum // monotona a media aritmetica
	}
}

func ToInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	default:
		return 0, false
	}
}

// toFloat effettua conversione sicura da interfaccia a float32
func ToFloat(v interface{}) (float32, bool) {
	switch n := v.(type) {
	case float64:
		return float32(n), true
	case float32:
		return n, true
	default:
		return 0, false
	}
}

func Round(x float32, digits int) float32 {
	ix := float32(math.Pow(10, float64(digits)))
	return float32(math.Round(float64(x)*float64(ix)) / float64(ix))
}
