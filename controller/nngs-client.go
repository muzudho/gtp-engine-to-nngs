package controller

import (
	"fmt"
	"io"
	"os"
	"regexp"

	e "github.com/muzudho/gtp-engine-to-nngs/entities"
	kwe "github.com/muzudho/kifuwarabe-gtp/entities"
	kwu "github.com/muzudho/kifuwarabe-gtp/usecases"
	"github.com/reiver/go-oi"
	"github.com/reiver/go-telnet"
)

// Spawn - クライアント接続
// * `engineStdin` - GTP Engine stdin
func Spawn(engineConf *kwe.EngineConf, entryConf *e.EntryConf, engineStdin *io.WriteCloser, engineStdout *io.ReadCloser) error {
	// NNGSクライアントの状態遷移図
	nngsClientStateDiagram := NngsClientStateDiagram{
		EngineStdin:  engineStdin,
		EngineStdout: engineStdout,
		engineConf:   *engineConf,
		entryConf:    *entryConf,
		// nngsClientStateDiagram: *new(NngsClientStateDiagram),
		index:                  0,
		regexCommand:           *regexp.MustCompile("^(\\d+) (.*)"),
		regexUseMatch:          *regexp.MustCompile("^Use <match"),
		regexUseMatchToRespond: *regexp.MustCompile("^Use <(.+?)> or <(.+?)> to respond."), // 頭の '9 ' は先に削ってあるから ここに含めない（＾～＾）
		regexMatchAccepted:     *regexp.MustCompile("^Match \\[.+?\\] with (\\S+?) in \\S+? accepted."),
		regexDecline1:          *regexp.MustCompile("declines your request for a match."),
		regexDecline2:          *regexp.MustCompile("You decline the match offer from"),
		regexOneSeven:          *regexp.MustCompile("1 7"),
		regexGame:              *regexp.MustCompile("Game (\\d+) ([a-zA-Z]): (\\S+) \\((\\S+) (\\S+) (\\S+)\\) vs (\\S+) \\((\\S+) (\\S+) (\\S+)\\)"),
		regexNngsMove:          *regexp.MustCompile("\\s*(\\d+)\\(([BWbw])\\): ([A-Z]\\d+|Pass)"),
		regexAcceptCommand:     *regexp.MustCompile("match \\S+ \\S+ (\\d+) "),
		regexEngineBestmove:    *regexp.MustCompile("= ([A-Z]\\d+|pass)")}
	return telnet.DialToAndCall(fmt.Sprintf("%s:%d", entryConf.Server.Host, entryConf.Server.Port), nngsClientStateDiagram)
}

// CallTELNET - 決まった形のメソッド。サーバーに対して読み書きできます
func (dia NngsClientStateDiagram) CallTELNET(ctx telnet.Context, w telnet.Writer, r telnet.Reader) {

	kwu.G.Chat.Trace("...GE2NNGS... 受信開始☆")
	lis := nngsClientStateDiagramListener{}

	dia.writerToServer = w
	dia.readerFromServer = r

	//go dia.read(&lis)
	dia.read(&lis)

	// scanner - 標準入力を監視します。
	// scanner := bufio.NewScanner(os.Stdin)

	kwu.G.Log.FlushAllLogs()

	/*
		Loop:
			for scanner.Scan() {
				bytes := scanner.Bytes()
				// 書き込みます。最後に改行を付けます。
				kwu.G.Chat.Notice("<--★GE2NNGS... [%s\n]\n", bytes)
				oi.LongWrite(dia.writerToServer, bytes)
				oi.LongWrite(dia.writerToServer, []byte("\n"))

				// "quit" メッセージを送信したエンジンは、その直後にアプリケーション終了するので、接続も切れる。
				// すると Scan() は "quit" メッセージを受け取れない。
				if string(bytes) == "quit" {
					// エンジンは終了しました
					kwu.G.Chat.Trace("...GE2NNGS... エンジンをQuitさせるぜ☆（＾～＾）")
					break Loop
				}

				kwu.G.Log.FlushAllLogs()
			}
	*/

	kwu.G.Chat.Trace("...GE2NNGS... Telnetを終了するぜ☆（＾～＾）")
}

// サーバーから送られてくるメッセージを待ち構えるループです。
func (dia *NngsClientStateDiagram) read(lis *nngsClientStateDiagramListener) {
	var buffer [1]byte // これが満たされるまで待つ。1バイト。
	p := buffer[:]

	for {
		n, err := dia.readerFromServer.Read(p) // 送られてくる文字がなければ、ここでブロックされます。

		if n > 0 {
			bytes := p[:n]
			dia.lineBuffer[dia.index] = bytes[0]
			dia.index++

			// if dia.newlineReadableState < 2 {
			// 	// サーバーから１文字送られてくるたび、表示。
			// 	// [受信] 割り込みで 改行がない行も届くので、改行が届くまで待つという処理ができません。
			// 	print(string(bytes))
			// }

			// 改行を受け取る前にパースしてしまおう☆（＾～＾）早とちりするかも知れないけど☆（＾～＾）
			if dia.parse(lis) {
				// このアプリを終了します
				kwu.G.Chat.Trace("...GE2NNGS... End read loop A\n")
				// 標準入力のスキャンに、 "quit" を送り付けます。
				fmt.Fprintf(os.Stdin, "quit\n")
				kwu.G.Log.Trace("...GE2NNGS... quit\n")
				return
			}

			// `Login:` のように 改行が送られてこないケースはあるが、
			// 対局が始まってしまえば、改行は送られてくると考えろだぜ☆（＾～＾）
			if bytes[0] == '\n' {
				dia.index = 0

				if dia.newlineReadableState == 1 {
					kwu.G.Chat.Trace("[行単位入力へ切替(^q^)]")
					dia.newlineReadableState = 2
					break // for文を抜ける
				}
			}
		}

		if nil != err {
			return // 相手が切断したなどの理由でエラーになるので、終了します。
		}
	}

	// サーバーから改行が送られてくるものと考えるぜ☆（＾～＾）
	// これで、１行ずつ読み込めるな☆（＾～＾）
	for {
		n, err := dia.readerFromServer.Read(p) // サーバーから送られてくる文字がなければ、ここでブロックされます。

		if nil != err {
			return // 相手が切断したなどの理由でエラーになるので、終了します。
		}

		if n > 0 {
			bytes := p[:n]

			if bytes[0] == '\r' {
				// Windows では、 \r\n と続いてくるものと想定します。
				// Linux なら \r はこないものと想定します。
				continue

			} else if bytes[0] == '\n' {
				// `Login:` のように 改行が送られてこないケースはあるが、
				// 対局が始まってしまえば、改行は送られてくると考えろだぜ☆（＾～＾）
				// 1行をパースします
				if dia.parse(lis) {
					// このアプリを終了します
					kwu.G.Chat.Trace("...GE2NNGS... End read loop B\n")
					// 標準入力のスキャンに、 "quit" を送り付けます。
					fmt.Fprintf(os.Stdin, "quit\n")
					kwu.G.Log.Trace("...GE2NNGS... quit\n")
					return
				}
				dia.index = 0

			} else {
				dia.lineBuffer[dia.index] = bytes[0]
				dia.index++
			}
		}
	}
}

// 簡易表示モードに切り替えます。
// Original code: NngsClient.rb/NNGSClient/`def login`
func setClientMode(writerToServer telnet.Writer) {
	message := "set client true\n"
	kwu.G.Chat.Notice("<--GE2NNGS... [%s]\n", message)
	oi.LongWrite(writerToServer, []byte(message))
}
