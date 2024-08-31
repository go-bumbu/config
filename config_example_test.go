package config_test

import (
	"fmt"
	"github.com/go-bumbu/config"
	"os"
)

type sampleConfig struct {
	Sub      submodule
	Number   int
	FromFile string `config:"from_file"`
	FromEnv  string `config:"EnvVarConfig"`
}

type submodule struct {
	Name  string
	Value float64
}

var defaultCfg = sampleConfig{
	Number: 42,
}

// this is a full example of all features
func ExampleConfig() {
	// this example ignores error handling

	_ = os.Setenv("ENVPREFIX_SUB_NAME", "Superman")

	cfg := sampleConfig{}
	_, _ = config.Load(
		config.Defaults{Item: defaultCfg},                                   // use default values
		config.Unmarshal{Item: &cfg},                                        // marshal result into cfg
		config.CfgFile{Path: "sampledata/example_test/example.config.json"}, // load config from file
		config.EnvVar{Prefix: "ENVPREFIX"},                                  // load a config value from an env
	)

	// print the output
	fmt.Println(cfg.Number)   // using default value
	fmt.Println(cfg.FromFile) // from config file
	fmt.Println(cfg.Sub.Name) // loaded from env var

	// Output:
	// 42
	// Spiderman
	// Superman

}
