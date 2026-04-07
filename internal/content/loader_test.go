package content_test

import (
	"testing"

	"github.com/fuchigta/roadmapper/internal/content"
)

func TestParse_withFrontmatter(t *testing.T) {
	raw := `---
title: HTML
links:
  - { title: "MDN", url: "https://developer.mozilla.org" }
---

## 学ぶこと

本文のテキスト。
`
	doc, err := content.Parse([]byte(raw), "html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Frontmatter.Title != "HTML" {
		t.Errorf("expected title=HTML, got %q", doc.Frontmatter.Title)
	}
	if len(doc.Frontmatter.Links) != 1 {
		t.Errorf("expected 1 link, got %d", len(doc.Frontmatter.Links))
	}
	if doc.Frontmatter.Links[0].URL != "https://developer.mozilla.org" {
		t.Errorf("unexpected link URL: %s", doc.Frontmatter.Links[0].URL)
	}
	if doc.Body == "" {
		t.Error("body should not be empty")
	}
	if doc.ID != "html" {
		t.Errorf("expected id=html, got %q", doc.ID)
	}
}

func TestParse_withoutFrontmatter(t *testing.T) {
	raw := "## 本文だけ\n\nテキスト。\n"
	doc, err := content.Parse([]byte(raw), "node1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Body != raw {
		t.Errorf("body should equal original, got %q", doc.Body)
	}
	if doc.Frontmatter.Title != "" {
		t.Error("frontmatter should be empty")
	}
}

func TestParse_emptyBody(t *testing.T) {
	raw := "---\ntitle: Empty\n---\n"
	doc, err := content.Parse([]byte(raw), "empty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Frontmatter.Title != "Empty" {
		t.Errorf("expected title=Empty, got %q", doc.Frontmatter.Title)
	}
	if doc.Body != "" {
		t.Errorf("expected empty body, got %q", doc.Body)
	}
}
