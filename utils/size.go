package utils

import (
	"fmt"
)

// FormatSize takes a list of file sizes in bytes and returns a formatted string
// representing the total size (e.g., "1023 bytes", "1.1 Kb", "7.89 Mb").
func FormatSize(fileSizes []int) string {
	var totalBytes int64
	for _, size := range fileSizes {
		totalBytes += int64(size)
	}

	// Define units and their thresholds
	units := []string{"bytes", "Kb", "Mb", "Gb", "Tb", "Pb"}
	threshold := int64(1024)
	size := float64(totalBytes)
	unitIndex := 0

	// Determine appropriate unit
	for size >= float64(threshold) && unitIndex < len(units)-1 {
		size /= float64(threshold)
		unitIndex++
	}

	// Format output
	if unitIndex == 0 {
		// Bytes: display as integer
		return fmt.Sprintf("%d %s", int64(size), units[unitIndex])
	}
	// Other units: display with up to 2 decimal places
	return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}