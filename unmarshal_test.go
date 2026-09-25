package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
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

// configWithPtr is used to test unmarshal case reflect.Pointer (ptr to scalar and ptr to struct).
type configWithPtr struct {
	PtrText   *string `config:"ptrText"`
	PtrNumber *int    `config:"ptrNumber"`
	PtrNested *Child  `config:"ptrNested"`
}

// configWithOptional has a pointer field of each supported kind; a nil pointer
// means "not set by any source", which a plain field cannot express.
type configWithOptional struct {
	Flag   *bool          `config:"flag"`
	Text   *string        `config:"text"`
	Number *int           `config:"number"`
	Float  *float64       `config:"float"`
	Nested *Child         `config:"nested"`
	Items  []optionalItem `config:"items"`
}

type optionalItem struct {
	Name string `config:"name"`
	Flag *bool  `config:"flag"`
}

// configWithUnhandledPtr has a pointer to a type not supported by unmarshal (int64), used to test error path.
type configWithUnhandledPtr struct {
	PtrInt64 *int64 `config:"ptrInt64"`
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
			opts: []any{CfgFile{"sampledata/testSingleFile.yaml", true}},

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
		{
			name: "env overrides bool default to false",
			opts: []any{Defaults{testConfig{Bol: true}}, EnvVar{}},
			envs: map[string]string{
				"BOL": "false",
			},
			expectParams: testConfig{Bol: false},
		},
		{
			name: "env overrides bool default to true",
			opts: []any{Defaults{testConfig{Bol: false}}, EnvVar{}},
			envs: map[string]string{
				"BOL": "true",
			},
			expectParams: testConfig{Bol: true},
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

func TestUnmarshalErrs(t *testing.T) {
	tcs := []struct {
		name      string
		envs      map[string]string
		expectErr string
	}{
		{
			name: "expect error if wrong boolean",

			envs: map[string]string{
				"TEST_BOL": "banana",
			},
			expectErr: "unable to convert env value to bool strconv.ParseBool: parsing \"banana\": invalid syntax",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envs {
				t.Setenv(k, v)
			}
			cfg, err := Load(EnvVar{Prefix: "TEST"})
			if err != nil {
				t.Fatal(err)
			}

			got := testConfig{}
			err = cfg.Unmarshal(&got)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != tc.expectErr {
				t.Errorf("expected '%s', got '%s'", tc.expectErr, err.Error())
			}

		})
	}
}

func TestUnmarshal_ptr(t *testing.T) {
	ptrNum42 := 42
	ptrNum99 := 99
	tcs := []struct {
		name         string
		opts         []any
		envs         map[string]string
		expectParams configWithPtr
	}{
		{
			name: "ptr from env",
			opts: []any{EnvVar{Prefix: "TEST"}},
			envs: map[string]string{
				"TEST_PTRTEXT":           "hello",
				"TEST_PTRNUMBER":         "42",
				"TEST_PTRNESTED_NUMBER":  "7",
				"TEST_PTRNESTED_TEXT":    "nested-text",
				"TEST_PTRNESTED_RENAMED": "from-env",
			},
			expectParams: configWithPtr{
				PtrText:   strPtr("hello"),
				PtrNumber: &ptrNum42,
				PtrNested: &Child{Number: 7, Text: "nested-text", AnotherName: "from-env"},
			},
		},
		{
			name: "ptr from file",
			opts: []any{CfgFile{"sampledata/testPtrConfig.yaml", true}},
			envs: nil,
			expectParams: configWithPtr{
				PtrText:   strPtr("from-file"),
				PtrNumber: &ptrNum99,
				PtrNested: &Child{Number: 11, Text: "nested-from-file", AnotherName: "ptr-nested-renamed"},
			},
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
			var got configWithPtr
			err = cfg.Unmarshal(&got)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(got, tc.expectParams); diff != "" {
				t.Errorf("unexpected value (-got +want)\n%s", diff)
			}
		})
	}
}

