package config

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

var (
	watchConfigLifecycleMu sync.Mutex
	watchConfigMu          sync.Mutex
	activeConfigWatcher    *configWatcher
)

type configWatcher struct {
	watcher    *fsnotify.Watcher
	configFile string
	realFile   string
	onChange   func()
	done       chan struct{}
	stopped    chan struct{}
	stopOnce   sync.Once
}

// WatchConfig 启动一个可停止、可等待的配置 watcher；只有日志配置会被热更新。
func WatchConfig(onChange func()) error {
	watchConfigLifecycleMu.Lock()
	defer watchConfigLifecycleMu.Unlock()
	stopWatchConfigLocked()

	if v == nil || v.ConfigFileUsed() == "" {
		return nil
	}
	configFile, err := filepath.Abs(v.ConfigFileUsed())
	if err != nil {
		return fmt.Errorf("resolve config file: %w", err)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create config watcher: %w", err)
	}
	if err := watcher.Add(filepath.Dir(configFile)); err != nil {
		_ = watcher.Close()
		return fmt.Errorf("watch config directory: %w", err)
	}
	realFile, _ := filepath.EvalSymlinks(configFile)
	running := &configWatcher{
		watcher:    watcher,
		configFile: filepath.Clean(configFile),
		realFile:   realFile,
		onChange:   onChange,
		done:       make(chan struct{}),
		stopped:    make(chan struct{}),
	}
	go running.run()
	watchConfigMu.Lock()
	activeConfigWatcher = running
	watchConfigMu.Unlock()
	return nil
}

// StopWatchConfig 停止 watcher，并等待正在执行的配置回调退出。
func StopWatchConfig() {
	watchConfigLifecycleMu.Lock()
	defer watchConfigLifecycleMu.Unlock()
	stopWatchConfigLocked()
}

func stopWatchConfigLocked() {
	watchConfigMu.Lock()
	running := activeConfigWatcher
	activeConfigWatcher = nil
	watchConfigMu.Unlock()
	if running != nil {
		running.stop()
	}
}

func (running *configWatcher) stop() {
	running.stopOnce.Do(func() { close(running.done) })
	<-running.stopped
}

func (running *configWatcher) run() {
	defer close(running.stopped)
	defer running.watcher.Close()
	for {
		select {
		case event, ok := <-running.watcher.Events:
			if !ok {
				return
			}
			if running.shouldReload(event) {
				running.reload(event)
			}
		case err, ok := <-running.watcher.Errors:
			if !ok {
				return
			}
			zap.L().Error("config watcher error", zap.Error(err))
		case <-running.done:
			return
		}
	}
}

func (running *configWatcher) shouldReload(event fsnotify.Event) bool {
	currentRealFile, _ := filepath.EvalSymlinks(running.configFile)
	realFileChanged := currentRealFile != "" && currentRealFile != running.realFile
	if realFileChanged {
		running.realFile = currentRealFile
	}
	sameFile := filepath.Clean(event.Name) == running.configFile
	return realFileChanged || (sameFile && (event.Has(fsnotify.Write) || event.Has(fsnotify.Create)))
}

func (running *configWatcher) reload(event fsnotify.Event) {
	zap.L().Info("Config file changed, reloading...",
		zap.String("file", event.Name),
		zap.String("op", event.Op.String()),
	)
	if err := v.ReadInConfig(); err != nil {
		zap.L().Error("Config reload failed", zap.Error(err))
		return
	}
	applyReloadableLogConfig()
	if running.onChange != nil {
		running.onChange()
	}
	zap.L().Info("Config reloaded successfully")
}
