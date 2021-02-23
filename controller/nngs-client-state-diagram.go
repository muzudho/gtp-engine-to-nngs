package controller

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/muzudho/gtp-engine-to-nngs/controller/clistat"
	e "github.com/muzudho/gtp-engine-to-nngs/entities"
	"github.com/muzudho/gtp-engine-to-nngs/entities/phase"
	kwe "github.com/muzudho/kifuwarabe-gtp/entities"
	kwphase "github.com/muzudho/kifuwarabe-gtp/entities/phase"
	kwu "github.com/muzudho/kifuwarabe-gtp/usecases"
	"github.com/reiver/go-oi"
	"github.com/reiver/go-telnet"
)

// NngsClientStateDiagram - NNGSクライアントの状態遷移図
type NngsClientStateDiagram struct {
	// EngineStdin - GTP Engine stdin
	EngineStdin *io.WriteCloser
	// EngineStdin - GTP Engine stdin
	EngineStdout *io.ReadCloser

	connectorConf e.ConnectorConf
	engineConf    kwe.EngineConf

	// 状態遷移
	state clistat.Clistat
	// 状態遷移の中の小さな区画
	promptState int
	// 末尾に改行が付いていると想定していいフェーズ。逆に、そうでない例は `Login:` とか
	newlineReadableState uint
	// 0: 白番は対局を承諾へ
	// 10: 黒番は着手へ
	// 20: 白番は盤更新へ
	// 30: 白番は着手へ
	// 40: 黒番は盤更新へ
	// turnState uint

	// NNGSへ書込み
	writerToServer telnet.Writer
	// NNGSから読込み
	readerFromServer telnet.Reader

	// NNGSクライアントの状態遷移図
	//nngsClientStateDiagram NngsClientStateDiagram

	// １行で 1024 byte は飛んでこないことをサーバーと決めておけだぜ☆（＾～＾）
	lineBuffer [1024]byte
	index      uint

	// 正規表現
	regexCommand           regexp.Regexp
	regexUseMatch          regexp.Regexp
	regexUseMatchToRespond regexp.Regexp
	regexMatchAccepted     regexp.Regexp
	regexDecline1          regexp.Regexp
	regexDecline2          regexp.Regexp
	regexOneSeven          regexp.Regexp
	regexGame              regexp.Regexp

	// Example: `15 Game 2 I: kifuwarabe (0 2289 -1) vs kifuwarabi (0 2298 -1)`.
	regexNngsMove      regexp.Regexp
	regexAcceptCommand regexp.Regexp
	// Example: `= A1`
	// Example: `= pass`
	regexEngineBestmove regexp.Regexp

	// MyColor - 自分の手番の色
	MyColor kwphase.Phase
	// Phase - これから指す方。局面の手番とは逆になる
	CurrentPhase kwphase.Phase

	// BoardSize - 何路盤。マッチを受け取ったときに確定
	BoardSize uint
	// MyMove - 自分の指し手
	MyMove string
	// OpponentMove - 相手の指し手
	OpponentMove string
	// CommandOfMatchAccept - 申し込まれた対局を受け入れるコマンド。人間プレイヤーの入力補助用
	CommandOfMatchAccept string
	// CommandOfMatchDecline - 申し込まれた対局をお断りするコマンド。人間プレイヤーの入力補助用
	CommandOfMatchDecline string
	// GameID - 対局番号☆（＾～＾） 1 から始まる数☆（＾～＾）
	GameID uint
	// GameType - なんだか分からないが少なくとも "I" とか入ってるぜ☆（＾～＾）
	GameType string
	// GameWName - 白手番の対局者アカウント名
	GameWName string
	// GameWField2 - 白手番の２番目のフィールド（用途不明）
	GameWField2 string
	// GameWAvailableSeconds - 白手番の残り時間（秒）
	GameWAvailableSeconds int
	// GameWField4 - 白手番の４番目のフィールド（用途不明）
	GameWField4 string
	// GameBName - 黒手番の対局者アカウント名
	GameBName string
	// GameBField2 - 黒手番の２番目のフィールド（用途不明）
	GameBField2 string
	// GameBAvailableSeconds - 白手番の残り時間（秒）
	GameBAvailableSeconds int
	// GameBField4 - 黒手番の４番目のフィールド（用途不明）
	GameBField4 string
}

