package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type ContextConfigKey string

// Feel free to override these variables in your application.
// config.ConfigKey = "MYTeamConfigKey"
var (
	ConfigKey ContextConfigKey = "ConfigKey"
)

// LoadConfig load configuration by searching yaml files in the given path.
// configPath must be a FOLDER path.
func LoadConfig[T any](configPath string) (config *T, err error) {
	if configPath == "" {
		zap.S().Info("no config path provided")
	}

	viper.AddConfigPath(configPath)
	files, _ := os.ReadDir(configPath)
	for _, file := range files {
		fileName := file.Name()
		extFile := filepath.Ext(file.Name())
		if extFile != ".yaml" && extFile != ".yml" {
			zap.S().Infow("File not in a yaml format, will be ignored", "filename", fileName)
			continue
		}
		lastDotIndex := strings.LastIndex(fileName, ".")
		viper.SetConfigName(fileName[:lastDotIndex])
		viper.SetConfigType("yaml")
		err = viper.MergeInConfig()
		if err != nil {
			return
		}
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	config = new(T)
	err = viper.Unmarshal(config)
	return config, errors.Wrap(err, "could not unmarshal config, check that you provided a valid yaml file")
}

// Watch config changes
func WatchConfig[T any](onConfigChange func(T)) {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		var cfg T
		viper.Unmarshal(&cfg)
		onConfigChange(cfg)
	})
}

func ConfigFromContextOrDiscard[T any](ctx context.Context) *T {
	if v, ok := ctx.Value(ConfigKey).(*T); ok {
		return v
	}
	return new(T)
}

func CreateContextWithConfig[T any](ctx context.Context, cfg *T) context.Context {
	return context.WithValue(ctx, ConfigKey, cfg)
}
