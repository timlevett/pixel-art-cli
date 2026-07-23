package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewLayerCmd creates the layer parent command with add/list/select/visible subcommands.
func NewLayerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "layer",
		Short: "Manage drawing layers (composited on export, base first)",
	}
	cmd.AddCommand(newLayerAddCmd())
	cmd.AddCommand(newLayerListCmd())
	cmd.AddCommand(newLayerSelectCmd())
	cmd.AddCommand(newLayerVisibleCmd())
	return cmd
}

func newLayerAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new blank, visible layer (\"base\" always exists and can't be re-added)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return invalidArgCount(1, len(args))
			}
			return sendCommandRequest(cmd, fmt.Sprintf("layer_add %s", args[0]))
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func newLayerListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List layer names in creation order (\"base\" first)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return invalidArgCount(0, len(args))
			}
			return sendCommandRequest(cmd, "layer_list")
		},
	}
	return cmd
}

func newLayerSelectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select <name>",
		Short: "Select the active layer: drawing, get_pixel, inspect, undo, and redo all act on it",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return invalidArgCount(1, len(args))
			}
			return sendCommandRequest(cmd, fmt.Sprintf("layer_select %s", args[0]))
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func newLayerVisibleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "visible <name> <true|false>",
		Short: "Include or exclude a layer when export flattens all layers",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return invalidArgCount(2, len(args))
			}
			if args[1] != "true" && args[1] != "false" {
				return invalidArgsf("visibility must be \"true\" or \"false\"")
			}
			return sendCommandRequest(cmd, fmt.Sprintf("layer_visible %s %s", args[0], args[1]))
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}
