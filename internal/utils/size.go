package utils

import (
	"fmt"
)

// FormatSize takes a list of file sizes in bytes and returns a formatted string
// representing the total size (e.g., "1023 bytes", "1.1 KB", "7.89 MB").
func FormatSize(fileSizes []int) string {
	var totalBytes int64
	for _, size := range fileSizes {
		totalBytes += int64(size)
	}
	return FormatBytes(totalBytes)
}

// FormatBytes formats a byte count into a human-readable string
func FormatBytes(bytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	threshold := int64(1024)
	size := float64(bytes)
	unitIndex := 0

	for size >= float64(threshold) && unitIndex < len(units)-1 {
		size /= float64(threshold)
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%d %s", int64(size), units[unitIndex])
	}
	return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}

// FormatCount formats a count with proper singular/plural form
func FormatCount(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %s", count, plural)
}