// // ChatTraceState - 状態をデバッグ表示
// func (dia *NngsClientStateDiagram) ChatTraceState() {
// 	kwu.G.Chat.Trace("...GE2NNGS... state=%d promptState=%d newlineReadableState=%d\n", dia.state, dia.promptState, dia.newlineReadableState)
// }

// アプリケーションを終了するなら 真 を返します
func (dia *NngsClientStateDiagram) promptDiagram(lis *nngsClientStateDiagramListener, subCode int) bool {
	switch subCode {
	// Info
	case 5:
		if dia.promptState == 0 {
			// このフラグを立てるのは初回だけ。
			dia.newlineReadableState = 1
		}

		if dia.promptState == 7 {
			lis.matchEnd() // 対局終了

			// このアプリを終了します
			kwu.G.Chat.Notice("...GE2NNGS... Match end\n")
			return true
		}
		dia.promptState = 5
	// PlayingGo
	case 6:
		if dia.promptState == 5 {
			lis.matchStart() // 対局成立
			dia.turn(lis)
		}
		dia.promptState = 6
	// Scoring
	case 7:
		if dia.promptState == 6 {
			lis.scoring() // 得点計算

			// 本来は 死に石 を選んだりするフェーズだが、
			// コンピューター囲碁大会では 思考エンジンの自己申告だけ聞き取るので、
			// このフェーズは飛ばします。
			dia.done(lis)

			// 思考エンジン・アプリケーションも終了させます。
			dia.quit(lis)
		}
		dia.promptState = 7
	default:
		// "1 1" とか来ても無視しろだぜ☆（＾～＾）
	}

	return false
}

