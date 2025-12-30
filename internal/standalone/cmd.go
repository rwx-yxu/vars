package standalone

import (
	"fmt"
	"os"
	"sort"

	"github.com/rwx-yxu/vars"
	"github.com/spf13/cobra"
)

func Execute() {
	root := cmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "vars",
		Short:         "Manage stateful properties for any application",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "init <name> [scope]",
		Short: "Initialize vars (Required before use)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(c *cobra.Command, args []string) error {
			ns, scope := parseArgs(args)
			v := vars.New(ns, scope...)
			if err := v.Init(); err != nil {
				return err
			}
			c.Println("Initialized vars properties")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <name> [scope] <key> <value>",
		Short: "Set a variable for a specific property",
		Args:  cobra.RangeArgs(3, 4),
		RunE: func(c *cobra.Command, args []string) error {
			key := args[len(args)-2]
			val := args[len(args)-1]
			ns, scope := parseArgs(args[:len(args)-2])
			return vars.New(ns, scope...).Set(key, val)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "unset <name> [scope] <key>",
		Short: "Unset a variable property key value",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(c *cobra.Command, args []string) error {
			key := args[len(args)-1]
			ns, scope := parseArgs(args[:len(args)-1])
			return vars.New(ns, scope...).Unset(key)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "data <name> [scope]",
		Short: "Prints all vars for given name",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(c *cobra.Command, args []string) error {
			ns, scope := parseArgs(args)
			data, err := vars.New(ns, scope...).All()
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
		Use:   "edit <name> [scope]",
		Short: "Edit vars file in default editor",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(c *cobra.Command, args []string) error {
			ns, scope := parseArgs(args)
			return vars.New(ns, scope...).Edit()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "keys <name> [scope]",
		Aliases: []string{"k"},
		Short:   "List all keys for given vars name",
		Args:    cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, scope := parseArgs(args)
			data, err := vars.New(ns, scope...).All()
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
		Use:   "get <name> [scope] <key>",
		Short: "Get a variable from a specific vars property value",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			key := args[len(args)-1]
			ns, scope := parseArgs(args[:len(args)-1])
			val, err := vars.New(ns, scope...).Get(key)
			if err != nil {
				return err
			}
			c.Println(val)
			return nil
		},
	})

	return cmd
}

func parseArgs(contextArgs []string) (namespace string, scope []string) {
	namespace = contextArgs[0]
	if len(contextArgs) > 1 {
		scope = []string{contextArgs[1]}
	}
	return namespace, scope
}
