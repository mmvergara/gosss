package api

import (
	"bytes"
	"fmt"
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
	// Check if key is empty
	if len(key) == 0 {
		return false, "key cannot be empty"
	}

	// Check maximum length (1024 bytes for most regions)
	if len(key) > 1024 {
		return false, "key length cannot exceed 1024 bytes"
	}

	// Check for invalid characters
	invalidChars := []byte{
		0x00, // NULL
		0x0A, // Line Feed
		0x0D, // Carriage Return
	}

	for _, char := range invalidChars {
		if bytes.Contains([]byte(key), []byte{char}) {
			return false, "key contains invalid control characters"
		}
	}

	// Check for invalid prefixes
	invalidPrefixes := []string{
		".",
		"..",
		"-",
		"_",
	}

	for _, prefix := range invalidPrefixes {
		if strings.HasPrefix(key, prefix) {
			return false, fmt.Sprintf("key cannot start with %s", prefix)
		}
	}

	// Check for specific invalid sequences
	invalidSequences := []string{
		"//", // Double forward slashes
		"\\", // Backslashes
	}

	for _, seq := range invalidSequences {
		if strings.Contains(key, seq) {
			return false, fmt.Sprintf("key cannot contain %s", seq)
		}
	}

	// Check if key ends with forward slash (directory style)
	if strings.HasSuffix(key, "/") {
		return false, "key cannot end with forward slash"
	}

	// Additional safe character validation using regex
	safePattern := regexp.MustCompile(`^[a-zA-Z0-9!-_.*'()/]+$`)
	if !safePattern.MatchString(key) {
		return false, "key contains invalid characters"
	}

	return true, ""
}