// サーバーから送られてくるメッセージを解析します
// アプリケーションを終了するなら 真 を返します
func (dia *NngsClientStateDiagram) parse(lis *nngsClientStateDiagramListener) bool {
	// 現在読み取り中の文字なので、早とちりするかも知れないぜ☆（＾～＾）
	line := string(dia.lineBuffer[:dia.index])

	if dia.newlineReadableState == 2 {
		kwu.G.Chat.Notice("-->GE2NNGS... [%s]\n", line)
	}

	switch dia.state {
	case clistat.None:
		// Original code: NngsClient.rb/NNGSClient/`def login`
		// Waitfor "Login: ".
		if line == "Login: " {
			// あなたの名前を入力してください。

			// 設定ファイルから自動で入力するぜ☆（＾ｑ＾）
			user := dia.engineConf.Profile.Name

			// 自動入力のときは、設定ミスなら強制終了しないと無限ループしてしまうぜ☆（＾～＾）
			if user == "" {
				panic(kwu.G.Chat.Fatal("...GE2NNGS... Need name (UserName)"))
			}

			kwu.G.Chat.Notice("<--GE2NNGS... [%s\n]\n", user)
			oi.LongWrite(dia.writerToServer, []byte(user))
			oi.LongWrite(dia.writerToServer, []byte("\n"))

			dia.state = clistat.EnteredMyName
		}
	// Original code: NngsClient.rb/NNGSClient/`def login`
	case clistat.EnteredMyName:
		if line == "1 1" {
			// パスワードを入れろだぜ☆（＾～＾）
			if dia.engineConf.Profile.Pass == "" {
				panic(kwu.G.Chat.Fatal("...GE2NNGS... Need password"))
			}

			kwu.G.Chat.Notice("<--GE2NNGS... [%s\n]\n", dia.engineConf.Profile.Pass)
			oi.LongWrite(dia.writerToServer, []byte(dia.engineConf.Profile.Pass))
			oi.LongWrite(dia.writerToServer, []byte("\n"))
			setClientMode(dia.writerToServer)
			dia.state = clistat.EnteredClientMode

		} else if line == "Password: " {
			// パスワードを入れろだぜ☆（＾～＾）
			if dia.engineConf.Profile.Pass == "" {
				panic(kwu.G.Chat.Fatal("...GE2NNGS... Need password"))
			}

			kwu.G.Chat.Notice("<--GE2NNGS... [%s\n]\n", dia.engineConf.Profile.Pass)
			oi.LongWrite(dia.writerToServer, []byte(dia.engineConf.Profile.Pass))
			oi.LongWrite(dia.writerToServer, []byte("\n"))
			dia.state = clistat.EnteredMyPasswordAndIAmWaitingToBePrompted

		} else if line == "#> " {
			setClientMode(dia.writerToServer)
			dia.state = clistat.EnteredClientMode
		}
		// 入力した名前が被っていれば、ここで無限ループしてるかも☆（＾～＾）

	// Original code: NngsClient.rb/NNGSClient/`def login`
	case clistat.EnteredMyPasswordAndIAmWaitingToBePrompted:
		if line == "#> " {
			setClientMode(dia.writerToServer)
			dia.state = clistat.EnteredClientMode
		}
	case clistat.EnteredClientMode:
		if dia.connectorConf.ApplyFromMe() {
			// 対局を申し込みます。
			_, configuredColorUpperCase, err := dia.connectorConf.MyColor()
			if err != nil {
				panic(kwu.G.Chat.Fatal(fmt.Sprintf("...GE2NNGS... %s", err)))
			}

			kwu.G.Log.Trace("...GE2NNGS... lis.MyColorを%sに変更☆（＾～＾）\n", configuredColorUpperCase)
			dia.MyColor = phase.ToNum(configuredColorUpperCase)

			message := fmt.Sprintf("match %s %s %d %d %d\n", dia.connectorConf.OpponentName(), configuredColorUpperCase, dia.connectorConf.BoardSize(), dia.connectorConf.AvailableTimeMinutes(), dia.connectorConf.CanadianTiming())
			kwu.G.Log.Trace("...GE2NNGS... 対局を申し込んだぜ☆（＾～＾）")
			kwu.G.Chat.Notice("<--GE2NNGS... [%s]\n", message)
			oi.LongWrite(dia.writerToServer, []byte(message))
		}
		dia.state = clistat.WaitingInInfo

	// '1 5' - Waiting
	case clistat.WaitingInInfo:
		// Example: 1 5
		matches := dia.regexCommand.FindSubmatch(dia.lineBuffer[:dia.index])

		//kwu.G.Chat.Trace("...GE2NNGS... m[%s]", matches)
		if 2 < len(matches) {
			commandCodeBytes := matches[1]
			commandCode := string(commandCodeBytes)
			promptStateBytes := matches[2]

			code, err := strconv.Atoi(commandCode)
			if err != nil {
				panic(kwu.G.Chat.Fatal(fmt.Sprintf("...GE2NNGS... %s", err)))
			}
			switch code {
			// Prompt
			case 1:
				promptState := string(promptStateBytes)
				promptStateNum, err := strconv.Atoi(promptState)
				if err == nil {
					if dia.promptDiagram(lis, promptStateNum) {
						// このアプリを終了します
						kwu.G.Chat.Trace("...GE2NNGS... End parse\n")
						return true
					}
				}
			// Info
			case 9:
				if dia.regexUseMatch.Match(promptStateBytes) {
					matches2 := dia.regexUseMatchToRespond.FindSubmatch(promptStateBytes)
					if 2 < len(matches2) {
						// 対局を申し込まれた方だけ、ここを通るぜ☆（＾～＾）
						kwu.G.Chat.Trace("...GE2NNGS... 対局を申し込まれたぜ☆（＾～＾）[%s] accept[%s],decline[%s]\n", string(promptStateBytes), matches2[1], matches2[2])

						// Example: `match kifuwarabi W 19 40 0`
						dia.CommandOfMatchAccept = string(matches2[1])
						// Example: `decline kifuwarabi`
						dia.CommandOfMatchDecline = string(matches2[2])

						// acceptコマンドを半角空白でスプリットした３番目が、自分の手番
						matchAcceptTokens := strings.Split(dia.CommandOfMatchAccept, " ")
						if len(matchAcceptTokens) < 6 {
							panic(kwu.G.Chat.Fatal("...GE2NNGS... Error matchAcceptTokens=[%s].", matchAcceptTokens))
						}

						opponentPlayerName := matchAcceptTokens[1]
						myColorString := matchAcceptTokens[2]
						myColorUppercase := strings.ToUpper(myColorString)
						// kwu.G.Chat.Trace("...GE2NNGS... MyColorを[%s]に変更☆（＾～＾）\n", myColorString)
						dia.MyColor = phase.ToNum(myColorString)

						boardSize, err := strconv.ParseUint(matchAcceptTokens[3], 10, 0)
						if err != nil {
							panic(kwu.G.Chat.Fatal(fmt.Sprintf("...GE2NNGS... %s", err)))
						}
						dia.BoardSize = uint(boardSize)
						// kwu.G.Chat.Trace("...GE2NNGS... ボードサイズは%d☆（＾～＾）", dia.BoardSize)

						configuredColor, _, err := dia.connectorConf.MyColor()
						if err != nil {
							panic(kwu.G.Chat.Fatal(fmt.Sprintf("...GE2NNGS... %s", err)))
						}

						if dia.MyColor != configuredColor {
							panic(kwu.G.Chat.Fatal("...GE2NNGS... Unexpected phase. MyColor=%d configuredColor=%d.", dia.MyColor, configuredColor))
						}

						// cmd_match
						message := fmt.Sprintf("match %s %s %d %d %d\n", opponentPlayerName, myColorUppercase, dia.connectorConf.BoardSize(), dia.connectorConf.AvailableTimeMinutes(), dia.connectorConf.CanadianTiming())
						kwu.G.Chat.Trace("...GE2NNGS... 対局を申し込むぜ☆（＾～＾）[%s]\n", message)
						oi.LongWrite(dia.writerToServer, []byte(message))
					}
				} else if dia.regexMatchAccepted.Match(promptStateBytes) {
					// 黒の手番から始まるぜ☆（＾～＾）
					dia.CurrentPhase = kwphase.Black
					// dia.turnState = 10

				} else if dia.regexDecline1.Match(promptStateBytes) {
					kwu.G.Chat.Trace("[対局はキャンセルされたぜ☆]")
				} else if dia.regexDecline2.Match(promptStateBytes) {
					kwu.G.Chat.Trace("[対局はキャンセルされたぜ☆]")
				} else if dia.regexOneSeven.Match(promptStateBytes) {
					kwu.G.Chat.Trace("[サブ遷移へ☆]")
					if dia.promptDiagram(lis, 7) {
						// このアプリを終了します
						kwu.G.Chat.Trace("...GE2NNGS... End parse\n")
						return true
					}
				} else {
					// "9 1 5" とか来るが、無視しろだぜ☆（＾～＾）
				}
			// Move
			// Example: `15 Game 2 I: kifuwarabe (0 2289 -1) vs kifuwarabi (0 2298 -1)`.
			// Example: `15   4(B): J4`.
			// A1 かもしれないし、 A12 かも知れず、いつコマンドが完了するか分からないので、２回以上実行されることはある。
			case 15:
				// 対局中、ゲーム情報は 指し手の前に毎回流れてくるぜ☆（＾～＾）
				// 自分が指すタイミングと、相手が指すタイミングのどちらでも流れてくるぜ☆（＾～＾）
				// とりあえずゲーム情報を全部変数に入れとけばあとで使える☆（＾～＾）
				matches2 := dia.regexGame.FindSubmatch(promptStateBytes)
				if 10 < len(matches2) {
					// 白 VS 黒 の順序固定なのか☆（＾～＾）？ それともマッチを申し込んだ方 VS 申し込まれた方 なのか☆（＾～＾）？
					// kwu.G.Chat.Trace("...GE2NNGS... 対局現在情報☆（＾～＾） gameid[%s], gametype[%s] white_user[%s][%s][%s][%s] black_user[%s][%s][%s][%s]", matches2[1], matches2[2], matches2[3], matches2[4], matches2[5], matches2[6], matches2[7], matches2[8], matches2[9], matches2[10])

					// ゲームID
					// Original code: @gameid
					gameID, err := strconv.ParseUint(string(matches2[1]), 10, 0)
					if err != nil {
						panic(kwu.G.Chat.Fatal(fmt.Sprintf("...GE2NNGS... %s", err)))
					}
					dia.GameID = uint(gameID)

					// ゲームの型？
					// Original code: @gametype
					dia.GameType = string(matches2[2])

					// 白手番の名前、フィールド２、残り時間（秒）、フィールド４
					// Original code: @white_user = [$3, $4, $5, $6]
					dia.GameWName = string(matches2[3])
					dia.GameWField2 = string(matches2[4])

					gameWAvailableSeconds, err := strconv.Atoi(string(matches2[5]))
					if err != nil {
						panic(kwu.G.Chat.Fatal(fmt.Sprintf("...GE2NNGS... %s", err)))
					}
					dia.GameWAvailableSeconds = gameWAvailableSeconds

					dia.GameWField4 = string(matches2[6])

					// 黒手番の名前、フィールド２、残り時間（秒）、フィールド４
					// Original code: @black_user = [$7, $8, $9, $10]
					dia.GameBName = string(matches2[7])
					dia.GameBField2 = string(matches2[8])

					gameBAvailableSeconds, err := strconv.Atoi(string(matches2[9]))
					if err != nil {
						panic(kwu.G.Chat.Fatal(fmt.Sprintf("...GE2NNGS... %s", err)))
					}
					dia.GameBAvailableSeconds = gameBAvailableSeconds

					dia.GameBField4 = string(matches2[10])

				} else {

					// 指し手はこっちだぜ☆（＾～＾）
					matches3 := dia.regexNngsMove.FindSubmatch(promptStateBytes)
					if 3 < len(matches3) {
						// Original code: @lastmove = [$1, $2, $3]

						// 相手の指し手を受信したのだから、手番はその逆だぜ☆（＾～＾）
						dia.CurrentPhase = phase.ToNum(phase.FlipColorString(string(matches3[2])))

						// kwu.G.Chat.Trace("...GE2NNGS... 指し手☆（＾～＾） code=%s color=%s move=%s MyColor=%s, CurrentPhase=%s\n", matches3[1], matches3[2], matches3[3], phase.ToString(dia.MyColor), phase.ToString(dia.CurrentPhase))
						if dia.MyColor == dia.CurrentPhase {
							// 自分の手番だぜ☆（＾～＾）！
							// kwu.G.Chat.Trace("...GE2NNGS... 相手の手を記憶☆（＾～＾） move=%s\n", matches3[3])
							dia.OpponentMove = string(matches3[3]) // 相手の指し手が付いてくるので記憶
							// kwu.G.Chat.Trace("...GE2NNGS... 自分の手番で一旦ブロッキング☆（＾～＾）")
							// 初回だけここを通るが、以後、ここには戻ってこないぜ☆（＾～＾）
							// dia.state = clistat.BlockingReceiver

							// Original code: nngsCUI.rb/announce class/update/`when 'my_turn'`.
							// Original code: nngsCUI.rb/engine  class/update/`when 'my_turn'`.

							// @gtp.time_left('WHITE', @nngs.white_user[2])
							// @gtp.time_left('BLACK', @nngs.black_user[2])
							//    mv, c = @gtp.genmove
							//    if mv.nil?
							//      mv = 'PASS'
							//    elsif mv == "resign"

							//    else
							//      i, j = mv
							//      mv = '' << 'ABCDEFGHJKLMNOPQRST'[i-1]
							//      mv = "#{mv}#{j}"
							//    end
							//    @nngs.input mv

							dia.play(lis)
							// TODO '= \n\n' が返ってくると思うが、 genmove と混線しない工夫が必要。

						} else {
							// 相手の手番だぜ☆（＾～＾）！
							// kwu.G.Chat.Trace("...GE2NNGS... 自分の手を記憶☆（＾～＾） move=%s\n", matches3[3])
							dia.MyMove = string(matches3[3]) // 自分の指し手が付いてくるので記憶
							// kwu.G.Chat.Trace("...GE2NNGS... 相手の手番で一旦ブロッキング☆（＾～＾）")
							// 初回だけここを通るが、以後、ここには戻ってこないぜ☆（＾～＾）
							// dia.state = clistat.BlockingSender

							// Original code: nngsCUI.rb/annouce class/update/`when 'his_turn'`.
							// Original code: nngsCUI.rb/engine  class/update/`when 'his_turn'`.

							// lis.
							//       mv = if move == 'Pass'
							//              nil
							//            elsif move.downcase[/resign/] == "resign"
							//              "resign"
							//            else
							//              i = move.upcase[0].ord - ?A.ord + 1
							// 	         i = i - 1 if i > ?I.ord - ?A.ord
							//              j = move[/[0-9]+/].to_i
							//              [i, j]
							//            end
							// #      p [mv, @his_color]
							//       @gtp.playmove([mv, @his_color])
						}

						dia.turn(lis)
					}
				}
			default:
				// 想定外のコードが来ても無視しろだぜ☆（＾～＾）
			}
		}
	case clistat.BlockingReceiver:
		// 申し込まれた方はブロック中です
		// kwu.G.Chat.Trace("...GE2NNGS... 申し込まれた方[%s]のブロッキング☆（＾～＾）", phase.ToString(dia.MyColor))
	case clistat.BlockingSender:
		// 申し込んだ方はブロック中です。
		// kwu.G.Chat.Trace("...GE2NNGS... 申し込んだ方[%s]のブロッキング☆（＾～＾）", phase.ToString(dia.MyColor))
	default:
		// 想定外の遷移だぜ☆（＾～＾）！
		panic(kwu.G.Chat.Fatal("...GE2NNGS... Unexpected state transition. state=%d", dia.state))
	}

	return false
}

