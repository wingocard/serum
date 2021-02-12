// Package envparser contains all functions for parsing environment variables and secrets
// from files and the process' environment.
package envparser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

const (
	commentToken  = "#"
	kvSeparator   = "="
	secretRegex   = `^!{(?P<secretval>.+)}$` //nolint:gosec
	emptySecret   = "!{}"
	kvSplitLength = 2
)

var secretRe *regexp.Regexp

func init() {
	secretRe = regexp.MustCompile(secretRegex)
}

type fsWrapper interface {
	Open(path string) (io.ReadCloser, error)
}

type osFS struct{}

func (o *osFS) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

// EnvVars contains the plain text key value mappings as well as the encrypted secret key value mappings
// parsed from an env file.
type EnvVars struct {
	Plain   map[string]string
	Secrets map[string]string
}

// ParseFile parses a .env file at path and returns the key value
// mappings for plain text variables and secret variables.
func ParseFile(path string) (*EnvVars, error) {
	return parseFile(&osFS{}, path)
}

func parseFile(fs fsWrapper, path string) (*EnvVars, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %s", path, err)
	}
	defer f.Close()

	envVars := &EnvVars{
		Plain:   make(map[string]string),
		Secrets: make(map[string]string),
	}

	var lp lineParser
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if err := lp.parse(envVars, line); err != nil {
			return nil, fmt.Errorf("error parsing line: %s: %s", line, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing file: %s", err)
	}

	return envVars, nil
}

type lineParser struct {
	multiline bool
	key       string
	value     string
}

func (p *lineParser) parse(envVars *EnvVars, l string) error {
	l = strings.TrimSpace(l)
	// ignore empty lines
	if l == "" {
		return nil
	}

	// ignore commented line
	// TODO: ignore inline comments
	if strings.HasPrefix(l, commentToken) {
		return nil
	}

	// handle multiline variables
	if p.multiline {
		// check if it's the end of a multiline var
		if strings.HasSuffix(l, `"`) {
			p.value += strings.TrimSuffix(l, `"`)
			p.multiline = false
			envVars.Plain[p.key] = p.value
			return nil
		}
		p.value += fmt.Sprintf("%s\n", l)
		return nil
	}

	// split line into two pieces (k,v) based on key value separator
	splits := strings.SplitN(l, kvSeparator, kvSplitLength)
	if len(splits) != kvSplitLength {
		return fmt.Errorf("invalid format %q", l)
	}

	// key is first index, value is second
	k := strings.TrimSpace(splits[0])
	v := strings.TrimSpace(splits[1])
	if k == "" || v == emptySecret {
		return fmt.Errorf("invalid format %q", l)
	}

	// check if value is encrypted secret
	if secretRe.MatchString(v) {
		// fill in secret value - replace template value with capture group "secretval"
		envVars.Secrets[k] = secretRe.ReplaceAllString(v, "$secretval")
		return nil
	}

	// check if value is the beginning of a multiline variable
	if strings.HasPrefix(v, `"`) && !strings.HasSuffix(v, `"`) {
		p.multiline = true
		p.key = k
		p.value = fmt.Sprintf("%s\n", strings.TrimPrefix(v, `"`))
		return nil
	}

	// not a secret, fill in plain text value
	envVars.Plain[k] = v
	return nil
}

// ParseEnv parses the process' environment for the specified
// keys and returns the key value mappings for plain text
// variables and secret variables.
func ParseEnv(keys []string) (*EnvVars, error) {
	envVars := &EnvVars{
		Plain:   make(map[string]string),
		Secrets: make(map[string]string),
	}

	for _, k := range keys {
		v, ok := os.LookupEnv(k)
		if !ok {
			return nil, fmt.Errorf("env variable %q not found", v)
		}

		// check if value is encrypted secret
		if secretRe.MatchString(v) {
			// fill in secret value - replace template value with capture group "secretval"
			envVars.Secrets[k] = secretRe.ReplaceAllString(v, "$secretval")
			continue
		}

		// not a secret, fill in plain text value
		envVars.Plain[k] = v
	}

	return envVars, nil
}
