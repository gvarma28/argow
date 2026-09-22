package main

import (
	"fmt"
	"os"
	"strings"
)

// splitCSV splits a comma-separated string into trimmed, non-empty entries.
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// dedupe removes repeated entries while keeping the first occurrence's position.
func dedupe[T comparable](s []T) []T {
	seen := make(map[T]bool, len(s))
	result := make([]T, 0, len(s))
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

func joinOps[T ~string](items []T, sep string) string {
	strs := make([]string, len(items))
	for i, v := range items {
		strs[i] = string(v)
	}
	return strings.Join(strs, sep)
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, Red+format+Reset+"\n", args...)
	os.Exit(1)
}
