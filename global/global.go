package global

import (
	l "github.com/muzudho/go-logger"
)

// Variables - グローバル変数。
type Variables struct {
	// Log - ロガー。
	Log l.Logger
	// Chat - チャッター。 標準出力とロガーを一緒にしただけです。
	Chat l.Chatter
	// StderrChat - チャッター。 標準エラー出力とロガーを一緒にしただけです。
	StderrChat l.StderrChatter
}

// G - グローバル変数。思い切った名前。
var G Variables
