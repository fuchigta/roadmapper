package command

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/fuchigta/roadmapper/internal/repo"
)

const githubPagesWorkflow = `name: Deploy to GitHub Pages
on:
  push:
    branches: [__BRANCH__]
permissions:
  contents: read
  pages: write
  id-token: write
concurrency:
  group: pages
  cancel-in-progress: true
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install roadmapper
        run: |
          curl -sSfL https://github.com/fuchigta/roadmapper/releases/latest/download/roadmapper-linux-amd64.tar.gz \
            | tar -xz -C /usr/local/bin roadmapper
          chmod +x /usr/local/bin/roadmapper
      - name: Build
        run: roadmapper build --base "/${{ github.event.repository.name }}/"
      - uses: actions/upload-pages-artifact@v3
        with:
          path: dist
  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deploy.outputs.page_url }}
    steps:
      - id: deploy
        uses: actions/deploy-pages@v4
`

const gitlabPagesCI = `pages:
  image: alpine:latest
  before_script:
    - apk add --no-cache curl
    - |
      curl -sSfL https://github.com/fuchigta/roadmapper/releases/latest/download/roadmapper-linux-amd64.tar.gz \
        | tar -xz -C /usr/local/bin roadmapper
      chmod +x /usr/local/bin/roadmapper
  script:
    - roadmapper build --base "/$CI_PROJECT_NAME/"
    - mv dist public
  artifacts:
    paths:
      - public
  rules:
    - if: $CI_COMMIT_BRANCH == "__BRANCH__"
`

// defaultDeployBranch はブランチ名を判定できないときのフォールバック。
const defaultDeployBranch = "main"

// branchNameRe は CI 定義 (YAML) へ安全に埋め込めるブランチ名。
var branchNameRe = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)

func NewDeployCmd() *cobra.Command {
	var target, branch string

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "CI/CD 用のワークフローファイルを生成する",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(target, branch)
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "デプロイ先 (github / gitlab)")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "デプロイ対象ブランチ (省略時は現在のブランチ、取得できなければ main)")
	_ = cmd.MarkFlagRequired("target")

	return cmd
}

// resolveBranch は CI の対象ブランチを決める。flag 指定を優先し、
// 未指定なら dir のリポジトリの .git/HEAD から現在のブランチを読む。
// 取得できなければ defaultDeployBranch にフォールバックし、その理由を note で返す。
func resolveBranch(flag, dir string) (branch, note string, err error) {
	if flag != "" {
		if !branchNameRe.MatchString(flag) {
			return "", "", fmt.Errorf("ブランチ名 %q に使用できない文字が含まれています", flag)
		}
		return flag, "", nil
	}
	root, ok, err := repo.FindRoot(dir)
	if err != nil {
		return "", "", fmt.Errorf("リポジトリルートの検出に失敗: %w", err)
	}
	if !ok {
		return defaultDeployBranch, "Git リポジトリが見つからないため", nil
	}
	b, ok, err := repo.CurrentBranch(root)
	switch {
	case err != nil:
		return defaultDeployBranch, fmt.Sprintf("ブランチ名を取得できないため (%v)", err), nil
	case !ok:
		return defaultDeployBranch, "detached HEAD のため", nil
	case !branchNameRe.MatchString(b):
		return defaultDeployBranch, fmt.Sprintf("ブランチ名 %q を CI 定義に埋め込めないため", b), nil
	}
	return b, "", nil
}

func runDeploy(target, branchFlag string) error {
	var path, tmpl, label string
	switch strings.ToLower(target) {
	case "github":
		path, tmpl, label = ".github/workflows/pages.yml", githubPagesWorkflow, "GitHub Actions"
	case "gitlab":
		path, tmpl, label = ".gitlab-ci.yml", gitlabPagesCI, "GitLab CI"
	default:
		return fmt.Errorf("不明なターゲット %q (github または gitlab を指定してください)", target)
	}

	branch, note, err := resolveBranch(branchFlag, ".")
	if err != nil {
		return err
	}
	if note != "" {
		fmt.Printf("%s %s にフォールバックします (--branch で指定可能)\n", note, branch)
	}
	fmt.Printf("対象ブランチ: %s\n", branch)

	return writeWorkflow(path, strings.ReplaceAll(tmpl, "__BRANCH__", branch), label)
}

func writeWorkflow(path, content, label string) error {
	// 既存ファイルの確認
	if _, err := os.Stat(path); err == nil {
		// diff を表示して確認
		existing, _ := os.ReadFile(path)
		if string(existing) == content {
			fmt.Printf("✓ %s は既に最新です: %s\n", label, path)
			return nil
		}
		fmt.Printf("既存ファイルが見つかりました: %s\n", path)
		showDiff(path, content)
		fmt.Print("上書きしますか? [y/N]: ")
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			fmt.Println("キャンセルしました")
			return nil
		}
	}

	// ディレクトリ作成 (サブディレクトリがある場合のみ)
	if idx := strings.LastIndex(path, "/"); idx > 0 {
		dir := path[:idx]
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("ディレクトリ作成失敗: %w", err)
		}
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("ファイル書き込み失敗: %w", err)
	}

	fmt.Printf("✓ %s ワークフローを生成しました: %s\n", label, path)
	fmt.Printf("\n次のステップ:\n")
	fmt.Printf("  git add %s\n", path)
	target := strings.ToLower(strings.Fields(label)[0])
	fmt.Printf("  git commit -m 'ci: add %s pages deployment'\n", target)
	fmt.Printf("  git push\n")
	return nil
}

// showDiff は unified diff を表示する (diff コマンドがあれば使用)。
func showDiff(existingPath, newContent string) {
	tmpFile, err := os.CreateTemp("", "roadmapper-*.yml")
	if err != nil {
		return
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(newContent)
	tmpFile.Close()

	out, err := exec.Command("diff", "-u", existingPath, tmpFile.Name()).Output()
	if err != nil || len(out) == 0 {
		return
	}
	fmt.Println("--- 差分 ---")
	fmt.Println(string(out))
	fmt.Println("-----------")
}
