package content

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_frontmatterUpdatedChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "html.md")
	src := "---\nupdated: 2026-10-05\nchanges:\n  - date: 2026-09-01\n    summary: 初版\n  - {date: \"2026-10-05\", summary: 改訂}\n---\n# HTML\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := Load(path, "html")
	if err != nil {
		t.Fatal(err)
	}
	fm := doc.Frontmatter
	if fm.Updated != "2026-10-05" {
		t.Errorf("Updated = %q", fm.Updated)
	}
	if len(fm.Changes) != 2 || fm.Changes[0] != (Change{"2026-09-01", "初版"}) || fm.Changes[1] != (Change{"2026-10-05", "改訂"}) {
		t.Errorf("Changes = %+v", fm.Changes)
	}
}
