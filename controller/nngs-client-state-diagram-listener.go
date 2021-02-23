package controller

import g "github.com/muzudho/gtp-engine-to-nngs/global"

// `github.com/reiver/go-telnet` ライブラリーの動作をリスニングします
type nngsClientStateDiagramListener struct {
}

func (lis *nngsClientStateDiagramListener) matchStart() {
	g.G.Chat.Trace("...GE2NNGS... 対局成立だぜ☆")
}
func (lis *nngsClientStateDiagramListener) matchEnd() {
	g.G.Chat.Trace("...GE2NNGS... 対局終了だぜ☆")
}
func (lis *nngsClientStateDiagramListener) scoring() {
	g.G.Chat.Trace("...GE2NNGS... 得点計算だぜ☆")
}

func (lis *nngsClientStateDiagramListener) myTurn(dia *NngsClientStateDiagram) {
	g.G.Chat.Trace("****** I am thinking now   ******\n")
}
func (lis *nngsClientStateDiagramListener) opponentTurn(dia *NngsClientStateDiagram) {
	g.G.Chat.Trace("****** wating for his move ******\n")
	// ここでは　相手の着手は　分からないぜ（＾～＾）
	// g.G.Chat.Trace("...GE2NNGS... MyMove=[%s] OpponentMove=[%s]\n", dia.MyMove, dia.OpponentMove)
}
