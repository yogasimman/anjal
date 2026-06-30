package env

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func LoadForCollection(collectionName string) (map[string]string, error) {
	vars := make(map[string]string)

	home, err := os.UserHomeDir()
	if err != nil {
		return vars, fmt.Errorf("could not determine home directory: %w", err)
	}

	workspace := filepath.Join(home, ".anjal")
	loadFile(vars, filepath.Join(workspace, ".env"))

	if collectionName != "" {
		collectionEnv := strings.TrimSuffix(collectionName, ".md") + ".env"
		loadFile(vars, filepath.Join(workspace, collectionEnv))
	}

	return vars, nil
}

func Load() (map[string]string, error) {
	return LoadForCollection("")
}

func Save(collectionName, key, value string) error {
	if collectionName == "" {
		return fmt.Errorf("collection name is required to save")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	workspace := filepath.Join(home, ".anjal")
	if err := os.MkdirAll(workspace, 0755); err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	collectionEnv := strings.TrimSuffix(collectionName, ".md") + ".env"
	path := filepath.Join(workspace, collectionEnv)

	lines, _ := readLines(path)

	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		idx := strings.Index(trimmed, "=")
		if idx == -1 {
			continue
		}
		if strings.TrimSpace(trimmed[:idx]) == key {
			lines[i] = key + "=" + value
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, key+"="+value)
	}

	return writeLines(path, lines)
}

func Delete(collectionName, key string) error {
	if collectionName == "" {
		return fmt.Errorf("collection name is required")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	collectionEnv := strings.TrimSuffix(collectionName, ".md") + ".env"
	path := filepath.Join(home, ".anjal", collectionEnv)

	lines, err := readLines(path)
	if err != nil {
		return nil
	}

	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			filtered = append(filtered, line)
			continue
		}
		idx := strings.Index(trimmed, "=")
		if idx != -1 && strings.TrimSpace(trimmed[:idx]) == key {
			continue
		}
		filtered = append(filtered, line)
	}

	return writeLines(path, filtered)
}

func loadFile(vars map[string]string, path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		idx := strings.Index(line, "=")
		if idx == -1 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[0] == value[len(value)-1] {
			value = value[1 : len(value)-1]
		}

		vars[key] = value
	}
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(path string, lines []string) error {
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

func Resolve(input string, vars map[string]string) string {
	if vars == nil {
		return input
	}
	for key, value := range vars {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		input = strings.ReplaceAll(input, placeholder, value)
	}
	return input
}
