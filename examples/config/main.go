package main

import (
	"fmt"
	"os"

	"github.com/athosone/golib/pkg/config"
	"github.com/spf13/viper"
)

type ExampleConfig struct {
	ServicePrincipal struct {
		ClientId     string `yaml:"clientId"`
		ClientSecret string `yaml:"clientSecret"`
		TenantId     string `yaml:"tenantId"`
	} `yaml:"servicePrincipal"`

	IsDebug bool `yaml:"isDebug"`
}

func LoadConfig() (*ExampleConfig, error) {
	_ = viper.BindEnv("servicePrincipal.clientId", "AZURE_CLIENT_ID")
	return config.LoadConfig[ExampleConfig](".")
}

func main() {
	os.Setenv("AZURE_CLIENT_ID", "overridden in env")
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", cfg)
}
