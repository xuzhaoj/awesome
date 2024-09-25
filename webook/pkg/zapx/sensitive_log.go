package zapx

import "go.uber.org/zap/zapcore"

type MyCore struct {
	zapcore.Core
}

func (c MyCore) Write(entry zapcore.Entry, fds []zapcore.Field) error {
	for _, fd := range fds {
		if fd.Key == "phone" {
			phone := fd.String
			fd.String = phone[:3] + "***" + phone[7:]
		}

	}
	return c.Core.Write(entry, fds)
}
