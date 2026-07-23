package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewCopyCmd creates the copy command.
func NewCopyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy <x> <y> <w> <h> [clipboard]",
		Short: "Capture a rectangle into a named clipboard slot (default clipboard if omitted)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 4 && len(args) != 5 {
				return invalidArgsf("expected 4 args (or 5 with a trailing clipboard name), got %d", len(args))
			}
			w, err := parseIntArg(args[2], "w")
			if err != nil {
				return err
			}
			if w <= 0 {
				return invalidArgsf("w must be > 0")
			}
			h, err := parseIntArg(args[3], "h")
			if err != nil {
				return err
			}
			if h <= 0 {
				return invalidArgsf("h must be > 0")
			}
			if _, err := parseIntArg(args[0], "x"); err != nil {
				return err
			}
			if _, err := parseIntArg(args[1], "y"); err != nil {
				return err
			}
			request := fmt.Sprintf("copy %s %s %s %s", args[0], args[1], args[2], args[3])
			if len(args) == 5 {
				request += " " + args[4]
			}
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}

// NewPasteCmd creates the paste command.
func NewPasteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "paste <x> <y> [clipboard]",
		Short: "Stamp a clipboard region with its top-left corner at (x,y)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 && len(args) != 3 {
				return invalidArgsf("expected 2 args (or 3 with a trailing clipboard name), got %d", len(args))
			}
			if _, err := parseIntArg(args[0], "x"); err != nil {
				return err
			}
			if _, err := parseIntArg(args[1], "y"); err != nil {
				return err
			}
			request := fmt.Sprintf("paste %s %s", args[0], args[1])
			if len(args) == 3 {
				request += " " + args[2]
			}
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}

// NewMoveCmd creates the move command.
func NewMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <x> <y> <w> <h> <dx> <dy>",
		Short: "Relocate a rectangle by an offset, clearing the source",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 6 {
				return invalidArgCount(6, len(args))
			}
			w, err := parseIntArg(args[2], "w")
			if err != nil {
				return err
			}
			if w <= 0 {
				return invalidArgsf("w must be > 0")
			}
			h, err := parseIntArg(args[3], "h")
			if err != nil {
				return err
			}
			if h <= 0 {
				return invalidArgsf("h must be > 0")
			}
			for i, name := range []string{"x", "y"} {
				if _, err := parseIntArg(args[i], name); err != nil {
					return err
				}
			}
			for i, name := range []string{"dx", "dy"} {
				if _, err := parseIntArg(args[4+i], name); err != nil {
					return err
				}
			}
			request := fmt.Sprintf("move %s %s %s %s %s %s", args[0], args[1], args[2], args[3], args[4], args[5])
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}

// NewMirrorCmd creates the mirror command.
func NewMirrorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mirror <x> <y> <w> <h> <horizontal|vertical>",
		Short: "Flip a rectangle in place (horizontal: left-right, vertical: top-bottom)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 5 {
				return invalidArgCount(5, len(args))
			}
			w, err := parseIntArg(args[2], "w")
			if err != nil {
				return err
			}
			if w <= 0 {
				return invalidArgsf("w must be > 0")
			}
			h, err := parseIntArg(args[3], "h")
			if err != nil {
				return err
			}
			if h <= 0 {
				return invalidArgsf("h must be > 0")
			}
			if _, err := parseIntArg(args[0], "x"); err != nil {
				return err
			}
			if _, err := parseIntArg(args[1], "y"); err != nil {
				return err
			}
			if args[4] != "horizontal" && args[4] != "vertical" {
				return invalidArgsf("axis must be \"horizontal\" or \"vertical\"")
			}
			request := fmt.Sprintf("mirror %s %s %s %s %s", args[0], args[1], args[2], args[3], args[4])
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}
