package kom

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestUsageDenominator_String(t *testing.T) {
	tests := []struct {
		name string
		d    UsageDenominator
		want string
	}{
		{"Auto", DenominatorAuto, "auto"},
		{"Node", DenominatorNode, "node"},
		{"Limit", DenominatorLimit, "limit"},
		{"Unknown", UsageDenominator(999), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.String(); got != tt.want {
				t.Errorf("UsageDenominator.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToTableData(t *testing.T) {
	// Setup mock data
	cpuReq := resource.MustParse("100m")
	cpuLim := resource.MustParse("200m")
	cpuReal := resource.MustParse("50m")
	cpuAlloc := resource.MustParse("1000m")

	memReq := resource.MustParse("100Mi")
	memLim := resource.MustParse("200Mi")
	memReal := resource.MustParse("50Mi")
	memAlloc := resource.MustParse("1000Mi")

	ephemeralReq := resource.MustParse("100Mi")
	ephemeralLim := resource.MustParse("200Mi")
	ephemeralReal := resource.MustParse("50Mi")
	ephemeralAlloc := resource.MustParse("1000Mi")

	result := &ResourceUsageResult{
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:              cpuReq,
			corev1.ResourceMemory:           memReq,
			corev1.ResourceEphemeralStorage: ephemeralReq,
		},
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:              cpuLim,
			corev1.ResourceMemory:           memLim,
			corev1.ResourceEphemeralStorage: ephemeralLim,
		},
		Realtime: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:              cpuReal,
			corev1.ResourceMemory:           memReal,
			corev1.ResourceEphemeralStorage: ephemeralReal,
		},
		Allocatable: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:              cpuAlloc,
			corev1.ResourceMemory:           memAlloc,
			corev1.ResourceEphemeralStorage: ephemeralAlloc,
		},
		UsageFractions: map[corev1.ResourceName]ResourceUsageFraction{
			corev1.ResourceCPU:              {RequestFraction: "10%", LimitFraction: "20%", RealtimeFraction: "5%"},
			corev1.ResourceMemory:           {RequestFraction: "10%", LimitFraction: "20%", RealtimeFraction: "5%"},
			corev1.ResourceEphemeralStorage: {RequestFraction: "10%", LimitFraction: "20%", RealtimeFraction: "5%"},
		},
	}

	tableData, err := convertToTableData(result)
	if err != nil {
		t.Fatalf("convertToTableData failed: %v", err)
	}

	if len(tableData) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(tableData))
	}

	// Verify CPU row
	cpuRow := tableData[0]
	if cpuRow.ResourceType != string(corev1.ResourceCPU) {
		t.Errorf("Expected first row to be CPU, got %s", cpuRow.ResourceType)
	}
	if cpuRow.RequestFraction != "10%" {
		t.Errorf("Expected CPU RequestFraction 10%%, got %s", cpuRow.RequestFraction)
	}

	// Test nil result
	_, err = convertToTableData(nil)
	if err == nil {
		t.Error("Expected error for nil result, got nil")
	}
}
