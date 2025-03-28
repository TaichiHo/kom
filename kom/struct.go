package kom

import (
	"fmt"

	"github.com/weibaohui/kom/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ResourceUsageFraction defines the usage ratio for a single resource type
type ResourceUsageFraction struct {
	RequestFraction float64 `json:"requestFraction"` // Request usage percentage relative to total allocatable value
	LimitFraction   float64 `json:"limitFraction"`   // Limit usage percentage relative to total allocatable value
}

// ResourceUsageResult defines the structure for resource usage
// Stores resource usage information for Pods and Nodes
type ResourceUsageResult struct {
	Requests       map[corev1.ResourceName]resource.Quantity     `json:"requests"`       // Requested usage
	Limits         map[corev1.ResourceName]resource.Quantity     `json:"limits"`         // Limited usage
	Allocatable    map[corev1.ResourceName]resource.Quantity     `json:"allocatable"`    // Node's real-time allocatable value
	UsageFractions map[corev1.ResourceName]ResourceUsageFraction `json:"usageFractions"` // Usage ratios
}

// ResourceUsageRow temporary structure for storing each row of data
type ResourceUsageRow struct {
	ResourceType    string `json:"resourceType"`
	Total           string `json:"total"`
	Request         string `json:"request"`
	RequestFraction string `json:"requestFraction"`
	Limit           string `json:"limit"`
	LimitFraction   string `json:"limitFraction"`
}

func convertToTableData(result *ResourceUsageResult) ([]*ResourceUsageRow, error) {
	if result == nil {
		return nil, fmt.Errorf("result is nil")
	}
	var tableData []*ResourceUsageRow

	// Iterate through resource types (CPU, Memory, etc.) and generate table rows
	for _, resourceType := range []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory, corev1.ResourceEphemeralStorage} {
		// Create a row of data
		alc := result.Allocatable[resourceType]
		req := result.Requests[resourceType]
		lit := result.Limits[resourceType]
		row := &ResourceUsageRow{
			ResourceType:    string(resourceType),
			Total:           utils.FormatResource(alc),
			Request:         utils.FormatResource(req),
			RequestFraction: fmt.Sprintf("%.2f", result.UsageFractions[resourceType].RequestFraction),
			Limit:           utils.FormatResource(lit),
			LimitFraction:   fmt.Sprintf("%.2f", result.UsageFractions[resourceType].LimitFraction),
		}
		// Add row to table data
		tableData = append(tableData, row)
	}

	return tableData, nil
}
