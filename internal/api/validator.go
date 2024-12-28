package api

import (
	"regexp"
	"strings"
)

func isValidBucketName(name string) bool {
	// Check length constraint: between 3 and 63 characters
	if len(name) < 3 || len(name) > 63 {
		return false
	}

	// Check for invalid characters: Only lowercase letters, numbers, hyphens, and periods are allowed
	matched, _ := regexp.MatchString("^[a-z0-9.-]+$", name)
	if !matched {
		return false
	}

	// Bucket name must start and end with a letter or number
	if !regexp.MustCompile("^[a-z0-9]").MatchString(name) || !regexp.MustCompile("[a-z0-9]$").MatchString(name) {
		return false
	}

	// Periods (.) cannot be adjacent to each other
	if strings.Contains(name, "..") {
		return false
	}

	// Hyphens (-) cannot be adjacent to each other or at the beginning or end
	if strings.Contains(name, "--") || strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return false
	}

	return true
}

func isValidObjectKey(key string) bool {
	// Check length: object key should be <= 1024 characters
	if len(key) > 1024 {
		return false
	}

	// Check for invalid characters (control characters are not allowed)
	matched, _ := regexp.MatchString("^[\\x20-\\x7E]*$", key) // Only printable ASCII characters
	if !matched {
		return false
	}

	// Keys cannot contain multiple consecutive slashes (for clarity, but not a hard rule)
	if strings.Contains(key, "//") {
		return false
	}

	return true
}
