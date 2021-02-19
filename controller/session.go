package controller

import (
	"fmt"
	"strings"

	"github.com/muzudho/gtp-engine-to-nngs/entities/phase"
)

func (dia *NngsClientStateDiagram) play(lis *nngsClientStateDiagramListener) {
	// Request
	message := strings.ToLower(fmt.Sprintf("play %s %s\n", phase.FlipColorString(phase.ToString(dia.MyColor)), dia.OpponentMove))
	fmt.Printf("<GE2NNGS> エンジンへ送信[%s]\n", message)
	(*dia.EngineStdin).Write([]byte(message))

	// Response
	dia.waitForPlayResponse(lis)
}

func (dia *NngsClientStateDiagram) genmove(lis *nngsClientStateDiagramListener) {
	// Request
	message := fmt.Sprintf("genmove %s\n", strings.ToLower(phase.ToString(dia.MyColor)))
	fmt.Printf("<GE2NNGS> エンジンへ送信[%s]\n", message)
	(*dia.EngineStdin).Write([]byte(message))

	// Response
	dia.waitForGenmoveResponse(lis)
}
