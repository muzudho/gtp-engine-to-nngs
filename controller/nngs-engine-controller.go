package controller

import (
	e "github.com/muzudho/gtp-engine-to-nngs/entities"
)

// NngsEngineController - NNGS からの受信メッセージをさばきます。
type NngsEngineController struct {
	// ConnectorConf - 参加設定
	ConnectorConf e.ConnectorConf
}
