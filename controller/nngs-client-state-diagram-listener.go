package controller

// `github.com/reiver/go-telnet` ライブラリーの動作をリスニングします
type nngsClientStateDiagramListener struct {
}

func (lis *nngsClientStateDiagramListener) matchStart() {
	print("<GE2NNGS> 対局成立だぜ☆")
}
func (lis *nngsClientStateDiagramListener) matchEnd() {
	print("<GE2NNGS> 対局終了だぜ☆")
}
func (lis *nngsClientStateDiagramListener) scoring() {
	print("<GE2NNGS> 得点計算だぜ☆")
}

func (lis *nngsClientStateDiagramListener) myTurn(dia *NngsClientStateDiagram) {
	print("****** I am thinking now   ******\n")
}
func (lis *nngsClientStateDiagramListener) opponentTurn(dia *NngsClientStateDiagram) {
	print("****** wating for his move ******\n")
	// ここでは　相手の着手は　分からないぜ（＾～＾）
	// u.G.Chat.Debug("<GE2NNGS> MyMove=[%s] OpponentMove=[%s]\n", dia.MyMove, dia.OpponentMove)
}
