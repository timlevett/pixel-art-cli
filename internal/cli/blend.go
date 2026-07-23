package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewBlendCmd creates the blend command.
func NewBlendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blend <color1> <color2> <ratio>",
		Short: "Compute a linearly interpolated color between two colors (0=color1, 1=color2)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 3 {
				return invalidArgCount(3, len(args))
			}
			request := fmt.Sprintf("blend %s %s %s", args[0], args[1], args[2])
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}
