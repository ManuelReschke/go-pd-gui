package uploadutil

import "testing"

func TestMultiFileProgress(t *testing.T) {
	tests := []struct {
		name      string
		fileIndex int
		fileTotal int
		fraction  float64
		want      float64
	}{
		{name: "first file half of two", fileIndex: 0, fileTotal: 2, fraction: 0.5, want: 0.25},
		{name: "second file done of two", fileIndex: 1, fileTotal: 2, fraction: 1, want: 1},
		{name: "single file complete", fileIndex: 0, fileTotal: 1, fraction: 1, want: 1},
		{name: "clamps fraction high", fileIndex: 0, fileTotal: 2, fraction: 2, want: 0.5},
		{name: "clamps fraction low", fileIndex: 1, fileTotal: 2, fraction: -1, want: 0.5},
		{name: "total zero treated as one", fileIndex: 0, fileTotal: 0, fraction: 0.5, want: 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MultiFileProgress(tt.fileIndex, tt.fileTotal, tt.fraction)
			if got != tt.want {
				t.Fatalf("MultiFileProgress(%d, %d, %v) = %v, want %v",
					tt.fileIndex, tt.fileTotal, tt.fraction, got, tt.want)
			}
		})
	}
}

func TestJoinURLs(t *testing.T) {
	if got := JoinURLs(nil); got != "" {
		t.Fatalf("nil: got %q", got)
	}
	if got := JoinURLs([]string{}); got != "" {
		t.Fatalf("empty: got %q", got)
	}
	if got := JoinURLs([]string{"https://a", "", "https://b"}); got != "https://a\nhttps://b" {
		t.Fatalf("got %q", got)
	}
}
