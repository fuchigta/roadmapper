package content_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fuchigta/roadmapper/internal/content"
)

func writeFile(t *testing.T, dir, rel, body string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadDir_recursive(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "intro.md", "# intro")
	writeFile(t, dir, "frontend/html.md", "# html")
	writeFile(t, dir, "frontend/css.md", "# css")

	docs, err := content.LoadDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, key := range []string{"intro", "frontend/html", "frontend/css"} {
		if docs[key] == nil {
			t.Errorf("docs[%q] が nil です", key)
		}
	}
	// フォールバックキー
	for _, key := range []string{"html", "css"} {
		if docs[key] == nil {
			t.Errorf("フォールバックキー docs[%q] が nil です", key)
		}
	}
	// intro はルート直下なのでフォールバック不要 (relID == base)
	if docs["intro"] == nil {
		t.Error("docs[\"intro\"] が nil です")
	}
	// フォールバックキーと相対パスキーは同じ実体を指す
	if docs["html"] != docs["frontend/html"] {
		t.Error("docs[\"html\"] と docs[\"frontend/html\"] が別の実体です")
	}
}

func TestLoadDir_ambiguousBase(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "frontend/html.md", "# html frontend")
	writeFile(t, dir, "backend/html.md", "# html backend")

	docs, err := content.LoadDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 完全パスキーは両方引ける
	if docs["frontend/html"] == nil {
		t.Error("docs[\"frontend/html\"] が nil です")
	}
	if docs["backend/html"] == nil {
		t.Error("docs[\"backend/html\"] が nil です")
	}
	// 曖昧なフォールバックキーは存在しない
	if docs["html"] != nil {
		t.Error("曖昧なフォールバックキー docs[\"html\"] は nil であるべきです")
	}
}

func TestLoadDir_rootWins(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "html.md", "# root html")
	writeFile(t, dir, "frontend/html.md", "# frontend html")

	docs, err := content.LoadDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if docs["html"] == nil {
		t.Fatal("docs[\"html\"] が nil です")
	}
	// ルート直下のファイルが優先される
	if docs["html"].ID != "html" {
		t.Errorf("docs[\"html\"].ID = %q, want \"html\"", docs["html"].ID)
	}
}

func TestLoadDir_deepNesting(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a/b/c.md", "# c")

	docs, err := content.LoadDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if docs["a/b/c"] == nil {
		t.Error("docs[\"a/b/c\"] が nil です")
	}
	if docs["c"] == nil {
		t.Error("末尾名フォールバック docs[\"c\"] が nil です")
	}
	if docs["c"] != docs["a/b/c"] {
		t.Error("フォールバックキーと相対パスキーが別実体です")
	}
}

func TestLoadDir_rootAfterSubdir(t *testing.T) {
	// WalkDir はアルファベット順なので foo/ は x.md より先に処理される。
	// ルート直下の x.md が後から書き込まれてもルート優先になることを確認。
	dir := t.TempDir()
	writeFile(t, dir, "foo/x.md", "# subdir x")
	writeFile(t, dir, "x.md", "# root x")

	docs, err := content.LoadDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if docs["x"] == nil {
		t.Fatal("docs[\"x\"] が nil です")
	}
	if docs["x"].ID != "x" {
		t.Errorf("docs[\"x\"].ID = %q, want \"x\" (ルート優先)", docs["x"].ID)
	}
	if docs["foo/x"] == nil {
		t.Error("docs[\"foo/x\"] が nil です")
	}
}

func TestLoadDir_missingDir(t *testing.T) {
	docs, err := content.LoadDir("/nonexistent/path/xyz")
	if err != nil {
		t.Fatalf("存在しないディレクトリはエラーを返すべきではない: %v", err)
	}
	if len(docs) != 0 {
		t.Errorf("空の map を期待しましたが %d 件あります", len(docs))
	}
}

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

func TestLoadDir_relDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "intro.md", "# intro")
	writeFile(t, dir, "frontend/html.md", "# html")
	writeFile(t, dir, "frontend/sub/css.md", "# css")

	docs, err := content.LoadDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if docs["intro"].RelDir != "" {
		t.Errorf("intro RelDir = %q, want \"\"", docs["intro"].RelDir)
	}
	if docs["frontend/html"].RelDir != "frontend" {
		t.Errorf("frontend/html RelDir = %q, want \"frontend\"", docs["frontend/html"].RelDir)
	}
	if docs["frontend/sub/css"].RelDir != "frontend/sub" {
		t.Errorf("frontend/sub/css RelDir = %q, want \"frontend/sub\"", docs["frontend/sub/css"].RelDir)
	}
}

func TestLoadAssets(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "intro.md", "# md, not asset")
	writeFile(t, dir, "frontend/images/dom.png", "PNGDATA")
	writeFile(t, dir, "frontend/notes.txt", "txt")
	writeFile(t, dir, "drafts/wip.png", "draft")
	writeFile(t, dir, "raw.psd", "psd")

	t.Run("no exclude", func(t *testing.T) {
		assets, err := content.LoadAssets(dir, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := map[string]bool{
			"frontend/images/dom.png": true,
			"frontend/notes.txt":      true,
			"drafts/wip.png":          true,
			"raw.psd":                 true,
		}
		if len(assets) != len(want) {
			t.Errorf("expected %d assets, got %d: %+v", len(want), len(assets), assets)
		}
		for _, a := range assets {
			if !want[a.RelPath] {
				t.Errorf("unexpected asset %q", a.RelPath)
			}
		}
	})

	t.Run("with exclude", func(t *testing.T) {
		assets, err := content.LoadAssets(dir, []string{"drafts/**", "*.psd"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := map[string]bool{}
		for _, a := range assets {
			got[a.RelPath] = true
		}
		if got["drafts/wip.png"] {
			t.Error("drafts/wip.png should be excluded")
		}
		if got["raw.psd"] {
			t.Error("raw.psd should be excluded")
		}
		if !got["frontend/images/dom.png"] {
			t.Error("frontend/images/dom.png should be included")
		}
	})

	t.Run("missing dir", func(t *testing.T) {
		assets, err := content.LoadAssets("/nonexistent/xyz", nil)
		if err != nil {
			t.Fatalf("missing dir should not error: %v", err)
		}
		if len(assets) != 0 {
			t.Errorf("expected no assets, got %d", len(assets))
		}
	})
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
