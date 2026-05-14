package content

import "testing"

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		{"*.psd", "foo.psd", true},
		{"*.psd", "sub/foo.psd", false},
		{"drafts/**", "drafts/foo.txt", true},
		{"drafts/**", "drafts/sub/foo.txt", true},
		{"drafts/**", "drafts", true},
		{"drafts/**", "other/foo.txt", false},
		{"**/*.png", "foo.png", true},
		{"**/*.png", "a/b/foo.png", true},
		{"**/raw/*", "frontend/raw/dom.png", true},
		{"**/raw/*", "frontend/raw/sub/dom.png", false},
		{"**/raw/**", "frontend/raw/sub/dom.png", true},
		{"foo/bar", "foo/bar", true},
		{"foo/bar", "foo/bar/baz", false},
		{"foo/*", "foo/bar", true},
		{"foo/*", "foo/bar/baz", false},
		{"**", "anything/at/any/depth.txt", true},
	}
	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.name, func(t *testing.T) {
			got := MatchGlob(tt.pattern, tt.name)
			if got != tt.want {
				t.Errorf("MatchGlob(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
			}
		})
	}
}

func TestMatchAny(t *testing.T) {
	patterns := []string{"*.psd", "drafts/**"}
	if !MatchAny(patterns, "foo.psd") {
		t.Error("foo.psd should match")
	}
	if !MatchAny(patterns, "drafts/x.md") {
		t.Error("drafts/x.md should match")
	}
	if MatchAny(patterns, "foo.png") {
		t.Error("foo.png should not match")
	}
}
