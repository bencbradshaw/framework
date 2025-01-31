package env

import (
	"bufio"
	"os"
	"strings"
)

type EnvVars struct {
	Vars map[string]interface{}
}

func LoadEnvFile(filePath string) (*EnvVars, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	envVars := &EnvVars{Vars: make(map[string]interface{})}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value)
		envVars.Vars[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return envVars, nil
}