func (dia *NngsClientStateDiagram) turn(lis *nngsClientStateDiagramListener) {
	// kwu.G.Chat.Trace("...GE2NNGS... ターン☆（＾～＾） MyColor=%s, CurrentPhase=%s\n", phase.ToString(dia.MyColor), phase.ToString(dia.CurrentPhase))
	if dia.MyColor == dia.CurrentPhase {
		// 自分の手番だぜ☆（＾～＾）！
		// kwu.G.Chat.Trace("...GE2NNGS... 自分の手番だぜ☆（＾～＾）！\n")

		dia.genmove(lis)

		lis.myTurn(dia)
	} else {
		// 相手の手番だぜ☆（＾～＾）！
		// kwu.G.Chat.Trace("...GE2NNGS... 相手の手番だぜ☆（＾～＾）！\n")
		lis.opponentTurn(dia)
	}
}

// play コマンドの応答を待ちます
func (dia *NngsClientStateDiagram) waitForPlayResponse(lis *nngsClientStateDiagramListener) {
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
				kwu.G.Chat.Error("...GE2NNGS<-- エラーだぜ☆（＾～＾）[%s]\n", err)
				return
			}
			// 送られてくる文字がなければ、ここをずっと通る？
			// kwu.G.Chat.Trace("...GE2NNGS... EOFだぜ☆（＾～＾）\n")
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

					// if dia.MyColor == dia.CurrentPhase {
					// 	kwu.G.Chat.Trace("...GE2NNGS<-- 空行(手番)。\n")
					// } else {
					// 	kwu.G.Chat.Trace("...GE2NNGS<-- 空行(相手番)。\n")
					// }

					// dia.ChatDebugState()

				} else {
					kwu.G.Chat.Notice("...GE2NNGS<-- [%s]\n", lineString)

					if lineString == "= " {
						// `play` の OK かも。
						// kwu.G.Chat.Trace("...GE2NNGS... playのOKかも☆（＾～＾）\n")
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

func (dia *NngsClientStateDiagram) waitForGenmoveResponse(lis *nngsClientStateDiagramListener) {
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
				kwu.G.Chat.Error("...GE2NNGS<-- エラーだぜ☆（＾～＾）[%s]\n", err)
				return
			}
			// 送られてくる文字がなければ、ここをずっと通る？
			// kwu.G.Chat.Trace("...GE2NNGS<-- EOFだぜ☆（＾～＾）\n")
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
						// Example: `= pass`
						if bestmove != "" {
							kwu.G.Chat.Notice("<--GE2NNGS... [%s\n]\n", bestmove)
							oi.LongWrite(dia.writerToServer, []byte(bestmove))
							oi.LongWrite(dia.writerToServer, []byte("\n"))
							// myTurn のループ終わり（＾～＾）！
							return
						}
						// bestmove 未決定時
						kwu.G.Chat.Trace("...GE2NNGS<-- 空行(手番)\n")
					} else {
						kwu.G.Chat.Trace("...GE2NNGS<-- 空行(相手番)\n")
					}

					// dia.ChatTraceState()

				} else {
					kwu.G.Chat.Notice("...GE2NNGS<-- [%s]\n", lineString)

					// if lineString == "= " {
					// 	// `play` の OK かも。
					// 	kwu.G.Chat.Trace("...GE2NNGS... playのOKかも☆（＾～＾） lineString=[%s]\n", lineString)
					// } else {
					// サーバーに着手を送信します。１行前の文字列を使います
					// Example: `= A1`
					// Example: `= pass`
					matches71 := dia.regexEngineBestmove.FindSubmatch(lineBuffer[:index])
					if 1 < len(matches71) {
						// 着手
						bestmove = string(matches71[1])
						// kwu.G.Chat.Trace("...GE2NNGS... bestmove=[%s]\n", bestmove)
						// } else {
						// 	// TODO 空行とは限らないだろ、変なコマンドかも（＾～＾）？
						// 	kwu.G.Chat.Trace("...GE2NNGS... 空行(手番)。line=[%s] bestmove=[%s] len=[%d]\n", string(lineBuffer[:index]), bestmove, len(matches71))
					}
					//}
				}

				index = 0
			} else {
				lineBuffer[index] = bytes[0]
				index++
			}
		}
	}
}
