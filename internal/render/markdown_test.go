package render_test

import (
	"strings"
	"testing"

	"github.com/fuchigta/roadmapper/internal/render"
)

func TestRenderMarkdown_heading(t *testing.T) {
	html, err := render.RenderMarkdown("## こんにちは\n\nテスト本文。\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "<h2") {
		t.Errorf("expected <h2> in output, got: %s", html)
	}
	if !strings.Contains(html, "こんにちは") {
		t.Errorf("expected heading text in output")
	}
}

func TestRenderMarkdown_codeblock(t *testing.T) {
	src := "```go\nfmt.Println(\"hello\")\n```\n"
	html, err := render.RenderMarkdown(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// chroma ハイライトされると <pre> に変換される
	if !strings.Contains(html, "<pre") {
		t.Errorf("expected <pre> in output, got: %s", html)
	}
	if !strings.Contains(html, "Println") {
		t.Errorf("expected code content in output")
	}
}

func TestRenderMarkdown_mermaid(t *testing.T) {
	src := "```mermaid\ngraph LR; A-->B\n```\n"
	html, err := render.RenderMarkdown(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, `class="mermaid"`) {
		t.Errorf("expected mermaid passthrough, got: %s", html)
	}
	if strings.Contains(html, "<code") {
		t.Errorf("mermaid should not be inside <code>")
	}
}

func TestRenderMarkdown_checklist_not_disabled(t *testing.T) {
	src := "- [ ] タスク1\n- [x] タスク2\n"
	html, err := render.RenderMarkdown(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(html, `disabled`) {
		t.Errorf("checklist should not have disabled attribute, got: %s", html)
	}
	if !strings.Contains(html, `type="checkbox"`) {
		t.Errorf("expected checkbox in output")
	}
}

func TestRenderMarkdown_hasMermaid(t *testing.T) {
	withMermaid := "```mermaid\ngraph LR; A-->B\n```\n"
	withoutMermaid := "## テスト\n\n本文。\n"

	h1, _ := render.RenderMarkdown(withMermaid)
	h2, _ := render.RenderMarkdown(withoutMermaid)

	if !render.HasMermaid(h1) {
		t.Error("expected HasMermaid=true for mermaid block")
	}
	if render.HasMermaid(h2) {
		t.Error("expected HasMermaid=false for non-mermaid content")
	}
}
