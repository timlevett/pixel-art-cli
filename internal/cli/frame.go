package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewFrameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frame",
		Short: "Manage animation frames (each with its own independent layer stack)",
	}
	cmd.AddCommand(newFrameAddCmd())
	cmd.AddCommand(newFrameListCmd())
	cmd.AddCommand(newFrameSelectCmd())
	cmd.AddCommand(newFrameGhostCmd())
	return cmd
}

func newFrameAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Create a new blank frame and print its index (frame 0 always exists and can't be re-added)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return invalidArgCount(0, len(args))
			}
			return sendCommandRequest(cmd, "frame_add")
		},
	}
	return cmd
}

func newFrameListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List frame indices in creation order (\"0\" first)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return invalidArgCount(0, len(args))
			}
			return sendCommandRequest(cmd, "frame_list")
		},
	}
	return cmd
}

func newFrameSelectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select <index>",
		Short: "Select the active frame: drawing, layer_*, get_pixel, inspect, undo, redo, and export all act on it",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return invalidArgCount(1, len(args))
			}
			if _, err := parseIntArg(args[0], "index"); err != nil {
				return err
			}
			return sendCommandRequest(cmd, fmt.Sprintf("frame_select %s", args[0]))
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func newFrameGhostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ghost <index> [opacity]",
		Short: "Dump a text grid (like inspect) of the active frame with another frame ghosted underneath at reduced opacity (default 0.35), for onion-skinning",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 && len(args) != 2 {
				return invalidArgsf("expected 1 arg (index) or 2 args (index opacity), got %d", len(args))
			}
			if _, err := parseIntArg(args[0], "index"); err != nil {
				return err
			}
			request := fmt.Sprintf("frame_ghost %s", args[0])
			if len(args) == 2 {
				request = fmt.Sprintf("frame_ghost %s %s", args[0], args[1])
			}

			socketPath, err := SocketPath(cmd)
			if err != nil {
				return err
			}
			cli, err := drawNewClient(socketPath)
			if err != nil {
				return err
			}
			resp, err := cli.Send(request)
			if err != nil {
				return formatClientError(err)
			}
			printInspectResponse(cmd, resp.Raw)
			return nil
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}
