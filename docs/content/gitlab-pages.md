---
links:
  - { title: "GitLab Pages ドキュメント", url: "https://docs.gitlab.com/ee/user/project/pages/" }
---

## deploy コマンドで GitLab CI ファイルを生成

```bash
roadmapper deploy --target gitlab -c my-roadmap/roadmap.yml
```

`.gitlab-ci.yml` が生成されます。

```yaml
image: golang:1.22

pages:
  stage: deploy
  script:
    - go install github.com/fuchigta/roadmapper/cmd/roadmapper@latest
    - roadmapper build -c roadmap.yml -o public
  artifacts:
    paths:
      - public
  only:
    - main
```

GitLab Pages はデフォルトで `public/` ディレクトリを公開するため、
出力先を `-o public` にしている点に注意してください。

## サブタスク

- [ ] `roadmapper deploy --target gitlab` で `.gitlab-ci.yml` を生成した
- [ ] GitLab リポジトリに push してパイプラインが通った
- [ ] Pages の URL でサイトを確認した
