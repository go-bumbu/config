package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
)

type myCfg struct {
	fl  float64 `config:"default_float"`
	fl2 float64 `config:"floatNum"` //nolint:unused // field is used in test unmarshal
}

var defCfg = myCfg{
	fl: 35,
}

var _ = spew.Dump

func TestLoad(t *testing.T) {
	tcs := []struct {
		name        string
		opts        []any
		envs        map[string]string
		expectVal   map[string]string
		expectedErr string
	}{
		{
			name: "load default",
			opts: []any{Defaults{defCfg}},
			expectVal: map[string]string{
				"default_float": "35",
			},
		},
		{
			name:        "expect Err if default is not struct",
			opts:        []any{Defaults{12}},
			expectedErr: "error loading default values: passed src is not a pointer or struct",
		},
		{
			name:        "expect Err if mandatory file is not found",
			opts:        []any{CfgFile{"sampledata/nonexistent.yaml", true}},
			expectedErr: "unable to find config FILE: \"sampledata/nonexistent.yaml\"",
		},
		{
			name: "load only defaults",
			opts: []any{Defaults{defCfg}},
			expectVal: map[string]string{
				"default_float": "35",
			},
		},
		{
			name: "load default and overlay with file",
			opts: []any{Defaults{defCfg}, CfgFile{"sampledata/testSingleFile.yaml", true}},
			expectVal: map[string]string{
				"default_float": "35",
				"floatNum":      "3.14",
			},
		},
		{
			name: "load default and overlay overrides",
			opts: []any{Defaults{defCfg}, Overrides{myCfg{fl: 7.15}}},
			expectVal: map[string]string{
				"default_float": "7.15",
			},
		},
		{
			name: "overrides should load even without default",
			opts: []any{Overrides{myCfg{fl: 7.15}}},
			expectVal: map[string]string{
				"default_float": "7.15",
			},
		},
		{
			name: "allow to load multiple files",
			opts: []any{Defaults{defCfg},
				CfgFile{"sampledata/testSingleFile.yaml", true},
				CfgFile{"sampledata/override.yaml", true},
			},
			expectVal: map[string]string{
				"default_float": "35",
				"floatNum":      "8.95",
			},
		},
		{
			name: "ignore non mandatory file",
			opts: []any{Defaults{defCfg}, CfgFile{"sampledata/nonexistent.yaml", false}},
			expectVal: map[string]string{
				"default_float": "35",
			},
		},
		{
			name: "load default and overlay with env",
			opts: []any{Defaults{defCfg}, CfgFile{"sampledata/testSingleFile.yaml", true}, EnvVar{"TEST"}},
			envs: map[string]string{
				"TEST_FLOATNUM": "6.65",
			},
			expectVal: map[string]string{
				"default_float": "35",
				"floatNum":      "6.65",
			},
		},
		{
			name: "load from file",
			opts: []any{CfgFile{"sampledata/testSingleFile.yaml", true}},
			// intentionally setting envs that do NOT apply because we did not set the Option
			envs: map[string]string{
				"IGNORE_FLOATNUM": "44.4",
				"IGNORE_NUMBER":   "9090",
			},
			expectVal: map[string]string{
				"floatNum":             "3.14",
				"nested.child.renamed": "renamedString",
				"bol":                  "true",
				"fileContent":          "mysecret",
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
			expectVal: map[string]string{
				"floatNum":             "6.65",
				"nested.child.renamed": "envValue",
				"bol":                  "true",
				"liststring.0":         "string 1",
				"fileContent":          "mysecret",
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
			expectVal: map[string]string{
				"floatNum":             "6.65",
				"nested.child.renamed": "envValue",
				"bol":                  "true",
				"fileContent":          "mysecret",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envs {
				t.Setenv(k, v)
			}
			cfg, err := Load(tc.opts...)

			if tc.expectedErr != "" {
				if err == nil {
					t.Fatal("expected error but none got")
				}
				if diff := cmp.Diff(err.Error(), tc.expectedErr); diff != "" {
					t.Errorf("unexpected error (-got +want)\n%s", diff)
				}

			} else {
				if err != nil {
					t.Fatal(err)
				}
				t.Run("get values", func(t *testing.T) {
					for k, v := range tc.expectVal {
						got, err := cfg.GetString(k)
						if err != nil {
							t.Fatal(err)
						}
						if diff := cmp.Diff(got, v); diff != "" {
							t.Errorf("unexpected value (-got +want)\n%s", diff)
						}
					}
				})
			}

		})
	}
}

func TestLoad_EnvFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "# comment\nTEST_FLOATNUM=99.99\n\nOTHER=ignored\n"
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(
		Defaults{defCfg},
		CfgFile{"sampledata/testSingleFile.yaml", true},
		EnvFile{Path: envPath, Mandatory: true},
		EnvVar{Prefix: "TEST"},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Value should come from .env (99.99), not from YAML (3.14)
	got, err := cfg.GetString("floatNum")
	if err != nil {
		t.Fatal(err)
	}
	if got != "99.99" {
		t.Errorf("floatNum: got %q, want 99.99 (from .env)", got)
	}
}

func TestLoad_EnvFileOptional(t *testing.T) {
	// Non-mandatory missing .env should not error
	_, err := Load(
		Defaults{defCfg},
		EnvFile{Path: filepath.Join(t.TempDir(), "nonexistent.env"), Mandatory: false},
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFlattenMap(t *testing.T) {
	byt, err := os.ReadFile("sampledata/testFlatten.yaml")
	if err != nil {
		t.Fatal(err)
	}
	in, err := readCfgBytes(byt, ExtYaml)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]interface{}{}
	flatten("", in, got)
	want := map[string]interface{}{
		"general.child1.list.0.name":         "item1",
		"general.child1.list.0.value":        true,
		"general.child1.list.1.name":         "item2",
		"general.child1.list.1.value":        "value2",
		"general.child1.list.1.data.subData": "my string",
		"general.child1.list.2.name":         "float",
		"general.child1.list.2.val":          2.5,
		"general.child2.sub1.sub2":           1,
		"top":                                "level",
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("unexpected value (-got +want)\n%s", diff)
	}

}
func TestFlattenStruct(t *testing.T) {
	in := DefaultCfg

	got := map[string]interface{}{}
	err := flattenStruct(in, got)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]interface{}{
		"number":              100,
		"floatNum":            100.1,
		"text":                "default text",
		"listString.0":        "default 1",
		"listString.1":        "default 2",
		"nested.child.number": 101,
		"nested.child.text":   "child text default",
		"userList.0.name":     "a1",
		"userList.0.pass":     "b1",
		"userList.1.name":     "a2",
		"userList.1.pass":     "b2",
		"userList.2.pass":     "b3",
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("unexpected value (-got +want)\n%s", diff)
	}

}
