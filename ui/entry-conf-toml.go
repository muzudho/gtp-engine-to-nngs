package ui

import (
	"fmt"
	"io/ioutil"

	e "github.com/muzudho/gtp-engine-to-nngs/entities"
	u "github.com/muzudho/gtp-engine-to-nngs/usecases"
	"github.com/pelletier/go-toml"
)

// LoadEntryConf - Toml形式の参加設定ファイルを読み込みます。
func LoadEntryConf(path string) e.EntryConf {

	// ファイル読込
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		u.G.Chat.Fatal("path=%s", path)
		panic(err)
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
	// fmt.Printf("<GE2NNGS> content=%s", string(fileData))

	// Toml解析
	tomlTree, err := toml.Load(string(fileData))
	if err != nil {
		panic(err)
	}
	fmt.Println("<GE2NNGS> Input:")
	fmt.Printf("<GE2NNGS> Server.Host=%s\n", tomlTree.Get("Server.Host").(string))
	fmt.Printf("<GE2NNGS> Server.Port=%d\n", tomlTree.Get("Server.Port").(int64))
	fmt.Printf("<GE2NNGS> User.Name=%s\n", tomlTree.Get("User.Name").(string))
	fmt.Printf("<GE2NNGS> User.Pass=%s\n", tomlTree.Get("User.Pass").(string))
	fmt.Printf("<GE2NNGS> User.InterfaceType=%s\n", tomlTree.Get("User.InterfaceType").(string))
	fmt.Printf("<GE2NNGS> User.EngineCommand=%s\n", tomlTree.Get("User.EngineCommand").(string))
	fmt.Printf("<GE2NNGS> User.EngineCommandOption=%s\n", tomlTree.Get("User.EngineCommandOption").(string))
	fmt.Printf("<GE2NNGS> MatchApplication.Phase=%s\n", tomlTree.Get("MatchApplication.Phase").(string))
	fmt.Printf("<GE2NNGS> MatchApplication.BoardSize=%d\n", tomlTree.Get("MatchApplication.BoardSize").(int64))
	fmt.Printf("<GE2NNGS> MatchApplication.AvailableTimeMinutes=%d\n", tomlTree.Get("MatchApplication.AvailableTimeMinutes").(int64))
	fmt.Printf("<GE2NNGS> MatchApplication.CanadianTiming=%d\n", tomlTree.Get("MatchApplication.CanadianTiming").(int64))
}

func debugPrintConfig(config e.EntryConf) {
	fmt.Println("<GE2NNGS> Memory:")
	fmt.Printf("<GE2NNGS> Server.Host=%s\n", config.Server.Host)
	fmt.Printf("<GE2NNGS> Server.Port=%d\n", config.Server.Port)
	fmt.Printf("<GE2NNGS> User.Name=%s\n", config.User.Name)
	fmt.Printf("<GE2NNGS> User.Pass=%s\n", config.User.Pass)
	fmt.Printf("<GE2NNGS> User.InterfaceType=%s\n", config.User.InterfaceType)
	fmt.Printf("<GE2NNGS> User.EngineCommand=%s\n", config.User.EngineCommand)
	fmt.Printf("<GE2NNGS> User.EngineCommandOption=%s\n", config.User.EngineCommandOption)
	fmt.Printf("<GE2NNGS> MatchApplication.Phase=%s\n", config.MatchApplication.Phase)
	fmt.Printf("<GE2NNGS> MatchApplication.BoardSize=%d\n", config.MatchApplication.BoardSize)
	fmt.Printf("<GE2NNGS> MatchApplication.AvailableTimeMinutes=%d\n", config.MatchApplication.AvailableTimeMinutes)
	fmt.Printf("<GE2NNGS> MatchApplication.CanadianTiming=%d\n", config.MatchApplication.CanadianTiming)
}
