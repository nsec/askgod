package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/goccy/go-yaml"
	"github.com/inconshreveable/log15"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

// Config represents the internal view of the configuration.
type Config struct {
	*api.Config
	logger   log15.Logger
	handlers []func(*Config)
}

// RegisterHandler makes it possible to register a function to be called on config changes.
func (c *Config) RegisterHandler(handler func(*Config)) error {
	c.handlers = append(c.handlers, handler)

	return nil
}

func parseConfig(configPath string, conf any) error {
	// Read the file's content
	content, err := os.ReadFile(configPath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	// Parse the yaml file
	err = yaml.Unmarshal(content, conf)
	if err != nil {
		return fmt.Errorf("failed to parse yaml: %w", err)
	}

	return nil
}

// ReadConfigFile will return a Config struct from the content of a yaml file.
func ReadConfigFile(configPath string, monitor bool, logger log15.Logger) (*Config, error) {
	if !utils.PathExists(configPath) {
		return nil, fmt.Errorf("the configuration file doesn't exist: %s", configPath)
	}

	logger.Info("Parsing configuration", log15.Ctx{"path": configPath})

	conf := Config{logger: logger}
	err := parseConfig(configPath, &conf.Config)
	if err != nil {
		return nil, err
	}

	// Watch for configuration changes
	if monitor {
		logger.Info("Setting up configuration watch", log15.Ctx{"path": configPath})

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, fmt.Errorf("unable to setup fsnotify: %w", err)
		}

		err = watcher.Add(filepath.Dir(configPath))
		if err != nil {
			return nil, fmt.Errorf("unable to setup fsnotify watch: %w", err)
		}

		pathDir := filepath.Dir(configPath)
		if pathDir == "" {
			pathDir = "./"
		}
		pathBase := filepath.Base(configPath)

		go func() {
			for {
				select {
				case ev := <-watcher.Events:
					if ev.Name != fmt.Sprintf("%s/%s", pathDir, pathBase) {
						continue
					}

					// Store the old config for comparison
					oldData, _ := yaml.Marshal(conf.Config)

					// Wait for 1s for ownership changes
					time.Sleep(time.Second)

					// Parse the new ocnfig
					err := parseConfig(configPath, conf.Config)
					if err != nil {
						logger.Error("failed to read the new configuration", log15.Ctx{"path": configPath, "error": err})
					}

					// Check if something changed
					newData, _ := yaml.Marshal(conf.Config)
					if string(oldData) == string(newData) {
						continue
					}

					logger.Info("Configuration file changed, reloading", log15.Ctx{"path": configPath})
					for _, handler := range conf.handlers {
						handler(&conf)
					}
				case err := <-watcher.Errors:
					logger.Error("got bad file notification", log15.Ctx{"error": err})
				}
			}
		}()
	}

	return &conf, nil
}
