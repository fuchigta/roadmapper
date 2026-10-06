// Package repo は Git リポジトリのルート検出とブランチ名の読み取りを行う。
// git コマンドは実行せず、.git の内容を直接読む (外部バイナリ依存ゼロ)。
package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindRoot は dir から上位に向かって .git (ディレクトリまたはファイル) を探し、
// 見つかったリポジトリルートの絶対パスを返す。見つからなければ ok=false。
func FindRoot(dir string) (root string, ok bool, err error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", false, fmt.Errorf("パスの絶対化に失敗: %w", err)
	}
	for cur := abs; ; {
		if _, statErr := os.Stat(filepath.Join(cur, ".git")); statErr == nil {
			return cur, true, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", false, nil
		}
		cur = parent
	}
}

// PathPrefix は dir のリポジトリルートからの相対パスを `/` 区切り・末尾スラッシュ付きで返す。
// 例: ルート直下なら "", <root>/docs なら "docs/"。リポジトリ外なら "" を返す。
func PathPrefix(dir string) (string, error) {
	root, ok, err := FindRoot(dir)
	if err != nil || !ok {
		return "", err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("パスの絶対化に失敗: %w", err)
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", fmt.Errorf("リポジトリルートからの相対パス計算に失敗: %w", err)
	}
	rel = filepath.ToSlash(rel)
	if rel == "." || rel == "" {
		return "", nil
	}
	return rel + "/", nil
}

// CurrentBranch は root の .git/HEAD から現在のブランチ名を返す。
// .git がファイル (worktree / submodule) の場合は gitdir を辿る。
// detached HEAD の場合は ok=false を返す。
func CurrentBranch(root string) (branch string, ok bool, err error) {
	gitDir := filepath.Join(root, ".git")
	fi, err := os.Stat(gitDir)
	if err != nil {
		return "", false, fmt.Errorf(".git が見つかりません: %w", err)
	}
	if !fi.IsDir() {
		data, err := os.ReadFile(gitDir)
		if err != nil {
			return "", false, fmt.Errorf(".git ファイルの読み込みに失敗: %w", err)
		}
		line := strings.TrimSpace(string(data))
		p, found := strings.CutPrefix(line, "gitdir:")
		if !found {
			return "", false, fmt.Errorf(".git ファイルの形式が不正です: %q", line)
		}
		p = strings.TrimSpace(p)
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, filepath.FromSlash(p))
		}
		gitDir = p
	}
	head, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return "", false, fmt.Errorf("HEAD の読み込みに失敗: %w", err)
	}
	ref, found := strings.CutPrefix(strings.TrimSpace(string(head)), "ref: refs/heads/")
	if !found || ref == "" {
		return "", false, nil
	}
	return ref, true, nil
}
