package vars

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// --- TEST: Core Logic & Edge Cases ---

func TestStrictScopes(t *testing.T) {
	tests := []struct {
		name      string
		scopeArgs []string
		wantPanic bool
		wantErr   bool
	}{
		{"Root Scope (e.g. 'pomo')", []string{}, false, false},
		{"Single Scope (e.g. 'timer')", []string{"timer"}, false, false},
		{"Explicit Empty Scope", []string{""}, false, false},
		{"Nested Scope (Variadic - PANIC)", []string{"timer", "work"}, true, false},
		{"Nested Scope (Slash - ERROR)", []string{"timer/work"}, false, true},
		{"Invalid Char Scope", []string{"timer!"}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wantPanic && r == nil {
					t.Errorf("Expected panic for args %v, but got none", tt.scopeArgs)
				} else if !tt.wantPanic && r != nil {
					t.Errorf("Unexpected panic: %v", r)
				}
			}()

			v := New("pomo-cli", tt.scopeArgs...)
			tempDir := t.TempDir()
			v.stateDir = func() (string, error) {
				return tempDir, nil
			}

			if tt.wantPanic {
				return
			}

			err := v.Init()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for scope %v, got nil", tt.scopeArgs)
				}
			} else {
				if err != nil {
					t.Errorf("Init failed for valid scope %v: %v", tt.scopeArgs, err)
				}
			}
		})
	}
}

func TestUninitializedAccess(t *testing.T) {
	v := New("weather-cli")
	tempDir := t.TempDir()
	v.stateDir = func() (string, error) {
		return tempDir, nil
	}
	if err := v.Set("api_key", "12345"); err == nil {
		t.Error("Set should fail if Init() hasn't been called")
	} else if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("Wrong error message: %v", err)
	}

	if _, err := v.Get("api_key"); err == nil {
		t.Error("Get should fail if Init() hasn't been called")
	}
}

func TestPersistenceAndEncoding(t *testing.T) {
	ns := "api"
	key := "secret_key"
	val := "-----BEGIN PUBLIC KEY-----\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\n-----END PUBLIC KEY-----"

	v1 := New(ns)
	tempDir := t.TempDir()
	v1.stateDir = func() (string, error) {
		return tempDir, nil
	}
	v1.Init()
	if err := v1.Set(key, val); err != nil {
		t.Fatal(err)
	}

	v2 := New(ns)
	v2.stateDir = func() (string, error) {
		return tempDir, nil
	}
	got, err := v2.Get(key)
	if err != nil {
		t.Fatal(err)
	}

	if got != val {
		t.Errorf("Encoding mismatch.\nWant: %q\nGot:  %q", val, got)
	}
}

func TestUnset(t *testing.T) {
	v := New("todo-app")
	tempDir := t.TempDir()
	v.stateDir = func() (string, error) {
		return tempDir, nil
	}
	v.Init()
	v.Set("theme", "dark")
	v.Set("show_completed", "true")

	if err := v.Unset("show_completed"); err != nil {
		t.Fatal(err)
	}

	if _, err := v.Get("show_completed"); err == nil {
		t.Error("Get should fail for unset key")
	}
	if val, _ := v.Get("theme"); val != "dark" {
		t.Error("Unset affected other keys")
	}
}

// --- TEST: Concurrency ---

func TestConcurrency(t *testing.T) {
	v := New("stock-ticker")
	tempDir := t.TempDir()
	v.stateDir = func() (string, error) {
		return tempDir, nil
	}
	v.Init()

	var wg sync.WaitGroup
	routines := 50

	for i := 0; i < routines; i++ {
		wg.Go(func() {
			v.Set(fmt.Sprintf("AAPL_%d", i), "150.00")
		})
	}

	for i := 0; i < routines; i++ {
		wg.Go(func() {
			v.All()
		})
	}

	wg.Wait()

	data, _ := v.All()
	if len(data) != routines {
		t.Errorf("Race condition suspected. Expected %d keys, got %d", routines, len(data))
	}
}

// --- TEST: CLI Integration (Checking Args & Wiring) ---

func TestEmbeddedCmdIntegration(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	rootCmd := NewCmd("pomo-test", "timer")

	exec := func(args ...string) error {
		rootCmd.SetArgs(args)
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		return rootCmd.Execute()
	}

	if err := exec("init"); err != nil {
		t.Fatalf("CLI Init failed: %v", err)
	}

	if err := exec("set", "default_duration", "25m"); err != nil {
		t.Fatalf("CLI Set failed: %v", err)
	}

	if err := exec("set", "default_duration"); err == nil {
		t.Error("CLI Set should fail with 1 arg, but succeeded")
	}

	if err := exec("get", "default_duration"); err != nil {
		t.Fatalf("CLI Get failed: %v", err)
	}

	if err := exec("unset", "default_duration"); err != nil {
		t.Errorf("CLI Unset failed (did you fix args?): %v", err)
	}

	if err := exec("data"); err != nil {
		t.Errorf("CLI Data failed (did you fix args?): %v", err)
	}
}
