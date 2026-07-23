package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"pxcli/internal/client"
)

type scriptSender interface {
	SendScript(lines []string) (client.Response, error)
}

type scriptClientFactory func(socketPath string) (scriptSender, error)

var scriptNewClient scriptClientFactory = func(socketPath string) (scriptSender, error) {
	return client.New(socketPath)
}

// NewScriptCmd creates the script command.
func NewScriptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "script [file]",
		Short: "Execute a batch of drawing commands from a file (or stdin) as one undoable step",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var source io.Reader
			if len(args) == 1 && args[0] != "-" {
				file, err := os.Open(args[0])
				if err != nil {
					return invalidArgsf("unable to open file: %v", err)
				}
				defer file.Close()
				source = file
			} else {
				source = cmd.InOrStdin()
			}

			lines, err := readScriptLines(source)
			if err != nil {
				return invalidArgsf("unable to read script: %v", err)
			}

			socketPath, err := SocketPath(cmd)
			if err != nil {
				return err
			}
			cli, err := scriptNewClient(socketPath)
			if err != nil {
				return err
			}
			resp, err := cli.SendScript(lines)
			if err != nil {
				return formatClientError(err)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), resp.Raw)
			return nil
		},
	}
	cmd.Flags().SetInterspersed(false)

	return cmd
}

func readScriptLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
