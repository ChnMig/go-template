package log

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

var (
	monitorLifecycleMu sync.Mutex
	activeMonitor      *monitorRun
)

type monitorRun struct {
	done chan struct{}
	wg   sync.WaitGroup
}

// StartMonitor starts the release-mode file-reopen and daily-rotation loops.
func StartMonitor() {
	value, ok := settingsSnapshot()
	if !ok || value.development {
		return
	}
	monitorLifecycleMu.Lock()
	defer monitorLifecycleMu.Unlock()
	if activeMonitor != nil {
		return
	}
	run := &monitorRun{done: make(chan struct{})}
	run.wg.Add(2)
	activeMonitor = run
	go func() {
		defer run.wg.Done()
		monitorFile(run.done, value.logDir)
	}()
	go func() {
		defer run.wg.Done()
		rotateDaily(run.done)
	}()
}

// StopMonitor keeps the example-project cleanup entry point.
func StopMonitor() error {
	return Close()
}

// Close stops log goroutines, flushes the zap cores, and closes file writers.
func Close() error {
	monitorLifecycleMu.Lock()
	if activeMonitor != nil {
		close(activeMonitor.done)
		activeMonitor.wg.Wait()
		activeMonitor = nil
	}
	monitorLifecycleMu.Unlock()
	return closeLoggers()
}

func monitorFile(done <-chan struct{}, directory string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		zap.L().Error("File listening error", zap.Error(err))
		return
	}
	defer func() {
		if closeErr := watcher.Close(); closeErr != nil {
			zap.L().Warn("close log file watcher failed", zap.Error(closeErr))
		}
	}()
	if err := watcher.Add(directory); err != nil {
		zap.L().Error("File listening error", zap.Error(err))
		return
	}
	var reopenTimer *time.Timer
	var reopen <-chan time.Time
	pendingPaths := make(map[string]struct{})
	defer func() {
		if reopenTimer != nil {
			reopenTimer.Stop()
		}
	}()
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if !event.Has(fsnotify.Remove) && !event.Has(fsnotify.Rename) || !isManagedLogPath(event.Name) {
				continue
			}
			pendingPaths[event.Name] = struct{}{}
			if reopenTimer == nil {
				reopenTimer = time.NewTimer(300 * time.Millisecond)
			} else {
				stopAndDrainTimer(reopenTimer)
				reopenTimer.Reset(300 * time.Millisecond)
			}
			reopen = reopenTimer.C
		case <-reopen:
			for path := range pendingPaths {
				if _, statErr := os.Stat(path); statErr != nil {
					zap.L().Warn("log file missing, reopening logger", zap.String("path", path))
					if resetErr := resetLogger(); resetErr != nil {
						zap.L().Error("reopen logger failed", zap.Error(resetErr))
					}
					break
				}
			}
			clear(pendingPaths)
			reopen = nil
		case watcherErr, ok := <-watcher.Errors:
			if !ok {
				return
			}
			zap.L().Error("file listening error", zap.Error(watcherErr))
		case <-done:
			return
		}
	}
}

func stopAndDrainTimer(timer *time.Timer) {
	if timer.Stop() {
		return
	}
	select {
	case <-timer.C:
	default:
	}
}

func rotateDaily(done <-chan struct{}) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		timer := time.NewTimer(time.Until(next))
		select {
		case <-timer.C:
			rotateAll()
		case <-done:
			stopAndDrainTimer(timer)
			return
		}
	}
}

func rotateAll() {
	loggerMu.RLock()
	businessCore := loggerCore
	accessCore := ginLoggerCore
	loggerMu.RUnlock()
	if businessCore != nil {
		if err := businessCore.rotate(); err != nil {
			zap.L().Warn("rotate business log failed", zap.Error(err))
		}
	}
	if accessCore != nil {
		if err := accessCore.rotate(); err != nil {
			zap.L().Warn("rotate Gin log failed", zap.Error(err))
		}
	}
}

func resetLogger() error {
	value, ok := settingsSnapshot()
	if !ok {
		return nil
	}
	return applySettings(value)
}

func settingsSnapshot() (settings, bool) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	if currentSettings == nil {
		return settings{}, false
	}
	return *currentSettings, true
}

func isManagedLogPath(path string) bool {
	value, ok := settingsSnapshot()
	return ok && (filepath.Clean(path) == filepath.Clean(value.logPath) ||
		filepath.Clean(path) == filepath.Clean(value.ginLogPath))
}

func closeLoggers() error {
	configureMu.Lock()
	defer configureMu.Unlock()
	loggerMu.Lock()
	businessCore := loggerCore
	accessCore := ginLoggerCore
	logger = nil
	loggerCore = nil
	ginLogger = nil
	ginErrorLogger = nil
	ginLoggerCore = nil
	currentSettings = nil
	zap.ReplaceGlobals(zap.NewNop())
	loggerMu.Unlock()
	return errors.Join(closeCore(businessCore), closeCore(accessCore))
}

func closeCore(core *managedCore) error {
	if core == nil {
		return nil
	}
	syncErr := core.Sync()
	if errors.Is(syncErr, syscall.EINVAL) || errors.Is(syncErr, syscall.EBADF) {
		syncErr = nil
	}
	return errors.Join(syncErr, core.close())
}
