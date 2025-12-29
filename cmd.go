package vars

import (
	"sort"

	"github.com/spf13/cobra"
)

func NewCmd(appName string) *cobra.Command {
	vars := New(appName)

	cmd := &cobra.Command{
		Use:           "vars",
		Short:         "Manage variables for " + appName,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "initialize empty vars file for <name>",
		Args:  cobra.ExactArgs(0),
		RunE: func(c *cobra.Command, _ []string) error {
			vars := New(appName)
			return vars.Init()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a variable",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return vars.Set(args[0], args[1])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "unset <key>",
		Short: "Unset a variable property key value",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return vars.Unset(args[1])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "data",
		Short: "Prints all vars",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			data, err := vars.All()
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
		Use:     "keys",
		Aliases: []string{"k"},
		Short:   "Prints all keys",
		Args:    cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			data, err := vars.All()
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
			val, err := vars.Get(args[0])
			if err != nil {
				return err
			}
			c.Println(val)
			return nil
		},
	})

	return cmd
}
