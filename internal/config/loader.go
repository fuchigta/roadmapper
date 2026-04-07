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
