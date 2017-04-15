package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/fsnotify.v0"
	"gopkg.in/inconshreveable/log15.v2"
	"gopkg.in/yaml.v2"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

// Config represents the internal view of the configuration
type Config struct {
	*api.Config
	logger log15.Logger
}

func parseConfig(configPath string, conf interface{}) error {
	// Read the file's content
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("Failed to read file content: %v", err)
	}

	// Parse the yaml file
	err = yaml.Unmarshal(content, conf)
	if err != nil {
		return fmt.Errorf("Failed to parse yaml: %v", err)
	}

	return nil
}

// ReadConfigFile will return a Config struct from the content of a yaml file
func ReadConfigFile(configPath string, monitor bool, logger log15.Logger) (*Config, error) {
	if !utils.PathExists(configPath) {
		return nil, fmt.Errorf("The configuration file doesn't exist: %s", configPath)
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
			return nil, fmt.Errorf("Unable to setup fsnotify: %v", err)
		}

		err = watcher.Watch(filepath.Dir(configPath))
		if err != nil {
			return nil, fmt.Errorf("Unable to setup fsnotify watch: %v", err)
		}

		pathDir := filepath.Dir(configPath)
		if pathDir == "" {
			pathDir = "./"
		}
		pathBase := filepath.Base(configPath)

		go func() {
			for {
				select {
				case ev := <-watcher.Event:
					if ev.Name != fmt.Sprintf("%s/%s", pathDir, pathBase) {
						continue
					}

					if !ev.IsModify() || ev.IsAttrib() {
						continue
					}

					logger.Info("Configuration file changed, reloading", log15.Ctx{"path": configPath})
					err := parseConfig(configPath, conf.Config)
					if err != nil {
						logger.Error("Failed to read the new configuration", log15.Ctx{"path": configPath, "error": err})
					}
				case err := <-watcher.Error:
					logger.Error("Got bad file notification", log15.Ctx{"error": err})
				}
			}
		}()
	}

	return &conf, nil
}
