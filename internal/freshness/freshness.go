// Package freshness provides utilities for checking data freshness.
package freshness

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Thresholds defines how old data can be before it's considered stale.
var Thresholds = map[string]time.Duration{
	"daily":   24 * time.Hour,
	"weekly":  7 * 24 * time.Hour,
	"monthly": 30 * 24 * time.Hour,
	"manual":  365 * 24 * time.Hour * 100, // ~100 years, effectively never stale
}

// Check determines if a file is fresh based on the freshness policy.
// Returns (isFresh, statusMessage).
func Check(path string, freshnessPolicy string) (bool, string) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, "missing"
	}
	if err != nil {
		return false, fmt.Sprintf("error: %v", err)
	}

	age := time.Since(info.ModTime())
	maxAge, ok := Thresholds[freshnessPolicy]
	if !ok {
		maxAge = Thresholds["manual"]
	}

	if age < maxAge {
		return true, formatAge(age, "fresh")
	}
	return false, formatAge(age, "stale")
}

// CheckWithRoot checks freshness using a path relative to projectRoot.
func CheckWithRoot(projectRoot, relativePath, freshnessPolicy string) (bool, string) {
	fullPath := filepath.Join(projectRoot, relativePath)
	return Check(fullPath, freshnessPolicy)
}

// formatAge returns a human-readable age string.
func formatAge(age time.Duration, prefix string) string {
	hours := int(age.Hours())
	days := hours / 24

	if days == 0 {
		if hours == 0 {
			mins := int(age.Minutes())
			return fmt.Sprintf("%s (%dm ago)", prefix, mins)
		}
		return fmt.Sprintf("%s (%dh ago)", prefix, hours)
	}
	return fmt.Sprintf("%s (%dd ago)", prefix, days)
}

// IsFresh is a convenience function that returns only the boolean.
func IsFresh(path string, freshnessPolicy string) bool {
	fresh, _ := Check(path, freshnessPolicy)
	return fresh
}

