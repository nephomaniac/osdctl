//go:build !integrationtest
// +build !integrationtest

package integration

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewCmdIntegrationTests returns nil when integrationtest build tag is not set
func NewCmdIntegrationTests(streams genericclioptions.IOStreams) *cobra.Command {
	return nil
}