func TestUnmarshal_ptrUnhandledType(t *testing.T) {
	t.Setenv("TEST_PTRINT64", "1")
	cfg, err := Load(EnvVar{Prefix: "TEST"})
	if err != nil {
		t.Fatal(err)
	}
	var got configWithUnhandledPtr
	err = cfg.Unmarshal(&got)
	if err == nil {
		t.Fatal("expected error for unhandled pointer-to type, got nil")
	}
	wantErr := "unhandled pointer-to type: \"int64\" in struct"
	if err.Error() != wantErr {
		t.Errorf("expected error %q, got %q", wantErr, err.Error())
	}
}

// TestUnmarshal_ptrUnset checks that a pointer field stays nil when no source
// sets its key (missing, blank or null), while an explicit zero value such as
// false, "" or 0 still produces a non-nil pointer.
func TestUnmarshal_ptrUnset(t *testing.T) {
	tcs := []struct {
		name         string
		file         string // the extension selects the format
		content      string
		envs         map[string]string
		preset       configWithOptional // struct content before unmarshal
		expectParams configWithOptional
	}{
		{
			name:    "keys missing",
			file:    "cfg.yaml",
			content: "other: 1\n",
		},
		{
			name:    "keys blank",
			file:    "cfg.yaml",
			content: "flag:\ntext:\nnumber:\nfloat:\nnested:\n",
		},
		{
			name:    "keys null",
			file:    "cfg.json",
			content: `{"flag": null, "text": null, "number": null, "float": null, "nested": null}`,
		},
		{
			name:    "explicit zero values",
			file:    "cfg.yaml",
			content: "flag: false\ntext: \"\"\nnumber: 0\nfloat: 0.0\nnested:\n  number: 0\n",
			expectParams: configWithOptional{
				Flag:   ptr(false),
				Text:   ptr(""),
				Number: ptr(0),
				Float:  ptr(0.0),
				Nested: &Child{},
			},
		},
		{
			name:         "env var fills a blank key",
			file:         "cfg.yaml",
			content:      "flag:\n",
			envs:         map[string]string{"TEST_FLAG": "false"},
			expectParams: configWithOptional{Flag: ptr(false)},
		},
		{
			name:    "list entries",
			file:    "cfg.yaml",
			content: "items:\n  - name: a\n  - name: b\n    flag:\n  - name: c\n    flag: false\n  - flag: true\n",
			expectParams: configWithOptional{Items: []optionalItem{
				{Name: "a"},
				{Name: "b"},
				{Name: "c", Flag: ptr(false)},
				{Flag: ptr(true)},
			}},
		},
		{
			name:         "preset pointer kept when key missing",
			file:         "cfg.yaml",
			content:      "other: 1\n",
			preset:       configWithOptional{Flag: ptr(true)},
			expectParams: configWithOptional{Flag: ptr(true)},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envs {
				t.Setenv(k, v)
			}
			path := filepath.Join(t.TempDir(), tc.file)
			if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(CfgFile{Path: path, Mandatory: true}, EnvVar{Prefix: "TEST"})
			if err != nil {
				t.Fatal(err)
			}
			got := tc.preset
			err = cfg.Unmarshal(&got)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(got, tc.expectParams); diff != "" {
				t.Errorf("unexpected value (-got +want)\n%s", diff)
			}
		})
	}
}

func TestUnmarshal_ptrInvalidValue(t *testing.T) {
	t.Setenv("TEST_FLAG", "banana")
	cfg, err := Load(EnvVar{Prefix: "TEST"})
	if err != nil {
		t.Fatal(err)
	}
	var got configWithOptional
	err = cfg.Unmarshal(&got)
	if err == nil {
		t.Fatal("expected error for invalid bool in pointer field, got nil")
	}
}

func strPtr(s string) *string { return &s }

func ptr[T any](v T) *T { return &v }
