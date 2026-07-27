package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type envBinding struct {
	key string
	env string
}

// LogWatcher owns one configuration-directory watcher and its goroutine.
type LogWatcher struct {
	watcher    *fsnotify.Watcher
	configFile string
	realFile   string
	onChange   func(LogConfig) error
	done       chan struct{}
	stopped    chan struct{}
	stopOnce   sync.Once
}

// WatchLogConfig watches only the source file's log section.
func WatchLogConfig(initial LogConfig, onChange func(LogConfig) error) (*LogWatcher, error) {
	running := &LogWatcher{
		onChange: onChange, done: make(chan struct{}), stopped: make(chan struct{}),
	}
	path := strings.TrimSpace(initial.sourcePath)
	if path == "" {
		close(running.stopped)
		return running, nil
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config file: %w", err)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create config watcher: %w", err)
	}
	if err := watcher.Add(filepath.Dir(absolutePath)); err != nil {
		return nil, errors.Join(fmt.Errorf("watch config directory: %w", err), watcher.Close())
	}
	running.watcher = watcher
	running.configFile = filepath.Clean(absolutePath)
	running.realFile, _ = filepath.EvalSymlinks(absolutePath)
	go running.run()
	return running, nil
}

// Close stops and joins the watcher goroutine.
func (running *LogWatcher) Close() error {
	if running == nil {
		return nil
	}
	running.stopOnce.Do(func() { close(running.done) })
	<-running.stopped
	return nil
}

func (running *LogWatcher) run() {
	defer close(running.stopped)
	defer func() {
		if err := running.watcher.Close(); err != nil {
			zap.L().Warn("close config watcher failed", zap.Error(err))
		}
	}()
	for {
		select {
		case event, ok := <-running.watcher.Events:
			if !ok {
				return
			}
			if running.shouldReload(event) {
				if err := running.reload(); err != nil {
					zap.L().Error("reload log configuration failed", zap.Error(err))
				}
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

func (running *LogWatcher) shouldReload(event fsnotify.Event) bool {
	currentRealFile, _ := filepath.EvalSymlinks(running.configFile)
	realFileChanged := currentRealFile != "" && currentRealFile != running.realFile
	if realFileChanged {
		running.realFile = currentRealFile
	}
	sameFile := filepath.Clean(event.Name) == running.configFile
	return realFileChanged || sameFile && (event.Has(fsnotify.Write) || event.Has(fsnotify.Create))
}

func (running *LogWatcher) reload() error {
	next, err := readLogConfig(running.configFile)
	if err != nil {
		return err
	}
	if running.onChange != nil {
		if err := running.onChange(next); err != nil {
			return fmt.Errorf("apply log configuration: %w", err)
		}
	}
	return nil
}

func readLogConfig(path string) (LogConfig, error) {
	loader := viper.New()
	loader.SetConfigFile(path)
	loader.SetDefault("log.level", "info")
	loader.SetDefault("log.gin_level", "")
	loader.SetDefault("log.max_size", 50)
	loader.SetDefault("log.max_age", 30)
	loader.SetEnvPrefix(environmentPrefix)
	loader.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	for _, binding := range []envBinding{
		{"log.level", "LOG_LEVEL"},
		{"log.gin_level", "LOG_GIN_LEVEL"},
		{"log.max_size", "LOG_MAX_SIZE"},
		{"log.max_age", "LOG_MAX_AGE"},
	} {
		if err := loader.BindEnv(binding.key, environmentPrefix+"_"+binding.env); err != nil {
			return LogConfig{}, fmt.Errorf("bind log environment: %w", err)
		}
	}
	if err := loader.ReadInConfig(); err != nil {
		return LogConfig{}, fmt.Errorf("read log configuration: %w", err)
	}
	var next LogConfig
	if err := loader.UnmarshalKey("log", &next); err != nil {
		return LogConfig{}, fmt.Errorf("decode log configuration: %w", err)
	}
	next.sourcePath = path
	if err := next.validate(); err != nil {
		return LogConfig{}, err
	}
	return next, nil
}
