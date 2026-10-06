---
links:
  - { title: "GitLab Pages ドキュメント", url: "https://docs.gitlab.com/ee/user/project/pages/" }
---

## deploy コマンドで GitLab CI ファイルを生成

```bash
roadmapper deploy --target gitlab
```

`.gitlab-ci.yml` が生成されます。リポジトリのルートで実行してください (`-c` / `-o` はありません)。

```yaml
pages:
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
    - if: $CI_COMMIT_BRANCH == "main"
```

GitLab Pages はデフォルトで `public/` ディレクトリを公開するため、
ビルド後に `dist` を `public` へ移動している点に注意してください。
`--base` にはプロジェクト名が渡されます。公開 URL が異なる場合 (グループ Pages など) は書き換えてください。
既定ブランチが `main` でない場合は `rules:` の条件を書き換えます。

## サブタスク

- [ ] `roadmapper deploy --target gitlab` で `.gitlab-ci.yml` を生成した
- [ ] GitLab リポジトリに push してパイプラインが通った
- [ ] Pages の URL でサイトを確認した
