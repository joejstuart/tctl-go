// Package util provides shared utility functions for tctl.
package util

import (
	"regexp"
	"strings"
)

// StopWords are common words to exclude from keyword extraction.
var StopWords = map[string]bool{
	"a": true, "an": true, "the": true, "is": true, "are": true,
	"was": true, "were": true, "be": true, "been": true,
	"and": true, "or": true, "but": true, "for": true, "to": true,
	"from": true, "with": true, "in": true, "on": true, "of": true,
	"at": true, "by": true, "this": true, "that": true, "it": true,
	"its": true, "does": true, "do": true,
}

// ExtractKeywords extracts meaningful words from text, excluding stop words.
func ExtractKeywords(text string) []string {
	re := regexp.MustCompile(`\b[a-zA-Z]{3,}\b`)
	words := re.FindAllString(strings.ToLower(text), -1)

	var result []string
	for _, w := range words {
		if !StopWords[w] {
			result = append(result, w)
		}
	}
	return result
}

// Min returns the smaller of two integers.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of two integers.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

