package logx

var DiscardLogger = &discardLogger{}

type discardLogger struct{}

func (l discardLogger) Debug(string, ...any) {}
func (l discardLogger) Info(string, ...any)  {}
func (l discardLogger) Warn(string, ...any)  {}
func (l discardLogger) Error(string, ...any) {}
func (l discardLogger) Fatal(string, ...any) {}
