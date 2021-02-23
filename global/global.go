package global

import (
	kwu "github.com/muzudho/kifuwarabe-gtp/usecases"
)

// GlobalVariables - グローバル変数。
type GlobalVariables struct {
	// Log - ロガー。
	Log kwu.Logger
	// Chat - チャッター。 標準出力とロガーを一緒にしただけです。
	Chat kwu.Chatter
	// StderrChat - チャッター。 標準エラー出力とロガーを一緒にしただけです。
	StderrChat kwu.StderrChatter
}

// G - グローバル変数。思い切った名前。
var G GlobalVariables
