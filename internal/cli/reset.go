package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newResetCommand(deps dependencies) *cobra.Command {
	return &cobra.Command{
		Use:           "reset",
		Short:         "Delete local settings and credentials",
		Long:          "Delete all nth entries from the system keyring. Environment variables, browser data, and server-side credentials are not changed.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			result, err := deps.credentials.Reset()
			if err != nil {
				return err
			}

			fmt.Fprintln(command.OutOrStdout(), "Removed nth settings and credentials from the system keyring.")
			if len(result.Environment) != 0 {
				fmt.Fprintf(
					command.ErrOrStderr(),
					"Warning: environment variables remain set: %s. Remove them separately to complete the reset.\n",
					strings.Join(result.Environment, ", "),
				)
			}

			return nil
		},
	}
}
