package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

var secretRe *regexp.Regexp

const secretRegex = `^(?P<key>.+)=\${(?P<secret>.+)}$`

func init() {
	//TODO: remove this global
	secretRe = regexp.MustCompile(secretRegex)
}

func main() {
	var env string

	inject := flag.NewFlagSet("inject", flag.ExitOnError)
	inject.StringVar(&env, "env", "local", "the current running environment")
	inject.StringVar(&env, "e", "local", "shorthand for env")

	flag.Parse()

	args := flag.Args()
	//TODO: args validation
	//0 is the inject arg
	filePath := args[1]

	if err := parseFile(filePath); err != nil {
		log.Fatalf("error injecting config: %s\n", err)
	}
}

func parseFile(fp string) error {
	f, err := os.Open(fp)
	if err != nil {
		return fmt.Errorf("error opening file %s: %s", fp, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		val, err := parseLine(line)
		if err != nil {
			return fmt.Errorf("error parsing line: %s: %s", line, err)
		}
		if val == "" {
			//line is a comment, skip it
			continue
		}

		fmt.Printf("parsed value: %s\n", val)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error parsing config file: %s", err)
	}

	return nil
}

func parseLine(l string) (string, error) {
	//ignore commented line
	if strings.HasPrefix(l, "#") {
		return "", nil
	}
	//TODO: ignore inline comments

	res := secretRe.FindStringSubmatch(l)
	if res == nil {
		return l, nil
	}

	//0 index is for the entire match
	key := res[1]
	secret := res[2]

	//lookup secret in secret manager
	c, err := secretmanager.NewClient(context.Background())
	if err != nil {
		return "", err
	}

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secret,
	}

	result, err := c.AccessSecretVersion(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %v", err)
	}

	return fmt.Sprintf("%s=%s", key, result.Payload.Data), nil
}

//split line into two pieces based on key value seperator "="
// 	splits := strings.SplitN(l, "=", 2)
// 	if len(splits) != 2 {
// 		return "", fmt.Errorf("invalid config format %s", l)
// 	}

// 	//key is first index, value is second
// 	key := strings.TrimSpace(splits[0])
// 	value := strings.TrimSpace(splits[1])
// 	//TODO: ignore inline comments

// 	//check if value matches a secret that needs to be replaced using the secret manager
// 	regex, err := regexp.Compile(secretRegex)
// 	if err != nil {
// 		return "", fmt.Errorf("error compiling secret regex")
// 	}

// 	if regex.Match([]byte(value)) {
// 		value, err := lookupSecret(value)
// 	}

// 	return fmt.Sprintf("")
// }
