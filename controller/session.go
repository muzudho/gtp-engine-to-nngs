package controller

import (
	"fmt"
	"strings"

	"github.com/muzudho/gtp-engine-to-nngs/entities/phase"
	kwu "github.com/muzudho/kifuwarabe-gtp/usecases"
	"github.com/reiver/go-oi"
)

func (dia *NngsClientStateDiagram) play(lis *nngsClientStateDiagramListener) {
	// Request
	message := strings.ToLower(fmt.Sprintf("play %s %s\n", phase.FlipColorString(phase.ToString(dia.MyColor)), dia.OpponentMove))

	// エンジンへ
	kwu.G.Chat.Notice("...GE2NNGS--> [%s]\n", message)
	(*dia.EngineStdin).Write([]byte(message))

	// Response
	dia.waitForPlayResponse(lis)
}

func (dia *NngsClientStateDiagram) genmove(lis *nngsClientStateDiagramListener) {
	// Request
	message := fmt.Sprintf("genmove %s\n", strings.ToLower(phase.ToString(dia.MyColor)))
	kwu.G.Chat.Notice("...GE2NNGS--> [%s]\n", message)
	(*dia.EngineStdin).Write([]byte(message))

	// Response
	dia.waitForGenmoveResponse(lis)
}

func (dia *NngsClientStateDiagram) done(lis *nngsClientStateDiagramListener) {
	message := "done\n"
	kwu.G.Chat.Trace("...GE2NNGS... 得点計算は飛ばすぜ☆（＾～＾）\n")

	// サーバーへ
	kwu.G.Chat.Notice("<--GE2NNGS... [%s]\n", message)
	oi.LongWrite(dia.writerToServer, []byte(message))
}

func (dia *NngsClientStateDiagram) quit(lis *nngsClientStateDiagramListener) {
	message := "quit\n"
	kwu.G.Chat.Trace("...GE2NNGS... アプリケーションも終了するぜ☆（＾～＾）\n")

	// エンジンへ。返無用。
	dia.quitEngine(lis)

	// サーバーへ
	kwu.G.Chat.Notice("<--GE2NNGS... [%s]\n", message)
	oi.LongWrite(dia.writerToServer, []byte(message))
}

func (dia *NngsClientStateDiagram) quitEngine(lis *nngsClientStateDiagramListener) {
	// エンジンへ。返無用。
	message := "quit\n"
	kwu.G.Chat.Notice("...GE2NNGS--> [%s]\n", message)
	(*dia.EngineStdin).Write([]byte(message))
}
