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
	print("<GE2NNGS> 対局成立だぜ☆")
}
func (lis *nngsClientStateDiagramListener) matchEnd() {
	print("<GE2NNGS> 対局終了だぜ☆")
}
func (lis *nngsClientStateDiagramListener) scoring() {
	print("<GE2NNGS> 得点計算だぜ☆")
}

// play コマンドの応答を待ちます
func (lis *nngsClientStateDiagramListener) waitForPlayResponse(dia *NngsClientStateDiagram) {
	print("****** Board updating ... ******\n")

	var buffer [1]byte // これが満たされるまで待つ。1バイト。

	// ただのライン・バッファー
	var lineBuffer [1024]byte
	index := 0
	p := buffer[:]

	for {
		// エンジンから送られてくる文字列
		n, err := (*dia.EngineStdout).Read(p) // ブロッキングしない？

		if nil != err {
			if fmt.Sprintf("%s", err) != "EOF" {
				fmt.Printf("<GE2NNGS> エラーだぜ☆（＾～＾）[%s]\n", err)
				return
			}
			// 送られてくる文字がなければ、ここをずっと通る？
			fmt.Printf("<GE2NNGS> EOFだぜ☆（＾～＾）\n")
			index = 0
			continue
		}

		if 0 < n {
			bytes := p[:n]

			// 思考エンジンから１文字送られてくるたび、表示。
			print(string(bytes))

			if bytes[0] == '\n' {
				// 思考エンジンから送られてきた１文字が、１行分 溜まるごとに表示。
				lineString := string(lineBuffer[:index])

				if lineString == "" {
					// 空行

					if dia.MyColor == dia.CurrentPhase {
						u.G.Chat.Debug("<GE2NNGS> 空行(手番)。\n")
					} else {
						u.G.Chat.Debug("<GE2NNGS> 空行(相手番)。\n")
					}

					dia.ChatDebugState()

				} else {
					u.G.Chat.Debug("<GE2NNGS> 受信行[%s]\n", lineString)

					if lineString == "= " {
						// `play` の OK かも。
						u.G.Chat.Debug("<GE2NNGS> playのOKかも☆（＾～＾）\n")
						return
					}
				}

				index = 0
			} else {
				lineBuffer[index] = bytes[0]
				index++
			}
		}
	}
}

func (lis *nngsClientStateDiagramListener) myTurn(dia *NngsClientStateDiagram) {
	print("****** I am thinking now   ******\n")

	message := fmt.Sprintf("genmove %s\n", strings.ToLower(phase.ToString(dia.MyColor)))
	fmt.Printf("<GE2NNGS> エンジンへ送信[%s]\n", message)
	(*dia.EngineStdin).Write([]byte(message))

	// if dia.CurrentPhase == phase.Black {
	// 	dia.turnState = 20
	// } else {
	// 	dia.turnState = 40
	// }

	var buffer [1]byte // これが満たされるまで待つ。1バイト。

	// 着手
	bestmove := ""

	// ただのライン・バッファー
	var lineBuffer [1024]byte
	index := 0
	p := buffer[:]

	for {
		// エンジンから送られてくる文字列
		n, err := (*dia.EngineStdout).Read(p) // ブロッキングしない？

		if nil != err {
			if fmt.Sprintf("%s", err) != "EOF" {
				fmt.Printf("<GE2NNGS> エラーだぜ☆（＾～＾）[%s]\n", err)
				return
			}
			// 送られてくる文字がなければ、ここをずっと通る？
			// fmt.Printf("<GE2NNGS> EOFだぜ☆（＾～＾）\n")
			index = 0
			continue
		}

		if 0 < n {
			bytes := p[:n]

			// 思考エンジンから１文字送られてくるたび、表示。
			// print(string(bytes))

			if bytes[0] == '\n' {
				// 思考エンジンから送られてきた１文字が、１行分 溜まるごとに表示。
				lineString := string(lineBuffer[:index])

				if lineString == "" {
					// 空行

					if dia.MyColor == dia.CurrentPhase {
						// サーバーに着手を送信します。１行前の文字列を使います
						// Example: `= A1`
						// Example: `= Pass`
						if bestmove != "" {
							u.G.Chat.Debug("<GE2NNGS> サーバーへ送信[%s\n]\n", bestmove)
							oi.LongWrite(dia.writerToServer, []byte(bestmove))
							oi.LongWrite(dia.writerToServer, []byte("\n"))
							// myTurn のループ終わり（＾～＾）！
							return
						}
						u.G.Chat.Debug("<GE2NNGS> 空行(手番)。bestmove未決定=[%s]\n", bestmove)
					} else {
						u.G.Chat.Debug("<GE2NNGS> 空行(相手番)。\n")
					}

					dia.ChatDebugState()

				} else {
					u.G.Chat.Debug("<GE2NNGS> 受信行[%s]\n", lineString)

					if lineString == "= " {
						// `play` の OK かも。
						u.G.Chat.Debug("<GE2NNGS> playのOKかも☆（＾～＾） lineString=[%s]\n", lineString)
					} else {
						// サーバーに着手を送信します。１行前の文字列を使います
						// Example: `= A1`
						// Example: `= Pass`
						matches71 := dia.regexBestmove.FindSubmatch(lineBuffer[:index])
						if 1 < len(matches71) {
							bestmove = string(matches71[1])
							u.G.Chat.Debug("<GE2NNGS> bestmove=[%s]\n", bestmove)
						} else {
							u.G.Chat.Debug("<GE2NNGS> 空行(手番)。line=[%s] bestmove=[%s] len=[%d]\n", string(lineBuffer[:index]), bestmove, len(matches71))
						}
					}
				}

				index = 0
			} else {
				lineBuffer[index] = bytes[0]
				index++
			}
		}
	}
}
func (lis *nngsClientStateDiagramListener) opponentTurn(dia *NngsClientStateDiagram) {
	print("****** wating for his move ******\n")
	// ここでは　相手の着手は　分からないぜ（＾～＾）
	u.G.Chat.Debug("<GE2NNGS> MyMove=[%s] OpponentMove=[%s]\n", dia.MyMove, dia.OpponentMove)
}
