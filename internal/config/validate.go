package config

import (
	"fmt"
	"strings"
)

// ValidationError は1件のバリデーションエラーを表す。
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors は複数バリデーションエラーのまとまり。
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	msgs := make([]string, len(ve))
	for i, e := range ve {
		msgs[i] = "  - " + e.Error()
	}
	return "バリデーションエラー:\n" + strings.Join(msgs, "\n")
}

// Validate は Config の整合性を検査する。
func Validate(cfg *Config) error {
	var errs ValidationErrors

	if cfg.Site.Title == "" {
		errs = append(errs, ValidationError{"site.title", "タイトルは必須です"})
	}

	if len(cfg.Roadmaps) == 0 {
		errs = append(errs, ValidationError{"roadmaps", "ロードマップが1つも定義されていません"})
	}

	roadmapIDs := map[string]bool{}
	for ri, rm := range cfg.Roadmaps {
		prefix := fmt.Sprintf("roadmaps[%d]", ri)

		if rm.ID == "" {
			errs = append(errs, ValidationError{prefix + ".id", "id は必須です"})
		} else if roadmapIDs[rm.ID] {
			errs = append(errs, ValidationError{prefix + ".id", fmt.Sprintf("id %q が重複しています", rm.ID)})
		} else {
			roadmapIDs[rm.ID] = true
		}

		if rm.Title == "" {
			errs = append(errs, ValidationError{prefix + ".title", "title は必須です"})
		}

		nodeIDs := map[string]bool{}
		collectNodeIDs(rm.Nodes, nodeIDs)
		validateNodes(rm.Nodes, nodeIDs, prefix+".nodes", &errs)
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func collectNodeIDs(nodes []*Node, ids map[string]bool) {
	for _, n := range nodes {
		ids[n.ID] = true
		collectNodeIDs(n.Children, ids)
	}
}

func validateNodes(nodes []*Node, allIDs map[string]bool, prefix string, errs *ValidationErrors) {
	seen := map[string]bool{}
	validateNodesRec(nodes, allIDs, seen, prefix, errs)
}

func validateNodesRec(nodes []*Node, allIDs, seen map[string]bool, prefix string, errs *ValidationErrors) {
	for i, n := range nodes {
		p := fmt.Sprintf("%s[%d]", prefix, i)

		if n.ID == "" {
			*errs = append(*errs, ValidationError{p + ".id", "id は必須です"})
		} else if seen[n.ID] {
			*errs = append(*errs, ValidationError{p + ".id", fmt.Sprintf("id %q が重複しています", n.ID)})
		} else {
			seen[n.ID] = true
		}

		if n.Title == "" {
			*errs = append(*errs, ValidationError{p + ".title", "title は必須です"})
		}

		for _, parentID := range n.Parents {
			if !allIDs[parentID] {
				*errs = append(*errs, ValidationError{
					p + ".parents",
					fmt.Sprintf("親ノード %q が存在しません", parentID),
				})
			}
		}

		switch n.Type {
		case NodeTypeRequired, NodeTypeOptional, NodeTypeAlternative:
		default:
			*errs = append(*errs, ValidationError{
				p + ".type",
				fmt.Sprintf("不正な type %q (required/optional/alternative のいずれかを指定)", n.Type),
			})
		}

		validateNodesRec(n.Children, allIDs, seen, p+".children", errs)
	}
}
