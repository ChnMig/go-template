package log

import (
	"http-services/utils/apperrors"

	"go.uber.org/zap/zapcore"
)

// cancellationCore 在统一输出边界过滤仅由请求取消导致的 Warn/Error，
// 包含 Logger.With 绑定的错误；放在 sampler 内侧以保留采样和日志阈值。
type cancellationCore struct {
	zapcore.Core
	hasCancellation bool
	hasOtherError   bool
}

func newCancellationCore(core zapcore.Core) zapcore.Core {
	return &cancellationCore{Core: core}
}

func (c *cancellationCore) With(fields []zapcore.Field) zapcore.Core {
	clone := *c
	clone.Core = c.Core.With(fields)
	clone.inspectErrors(fields)
	return &clone
}

func (c *cancellationCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		// 登记当前 core，确保写入阶段会执行取消判断。
		return checked.AddCore(entry, c)
	}
	return checked
}

func (c *cancellationCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if entry.Level == zapcore.WarnLevel || entry.Level == zapcore.ErrorLevel {
		state := *c
		state.inspectErrors(fields)
		if state.hasCancellation && !state.hasOtherError {
			return nil
		}
	}
	return c.Core.Write(entry, fields)
}

func (c *cancellationCore) inspectErrors(fields []zapcore.Field) {
	for _, field := range fields {
		if field.Type != zapcore.ErrorType {
			continue
		}
		err, ok := field.Interface.(error)
		if !ok || err == nil {
			continue
		}
		if isCancellationLogError(err) {
			c.hasCancellation = true
		} else {
			c.hasOtherError = true
		}
	}
}

// 第三方错误的 Is/Unwrap 可能 panic；分类失败时保留日志并沿用 Zap 原错误编码。
func isCancellationLogError(err error) (canceled bool) {
	defer func() {
		if recover() != nil {
			canceled = false
		}
	}()
	return apperrors.IsClientCancellation(err)
}
