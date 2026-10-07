package config

// Site はサイト全体のメタ情報を保持する。
type Site struct {
	Title         string        `yaml:"title"`
	Description   string        `yaml:"description"`
	BrandColor    string        `yaml:"brandColor"`
	Author        string        `yaml:"author"`
	License       string        `yaml:"license"`
	Repo          string        `yaml:"repo"`
	EditBranch    string        `yaml:"editBranch"`
	BasePath      string        `yaml:"basePath"`
	SiteURL       string        `yaml:"siteUrl"` // 公開URL (sitemap/RSS/OGP 用, 例: https://example.com)
	Layout        Layout        `yaml:"layout"`
	ProgressSync  ProgressSync  `yaml:"progressSync"`
	Analytics     Analytics     `yaml:"analytics"`
	Panel         Panel         `yaml:"panel"`
	ContentAssets ContentAssets `yaml:"contentAssets"`
}

// ContentAssets は content/ 配下の静的ファイル（画像など）コピー設定。
type ContentAssets struct {
	// Exclude は content/ からの相対パスに対するコピー除外 glob パターン。
	// `*` は単一セグメント、`**` は複数セグメントをマッチする。
	Exclude []string `yaml:"exclude"`
}

// Layout は dagre に渡すレイアウトパラメータ。
type Layout struct {
	RankDir string  `yaml:"rankDir"` // TB / LR / BT / RL
	NodeSep float64 `yaml:"nodeSep"`
	RankSep float64 `yaml:"rankSep"`
}

// Panel は記事サイドパネルの幅設定 (単位: px)。
type Panel struct {
	Width    int `yaml:"width"`    // 初期幅 (既定 520)
	MinWidth int `yaml:"minWidth"` // 最小幅 (既定 320)
	MaxWidth int `yaml:"maxWidth"` // 最大幅 (既定 960)
}

// ProgressSync は進捗バックエンド同期の設定。
type ProgressSync struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"` // 末尾スラッシュなしのベース URL
}

// アクセス解析 provider。
const (
	AnalyticsUmami       = "umami"
	AnalyticsPlausible   = "plausible"
	AnalyticsGoatCounter = "goatcounter"
	AnalyticsCustom      = "custom"
)

// Analytics はアクセス解析の設定。Provider が空なら無効。
type Analytics struct {
	Provider      string   `yaml:"provider"`      // umami | plausible | goatcounter | custom
	ScriptURL     string   `yaml:"scriptUrl"`     // 解析スクリプトの URL
	SiteID        string   `yaml:"siteId"`        // umami: data-website-id / plausible: data-domain / goatcounter: data-goatcounter
	Domains       []string `yaml:"domains"`       // umami の data-domains (計測を許可するホスト名)
	Events        *bool    `yaml:"events"`        // カスタムイベント送信 (既定 true)
	ExcludeSearch *bool    `yaml:"excludeSearch"` // クエリ文字列を記録しない (既定 true, umami のみ)
	Head          string   `yaml:"head"`          // provider: custom のとき <head> にそのまま挿入する HTML
}

// Enabled は解析が有効 (provider 指定あり) かを返す。
func (a Analytics) Enabled() bool { return a.Provider != "" }

// EventsEnabled はカスタムイベント送信が有効かを返す (nil は true)。
func (a Analytics) EventsEnabled() bool { return a.Events == nil || *a.Events }

// ExcludeSearchEnabled はクエリ文字列の除外が有効かを返す (nil は true)。
func (a Analytics) ExcludeSearchEnabled() bool { return a.ExcludeSearch == nil || *a.ExcludeSearch }

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
	Content       string     `yaml:"content"`       // content/<path>.md を明示指定 (任意、拡張子なし)
	Draft         bool       `yaml:"draft"`         // 下書きノード (任意。公開ビルドでは非活性表示にする)
}

// Roadmap は1つのロードマップ全体を表す。
type Roadmap struct {
	ID          string  `yaml:"id"`
	Title       string  `yaml:"title"`
	Description string  `yaml:"description"`
	Nodes       []*Node `yaml:"nodes"`
	// Changelog はロードマップ全体の改版履歴 (任意)。
	Changelog []ChangelogEntry `yaml:"changelog"`
}

// ChangelogEntry は改版履歴の 1 項目。
type ChangelogEntry struct {
	Date    string   `yaml:"date"` // "2006-01-02" 形式
	Summary string   `yaml:"summary"`
	Nodes   []string `yaml:"nodes"` // 関連ノード ID (任意)
}

// ChangelogDateLayout は改版履歴の日付フォーマット。
const ChangelogDateLayout = "2006-01-02"

// Config は roadmap.yml 全体のルート構造体。
type Config struct {
	Site     Site      `yaml:"site"`
	Roadmaps []Roadmap `yaml:"roadmaps"`
}
