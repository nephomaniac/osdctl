//go:build integrationtest
// +build integrationtest

package integration

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewCmdIntegrationTests implements the base integrationtests command
func NewCmdIntegrationTests(streams genericclioptions.IOStreams) *cobra.Command {
	integrationTestsCmd := &cobra.Command{
		Use:               "integrationtests",
		Short:             "Integration test utilities for OSDCTL",
		Args:              cobra.NoArgs,
		DisableAutoGenTag: true,
	}

	integrationTestsCmd.AddCommand(newCmdLoginTests())
	return integrationTestsCmd
}
