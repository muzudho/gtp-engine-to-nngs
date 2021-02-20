package ui

import (
	"io/ioutil"

	e "github.com/muzudho/gtp-engine-to-nngs/entities"
	kwu "github.com/muzudho/kifuwarabe-gtp/usecases"
	"github.com/pelletier/go-toml"
)

// LoadEntryConf - Toml形式の参加設定ファイルを読み込みます。
func LoadEntryConf(path string) e.EntryConf {

	// ファイル読込
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		panic(kwu.G.Chat.Fatal("path=%s err=%s", path, err))
	}

	debugPrintToml(fileData)

	// Toml解析
	binary := []byte(string(fileData))
	config := e.EntryConf{}
	toml.Unmarshal(binary, &config)

	debugPrintConfig(config)

	return config
}

func debugPrintToml(fileData []byte) {
	// kwu.G.Chat.Trace("...GE2NNGS... content=%s", string(fileData))

	// Toml解析
	tomlTree, err := toml.Load(string(fileData))
	if err != nil {
		panic(kwu.G.Chat.Fatal(err.Error()))
	}
	kwu.G.Chat.Trace("...GE2NNGS... Input:\n")
	kwu.G.Chat.Trace("...GE2NNGS... Server.Host=%s\n", tomlTree.Get("Server.Host").(string))
	kwu.G.Chat.Trace("...GE2NNGS... Server.Port=%d\n", tomlTree.Get("Server.Port").(int64))
	kwu.G.Chat.Trace("...GE2NNGS... User.InterfaceType=%s\n", tomlTree.Get("User.InterfaceType").(string))
	kwu.G.Chat.Trace("...GE2NNGS... User.EngineCommand=%s\n", tomlTree.Get("User.EngineCommand").(string))
	kwu.G.Chat.Trace("...GE2NNGS... User.EngineCommandOption=%s\n", tomlTree.Get("User.EngineCommandOption").(string))
	kwu.G.Chat.Trace("...GE2NNGS... MatchApplication.Phase=%s\n", tomlTree.Get("MatchApplication.Phase").(string))
	kwu.G.Chat.Trace("...GE2NNGS... MatchApplication.BoardSize=%d\n", tomlTree.Get("MatchApplication.BoardSize").(int64))
	kwu.G.Chat.Trace("...GE2NNGS... MatchApplication.AvailableTimeMinutes=%d\n", tomlTree.Get("MatchApplication.AvailableTimeMinutes").(int64))
	kwu.G.Chat.Trace("...GE2NNGS... MatchApplication.CanadianTiming=%d\n", tomlTree.Get("MatchApplication.CanadianTiming").(int64))
}

func debugPrintConfig(config e.EntryConf) {
	kwu.G.Chat.Trace("...GE2NNGS... Memory:\n")
	kwu.G.Chat.Trace("...GE2NNGS... Server.Host=%s\n", config.Server.Host)
	kwu.G.Chat.Trace("...GE2NNGS... Server.Port=%d\n", config.Server.Port)
	kwu.G.Chat.Trace("...GE2NNGS... User.InterfaceType=%s\n", config.User.InterfaceType)
	kwu.G.Chat.Trace("...GE2NNGS... User.EngineCommand=%s\n", config.User.EngineCommand)
	kwu.G.Chat.Trace("...GE2NNGS... User.EngineCommandOption=%s\n", config.User.EngineCommandOption)
	kwu.G.Chat.Trace("...GE2NNGS... MatchApplication.Phase=%s\n", config.MatchApplication.Phase)
	kwu.G.Chat.Trace("...GE2NNGS... MatchApplication.BoardSize=%d\n", config.MatchApplication.BoardSize)
	kwu.G.Chat.Trace("...GE2NNGS... MatchApplication.AvailableTimeMinutes=%d\n", config.MatchApplication.AvailableTimeMinutes)
	kwu.G.Chat.Trace("...GE2NNGS... MatchApplication.CanadianTiming=%d\n", config.MatchApplication.CanadianTiming)
}
