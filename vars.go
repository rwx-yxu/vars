package vars

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Vars struct {
	namespace string
	scope     string
	mu        sync.RWMutex
	stateDir  func() (string, error)
}

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func defaultStateDir() (string, error) {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return xdg, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "state"), nil
}

func New(ns string, scope ...string) *Vars {
	s := ""
	if len(scope) > 0 {
		if len(scope) > 1 {
			// Fail fast for developer error
			panic("vars: strict mode allows only a single level of scope (no nesting)")
		}
		s = scope[0]
	}
	return &Vars{
		namespace: ns,
		scope:     s,
		stateDir:  defaultStateDir,
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
	path, err := v.basePath()
	if err != nil {
		return nil, err
	}

	root, err := os.OpenRoot(path)
	if os.IsNotExist(err) {
		target := v.namespace
		if v.scope != "" {
			target = filepath.Join(target, v.scope)
		}
		return nil, fmt.Errorf("vars not initialized for %q (run 'init' first)", target)
	}
	if err != nil {
		return nil, err
	}
	return root, nil

}

func (v *Vars) basePath() (string, error) {
	if v.namespace == "" {
		return "", fmt.Errorf("namespace cannot be empty")
	}

	if !validNameRegex.MatchString(v.namespace) {
		return "", fmt.Errorf("invalid namespace %q", v.namespace)
	}

	if v.scope != "" {
		if strings.ContainsAny(v.scope, `/\`) {
			return "", fmt.Errorf("invalid scope %q: nesting is not allowed", v.scope)
		}
		if !validNameRegex.MatchString(v.scope) {
			return "", fmt.Errorf("invalid scope %q", v.scope)
		}
	}

	rootDir, err := v.stateDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(rootDir, v.namespace, v.scope), nil
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

func (v *Vars) Edit() error {
	path, err := v.basePath()
	if err != nil {
		return err
	}

	filePath := filepath.Join(path, "vars.properties")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("vars not initialized for %q (run 'init' first)", v.namespace)
	}

	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}

	parts := strings.Fields(editor)
	executable := parts[0]
	args := parts[1:]
	args = append(args, filePath)

	cmd := exec.Command(executable, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
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
