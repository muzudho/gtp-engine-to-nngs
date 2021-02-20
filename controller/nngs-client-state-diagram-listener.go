package controller

import kwu "github.com/muzudho/kifuwarabe-gtp/usecases"

// `github.com/reiver/go-telnet` ライブラリーの動作をリスニングします
type nngsClientStateDiagramListener struct {
}

func (lis *nngsClientStateDiagramListener) matchStart() {
	kwu.G.Chat.Trace("...GE2NNGS... 対局成立だぜ☆")
}
func (lis *nngsClientStateDiagramListener) matchEnd() {
	kwu.G.Chat.Trace("...GE2NNGS... 対局終了だぜ☆")
}
func (lis *nngsClientStateDiagramListener) scoring() {
	kwu.G.Chat.Trace("...GE2NNGS... 得点計算だぜ☆")
}

func (lis *nngsClientStateDiagramListener) myTurn(dia *NngsClientStateDiagram) {
	kwu.G.Chat.Trace("****** I am thinking now   ******\n")
}
func (lis *nngsClientStateDiagramListener) opponentTurn(dia *NngsClientStateDiagram) {
	kwu.G.Chat.Trace("****** wating for his move ******\n")
	// ここでは　相手の着手は　分からないぜ（＾～＾）
	// kwu.G.Chat.Trace("...GE2NNGS... MyMove=[%s] OpponentMove=[%s]\n", dia.MyMove, dia.OpponentMove)
}
