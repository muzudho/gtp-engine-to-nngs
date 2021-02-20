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
	kwu "github.com/muzudho/kifuwarabe-gtp/usecases"
)

func main() {
	// Working directory
	wdir, err := os.Getwd()
	if err != nil {
		// ここでは、ログはまだ設定できてない
		panic(fmt.Sprintf("<GE2NNGS> wdir=%s", wdir))
	}
	fmt.Printf("<GE2NNGS> wdir=%s\n", wdir)

	// コマンドライン引数
	workdir := flag.String("workdir", wdir, "Working directory path.")
	flag.Parse()
	fmt.Printf("<GE2NNGS> flag.Args()=%s\n", flag.Args())
	fmt.Printf("<GE2NNGS> workdir=%s\n", *workdir)
	entryConfPath := filepath.Join(*workdir, "input/default.entryConf.toml")
	fmt.Printf("<GE2NNGS> entryConfPath=%s\n", entryConfPath)

	// グローバル変数の作成
	u.G = *new(u.GlobalVariables)

	// ロガーの作成。
	kwu.G.Log = *kwu.NewLogger(
		filepath.Join(*workdir, "output/trace.log"),
		filepath.Join(*workdir, "output/debug.log"),
		filepath.Join(*workdir, "output/info.log"),
		filepath.Join(*workdir, "output/notice.log"),
		filepath.Join(*workdir, "output/warn.log"),
		filepath.Join(*workdir, "output/error.log"),
		filepath.Join(*workdir, "output/fatal.log"),
		filepath.Join(*workdir, "output/print.log"))

	// チャッターの作成。 標準出力とロガーを一緒にしただけです。
	kwu.G.Chat = *kwu.NewChatter(kwu.G.Log)
	kwu.G.StderrChat = *kwu.NewStderrChatter(kwu.G.Log)

	// fmt.Println("<GE2NNGS> 設定ファイルを読み込んだろ☆（＾～＾）")
	entryConf := ui.LoadEntryConf(entryConfPath) // "./input/default.entryConf.toml"

	// NNGSからのメッセージ受信に対応するプログラムを指定したろ☆（＾～＾）
	fmt.Printf("<GE2NNGS> (^q^) プレイヤーのタイプ☆ [%s]", entryConf.User.InterfaceType)

	// 思考エンジンを起動
	engineStdin, engineStdout := startEngine(entryConf, workdir)

	fmt.Println("<GE2NNGS> (^q^) 何か文字を打てだぜ☆ 終わりたかったら [Ctrl]+[C]☆")
	c.Spawn(entryConf, &engineStdin, &engineStdout)

	engineStdin.Close()
	fmt.Println("<GE2NNGS> (^q^) おわり☆！")
}

// 思考エンジンを起動
func startEngine(entryConf e.EntryConf, workdir *string) (io.WriteCloser, io.ReadCloser) {
	parameters := strings.Split("--workdir "+*workdir+" "+entryConf.User.EngineCommandOption, " ")
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
