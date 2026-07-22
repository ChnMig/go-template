package log

import (
	"io"
	"sync"

	"go.uber.org/zap/zapcore"
)

// managedCore 让已经派生出的 logger 在底层 writer 切换后继续使用最新 core。
// 替换操作与 Write 互斥，因此旧 writer 只会在所有进行中的写入结束后关闭。
type managedCore struct {
	state  *managedCoreState
	fields []zapcore.Field
}

type managedCoreState struct {
	mu     sync.RWMutex
	core   zapcore.Core
	closer io.Closer
}

func newManagedCore() *managedCore {
	return &managedCore{
		state: &managedCoreState{core: zapcore.NewNopCore()},
	}
}

func (core *managedCore) Enabled(level zapcore.Level) bool {
	core.state.mu.RLock()
	defer core.state.mu.RUnlock()
	return core.state.core.Enabled(level)
}

func (core *managedCore) With(fields []zapcore.Field) zapcore.Core {
	combined := make([]zapcore.Field, 0, len(core.fields)+len(fields))
	combined = append(combined, core.fields...)
	combined = append(combined, fields...)
	return &managedCore{state: core.state, fields: combined}
}

func (core *managedCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if core.Enabled(entry.Level) {
		return checked.AddCore(entry, core)
	}
	return checked
}

func (core *managedCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	core.state.mu.RLock()
	defer core.state.mu.RUnlock()
	return core.state.core.With(core.fields).Write(entry, fields)
}

func (core *managedCore) Sync() error {
	core.state.mu.RLock()
	defer core.state.mu.RUnlock()
	return core.state.core.Sync()
}

func (core *managedCore) replace(next zapcore.Core, closer io.Closer) error {
	if next == nil {
		next = zapcore.NewNopCore()
	}

	core.state.mu.Lock()
	defer core.state.mu.Unlock()
	previousCloser := core.state.closer
	core.state.core = next
	core.state.closer = closer
	if previousCloser != nil {
		return previousCloser.Close()
	}
	return nil
}

func (core *managedCore) close() error {
	return core.replace(zapcore.NewNopCore(), nil)
}

func (core *managedCore) rotate() error {
	core.state.mu.RLock()
	defer core.state.mu.RUnlock()
	rotator, ok := core.state.closer.(interface{ Rotate() error })
	if !ok {
		return nil
	}
	return rotator.Rotate()
}
