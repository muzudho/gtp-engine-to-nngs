package controller

import (
	"fmt"
	"strings"

	"github.com/muzudho/gtp-engine-to-nngs/entities/phase"
)

// `github.com/reiver/go-telnet` ライブラリーの動作をリスニングします
type nngsClientStateDiagramListener struct {
}

func (lis *nngsClientStateDiagramListener) matchStart() {
	print("[情報] 対局成立だぜ☆")
}
func (lis *nngsClientStateDiagramListener) matchEnd() {
	print("[情報] 対局終了だぜ☆")
}
func (lis *nngsClientStateDiagramListener) scoring() {
	print("[情報] 得点計算だぜ☆")
}

func (lis *nngsClientStateDiagramListener) myTurn(dia *NngsClientStateDiagram) {
	print("****** I am thinking now   ******")
	message := fmt.Sprintf("genmove %s\n", strings.ToLower(phase.ToString(dia.MyColor)))
	fmt.Printf("[情報] エンジンにメッセージ送ったろ☆（＾～＾）[%s]", message)
	(*dia.EngineStdin).Write([]byte(message))

	var buffer [1]byte // これが満たされるまで待つ。1バイト。
	var lineBuffer [1024]byte
	index := 0
	p := buffer[:]

	for {
		n, err := (*dia.EngineStdout).Read(p) // ブロッキングしない？

		if nil != err {
			if fmt.Sprintf("%s", err) != "EOF" {
				fmt.Printf("[情報] エラーだぜ☆（＾～＾）[%s]\n", err)
				return
			}
			// 送られてくる文字がなければ、ここをずっと通る？
			// fmt.Printf("[情報] EOFだぜ☆（＾～＾）\n")
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

				index = 0
				// 終わりかどうか分からん。

				if lineString == "" {
					// 空行
					fmt.Printf("<情報> 空行。\n")
				} else {
					fmt.Printf("<情報> 受信行[%s]\n", lineString)
				}
			} else {
				lineBuffer[index] = bytes[0]
				index++
			}
		}
	}

}
func (lis *nngsClientStateDiagramListener) opponentTurn(dia *NngsClientStateDiagram) {
	print("****** wating for his move ******")
}
