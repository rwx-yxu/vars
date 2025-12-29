package standalone

import (
	"fmt"
	"os"

	"github.com/rwx-yxu/vars"
	"github.com/spf13/cobra"
)

func Execute() {
	root := newStandaloneCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newStandaloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vars",
		Short: "Manage stateful properties for any application",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "init <name>",
		Short: "initialize empty vars file for <name>",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			vars := vars.New(args[0])
			return vars.Init()
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
