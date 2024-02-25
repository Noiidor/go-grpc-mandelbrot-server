package config

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type ViperConfig struct {
	*viper.Viper
}

func NewViperConfig(pathToConfigFile string) Config {
	vc := &ViperConfig{viper.New()}

	vc.loadConfig(pathToConfigFile)

	return vc
}

func (vc ViperConfig) loadConfig(pathToConfigFile string) {
	filename := filepath.Base(pathToConfigFile)

	vc.SetConfigName(filepath.Base(filename))
	vc.SetConfigType(strings.Split(filename, ".")[1])
	vc.AddConfigPath(filepath.Dir(pathToConfigFile))

	if err := vc.ReadInConfig(); err != nil {
		log.Fatalf("error while load config: %v\n", err)
	}
}

func (vc ViperConfig) GetFromGRPC(key string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", fmt.Errorf("key cannot be an empty\n")
	}

	value := vc.GetString(fmt.Sprintf("GRPC.%v", key))
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf(
			"an empty value returned, perhaps it's not defined or is missing\nKey: %s\n",
			key,
		)
	}

	return value, nil
}
