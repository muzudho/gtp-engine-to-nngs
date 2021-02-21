package controller

import (
	"fmt"
	"strings"

	"github.com/muzudho/gtp-engine-to-nngs/entities/phase"
	kwu "github.com/muzudho/kifuwarabe-gtp/usecases"
)

func (dia *NngsClientStateDiagram) play(lis *nngsClientStateDiagramListener) {
	// Request
	message := strings.ToLower(fmt.Sprintf("play %s %s\n", phase.FlipColorString(phase.ToString(dia.MyColor)), dia.OpponentMove))
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
}

func (dia *NngsClientStateDiagram) quit(lis *nngsClientStateDiagramListener) {
}
