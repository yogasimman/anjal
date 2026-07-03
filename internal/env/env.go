package env

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LoadForCollection reads env vars from both global and local workspaces.
// Global (~/.anjal/) is loaded first as the base, then local (cwd/.anjal/)
// overlays it.
func LoadForCollection(collectionName string) (map[string]string, error) {
	vars := make(map[string]string)
	name := strings.TrimSuffix(collectionName, ".md")

	// 1. Global workspace (base — always loaded)
	home, err := os.UserHomeDir()
	if err == nil {
		global := filepath.Join(home, ".anjal")
		loadFile(vars, filepath.Join(global, ".env"))
		if name != "" {
			loadFile(vars, filepath.Join(global, "."+name+".env"))
		}
	}

	// 2. Local workspace (overlay — overrides global keys)
	// Sync'd with workspace.go: .anjal/ first, then anjal/
	cwd, err := os.Getwd()
	if err == nil {
		// Check cwd/.anjal/
		if info, err := os.Stat(filepath.Join(cwd, ".anjal")); err == nil && info.IsDir() {
			loadFile(vars, filepath.Join(cwd, ".anjal", ".env"))
			if name != "" {
				loadFile(vars, filepath.Join(cwd, ".anjal", "."+name+".env"))
			}
		}
		// Check cwd/anjal/
		if info, err := os.Stat(filepath.Join(cwd, "anjal")); err == nil && info.IsDir() {
			loadFile(vars, filepath.Join(cwd, "anjal", ".env"))
			if name != "" {
				loadFile(vars, filepath.Join(cwd, "anjal", "."+name+".env"))
			}
		}
	}

	return vars, nil
}

// Load is a convenience wrapper that loads only the global .env.
func Load() (map[string]string, error) {
	return LoadForCollection("")
}

// ResolveWriteDir returns the directory where env files should be written.
// Priority matches workspace.go: cwd/.anjal/ → cwd/anjal/ → ~/.anjal/
func ResolveWriteDir() (string, error) {
	cwd, err := os.Getwd()
	if err == nil {
		for _, name := range []string{".anjal", "anjal"} {
			dir := filepath.Join(cwd, name)
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				return dir, nil
			}
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	dir := filepath.Join(home, ".anjal")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create workspace: %w", err)
	}
	return dir, nil
}

// Save writes a single key=value to a collection's .env file.
// Writes to the local workspace if one exists, otherwise global ~/.anjal/.
func Save(collectionName, key, value string) error {
	if collectionName == "" {
		return fmt.Errorf("collection name is required to save")
	}

	ws, err := ResolveWriteDir()
	if err != nil {
		return err
	}

	name := strings.TrimSuffix(collectionName, ".md")
	path := filepath.Join(ws, "."+name+".env")

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

// Delete removes a key from a collection's .env file.
// Targets the local workspace if one exists, otherwise global ~/.anjal/.
func Delete(collectionName, key string) error {
	if collectionName == "" {
		return fmt.Errorf("collection name is required")
	}

	ws, err := ResolveWriteDir()
	if err != nil {
		return err
	}

	name := strings.TrimSuffix(collectionName, ".md")
	path := filepath.Join(ws, "."+name+".env")

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

// ---------- internal helpers ----------

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

// resolveRe matches {{.KEY}}, {{ .KEY }}, {{.KEY }}, etc.
var resolveRe = regexp.MustCompile(`\{\{\s*\.\s*(\w+)\s*\}\}`)

// Resolve replaces {{.VAR_NAME}} placeholders (whitespace-tolerant) with
// values from the vars map.
func Resolve(input string, vars map[string]string) string {
	if vars == nil {
		return input
	}
	return resolveRe.ReplaceAllStringFunc(input, func(match string) string {
		// Extract just the key name from the match
		sub := resolveRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		key := sub[1]
		if value, ok := vars[key]; ok {
			return value
		}
		return match // leave unresolved placeholders as-is
	})
}
