package main

import (
	"flag"
	"fmt"
	"io"
	"os/exec"
	"strings"

	c "github.com/muzudho/gtp-engine-to-nngs/controller"
	e "github.com/muzudho/gtp-engine-to-nngs/entities"
	"github.com/muzudho/gtp-engine-to-nngs/ui"
	u "github.com/muzudho/gtp-engine-to-nngs/usecases"
)

func main() {
	// コマンドライン引数
	entryConfPath := flag.String("entry", "./input/default.entryConf.toml", "*.entryConf.toml file path.")
	flag.Parse()
	fmt.Printf("[情報] flag.Args()=%s\n", flag.Args())
	fmt.Printf("[情報] entryConfPath=%s\n", *entryConfPath)

	// グローバル変数の作成
	u.G = *new(u.GlobalVariables)

	// ロガーの作成。
	u.G.Log = *u.NewLogger(
		"output/trace.log",
		"output/debug.log",
		"output/info.log",
		"output/notice.log",
		"output/warn.log",
		"output/error.log",
		"output/fatal.log",
		"output/print.log")

	// チャッターの作成。 標準出力とロガーを一緒にしただけです。
	u.G.Chat = *u.NewChatter(u.G.Log)

	// fmt.Println("[情報] 設定ファイルを読み込んだろ☆（＾～＾）")
	entryConf := ui.LoadEntryConf(*entryConfPath) // "./input/default.entryConf.toml"

	// NNGSからのメッセージ受信に対応するプログラムを指定したろ☆（＾～＾）
	fmt.Printf("[情報] (^q^) プレイヤーのタイプ☆ [%s]", entryConf.User.InterfaceType)

	// 思考エンジンを起動
	engineStdin, engineStdout := startEngine(entryConf)

	fmt.Println("[情報] (^q^) 何か文字を打てだぜ☆ 終わりたかったら [Ctrl]+[C]☆")
	c.Spawn(entryConf, &engineStdin, &engineStdout)

	engineStdin.Close()
	fmt.Println("[情報] (^q^) おわり☆！")
}

// 思考エンジンを起動
func startEngine(entryConf e.EntryConf) (io.WriteCloser, io.ReadCloser) {
	parameters := strings.Split(entryConf.User.EngineCommandOption, " ")
	fmt.Printf("(^q^) GTP対応の思考エンジンを起動するぜ☆ [%s] [%s]", entryConf.User.EngineCommand, strings.Join(parameters, " "))
	cmd := exec.Command(entryConf.User.EngineCommand, parameters...)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	return stdin, stdout
}
