package controller

import (
	"fmt"
	"strings"

	"github.com/muzudho/gtp-engine-to-nngs/entities/phase"
	u "github.com/muzudho/gtp-engine-to-nngs/usecases"
	"github.com/reiver/go-oi"
)

// `github.com/reiver/go-telnet` ライブラリーの動作をリスニングします
type nngsClientStateDiagramListener struct {
}

func (lis *nngsClientStateDiagramListener) matchStart() {
	print("<情報> 対局成立だぜ☆")
}
func (lis *nngsClientStateDiagramListener) matchEnd() {
	print("<情報> 対局終了だぜ☆")
}
func (lis *nngsClientStateDiagramListener) scoring() {
	print("<情報> 得点計算だぜ☆")
}

func (lis *nngsClientStateDiagramListener) myTurn(dia *NngsClientStateDiagram) {
	print("****** I am thinking now   ******")
	message := fmt.Sprintf("genmove %s\n", strings.ToLower(phase.ToString(dia.MyColor)))

	fmt.Printf("<情報> エンジンへ送信[%s]\n", message)
	(*dia.EngineStdin).Write([]byte(message))

	var buffer [1]byte // これが満たされるまで待つ。1バイト。

	// めっちゃ 難しいが、
	// [0] Current [1] Previous
	// [0] Previous [1] Current
	// を交互にやっている。
	var lineBuffer [2][1024]byte
	indexX := []uint{0, 0}
	indexY := 0
	p := buffer[:]

	for {
		n, err := (*dia.EngineStdout).Read(p) // ブロッキングしない？

		if nil != err {
			if fmt.Sprintf("%s", err) != "EOF" {
				fmt.Printf("<情報> エラーだぜ☆（＾～＾）[%s]\n", err)
				return
			}
			// 送られてくる文字がなければ、ここをずっと通る？
			// fmt.Printf("<情報> EOFだぜ☆（＾～＾）\n")
			indexY = (indexY + 1) % 2
			indexX[indexY] = 0
			continue
		}

		if 0 < n {
			bytes := p[:n]

			// 思考エンジンから１文字送られてくるたび、表示。
			// print(string(bytes))

			if bytes[0] == '\n' {
				// 思考エンジンから送られてきた１文字が、１行分 溜まるごとに表示。
				lineString := string(lineBuffer[indexY][:indexX[indexY]])

				if lineString == "" {
					// 空行

					if dia.MyColor == dia.CurrentPhase {
						// サーバーに着手を送信します。１行前の文字列を使います
						// Example: `= A1`
						// Example: `= Pass`
						matches71 := dia.regexBestmove.FindSubmatch(lineBuffer[(indexY+1)%2][:indexX[(indexY+1)%2]])
						if 1 < len(matches71) {
							u.G.Chat.Debug("<情報> サーバーへ送信[%s\n]\n", matches71[1])
							oi.LongWrite(dia.writerToServer, []byte(matches71[1]))
							oi.LongWrite(dia.writerToServer, []byte("\n"))

							// myTurn のループ終わり（＾～＾）！
							return
						}
						u.G.Chat.Debug("<情報> 空行(手番)。line=[%s] pre-line=[%s] len=[%d]\n", string(lineBuffer[indexY][:indexX[(indexY+1)%2]]), string(lineBuffer[(indexY+1)%2][:indexX[(indexY+1)%2]]), len(matches71))

					} else {
						u.G.Chat.Debug("<情報> 空行(相手番)。\n")
					}
					dia.ChatDebugState()

				} else {
					u.G.Chat.Debug("<情報> 受信行[%s]\n", lineString)
				}

				indexY = (indexY + 1) % 2
				indexX[indexY] = 0
			} else {
				lineBuffer[indexY][indexX[indexY]] = bytes[0]
				indexX[indexY]++
			}
		}
	}

}
func (lis *nngsClientStateDiagramListener) opponentTurn(dia *NngsClientStateDiagram) {
	print("****** wating for his move ******\n")
	u.G.Chat.Debug("<情報> MyMove=[%s] OpponentMove=[%s]\n", dia.MyMove, dia.OpponentMove)

	if dia.OpponentMove != "" {
		message := strings.ToLower(fmt.Sprintf("play %s %s", phase.FlipColorString(phase.ToString(dia.MyColor)), dia.OpponentMove))
		fmt.Printf("<情報> エンジンへ送信[%s]\n", message)
		(*dia.EngineStdin).Write([]byte(message))
	}
}
