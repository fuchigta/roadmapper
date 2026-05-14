// Package content は content/<id>.md を読み込み、frontmatter と本文に分離する。
package content

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter は content/<id>.md のヘッダ部分。
type Frontmatter struct {
	Title string `yaml:"title"` // 省略可 (roadmap.yml を正とする)
	Links []Link `yaml:"links"`
}

// Link は参考資料リンク。
type Link struct {
	Title string `yaml:"title" json:"title"`
	URL   string `yaml:"url"   json:"url"`
}

// Doc は1つのノードコンテンツを表す。
type Doc struct {
	ID          string
	Frontmatter Frontmatter
	Body        string // frontmatter 除去後の Markdown 本文
	// RelDir は content/ からの相対ディレクトリ (`/` 区切り)。
	// ルート直下のファイルは空文字列。例: content/frontend/html.md → "frontend"
	RelDir string
}

// Asset は content/ 配下にある非 .md の静的ファイル。
type Asset struct {
	SrcPath string // 実ファイルパス (filepath.Join 区切り)
	RelPath string // content/ からの相対パス (`/` 区切り)
}

// LoadDir は dir/ 以下を再帰的にスキャンして <id>.md ファイルをすべて読み込む。
//
// map のキーは 2 種類登録される:
//  1. 相対パスキー: "frontend/html" のようなスラッシュ区切りの相対パス (常に登録)
//  2. 末尾名フォールバックキー: "html" のようなファイル名のみ (一意な場合のみ登録)
//
// 同じファイル名が複数ディレクトリに存在する場合、フォールバックキーは曖昧として登録されない。
// ルート直下に同名ファイルが存在する場合はルートが優先される。
func LoadDir(dir string) (map[string]*Doc, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return map[string]*Doc{}, nil
	}

	docs := map[string]*Doc{}
	// base → 登録済み relID。空文字列は「曖昧」マーク
	baseOwners := map[string]string{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// 隠しディレクトリはスキップ
		if d.IsDir() && path != dir && strings.HasPrefix(d.Name(), ".") {
			return fs.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		relID := strings.TrimSuffix(filepath.ToSlash(rel), ".md")

		doc, err := Load(path, relID)
		if err != nil {
			return err
		}
		// 相対ディレクトリ (`/` 区切り、ルートは空文字列)
		relDir := filepath.ToSlash(filepath.Dir(rel))
		if relDir == "." {
			relDir = ""
		}
		doc.RelDir = relDir
		docs[relID] = doc

		// フォールバックキー (ルートファイルは relID == base なので登録不要)
		base := strings.TrimSuffix(d.Name(), ".md")
		if base == relID {
			return nil
		}
		switch owner, seen := baseOwners[base]; {
		case !seen:
			// ルート直下に同名ファイルがあれば docs[base] が既にキー base で登録済み
			// (ルートファイルは relID == base なので baseOwners には記録されない)
			if _, rootExists := docs[base]; !rootExists {
				docs[base] = doc
				baseOwners[base] = relID
			}
		case owner != "":
			// 2 件目の衝突 → 曖昧化してフォールバックキーを削除
			delete(docs, base)
			baseOwners[base] = ""
			// owner == "" は既に曖昧確定、何もしない
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("content ディレクトリを読み込めません: %w", err)
	}
	return docs, nil
}

// LoadAssets は dir/ 配下の非 .md ファイルを再帰スキャンしてアセットを返す。
// 隠しディレクトリ (`.` で始まる) と exclude にマッチするものはスキップする。
// 結果は RelPath 昇順でソートされる。
func LoadAssets(dir string, exclude []string) ([]Asset, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}

	var assets []Asset
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != dir && strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		if MatchAny(exclude, relSlash) {
			return nil
		}
		assets = append(assets, Asset{SrcPath: p, RelPath: relSlash})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("content アセットのスキャンに失敗: %w", err)
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].RelPath < assets[j].RelPath })
	return assets, nil
}

// Load は指定ファイルを読み込み、frontmatter と本文に分離して Doc を返す。
func Load(path, id string) (*Doc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s の読み込みに失敗: %w", path, err)
	}
	return Parse(data, id)
}

// Parse は raw バイト列を frontmatter と本文に分離して Doc を返す。
func Parse(data []byte, id string) (*Doc, error) {
	doc := &Doc{ID: id}

	// frontmatter は --- で囲まれた先頭ブロック
	if bytes.HasPrefix(data, []byte("---\n")) || bytes.HasPrefix(data, []byte("---\r\n")) {
		end := findFrontmatterEnd(data)
		if end > 0 {
			fm := data[4:end] // "---\n" の後から
			if err := yaml.Unmarshal(fm, &doc.Frontmatter); err != nil {
				return nil, fmt.Errorf("%s の frontmatter パースに失敗: %w", id, err)
			}
			// 閉じ --- の後の改行をスキップ
			rest := data[end+3:]
			if len(rest) > 0 && rest[0] == '\n' {
				rest = rest[1:]
			} else if len(rest) > 1 && rest[0] == '\r' && rest[1] == '\n' {
				rest = rest[2:]
			}
			doc.Body = string(rest)
			return doc, nil
		}
	}

	doc.Body = string(data)
	return doc, nil
}

// findFrontmatterEnd は "---\n" 開始後の閉じ "---" の先頭インデックスを返す。
// 見つからない場合は -1 を返す。
func findFrontmatterEnd(data []byte) int {
	// 最初の "---\n" をスキップ
	start := 4
	for i := start; i < len(data)-2; i++ {
		if data[i] == '-' && data[i+1] == '-' && data[i+2] == '-' {
			if i == 0 || data[i-1] == '\n' {
				return i
			}
		}
	}
	return -1
}
