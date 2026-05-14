package content

import (
	"path"
	"strings"
)

// MatchGlob は doublestar 風の glob パターンマッチを行う。
//   - `*`  は単一セグメント内の任意文字列にマッチ (path.Match と同じ)
//   - `**` は 0 個以上のセグメントにマッチ
//
// pattern と name はいずれも `/` 区切り。
func MatchGlob(pattern, name string) bool {
	pp := strings.Split(pattern, "/")
	np := strings.Split(name, "/")
	return matchSegments(pp, np)
}

// MatchAny は patterns のいずれかが name にマッチすれば true を返す。
func MatchAny(patterns []string, name string) bool {
	for _, p := range patterns {
		if MatchGlob(p, name) {
			return true
		}
	}
	return false
}

func matchSegments(pp, np []string) bool {
	for len(pp) > 0 {
		p := pp[0]
		if p == "**" {
			// 末尾の "**" 連続は 1 個分にまとめる
			for len(pp) > 1 && pp[1] == "**" {
				pp = pp[1:]
			}
			rest := pp[1:]
			for i := 0; i <= len(np); i++ {
				if matchSegments(rest, np[i:]) {
					return true
				}
			}
			return false
		}
		if len(np) == 0 {
			return false
		}
		ok, err := path.Match(p, np[0])
		if err != nil || !ok {
			return false
		}
		pp = pp[1:]
		np = np[1:]
	}
	return len(np) == 0
}
