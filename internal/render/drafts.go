package render

// Drafts は下書きノードの扱いを表す。
//
// 公開ビルドでは IDs のノードを非活性表示にして本文を出力しない (locked)。
// Preview が true (dev / build --drafts) のときは通常ノードと同じく操作可能にし、「下書き」バッジだけを付ける。
type Drafts struct {
	// IDs は下書きノードの ID 集合 (nil 可)。
	IDs map[string]bool
	// Preview は下書きを通常ノードとして扱い、バッジのみ付けるプレビュー表示かどうか。
	Preview bool
}

// locked は id が公開ビルドで非活性表示 (本文非公開) になるノードかを返す。
func (d Drafts) locked(id string) bool { return d.IDs[id] && !d.Preview }

// marked は id がプレビュー表示で「下書き」バッジを付けるノードかを返す。
func (d Drafts) marked(id string) bool { return d.IDs[id] && d.Preview }
