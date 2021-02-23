package controller

import (
	"fmt"
	"io"
	"regexp"

	e "github.com/muzudho/gtp-engine-to-nngs/entities"
	g "github.com/muzudho/gtp-engine-to-nngs/global"
	kwe "github.com/muzudho/kifuwarabe-gtp/entities"
	"github.com/reiver/go-oi"
	"github.com/reiver/go-telnet"
)

// SpawnServerConnection - サーバーとのTelnet接続
// * `engineStdin` - GTP Engine stdin
func SpawnServerConnection(engineConf *kwe.EngineConf, connectorConf *e.ConnectorConf, engineStdin *io.WriteCloser, engineStdout *io.ReadCloser) error {
	// NNGSクライアントの状態遷移図
	nngsClientStateDiagram := NngsClientStateDiagram{
		EngineStdin:   engineStdin,
		EngineStdout:  engineStdout,
		engineConf:    *engineConf,
		connectorConf: *connectorConf,
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

	g.G.Chat.Trace("...GE2NNGS... telnet.DialToAndCall\n")
	return telnet.DialToAndCall(fmt.Sprintf("%s:%d", connectorConf.Server.Host, connectorConf.Server.Port), nngsClientStateDiagram)
}

// CallTELNET - 決まった形のメソッド。サーバーに対して読み書きできます
func (dia NngsClientStateDiagram) CallTELNET(ctx telnet.Context, w telnet.Writer, r telnet.Reader) {

	g.G.Chat.Trace("...GE2NNGS... Start Telnet\n")
	lis := nngsClientStateDiagramListener{}

	dia.writerToServer = w
	dia.readerFromServer = r

	dia.read(&lis)

	g.G.Log.FlushAllLogs()

	// プロセスが終了しても、子プロセスは自動的には終了しませんので、
	// エンジンの終了を試みます
	g.G.Chat.Trace("...GE2NNGS... Try quit engine\n")
	dia.quitEngine(&lis)

	g.G.Chat.Trace("...GE2NNGS... End Telnet\n")
}

// サーバーから送られてくるメッセージを待ち構えるループです。
func (dia *NngsClientStateDiagram) read(lis *nngsClientStateDiagramListener) {
	var buffer [1]byte // これが満たされるまで待つ。1バイト。
	p := buffer[:]

	for {
		g.G.Log.FlushAllLogs()

		n, err := dia.readerFromServer.Read(p) // 送られてくる文字がなければ、ここでブロックされます。

		if n > 0 {
			bytes := p[:n]
			dia.lineBuffer[dia.index] = bytes[0]
			dia.index++

			// 改行を受け取る前にパースしてしまおう☆（＾～＾）早とちりするかも知れないけど☆（＾～＾）
			if dia.parse(lis) {
				// このアプリを終了します
				g.G.Chat.Trace("...GE2NNGS... End read loop A\n")
				return
			}

			// `Login:` のように 改行が送られてこないケースはあるが、
			// 対局が始まってしまえば、改行は送られてくると考えろだぜ☆（＾～＾）
			if bytes[0] == '\n' {
				dia.index = 0

				if dia.newlineReadableState == 1 {
					g.G.Chat.Trace("[行単位入力へ切替(^q^)]")
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
		g.G.Log.FlushAllLogs()

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
					g.G.Chat.Trace("...GE2NNGS... End read loop B\n")
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
	g.G.Chat.Notice("<--GE2NNGS... [%s]\n", message)
	oi.LongWrite(writerToServer, []byte(message))
}
