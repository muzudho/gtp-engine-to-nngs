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
	u "github.com/muzudho/gtp-engine-to-nngs/usecases"
	"github.com/reiver/go-oi"
	"github.com/reiver/go-telnet"
)

// NngsClientStateDiagram - NNGSクライアントの状態遷移図
type NngsClientStateDiagram struct {
	// EngineStdin - GTP Engine stdin
	EngineStdin *io.WriteCloser
	// EngineStdin - GTP Engine stdin
	EngineStdout *io.ReadCloser

	entryConf e.EntryConf

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
	regexMove          regexp.Regexp
	regexAcceptCommand regexp.Regexp
	// Example: `= A1`
	regexBestmove regexp.Regexp

	// MyColor - 自分の手番の色
	MyColor phase.Phase
	// Phase - これから指す方。局面の手番とは逆になる
	CurrentPhase phase.Phase

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

// ChatDebugState - 状態をデバッグ表示
func (dia *NngsClientStateDiagram) ChatDebugState() {
	u.G.Chat.Debug("<GE2NNGS> state=%d promptState=%d newlineReadableState=%d\n", dia.state, dia.promptState, dia.newlineReadableState)
}

func (dia *NngsClientStateDiagram) promptDiagram(lis *nngsClientStateDiagramListener, subCode int) {
	switch subCode {
	// Info
	case 5:
		if dia.promptState == 0 {
			// このフラグを立てるのは初回だけ。
			dia.newlineReadableState = 1
		}

		if dia.promptState == 7 {
			lis.matchEnd() // 対局終了
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
			message := "done\nquit\n"
			fmt.Printf("<GE2NNGS> 得点計算は飛ばすぜ☆（＾～＾）対局も終了するぜ☆（＾～＾）\n")
			u.G.Chat.Debug("<GE2NNGS> サーバーへ送信[%s]\n", message)
			oi.LongWrite(dia.writerToServer, []byte(message))
		}
		dia.promptState = 7
	default:
		// "1 1" とか来ても無視しろだぜ☆（＾～＾）
	}
}

func (dia *NngsClientStateDiagram) parse(lis *nngsClientStateDiagramListener) {
	// 現在読み取り中の文字なので、早とちりするかも知れないぜ☆（＾～＾）
	line := string(dia.lineBuffer[:dia.index])

	if dia.newlineReadableState == 2 {
		print(fmt.Sprintf("受信[%s]\n", line))
	}

	switch dia.state {
	case clistat.None:
		// Original code: NngsClient.rb/NNGSClient/`def login`
		// Waitfor "Login: ".
		if line == "Login: " {
			// あなたの名前を入力してください。

			// 設定ファイルから自動で入力するぜ☆（＾ｑ＾）
			user := dia.entryConf.UserName()

			// 自動入力のときは、設定ミスなら強制終了しないと無限ループしてしまうぜ☆（＾～＾）
			if user == "" {
				panic("Need name (UserName)")
			}

			u.G.Chat.Debug("<GE2NNGS> サーバーへ送信[%s\n]\n", user)
			oi.LongWrite(dia.writerToServer, []byte(user))
			oi.LongWrite(dia.writerToServer, []byte("\n"))

			dia.state = clistat.EnteredMyName
		}
	// Original code: NngsClient.rb/NNGSClient/`def login`
	case clistat.EnteredMyName:
		if line == "1 1" {
			// パスワードを入れろだぜ☆（＾～＾）
			if dia.entryConf.Pass() == "" {
				panic("Need password")
			}

			u.G.Chat.Debug("<GE2NNGS> サーバーへ送信[%s\n]\n", dia.entryConf.User.Pass)
			oi.LongWrite(dia.writerToServer, []byte(dia.entryConf.User.Pass))
			oi.LongWrite(dia.writerToServer, []byte("\n"))
			setClientMode(dia.writerToServer)
			dia.state = clistat.EnteredClientMode

		} else if line == "Password: " {
			// パスワードを入れろだぜ☆（＾～＾）
			if dia.entryConf.Pass() == "" {
				panic("Need password")
			}

			u.G.Chat.Debug("<GE2NNGS> サーバーへ送信[%s\n]\n", dia.entryConf.User.Pass)
			oi.LongWrite(dia.writerToServer, []byte(dia.entryConf.User.Pass))
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
		if dia.entryConf.ApplyFromMe() {
			// 対局を申し込みます。
			_, configuredColorUpperCase := dia.entryConf.MyColor()

			fmt.Printf("<GE2NNGS> lis.MyColorを%sに変更☆（＾～＾）\n", configuredColorUpperCase)
			dia.MyColor = phase.ToNum(configuredColorUpperCase)

			message := fmt.Sprintf("match %s %s %d %d %d\n", dia.entryConf.OpponentName(), configuredColorUpperCase, dia.entryConf.BoardSize(), dia.entryConf.AvailableTimeMinutes(), dia.entryConf.CanadianTiming())
			fmt.Printf("<GE2NNGS> 対局を申し込んだぜ☆（＾～＾）")
			u.G.Chat.Debug("<GE2NNGS> サーバーへ送信[%s]\n", message)
			oi.LongWrite(dia.writerToServer, []byte(message))
		}
		dia.state = clistat.WaitingInInfo

	// '1 5' - Waiting
	case clistat.WaitingInInfo:
		// Example: 1 5
		matches := dia.regexCommand.FindSubmatch(dia.lineBuffer[:dia.index])

		//fmt.Printf("<GE2NNGS> m[%s]", matches)
		if 2 < len(matches) {
			commandCodeBytes := matches[1]
			commandCode := string(commandCodeBytes)
			promptStateBytes := matches[2]

			code, err := strconv.Atoi(commandCode)
			if err != nil {
				panic(err) // 想定外の遷移だぜ☆（＾～＾）！
			}
			switch code {
			// Prompt
			case 1:
				promptState := string(promptStateBytes)
				promptStateNum, err := strconv.Atoi(promptState)
				if err == nil {
					dia.promptDiagram(lis, promptStateNum)
				}
			// Info
			case 9:
				if dia.regexUseMatch.Match(promptStateBytes) {
					matches2 := dia.regexUseMatchToRespond.FindSubmatch(promptStateBytes)
					if 2 < len(matches2) {
						// 対局を申し込まれた方だけ、ここを通るぜ☆（＾～＾）
						fmt.Printf("<GE2NNGS> 対局を申し込まれたぜ☆（＾～＾）[%s] accept[%s],decline[%s]\n", string(promptStateBytes), matches2[1], matches2[2])

						// Example: `match kifuwarabi W 19 40 0`
						dia.CommandOfMatchAccept = string(matches2[1])
						// Example: `decline kifuwarabi`
						dia.CommandOfMatchDecline = string(matches2[2])

						// acceptコマンドを半角空白でスプリットした３番目が、自分の手番
						matchAcceptTokens := strings.Split(dia.CommandOfMatchAccept, " ")
						if len(matchAcceptTokens) < 6 {
							panic(fmt.Sprintf("Error matchAcceptTokens=[%s].", matchAcceptTokens))
						}

						opponentPlayerName := matchAcceptTokens[1]
						myColorString := matchAcceptTokens[2]
						myColorUppercase := strings.ToUpper(myColorString)
						fmt.Printf("<GE2NNGS> MyColorを[%s]に変更☆（＾～＾）\n", myColorString)
						dia.MyColor = phase.ToNum(myColorString)

						boardSize, err := strconv.ParseUint(matchAcceptTokens[3], 10, 0)
						if err != nil {
							panic(err)
						}
						dia.BoardSize = uint(boardSize)
						fmt.Printf("<GE2NNGS> ボードサイズは%d☆（＾～＾）", dia.BoardSize)

						configuredColor, _ := dia.entryConf.MyColor()

						if dia.MyColor != configuredColor {
							panic(fmt.Sprintf("(Err.195) Unexpected phase. MyColor=%d configuredColor=%d.", dia.MyColor, configuredColor))
						}

						// cmd_match
						message := fmt.Sprintf("match %s %s %d %d %d\n", opponentPlayerName, myColorUppercase, dia.entryConf.BoardSize(), dia.entryConf.AvailableTimeMinutes(), dia.entryConf.CanadianTiming())
						fmt.Printf("<GE2NNGS> 対局を申し込むぜ☆（＾～＾）[%s]\n", message)
						oi.LongWrite(dia.writerToServer, []byte(message))
					}
				} else if dia.regexMatchAccepted.Match(promptStateBytes) {
					// 黒の手番から始まるぜ☆（＾～＾）
					dia.CurrentPhase = phase.Black
					// dia.turnState = 10

				} else if dia.regexDecline1.Match(promptStateBytes) {
					print("[対局はキャンセルされたぜ☆]")
				} else if dia.regexDecline2.Match(promptStateBytes) {
					print("[対局はキャンセルされたぜ☆]")
				} else if dia.regexOneSeven.Match(promptStateBytes) {
					print("[サブ遷移へ☆]")
					dia.promptDiagram(lis, 7)
				} else {
					// "9 1 5" とか来るが、無視しろだぜ☆（＾～＾）
				}
			// Move
			// Example: `15 Game 2 I: kifuwarabe (0 2289 -1) vs kifuwarabi (0 2298 -1)`.
			// Example: `15   4(B): J4`.
			// A1 かもしれないし、 A12 かも知れず、いつコマンドが完了するか分からないので、２回以上実行されることはある。
			case 15:
				// print("15だぜ☆")

				// 対局中、ゲーム情報は 指し手の前に毎回流れてくるぜ☆（＾～＾）
				// 自分が指すタイミングと、相手が指すタイミングのどちらでも流れてくるぜ☆（＾～＾）
				// とりあえずゲーム情報を全部変数に入れとけばあとで使える☆（＾～＾）
				matches2 := dia.regexGame.FindSubmatch(promptStateBytes)
				if 10 < len(matches2) {
					// 白 VS 黒 の順序固定なのか☆（＾～＾）？ それともマッチを申し込んだ方 VS 申し込まれた方 なのか☆（＾～＾）？
					fmt.Printf("<GE2NNGS> 対局現在情報☆（＾～＾） gameid[%s], gametype[%s] white_user[%s][%s][%s][%s] black_user[%s][%s][%s][%s]", matches2[1], matches2[2], matches2[3], matches2[4], matches2[5], matches2[6], matches2[7], matches2[8], matches2[9], matches2[10])

					// ゲームID
					// Original code: @gameid
					gameID, err := strconv.ParseUint(string(matches2[1]), 10, 0)
					if err != nil {
						panic(err)
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
						panic(err)
					}
					dia.GameWAvailableSeconds = gameWAvailableSeconds

					dia.GameWField4 = string(matches2[6])

					// 黒手番の名前、フィールド２、残り時間（秒）、フィールド４
					// Original code: @black_user = [$7, $8, $9, $10]
					dia.GameBName = string(matches2[7])
					dia.GameBField2 = string(matches2[8])

					gameBAvailableSeconds, err := strconv.Atoi(string(matches2[9]))
					if err != nil {
						panic(err)
					}
					dia.GameBAvailableSeconds = gameBAvailableSeconds

					dia.GameBField4 = string(matches2[10])

				} else {

					// 指し手はこっちだぜ☆（＾～＾）
					matches3 := dia.regexMove.FindSubmatch(promptStateBytes)
					if 3 < len(matches3) {
						// Original code: @lastmove = [$1, $2, $3]

						// 相手の指し手を受信したのだから、手番はその逆だぜ☆（＾～＾）
						dia.CurrentPhase = phase.ToNum(phase.FlipColorString(string(matches3[2])))

						fmt.Printf("<GE2NNGS> 指し手☆（＾～＾） code=%s color=%s move=%s MyColor=%s, CurrentPhase=%s\n", matches3[1], matches3[2], matches3[3], phase.ToString(dia.MyColor), phase.ToString(dia.CurrentPhase))
						if dia.MyColor == dia.CurrentPhase {
							// 自分の手番だぜ☆（＾～＾）！
							fmt.Printf("<GE2NNGS> 相手の手を記憶☆（＾～＾） move=%s\n", matches3[3])
							dia.OpponentMove = string(matches3[3]) // 相手の指し手が付いてくるので記憶
							// fmt.Printf("<GE2NNGS> 自分の手番で一旦ブロッキング☆（＾～＾）")
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

							//if dia.OpponentMove != "" {
							message := strings.ToLower(fmt.Sprintf("play %s %s\n", phase.FlipColorString(phase.ToString(dia.MyColor)), dia.OpponentMove))
							fmt.Printf("<GE2NNGS> エンジンへ送信[%s]\n", message)
							(*dia.EngineStdin).Write([]byte(message))

							// if dia.CurrentPhase == phase.Black {
							// 	dia.turnState = 30
							// } else {
							// 	dia.turnState = 10
							// }

							// TODO '= \n\n' が返ってくると思うが、 genmove と混線しない工夫が必要。
							//}
							lis.waitForPlayResponse(dia)

						} else {
							// 相手の手番だぜ☆（＾～＾）！
							fmt.Printf("<GE2NNGS> 自分の手を記憶☆（＾～＾） move=%s\n", matches3[3])
							dia.MyMove = string(matches3[3]) // 自分の指し手が付いてくるので記憶
							// fmt.Printf("<GE2NNGS> 相手の手番で一旦ブロッキング☆（＾～＾）")
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
		fmt.Printf("<GE2NNGS> 申し込まれた方[%s]のブロッキング☆（＾～＾）", phase.ToString(dia.MyColor))
	case clistat.BlockingSender:
		// 申し込んだ方はブロック中です。
		fmt.Printf("<GE2NNGS> 申し込んだ方[%s]のブロッキング☆（＾～＾）", phase.ToString(dia.MyColor))
	default:
		// 想定外の遷移だぜ☆（＾～＾）！
		panic(fmt.Sprintf("Unexpected state transition. state=%d", dia.state))
	}
}

func (dia *NngsClientStateDiagram) turn(lis *nngsClientStateDiagramListener) {
	fmt.Printf("<GE2NNGS> ターン☆（＾～＾） MyColor=%s, CurrentPhase=%s\n", phase.ToString(dia.MyColor), phase.ToString(dia.CurrentPhase))
	if dia.MyColor == dia.CurrentPhase {
		// 自分の手番だぜ☆（＾～＾）！
		fmt.Printf("<GE2NNGS> 自分の手番だぜ☆（＾～＾）！\n")
		lis.myTurn(dia)
	} else {
		// 相手の手番だぜ☆（＾～＾）！
		fmt.Printf("<GE2NNGS> 相手の手番だぜ☆（＾～＾）！\n")
		lis.opponentTurn(dia)
	}
}
