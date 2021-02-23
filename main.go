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
	g "github.com/muzudho/gtp-engine-to-nngs/global"
	"github.com/muzudho/gtp-engine-to-nngs/ui"
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
	g.G = *new(g.GlobalVariables)

	// ロガーの作成。
	g.G.Log = *kwu.NewLogger(
		filepath.Join(*workdir, "output/connector/trace.log"),
		filepath.Join(*workdir, "output/connector/debug.log"),
		filepath.Join(*workdir, "output/connector/info.log"),
		filepath.Join(*workdir, "output/connector/notice.log"),
		filepath.Join(*workdir, "output/connector/warn.log"),
		filepath.Join(*workdir, "output/connector/error.log"),
		filepath.Join(*workdir, "output/connector/fatal.log"),
		filepath.Join(*workdir, "output/connector/print.log"))

	// 既存のログ・ファイルを削除。
	g.G.Log.RemoveAllOldLogs()

	// // ログ・ファイルの開閉
	g.G.Log.OpenAllLogs()
	defer g.G.Log.CloseAllLogs()

	// チャッターの作成も、エンジンが起動時に行う
	g.G.Chat = *kwu.NewChatter(g.G.Log)
	g.G.StderrChat = *kwu.NewStderrChatter(g.G.Log)

	g.G.Chat.Trace("...GE2NNGS... Start program\n")

	// 設定ファイル読込
	engineConf, err := kwui.LoadEngineConf(engineConfPath)
	if err != nil {
		panic(g.G.Chat.Fatal("...GE2NNGS... engineConfPath=[%s] err=[%s]\n", engineConfPath, err))
	}

	connectorConf, err := ui.LoadConnectorConf(connectorConfPath)
	if err != nil {
		panic(g.G.Chat.Fatal("...GE2NNGS... connectorConfPath=[%s] err=[%s]\n", connectorConfPath, err))
	}

	g.G.Chat.Trace("...GE2NNGS... (^q^) プレイヤーのタイプ☆ [%s]\n", connectorConf.User.InterfaceType)

	// 思考エンジンを起動
	startEngine(engineConf, connectorConf, workdir)

	g.G.Chat.Trace("...GE2NNGS... End program\n")
}

// 思考エンジンを起動
func startEngine(engineConf *kwe.EngineConf, connectorConf *e.ConnectorConf, workdir *string) {
	parameters := strings.Split("--workdir "+*workdir+" "+connectorConf.User.EngineCommandOption, " ")
	parametersString := strings.Join(parameters, " ")
	parametersString = strings.TrimRight(parametersString, " ")
	g.G.Chat.Trace("...GE2NNGS... (^q^) GTP対応の思考エンジンを起動するぜ☆ 途中で終わりたかったら [Ctrl]+[C]\n")
	g.G.Chat.Trace("...GE2NNGS... (^q^) command=[%s] argumentList=[%s]\n", connectorConf.User.EngineCommand, parametersString)
	cmd := exec.Command(connectorConf.User.EngineCommand, parameters...)

	engineStdin, _ := cmd.StdinPipe()
	defer engineStdin.Close()

	engineStdout, _ := cmd.StdoutPipe()
	defer engineStdout.Close()

	err := cmd.Start()
	if err != nil {
		panic(g.G.Chat.Fatal(fmt.Sprintf("...GE2NNGS... cmd.Start() --> %s", err)))
	}

	c.SpawnServerConnection(engineConf, connectorConf, &engineStdin, &engineStdout)
	cmd.Wait()
}
