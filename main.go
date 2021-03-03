package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	l "github.com/muzudho/go-logger"
	c "github.com/muzudho/gtp-engine-to-nngs/controller"
	e "github.com/muzudho/gtp-engine-to-nngs/entities"
	g "github.com/muzudho/gtp-engine-to-nngs/global"
	"github.com/muzudho/gtp-engine-to-nngs/ui"
	kwe "github.com/muzudho/kifuwarabe-gtp/entities"
	kwui "github.com/muzudho/kifuwarabe-gtp/ui"
)

func main() {
	// Working directory
	dwd, err := os.Getwd()
	if err != nil {
		// ここでは、ログはまだ設定できてない
		panic(fmt.Sprintf("...GE2NNGS... DefaultWorkingDirectory=%s", dwd))
	}
	fmt.Printf("...GE2NNGS... DefaultWorkingDirectory=%s\n", dwd)

	// コマンドライン引数登録
	wd := flag.String("workdir", dwd, "Working directory path.")
	// 解析
	flag.Parse()

	connectorConfPath := filepath.Join(*wd, "input/connector.conf.toml")
	engineConfPath := filepath.Join(*wd, "input/engine.conf.toml")
	fmt.Printf("...GE2NNGS... WorkingDirectory=[%s] ConnectorConf=[%s] EngineConf=[%s]\n", *wd, connectorConfPath, engineConfPath)

	// グローバル変数の作成
	g.G = *new(g.Variables)

	// ロガーの作成。
	g.G.Log = *l.NewLogger(
		filepath.Join(*wd, "output/connector/trace.log"),
		filepath.Join(*wd, "output/connector/debug.log"),
		filepath.Join(*wd, "output/connector/info.log"),
		filepath.Join(*wd, "output/connector/notice.log"),
		filepath.Join(*wd, "output/connector/warn.log"),
		filepath.Join(*wd, "output/connector/error.log"),
		filepath.Join(*wd, "output/connector/fatal.log"),
		filepath.Join(*wd, "output/connector/print.log"))

	// 既存のログ・ファイルを削除。
	g.G.Log.RemoveAllOldLogs()

	// ログ・ファイルの開閉
	err = g.G.Log.OpenAllLogs()
	if err != nil {
		// ログ・ファイルを開くのに失敗したのだから、ログ・ファイルへは書き込めません
		panic(fmt.Sprintf("...GE2NNGS... %s", err))
	}

	defer g.G.Log.CloseAllLogs()

	// チャッターの作成も、エンジンが起動時に行う
	g.G.Chat = *l.NewChatter(g.G.Log)
	g.G.StderrChat = *l.NewStderrChatter(g.G.Log)

	g.G.Chat.Trace("...GE2NNGS... Start program\n")

	// 設定ファイル読込
	engineConf, err := kwui.LoadEngineConf(engineConfPath)
	if err != nil {
		panic(g.G.Chat.Fatal("...GE2NNGS... engineConf err=[%s]\n", err))
	}

	connectorConf, err := ui.LoadConnectorConf(connectorConfPath)
	if err != nil {
		panic(g.G.Chat.Fatal("...GE2NNGS... connectorConf err=[%s]\n", err))
	}

	g.G.Chat.Trace("...GE2NNGS... (^q^) プレイヤーのタイプ☆ [%s]\n", connectorConf.User.InterfaceType)

	// 思考エンジンを起動
	startEngine(engineConf, connectorConf, wd)

	g.G.Chat.Trace("...GE2NNGS... End program\n")
}

// 思考エンジンを起動
// `wd` - Working directory
func startEngine(engineConf *kwe.EngineConf, connectorConf *e.ConnectorConf, wd *string) {
	parameters := strings.Split("--workdir "+*wd+" "+connectorConf.User.EngineCommandOption, " ")
	parametersString := strings.Join(parameters, " ")
	parametersString = strings.TrimRight(parametersString, " ")
	g.G.Chat.Trace("...GE2NNGS... (^q^) GTP対応の思考エンジンを起動するぜ☆ 途中で終わりたかったら [Ctrl]+[C]\n")
	g.G.Chat.Trace("...GE2NNGS... (^q^) command=[%s] argumentList=[%s]\n", connectorConf.User.EngineCommand, parametersString)
	cmd := exec.Command(connectorConf.User.EngineCommand, parameters...)

	engineStdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	defer engineStdin.Close()

	engineStdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	defer engineStdout.Close()

	err = cmd.Start()
	if err != nil {
		panic(g.G.Chat.Fatal(fmt.Sprintf("...GE2NNGS... cmd.Start() --> %s", err)))
	}

	g.G.Chat.Trace("...GE2NNGS... SpawnServerConnection Begin\n")
	c.SpawnServerConnection(engineConf, connectorConf, &engineStdin, &engineStdout)
	g.G.Chat.Trace("...GE2NNGS... SpawnServerConnection End\n")

	cmd.Wait()
}
