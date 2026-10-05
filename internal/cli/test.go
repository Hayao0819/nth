package cli

import "github.com/spf13/cobra"

func newTestCommand(version string, deps dependencies) *cobra.Command {
	return &cobra.Command{
		Use:           "test",
		Short:         "Test live API access",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return runTest(command, version, deps)
		},
	}
}

func runTest(command *cobra.Command, version string, deps dependencies) error {
	settings, proceed, err := configure(command.Context(), deps, false)
	if err != nil || !proceed {
		return err
	}
	session, err := openSession(command.Context(), version, deps.credentials, settings)
	if err != nil {
		return err
	}

	return deps.diagnose(command.Context(), session.api)
}
