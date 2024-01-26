package env

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Utilities for getting environment variables, to be used by the DNS providers in tmpim/dnsproviders. These utility
// functions are based on lego's config/env/env.go, which is licensed under the MIT license. See
// https://github.com/go-acme/lego/blob/master/LICENSE for more details.

// GetWithFallback Get environment variable values.
// The first name in each group is use as key in the result map.
//
// case 1:
//
//	// LEGO_ONE="ONE"
//	// LEGO_TWO="TWO"
//	env.GetWithFallback([]string{"LEGO_ONE", "LEGO_TWO"})
//	// => "LEGO_ONE" = "ONE"
//
// case 2:
//
//	// LEGO_ONE=""
//	// LEGO_TWO="TWO"
//	env.GetWithFallback([]string{"LEGO_ONE", "LEGO_TWO"})
//	// => "LEGO_ONE" = "TWO"
//
// case 3:
//
//	// LEGO_ONE=""
//	// LEGO_TWO=""
//	env.GetWithFallback([]string{"LEGO_ONE", "LEGO_TWO"})
//	// => error
func GetWithFallback(groups ...[]string) (map[string]string, error) {
	values := map[string]string{}

	var missingEnvVars []string
	for _, names := range groups {
		if len(names) == 0 {
			return nil, errors.New("undefined environment variable names")
		}

		value, envVar := getOneWithFallback(names[0], names[1:]...)
		if len(value) == 0 {
			missingEnvVars = append(missingEnvVars, envVar)
			continue
		}
		values[envVar] = value
	}

	if len(missingEnvVars) > 0 {
		return nil, fmt.Errorf("some credentials information are missing: %s", strings.Join(missingEnvVars, ","))
	}

	return values, nil
}

func getOneWithFallback(main string, names ...string) (string, string) {
	value := GetOrFile(main)
	if len(value) > 0 {
		return value, main
	}

	for _, name := range names {
		value := GetOrFile(name)
		if len(value) > 0 {
			return value, main
		}
	}

	return "", main
}

// GetOrFile Attempts to resolve 'key' as an environment variable.
// Failing that, it will check to see if '<key>_FILE' exists.
// If so, it will attempt to read from the referenced file to populate a value.
func GetOrFile(envVar string) string {
	envVarValue := os.Getenv(envVar)
	if envVarValue != "" {
		return envVarValue
	}

	fileVar := envVar + "_FILE"
	fileVarValue := os.Getenv(fileVar)
	if fileVarValue == "" {
		return envVarValue
	}

	fileContents, err := ioutil.ReadFile(fileVarValue)
	if err != nil {
		log.Printf("Failed to read the file %s (defined by env var %s): %s", fileVarValue, fileVar, err)
		return ""
	}

	return strings.TrimSuffix(string(fileContents), "\n")
}
