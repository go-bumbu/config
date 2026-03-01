package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// EnvFile loads environment variables from a .env-style file (KEY=VALUE per line).
// Variables are set in the process environment so they are visible to EnvVar when unmarshalling.
// Use Mandatory: false so a missing file is ignored (e.g. for optional .env in development).
type EnvFile struct {
	Path      string
	Mandatory bool
}

// loadEnvFile reads path and sets os.Setenv for each KEY=VALUE line.
// Only sets variables that are not already in the environment, so explicit
// env vars (e.g. from shell) override .env. Supports # comments, empty lines,
// and optional double/single-quoted values.
func loadEnvFile(c *CfgHandler, path string, mandatory bool) error {
	// #nosec G304 -- path is from config option (EnvFile.Path), not arbitrary user input
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && !mandatory {
			c.info(fmt.Sprintf("ENV file \"%s\" not found, skipping (not mandatory)", path))
			return nil
		}
		return fmt.Errorf("unable to open ENV file %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	c.info(fmt.Sprintf("loading ENV from file \"%s\"", path))
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if key == "" {
			continue
		}
		val = unquoteEnvVal(val)
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, val); err != nil {
				return fmt.Errorf("ENV file %s line %d: setenv %q: %w", path, lineNum, key, err)
			}
		}
	}
	return scanner.Err()
}

// unquoteEnvVal removes surrounding double or single quotes and unescapes double-quoted content.
func unquoteEnvVal(s string) string {
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) && len(s) >= 2 {
		return unescapeDoubleQuoted(s[1 : len(s)-1])
	}
	if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") && len(s) >= 2 {
		return s[1 : len(s)-1]
	}
	// Inline comment after value (unquoted)
	if i := strings.Index(s, " #"); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

func unescapeDoubleQuoted(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			default:
				b.WriteByte(s[i])
				b.WriteByte(s[i+1])
			}
			i++
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
