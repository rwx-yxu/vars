package vars

import (
	"sort"

	"github.com/spf13/cobra"
)

// NewCmd returns a [cobra.Command] for managing persistent variables.
//
// The namespace argument creates a root directory for the application in the
// user's state home (e.g., ~/.local/state/my-app). If an optional scope is
// provided, variables are stored in a subdirectory of that namespace
// (e.g., ~/.local/state/my-app/ingest).
//
// The returned command contains subcommands for standard operations:
//  1. init: Initialize the storage.
//  2. set/unset: Write changes to the store.
//  3. get/data/keys: Read values from the store.
//  4. edit: Open the store in the user's preferred editor.
func NewCmd(namespace string, scope ...string) *cobra.Command {
	if len(scope) > 1 {
		panic("vars: strict mode allows only a single level of scope")
	}

	currentScope := ""
	if len(scope) > 0 {
		currentScope = scope[0]
	}

	desc := namespace
	if currentScope != "" {
		desc += "/" + currentScope
	}

	v := New(namespace, scope...)

	cmd := &cobra.Command{
		Use:           "vars",
		Short:         "Manage variables for " + desc,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "initialize empty vars file for <name>",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return v.Init()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a variable",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return v.Set(args[0], args[1])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "unset <key>",
		Short: "Unset a variable property key value",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return v.Unset(args[0])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "data",
		Short: "Prints all vars",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			data, err := v.All()
			if err != nil {
				return err
			}

			keys := make([]string, 0, len(data))
			for k := range data {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				c.Printf("%s=%s\n", k, data[k])
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "edit",
		Short: "Open vars file in default editor",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return v.Edit()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "keys",
		Aliases: []string{"k"},
		Short:   "Prints all keys",
		Args:    cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			data, err := v.All()
			if err != nil {
				return err
			}

			keys := make([]string, 0, len(data))
			for k := range data {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				c.Printf("%s\n", k)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get <key>",
		Short: "Get a variable",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			val, err := v.Get(args[0])
			if err != nil {
				return err
			}
			c.Println(val)
			return nil
		},
	})

	return cmd
}
