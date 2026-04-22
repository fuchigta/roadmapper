package config

// Site はサイト全体のメタ情報を保持する。
type Site struct {
	Title        string       `yaml:"title"`
	Description  string       `yaml:"description"`
	BrandColor   string       `yaml:"brandColor"`
	Author       string       `yaml:"author"`
	License      string       `yaml:"license"`
	Repo         string       `yaml:"repo"`
	EditBranch   string       `yaml:"editBranch"`
	BasePath     string       `yaml:"basePath"`
	SiteURL      string       `yaml:"siteUrl"` // 公開URL (sitemap/RSS/OGP 用, 例: https://example.com)
	Layout       Layout       `yaml:"layout"`
	ProgressSync ProgressSync `yaml:"progressSync"`
}

// Layout は dagre に渡すレイアウトパラメータ。
type Layout struct {
	RankDir string  `yaml:"rankDir"` // TB / LR / BT / RL
	NodeSep float64 `yaml:"nodeSep"`
	RankSep float64 `yaml:"rankSep"`
}

// ProgressSync は進捗バックエンド同期の設定。
type ProgressSync struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"` // 末尾スラッシュなしのベース URL
}

// NodeType はノードの重要度を表す。
type NodeType string

const (
	NodeTypeRequired    NodeType = "required"
	NodeTypeOptional    NodeType = "optional"
	NodeTypeAlternative NodeType = "alternative"
)

// Difficulty はノードの難易度を表す。
type Difficulty string

const (
	DifficultyBeginner     Difficulty = "beginner"
	DifficultyIntermediate Difficulty = "intermediate"
	DifficultyAdvanced     Difficulty = "advanced"
)

// Link は参考資料リンク。
type Link struct {
	Title string `yaml:"title" json:"title"`
	URL   string `yaml:"url"   json:"url"`
}

// Node は1つの学習トピックを表す。
type Node struct {
	ID            string     `yaml:"id"`
	Title         string     `yaml:"title"`
	Type          NodeType   `yaml:"type"`
	X             *float64   `yaml:"x"` // 手動座標オーバーライド (任意)
	Y             *float64   `yaml:"y"`
	Parents       []string   `yaml:"parents"`       // 複数親 (DAG)
	Children      []*Node    `yaml:"children"`      // 子ノードは再帰的にネスト
	Links         []Link     `yaml:"links"`         // 参考資料リンク
	Difficulty    Difficulty `yaml:"difficulty"`    // 難易度 (任意)
	EstimatedTime string     `yaml:"estimatedTime"` // 推定所要時間 (任意, 例: "2h", "3d")
}

// Roadmap は1つのロードマップ全体を表す。
type Roadmap struct {
	ID          string  `yaml:"id"`
	Title       string  `yaml:"title"`
	Description string  `yaml:"description"`
	Nodes       []*Node `yaml:"nodes"`
}

// Config は roadmap.yml 全体のルート構造体。
type Config struct {
	Site     Site      `yaml:"site"`
	Roadmaps []Roadmap `yaml:"roadmaps"`
}
