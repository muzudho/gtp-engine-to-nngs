package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	c "github.com/muzudho/gtp-engine-to-nngs/controller"
	e "github.com/muzudho/gtp-engine-to-nngs/entities"
	"github.com/muzudho/gtp-engine-to-nngs/ui"
	u "github.com/muzudho/gtp-engine-to-nngs/usecases"
	kwe "github.com/muzudho/kifuwarabe-gtp/entities"
	kwui "github.com/muzudho/kifuwarabe-gtp/ui"
	kwu "github.com/muzudho/kifuwarabe-gtp/usecases"
)

func main() {
	// Working directory
	wdir, err := os.Getwd()
	if err != nil {
		// ここでは、ログはまだ設定できてない
		panic(fmt.Sprintf("...GE2NNGS... wdir=%s", wdir))
	}
	fmt.Printf("...GE2NNGS... wdir=%s\n", wdir)

	// コマンドライン引数
	workdir := flag.String("workdir", wdir, "Working directory path.")
	flag.Parse()
	fmt.Printf("...GE2NNGS... flag.Args()=%s\n", flag.Args())
	fmt.Printf("...GE2NNGS... workdir=%s\n", *workdir)
	connectorConfPath := filepath.Join(*workdir, "input/connector.conf.toml")
	engineConfPath := filepath.Join(*workdir, "input/engine.conf.toml")
	fmt.Printf("...GE2NNGS... connectorConfPath=%s\n", connectorConfPath)
	fmt.Printf("...GE2NNGS... engineConfPath=%s\n", engineConfPath)

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

	// 既存のログ・ファイルを削除。エンジンが起動時に行う

	// ログ・ファイルの開閉
	kwu.G.Log.OpenAllLogs()
	defer kwu.G.Log.CloseAllLogs()

	// チャッターの作成。 標準出力とロガーを一緒にしただけです。
	kwu.G.Chat = *kwu.NewChatter(kwu.G.Log)
	kwu.G.StderrChat = *kwu.NewStderrChatter(kwu.G.Log)

	kwu.G.Chat.Trace("...GE2NNGS... Start program\n")

	// 設定ファイル読込
	engineConf, err := kwui.LoadEngineConf(engineConfPath)
	if err != nil {
		panic(kwu.G.Chat.Fatal("...GE2NNGS... engineConfPath=[%s] err=[%s]\n", engineConfPath, err))
	}

	connectorConf, err := ui.LoadConnectorConf(connectorConfPath)
	if err != nil {
		panic(kwu.G.Chat.Fatal("...GE2NNGS... connectorConfPath=[%s] err=[%s]\n", connectorConfPath, err))
	}

	kwu.G.Chat.Trace("...GE2NNGS... (^q^) プレイヤーのタイプ☆ [%s]\n", connectorConf.User.InterfaceType)

	// 思考エンジンを起動
	startEngine(engineConf, connectorConf, workdir)

	kwu.G.Chat.Trace("...GE2NNGS... End program\n")
}

// 思考エンジンを起動
func startEngine(engineConf *kwe.EngineConf, connectorConf *e.ConnectorConf, workdir *string) {
	parameters := strings.Split("--workdir "+*workdir+" "+connectorConf.User.EngineCommandOption, " ")
	kwu.G.Chat.Trace("...GE2NNGS... (^q^) GTP対応の思考エンジンを起動するぜ☆ 途中で終わりたかったら [Ctrl]+[C]\n")
	kwu.G.Chat.Trace("...GE2NNGS... (^q^) command=[%s] argumentList=[%s]\n", connectorConf.User.EngineCommand, strings.Join(parameters, " "))
	cmd := exec.Command(connectorConf.User.EngineCommand, parameters...)

	engineStdin, _ := cmd.StdinPipe()
	defer engineStdin.Close()

	engineStdout, _ := cmd.StdoutPipe()
	defer engineStdout.Close()

	err := cmd.Start()
	if err != nil {
		panic(kwu.G.Chat.Fatal(fmt.Sprintf("...GE2NNGS... %s", err)))
	}

	c.Spawn(engineConf, connectorConf, &engineStdin, &engineStdout)
	// cmd.Wait()
}
