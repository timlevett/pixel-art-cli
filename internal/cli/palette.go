package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// NewPaletteCmd creates the palette parent command with add/list/use subcommands.
func NewPaletteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "palette",
		Short: "Manage named color palettes",
	}
	cmd.AddCommand(newPaletteAddCmd())
	cmd.AddCommand(newPaletteListCmd())
	cmd.AddCommand(newPaletteUseCmd())
	return cmd
}

func newPaletteAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <color...>",
		Short: "Define or replace a named palette",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return invalidArgsf("expected a name and at least one color, got %d args", len(args))
			}
			request := fmt.Sprintf("palette_add %s", strings.Join(args, " "))
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func newPaletteListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [name]",
		Short: "List palette names, or the colors in a named palette",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return invalidArgCount(1, len(args))
			}
			request := "palette_list"
			if len(args) == 1 {
				request = fmt.Sprintf("palette_list %s", args[0])
			}
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func newPaletteUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <name>",
		Short: "Select the active palette for p:<index> references",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return invalidArgCount(1, len(args))
			}
			request := fmt.Sprintf("palette_use %s", args[0])
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}
