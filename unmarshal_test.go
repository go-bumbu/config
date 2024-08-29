package config

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

type testConfig struct {
	Number      int          `config:"number"`
	FloatNum    float64      `config:"floatNum"`
	Text        string       `config:"text"`
	FileContent string       `config:"fileContent"`
	Bol         bool         `config:"bol"`
	StringList  []string     `config:"listString"`
	StructList  []userData   `config:"userList"`
	Nested      NestedConfig `config:"nested"`
}
type NestedConfig struct {
	Child  Child `config:"child"`
	Child2 struct {
		Number int `config:"number"`
	} `config:"child_2"`
}
type userData struct {
	Name string `config:"name"`
	Pass string `config:"pass"`
}

type Child struct {
	Number      int    `config:"number"`
	Text        string `config:"text"`
	AnotherName string `config:"renamed"`
}

var DefaultCfg = testConfig{
	Number:     100,
	FloatNum:   100.1,
	Text:       "default text",
	Bol:        false,
	StringList: []string{"default 1", "default 2"},
	StructList: []userData{
		{Name: "a1", Pass: "b1"},
		{Name: "a2", Pass: "b2"},
		{Pass: "b3"},
	},
	Nested: NestedConfig{
		Child: Child{
			Number:      101,
			Text:        "child text default",
			AnotherName: "",
		},
	},
}

// todo set own struct annotations like `config:fieldName, required`

func TestUnmarshal(t *testing.T) {
	tcs := []struct {
		name         string
		opts         []any
		envs         map[string]string
		expectParams testConfig
	}{
		{
			name: "load from file",
			opts: []any{CfgFile{"sampledata/testSingleFile.yaml"}},

			// intentionally setting envs that do NOT apply because we did not set the Option
			envs: map[string]string{
				"TEST_ISDEVMODE":    "false",
				"TEST_GENERAL.PORT": "9090",
			},
			expectParams: testConfig{
				Number:      60,
				FloatNum:    3.14,
				Text:        "this is a string",
				FileContent: "mysecret",
				Bol:         true,
				StringList:  []string{"sting 1", "string 2"},
				StructList: []userData{
					{Name: "u1", Pass: "p1"},
					{Name: "u2", Pass: "p2"},
					{Pass: "p3"},
				},
				Nested: NestedConfig{
					Child: Child{
						Number:      61,
						Text:        "this is a string 2",
						AnotherName: "renamedString",
					},
					Child2: struct {
						Number int `config:"number"`
					}(struct{ Number int }{
						Number: 62,
					}),
				},
			},
		},
		{
			name: "12 factor only envs no prefix",
			opts: []any{EnvVar{}},

			// intentionally setting envs that do NOT apply because we did not set the Option
			envs: map[string]string{
				"NUMBER":               "60",
				"FLOATNUM":             "6.65",
				"TEXT":                 "this is a string",
				"FILECONTENT":          "@./sampledata/secretfile",
				"BOL":                  "true",
				"LISTSTRING_0":         "string 1",
				"LISTSTRING_1":         "string 2",
				"NESTED_CHILD_RENAMED": "envValue",
			},
			expectParams: testConfig{
				Number:      60,
				FloatNum:    6.65,
				Text:        "this is a string",
				FileContent: "mysecret",
				Bol:         true,
				StringList:  []string{"string 1", "string 2"},
				StructList:  nil,
				Nested: NestedConfig{
					Child: Child{
						AnotherName: "envValue",
					},
				},
			},
		},
		{
			name: "12 factor only envs with prefix",
			opts: []any{EnvVar{"TEST"}},

			// intentionally setting envs that do NOT apply because we did not set the Option
			envs: map[string]string{
				"TEST_NUMBER":               "60",
				"TEST_FLOATNUM":             "6.65",
				"TEST_TEXT":                 "this is a string",
				"TEST_FILECONTENT":          "@./sampledata/secretfile",
				"TEST_BOL":                  "true",
				"TEST_LISTSTRING_0":         "string 1",
				"TEST_LISTSTRING_1":         "string 2",
				"TEST_NESTED_CHILD_RENAMED": "envValue",
			},
			expectParams: testConfig{
				Number:      60,
				FloatNum:    6.65,
				Text:        "this is a string",
				FileContent: "mysecret",
				Bol:         true,
				StringList:  []string{"string 1", "string 2"},
				StructList:  nil,
				Nested: NestedConfig{
					Child: Child{
						AnotherName: "envValue",
					},
				},
			},
		},
		{
			name:         "preload defaults",
			opts:         []any{Defaults{DefaultCfg}, EnvVar{}},
			expectParams: DefaultCfg,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envs {
				t.Setenv(k, v)
			}
			cfg, err := Load(tc.opts...)
			if err != nil {
				t.Fatal(err)
			}

			t.Run("unmarshal", func(t *testing.T) {

				got := testConfig{}
				err = cfg.Unmarshal(&got)
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(got, tc.expectParams); diff != "" {
					t.Errorf("unexpected value (-got +want)\n%s", diff)
				}
			})
		})
	}
}
