package api

import (
	"net"
	"regexp"
	"strings"
)

func isValidBucketName(name string) (bool, string) {
	// Check length constraint: between 3 and 63 characters
	if len(name) < 3 || len(name) > 63 {
		return false, "Bucket name must be between 3 and 63 characters"
	}

	// Check for invalid characters: Only lowercase letters, numbers, hyphens, and periods are allowed
	matched, _ := regexp.MatchString("^[a-z0-9.-]+$", name)
	if !matched {
		return false, "Bucket name can only contain lowercase letters, numbers, hyphens, and periods"
	}

	// Bucket name must start and end with a letter or number
	if !regexp.MustCompile("^[a-z0-9]").MatchString(name) || !regexp.MustCompile("[a-z0-9]$").MatchString(name) {
		return false, "Bucket name must start and end with a letter or number"
	}

	// Periods (.) cannot be adjacent to each other
	if strings.Contains(name, "..") {
		return false, "Periods (.) cannot be adjacent to each other"
	}

	// Hyphens (-) cannot be adjacent to each other or at the beginning or end
	if strings.Contains(name, "--") || strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return false, "Hyphens (-) cannot be adjacent to each other or at the beginning or end"
	}

	// Check if it's a valid IP address (IPv4 or IPv6)
	if net.ParseIP(name) != nil {
		return false, "Bucket name cannot be an IP address"
	}

	// Check if it's a DNS-compliant name
	// A DNS name must only contain letters, numbers, hyphens, and dots
	// It cannot start or end with a hyphen
	re := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])*(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])*)*$`)
	isDNSCompliant := re.MatchString(name)
	if !isDNSCompliant {
		return false, "Bucket name must be a valid DNS-compliant name, containing only letters, numbers, hyphens, and periods. It cannot start or end with a hyphen or period."
	}

	return true, ""
}

func isValidObjectKey(key string) (bool, string) {
	// Check length: object key should be <= 1024 characters
	if len(key) > 1024 {
		return false, "Object key must be less than or equal to 1024 characters"
	}

	// Check for invalid characters (control characters are not allowed)
	matched, _ := regexp.MatchString("^[\\x20-\\x7E]*$", key) // Only printable ASCII characters
	if !matched {
		return false, "Object key can only contain printable ASCII characters"
	}

	// Keys cannot contain multiple consecutive slashes (for clarity, but not a hard rule)
	if strings.Contains(key, "//") {
		return false, "Object key cannot contain consecutive slashes"
	}

	// Object key must not start or end with a slash
	if strings.HasPrefix(key, "/") || strings.HasSuffix(key, "/") {
		return false, "Object key cannot start or end with a slash"
	}

	// Check if it's a valid IP address (IPv4 or IPv6)
	if net.ParseIP(key) != nil {
		return false, "Object key cannot be an IP address"
	}

	// Check if it's a DNS-compliant name (for objects stored in DNS-like paths)
	re := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])*(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])*)*$`)
	isDNSCompliant := re.MatchString(key)
	if !isDNSCompliant {
		return false, "Object key must be a valid DNS-compliant name, containing only letters, numbers, hyphens, and periods. It cannot start or end with a hyphen or period."
	}

	return true, ""
}
