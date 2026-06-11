package cluster

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/openshift/osdctl/internal/utils/globalflags"
)

func newCmdPullSecret(streams genericclioptions.IOStreams, globalOpts *globalflags.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "pull-secret",
		Short:             "Diagnose and manage cluster pull secrets",
		Long:              "Subcommands for inspecting and replacing cluster pull secrets.",
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newCmdPullSecretAudit(streams, globalOpts))
	cmd.AddCommand(newCmdPullSecretUpdate(streams, globalOpts))
	cmd.AddCommand(newCmdPullSecretValidate())

	return cmd
}

func newCmdPullSecretValidate() *cobra.Command {
	cmd := newCmdValidatePullSecretExt()
	cmd.Use = "validate"
	return cmd
}
