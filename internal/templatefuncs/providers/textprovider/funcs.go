package textprovider

import (
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

// IsEmptyString checks if a string is empty.
func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsValidValue checks if a given value is in the allowedValues list.
func IsValidValue(value string, allowedValues []string) bool {
	return slices.Contains(allowedValues, value)
}

// NormalizePath normalizes a path by removing dots, leading/trailing slashes, and white spaces.
func NormalizePath(input string) string {
	// Trim spaces
	input = strings.TrimSpace(input)

	// Remove sequences of only dots
	if strings.Trim(input, ".") == "" {
		return ""
	}

	// Clean the path to resolve ./ and ../
	cleaned := filepath.Clean(input)

	// Remove leading and trailing slashes
	return strings.Trim(cleaned, "/")
}

// TitleCase capitalizes the first letter of a word and preserves the rest of the word as-is.
func TitleCase(word string) string {
	if len(word) == 0 {
		return ""
	}

	runes := []rune(word)
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}

// SnakeToTitle converts a snake_case string to Title Case.
func SnakeToTitle(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		words[i] = TitleCase(word)
	}
	return strings.Join(words, " ")
}
