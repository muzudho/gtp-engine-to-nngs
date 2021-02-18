package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	c "github.com/muzudho/gtp-engine-to-nngs/controller"
	e "github.com/muzudho/gtp-engine-to-nngs/entities"
	"github.com/muzudho/gtp-engine-to-nngs/ui"
	u "github.com/muzudho/gtp-engine-to-nngs/usecases"
)

func main() {
	// Working directory
	wdir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("[情報] wdir=%s", wdir))
	}
	fmt.Printf("[情報] wdir=%s\n", wdir)

	// コマンドライン引数
	workdir := flag.String("workdir", wdir, "Working directory path.")
	flag.Parse()
	fmt.Printf("[情報] flag.Args()=%s\n", flag.Args())
	fmt.Printf("[情報] workdir=%s\n", *workdir)
	entryConfPath := filepath.Join(*workdir, "input/default.entryConf.toml")
	fmt.Printf("[情報] entryConfPath=%s\n", entryConfPath)

	// グローバル変数の作成
	u.G = *new(u.GlobalVariables)

	// ロガーの作成。
	u.G.Log = *u.NewLogger(
		filepath.Join(*workdir, "output/trace.log"),
		filepath.Join(*workdir, "output/debug.log"),
		filepath.Join(*workdir, "output/info.log"),
		filepath.Join(*workdir, "output/notice.log"),
		filepath.Join(*workdir, "output/warn.log"),
		filepath.Join(*workdir, "output/error.log"),
		filepath.Join(*workdir, "output/fatal.log"),
		filepath.Join(*workdir, "output/print.log"))

	// チャッターの作成。 標準出力とロガーを一緒にしただけです。
	u.G.Chat = *u.NewChatter(u.G.Log)

	// fmt.Println("[情報] 設定ファイルを読み込んだろ☆（＾～＾）")
	entryConf := ui.LoadEntryConf(entryConfPath) // "./input/default.entryConf.toml"

	// NNGSからのメッセージ受信に対応するプログラムを指定したろ☆（＾～＾）
	fmt.Printf("[情報] (^q^) プレイヤーのタイプ☆ [%s]", entryConf.User.InterfaceType)

	// 思考エンジンを起動
	engineStdin, engineStdout := startEngine(entryConf, workdir)

	fmt.Println("[情報] (^q^) 何か文字を打てだぜ☆ 終わりたかったら [Ctrl]+[C]☆")
	c.Spawn(entryConf, &engineStdin, &engineStdout)

	engineStdin.Close()
	fmt.Println("[情報] (^q^) おわり☆！")
}

// 思考エンジンを起動
func startEngine(entryConf e.EntryConf, workdir *string) (io.WriteCloser, io.ReadCloser) {
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
