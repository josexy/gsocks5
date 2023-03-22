package util

import "github.com/josexy/logx"

var Logger logx.Logger

func init() {
	Logger = logx.NewDevelopment(
		logx.WithLevel(true, true),
		logx.WithColor(true),
		logx.WithCaller(true, true, false, false),
		logx.WithSimpleEncoder(),
		logx.WithTime(true, nil),
	)
}
