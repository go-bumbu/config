package config

import (
	"testing"
)

func Test_unescapeDoubleQuoted(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"no escapes", "plain", "plain"},
		{"newline", `\n`, "\n"},
		{"tab", `\t`, "\t"},
		{"carriage return", `\r`, "\r"},
		{"double quote", `\"`, `"`},
		{"backslash", `\\`, `\`},
		{"unknown escape kept", `\x`, `\x`},
		{"mixed", `a\nb\tc`, "a\nb\tc"},
		{"quoted in middle", `say \"hi\"`, `say "hi"`},
		{"backslash at end", `path\`, `path\`},
		{"multiple backslashes", `\\\\`, `\\`},
		{"newline and tab", `\n\t`, "\n\t"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unescapeDoubleQuoted(tt.in)
			if got != tt.want {
				t.Errorf("unescapeDoubleQuoted(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func Test_unquoteEnvVal(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"unquoted", "value", "value"},
		{"double quoted", `"value"`, "value"},
		{"double quoted with escape", `"say \"hi\""`, `say "hi"`},
		{"single quoted", "'value'", "value"},
		{"single quoted preserves dollar", "'$VAR'", "$VAR"},
		{"inline comment", "value # comment", "value"},
		{"only two chars double quote", `""`, ""},
		{"only two chars single quote", "''", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unquoteEnvVal(tt.in)
			if got != tt.want {
				t.Errorf("unquoteEnvVal(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
