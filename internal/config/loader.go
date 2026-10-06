package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load は指定パスの roadmap.yml を読み込み、Config を返す。
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("roadmap.yml を読み込めません: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("roadmap.yml の解析に失敗しました: %w", err)
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Site.BrandColor == "" {
		cfg.Site.BrandColor = "#4f46e5"
	}
	if cfg.Site.EditBranch == "" {
		cfg.Site.EditBranch = "main"
	}
	if cfg.Site.Layout.RankDir == "" {
		cfg.Site.Layout.RankDir = "TB"
	}
	if cfg.Site.Layout.NodeSep == 0 {
		cfg.Site.Layout.NodeSep = 50
	}
	if cfg.Site.Layout.RankSep == 0 {
		cfg.Site.Layout.RankSep = 80
	}
	if cfg.Site.Analytics.Events == nil {
		cfg.Site.Analytics.Events = boolPtr(true)
	}
	if cfg.Site.Analytics.ExcludeSearch == nil {
		cfg.Site.Analytics.ExcludeSearch = boolPtr(true)
	}
	// 未指定の項目は、指定済みの値と矛盾しないよう既定値を範囲内に収めて補完する
	p := &cfg.Site.Panel
	if p.MaxWidth == 0 {
		p.MaxWidth = max(960, p.Width, p.MinWidth)
	}
	if p.MinWidth == 0 {
		p.MinWidth = 320
		if p.Width > 0 {
			p.MinWidth = min(p.MinWidth, p.Width)
		}
		p.MinWidth = min(p.MinWidth, p.MaxWidth)
	}
	if p.Width == 0 {
		p.Width = min(max(520, p.MinWidth), p.MaxWidth)
	}

	for ri := range cfg.Roadmaps {
		applyNodeDefaults(cfg.Roadmaps[ri].Nodes)
	}
}

func applyNodeDefaults(nodes []*Node) {
	for _, n := range nodes {
		if n.Type == "" {
			n.Type = NodeTypeRequired
		}
		applyNodeDefaults(n.Children)
	}
}

func boolPtr(b bool) *bool { return &b }
