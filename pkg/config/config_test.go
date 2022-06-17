package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/athosone/golib/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type TestConfig struct {
	Value string `yaml:"value"`
}

var _ = Describe("Load config", Label("Unit"), func() {
	var (
		configTestPath = ""
		configTestFile = ""
		configFullPath = ""
		srcCfg         = TestConfig{
			Value: "test",
		}
	)
	Context("Loading config from file", func() {
		BeforeEach(func() {
			configTestFile = utils.GenerateRandomNameWithPrefix("config") + ".yaml"
			configTestPath = utils.GenerateRandomNameWithPrefix("configpath")
			configFullPath = fmt.Sprintf("%s/%s", configTestPath, configTestFile)

			By(fmt.Sprintf("Creating a dummy config file at %s", configFullPath))
			err := os.MkdirAll(configTestPath, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
			file, err := os.Create(configFullPath)
			defer file.Close()
			Expect(err).NotTo(HaveOccurred())

			By("Writing dummy config to file")
			data, err := yaml.Marshal(srcCfg)
			Expect(err).NotTo(HaveOccurred())

			_, err = file.Write(data)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(configTestPath)).To(Succeed())
			viper.Reset()
		})

		It("Should load the config fields properly", func() {
			cfg, err := LoadConfig[TestConfig](configTestPath)
			Expect(err).NotTo(HaveOccurred())

			By("Checking loaded config match config src")
			Expect(*cfg).To(Equal(srcCfg))
		})

		When("Config file changes during runtime", func() {
			var config *TestConfig
			newValue := "new"
			mu := sync.RWMutex{}

			BeforeEach(func() {
				config, _ = LoadConfig[TestConfig](configTestPath)
				WatchConfig(func(tc TestConfig) {
					mu.Lock()
					config = &tc
					mu.Unlock()
				})
				file, _ := os.OpenFile(configFullPath, os.O_RDWR, 0666)
				defer file.Close()
				srcCfg.Value = newValue
				data, _ := yaml.Marshal(srcCfg)
				_, err := file.Write(data)
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should hot-reload the config fields properly", func() {
				Eventually(func() string {
					mu.RLock()
					defer mu.RUnlock()
					return config.Value
				}, 1, 0.1).Should(Equal(newValue))
			})
		})

		When("Environment variable is set", func() {
			expectedValue := "test-env"

			BeforeEach(func() {
				_ = viper.BindEnv("value", "AN_ENV_VAR")
				os.Setenv("AN_ENV_VAR", expectedValue)
			})
			AfterEach(func() { os.Unsetenv("AN_ENV_VAR") })

			It("Loads the config with value overridden by env variables", func() {
				cfg, err := LoadConfig[TestConfig](configTestPath)
				Expect(err).NotTo(HaveOccurred())

				By("Checking loaded config match config src")
				Expect(cfg.Value).To(Equal(expectedValue))
			})
		})
	})

	Context("Loading config from env variables without config file", func() {
		expectedValue := "test-env"

		BeforeEach(func() {
			_ = viper.BindEnv("value", "AN_ENV_VAR")
			os.Setenv("AN_ENV_VAR", expectedValue)
		})
		AfterEach(func() { os.Unsetenv("AN_ENV_VAR") })

		It("Should load the config fields properly", func() {
			cfg, err := LoadConfig[TestConfig](configTestPath)
			Expect(err).NotTo(HaveOccurred())

			By("Checking loaded config match config src")
			Expect(cfg.Value).To(Equal(expectedValue))
		})
	})
})
