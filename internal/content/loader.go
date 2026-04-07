// Package content は content/<id>.md を読み込み、frontmatter と本文に分離する。
package content

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter は content/<id>.md のヘッダ部分。
type Frontmatter struct {
	Title string   `yaml:"title"` // 省略可 (roadmap.yml を正とする)
	Links []Link   `yaml:"links"`
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
}

// LoadDir は dir/ 以下の <id>.md ファイルをすべて読み込み、map[id]Doc を返す。
func LoadDir(dir string) (map[string]*Doc, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return map[string]*Doc{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("content ディレクトリを読み込めません: %w", err)
	}

	docs := map[string]*Doc{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".md")
		path := filepath.Join(dir, e.Name())
		doc, err := Load(path, id)
		if err != nil {
			return nil, err
		}
		docs[id] = doc
	}
	return docs, nil
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
