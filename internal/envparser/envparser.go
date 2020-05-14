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
	commentToken = "#"
	kvSeparator  = "="
	secretRegex  = `^!{(?P<secretval>.+)}$`
	emptySecret  = "!{}"
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
// parsed from an env file
type EnvVars struct {
	Plain   map[string]string
	Secrets map[string]string
}

// ParseFile parses a .env file at path and returns the key value
// mappings for plain text variables and secret variables
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

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if err := parseLine(envVars, line); err != nil {
			return nil, fmt.Errorf("error parsing line: %s: %s", line, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing file: %s", err)
	}

	return envVars, nil
}

func parseLine(envVars *EnvVars, l string) error {
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

	// split line into two pieces (k,v) based on key value seperator
	splits := strings.SplitN(l, kvSeparator, 2)
	if len(splits) != 2 {
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

	// not a secret, fill in plain text value
	envVars.Plain[k] = v
	return nil
}
