package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func configCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config [key] [value]",
		Aliases: []string{"c"},
		Short:   "Set cylinder configuration environment",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set(args[0], args[1])
			return viper.WriteConfig()
		},
	}

	return cmd
}
