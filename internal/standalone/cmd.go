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
		Use:   "init <name>",
		Short: "Initialize storage (Required before use)",
		Args:  cobra.ExactArgs(0),
		RunE: func(c *cobra.Command, args []string) error {
			v := vars.New(args[0])
			if err := v.Init(); err != nil {
				return err
			}
			c.Println("Initialized vars properties")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <name> <key> <value>",
		Short: "Set a variable for a specific property",
		Args:  cobra.ExactArgs(3),
		RunE: func(c *cobra.Command, args []string) error {
			store := vars.New(args[0])
			return store.Set(args[1], args[2])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "unset <name> <key>",
		Short: "Unset a variable property key value",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			v := vars.New(args[0])
			return v.Unset(args[1])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "data <name>",
		Short: "Prints all vars for given name",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			v := vars.New(args[0])
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
		Use:     "keys <name>",
		Aliases: []string{"k"},
		Short:   "List all keys for given vars name",
		Args:    cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			v := vars.New(args[0])
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
		Use:   "get <name> <key>",
		Short: "Get a variable from a specific vars property value",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			store := vars.New(args[0])
			val, err := store.Get(args[1])
			if err != nil {
				return err
			}
			c.Println(val)
			return nil
		},
	})

	return cmd
}
