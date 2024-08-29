package config_test

import (
	"os"
)

func ExampleReadConfig() {
	// this example ignores error handling

	os.Setenv("EX_SOMENUMBER", "999")

	//// Load config from sample.yaml and set ENV variable refix to EX
	//cfg, _ := config.Load("./sampledata/sample.yaml", "EX")
	//
	//// read individual values
	//n := cfg.Viper().GetInt("SomeNumber")
	//fmt.Println(n)
	//
	//// unmarshal a struct
	//type cfgData struct {
	//	Toplevel   string `mapstructure:"Toplevel"`
	//	SomeNumber int    `mapstructure:"SomeNumber" validate:"required"`
	//}
	//data := cfgData{}
	//_ = cfg.Unmarshal(&data)
	//fmt.Println(data.Toplevel)
	//
	//// Output: 999
	//// banana
}
