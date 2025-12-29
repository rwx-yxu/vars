package vars

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Vars struct {
	name string
	mu   sync.RWMutex
}

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func New(name string) *Vars {
	return &Vars{
		name: name,
	}
}

func (v *Vars) Init() error {

	path, err := v.basePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("failed to create state dir: %w", err)
	}

	root, err := os.OpenRoot(path)
	if err != nil {
		return fmt.Errorf("failed to open root: %w", err)
	}

	defer root.Close()

	f, err := root.OpenFile("vars.properties", os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer f.Close()

	return nil
}

func (v *Vars) root() (*os.Root, error) {
	if v.name == "" {
		return nil, fmt.Errorf("vars name cannot be empty")
	}

	if !validNameRegex.MatchString(v.name) {
		return nil, fmt.Errorf("invalid app name %q: must be alphanumeric/safe", v.name)
	}

	path, err := v.basePath()
	if err != nil {
		return nil, err
	}

	root, err := os.OpenRoot(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("vars not initialized for %q (run 'init' first)", v.name)
	}
	if err != nil {
		return nil, err
	}
	return root, nil

}

func (v *Vars) basePath() (string, error) {
	// XDG Compliance: Check env var first
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, v.name), nil
	}

	// Fallback to ~/.local/state
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "state", v.name), nil
}

func (v *Vars) Get(key string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	m, err := v.load()
	if err != nil {
		return "", err
	}
	val, ok := m[key]
	if !ok {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return val, nil
}

func (v *Vars) Set(key, val string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	m, err := v.load()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if m == nil {
		m = make(map[string]string)
	}

	m[key] = val
	return v.save(m)
}

func (v *Vars) Unset(key string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	m, err := v.load()
	if err != nil {
		return err
	}
	delete(m, key)
	return v.save(m)
}

func (v *Vars) All() (map[string]string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.load()
}

func (v *Vars) load() (map[string]string, error) {
	data := make(map[string]string)

	root, err := v.root()
	if err != nil {
		return nil, fmt.Errorf("unable to construct vars.properties path: %w", err)
	}

	file, err := root.Open("vars.properties")
	if os.IsNotExist(err) {
		return data, fmt.Errorf("vars has not been initialized")
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := unescape(strings.TrimSpace(parts[1]))
			data[key] = val
		}
	}
	return data, scanner.Err()
}

func (v *Vars) save(data map[string]string) error {
	var buf bytes.Buffer

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("%s=%s\n", k, escape(data[k])))
	}

	root, err := v.root()
	if err != nil {
		return fmt.Errorf("unable to construct vars.properties path: %w", err)
	}

	return root.WriteFile("vars.properties", buf.Bytes(), 0600)
}

func escape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", "\\n"), "\r", "\\r")
}

func unescape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\\n", "\n"), "\\r", "\r")
}
