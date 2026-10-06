package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fuchigta/roadmapper/internal/repo"
)

func TestResolveBranch(t *testing.T) {
	tests := []struct {
		name     string
		flag     string
		head     string // 空なら .git を作らない
		wantBr   string
		wantNote bool
		wantErr  bool
	}{
		{"フラグ優先", "release", "ref: refs/heads/master\n", "release", false, false},
		{"フラグ不正", "a b", "", "", false, true},
		{"HEAD から取得", "", "ref: refs/heads/master\n", "master", false, false},
		{"detached はフォールバック", "", "0123456789abcdef0123456789abcdef01234567\n", "main", true, false},
		{"不正な文字を含むブランチはフォールバック", "", "ref: refs/heads/a#b\n", "main", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.head != "" {
				if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte(tt.head), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			got, note, err := resolveBranch(tt.flag, dir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantBr {
				t.Errorf("branch = %q, want %q", got, tt.wantBr)
			}
			if (note != "") != tt.wantNote {
				t.Errorf("note = %q, wantNote %v", note, tt.wantNote)
			}
		})
	}
}

func TestResolveBranch_noRepo(t *testing.T) {
	dir := t.TempDir()
	if _, ok, _ := repo.FindRoot(dir); ok {
		t.Skip("上位ディレクトリに .git が存在する")
	}
	got, note, err := resolveBranch("", dir)
	if err != nil || got != "main" || note == "" {
		t.Errorf("got (%q,%q,%v), want main with note", got, note, err)
	}
}
