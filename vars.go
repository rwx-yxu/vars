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
	// check XDG_STATE_HOME env path first

	root, err := v.root()
	if err != nil {
		return fmt.Errorf("failed to open root: %w", err)
	}

	defer root.Close()

	f, err := root.OpenFile("vars.properties", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

func (v *Vars) root() (*os.Root, error) {
	if v.name == "" {
		return nil, fmt.Errorf("vars name cannot be empty")
	}

	if !validNameRegex.MatchString(v.name) {
		return nil, fmt.Errorf("invalid app name %q: must be alphanumeric/safe", v.name)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	homeRoot, err := os.OpenRoot(home)
	if err != nil {
		return nil, err
	}
	defer homeRoot.Close()

	relPath := filepath.Join(".local", "state", v.name)

	if err := homeRoot.MkdirAll(relPath, 0700); err != nil {
		return nil, err
	}

	return homeRoot.OpenRoot(relPath)
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

func (v *Vars) Delete(key string) error {
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
