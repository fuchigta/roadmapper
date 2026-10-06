package repo

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, path, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPathPrefix(t *testing.T) {
	tests := []struct {
		name  string
		setup func(root string)
		sub   string
		want  string
	}{
		{"ルート直下", func(r string) { _ = os.MkdirAll(filepath.Join(r, ".git"), 0o755) }, "", ""},
		{".git ディレクトリ + サブディレクトリ", func(r string) { _ = os.MkdirAll(filepath.Join(r, ".git"), 0o755) }, "docs", "docs/"},
		{".git ファイル (worktree)", func(r string) { write(t, filepath.Join(r, ".git"), "gitdir: /x/y\n") }, "a/b", "a/b/"},
		{".git なし", func(r string) {}, "docs", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(root)
			dir := filepath.Join(root, filepath.FromSlash(tt.sub))
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			got, err := PathPrefix(dir)
			if err != nil {
				t.Fatal(err)
			}
			if tt.name == ".git なし" {
				// TempDir の上位に .git がある環境では検証できないためスキップ
				if _, ok, _ := FindRoot(dir); ok {
					t.Skip("上位ディレクトリに .git が存在する")
				}
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCurrentBranch(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(root string)
		want    string
		wantOK  bool
		wantErr bool
	}{
		{"master", func(r string) { write(t, filepath.Join(r, ".git", "HEAD"), "ref: refs/heads/master\n") }, "master", true, false},
		{"スラッシュ入り", func(r string) { write(t, filepath.Join(r, ".git", "HEAD"), "ref: refs/heads/feat/x\n") }, "feat/x", true, false},
		{"detached", func(r string) {
			write(t, filepath.Join(r, ".git", "HEAD"), "0123456789abcdef0123456789abcdef01234567\n")
		}, "", false, false},
		{"gitdir 相対", func(r string) {
			write(t, filepath.Join(r, ".git"), "gitdir: wt\n")
			write(t, filepath.Join(r, "wt", "HEAD"), "ref: refs/heads/develop\n")
		}, "develop", true, false},
		{"gitdir 絶対", func(r string) {
			write(t, filepath.Join(r, ".git"), "gitdir: "+filepath.ToSlash(filepath.Join(r, "abs"))+"\n")
			write(t, filepath.Join(r, "abs", "HEAD"), "ref: refs/heads/trunk\n")
		}, "trunk", true, false},
		{"HEAD なし", func(r string) { _ = os.MkdirAll(filepath.Join(r, ".git"), 0o755) }, "", false, true},
		{".git なし", func(r string) {}, "", false, true},
		{".git ファイル不正", func(r string) { write(t, filepath.Join(r, ".git"), "garbage") }, "", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(root)
			got, ok, err := CurrentBranch(root)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("got (%q,%v), want (%q,%v)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
