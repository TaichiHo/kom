package utils

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

// FormatResource formats resource.Quantity into human-readable format
// Example:
// Memory formatting example:
// q1 := resource.MustParse("8127096Ki")
// fmt.Println("Formatted memory:", utils.FormatResource(q1))
//
// // Memory formatting example (greater than Gi)
// q2 := resource.MustParse("256Gi")
// fmt.Println("Formatted memory:", utils.FormatResource(q2))
//
// // CPU formatting example
// q3 := resource.MustParse("500m") // CPU 0.01 core
// fmt.Println("Formatted CPU:", q3.String()) // For CPU, use original format directly
func FormatResource(q resource.Quantity) string {
	value := q.Value()
	format := q.Format

	switch format {
	case resource.BinarySI: // Ki, Mi, Gi, etc.
		return formatBinarySI(value)
	case resource.DecimalSI: // K, M, G, etc.
		return formatDecimalSI(value)
	default:
		return q.String() // Return original format
	}
}

// formatBinarySI converts binary format to readable format (Ki, Mi, Gi)
func formatBinarySI(value int64) string {
	const (
		Ki = 1024
		Mi = Ki * 1024
		Gi = Mi * 1024
		Ti = Gi * 1024
	)
	switch {
	case value >= Ti:
		return fmt.Sprintf("%.2fTi", float64(value)/float64(Ti))
	case value >= Gi:
		return fmt.Sprintf("%.2fGi", float64(value)/float64(Gi))
	case value >= Mi:
		return fmt.Sprintf("%.2fMi", float64(value)/float64(Mi))
	case value >= Ki:
		return fmt.Sprintf("%.2fKi", float64(value)/float64(Ki))
	default:
		return fmt.Sprintf("%d", value)
	}
}

// formatDecimalSI converts decimal format to readable format (K, M, G)
func formatDecimalSI(value int64) string {
	const (
		K = 1000
		M = K * 1000
		G = M * 1000
		T = G * 1000
	)
	switch {
	case value >= T:
		return fmt.Sprintf("%.2fT", float64(value)/float64(T))
	case value >= G:
		return fmt.Sprintf("%.2fG", float64(value)/float64(G))
	case value >= M:
		return fmt.Sprintf("%.2fM", float64(value)/float64(M))
	case value >= K:
		return fmt.Sprintf("%.2fK", float64(value)/float64(K))
	default:
		return fmt.Sprintf("%d", value)
	}
}
