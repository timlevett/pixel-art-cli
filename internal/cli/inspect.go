package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// NewInspectCmd creates the inspect command.
func NewInspectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect [x y w h]",
		Short: "Dump the canvas (or a sub-region) as a text grid of colors, one row per line",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 && len(args) != 4 {
				return invalidArgsf("expected 0 args (whole canvas) or 4 args (x y w h), got %d", len(args))
			}
			for i, name := range []string{"x", "y", "w", "h"} {
				if i >= len(args) {
					break
				}
				if _, err := parseIntArg(args[i], name); err != nil {
					return err
				}
			}

			request := "inspect"
			if len(args) == 4 {
				request = fmt.Sprintf("inspect %s %s %s %s", args[0], args[1], args[2], args[3])
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

// printInspectResponse renders an "ok <row>;<row>;..." response as one
// space-separated row of colors per output line. Non-"ok" responses (daemon
// errors) are printed as-is.
func printInspectResponse(cmd *cobra.Command, raw string) {
	out := cmd.OutOrStdout()
	payload := strings.TrimPrefix(raw, "ok")
	if payload == raw || strings.TrimSpace(payload) == "" {
		_, _ = fmt.Fprintln(out, raw)
		return
	}
	for _, row := range strings.Split(strings.TrimSpace(payload), ";") {
		_, _ = fmt.Fprintln(out, strings.ReplaceAll(row, ",", " "))
	}
}
