// Package uploadutil holds pure helpers for multi-file upload progress and results.
package uploadutil

import "strings"

// MultiFileProgress maps a single-file fraction into an overall multi-file bar.
// fileIndex is 0-based; fraction is in [0,1].
func MultiFileProgress(fileIndex, fileTotal int, fraction float64) float64 {
	if fileTotal < 1 {
		fileTotal = 1
	}
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	return (float64(fileIndex) + fraction) / float64(fileTotal)
}

// JoinURLs returns newline-separated non-empty URLs for clipboard copy of a batch.
func JoinURLs(urls []string) string {
	if len(urls) == 0 {
		return ""
	}
	out := make([]string, 0, len(urls))
	for _, u := range urls {
		if u != "" {
			out = append(out, u)
		}
	}
	return strings.Join(out, "\n")
}
