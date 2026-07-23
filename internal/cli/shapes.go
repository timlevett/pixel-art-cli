package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewCircleCmd creates the circle command.
func NewCircleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "circle <cx> <cy> <r> <color> [fill]",
		Short: "Draw a circle outline, or a filled disk with the optional fill argument",
		RunE: func(cmd *cobra.Command, args []string) error {
			filled := len(args) == 5 && args[4] == "fill"
			if len(args) != 4 && !filled {
				return invalidArgsf("expected 4 args (or 5 with trailing \"fill\"), got %d", len(args))
			}
			if _, err := parseIntArg(args[0], "cx"); err != nil {
				return err
			}
			if _, err := parseIntArg(args[1], "cy"); err != nil {
				return err
			}
			r, err := parseIntArg(args[2], "r")
			if err != nil {
				return err
			}
			if r <= 0 {
				return invalidArgsf("r must be > 0")
			}
			request := fmt.Sprintf("circle %s %s %s %s", args[0], args[1], args[2], args[3])
			if filled {
				request += " fill"
			}
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}

// NewEllipseCmd creates the ellipse command.
func NewEllipseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ellipse <cx> <cy> <rx> <ry> <color> [fill]",
		Short: "Draw an ellipse outline, or a filled region with the optional fill argument",
		RunE: func(cmd *cobra.Command, args []string) error {
			filled := len(args) == 6 && args[5] == "fill"
			if len(args) != 5 && !filled {
				return invalidArgsf("expected 5 args (or 6 with trailing \"fill\"), got %d", len(args))
			}
			if _, err := parseIntArg(args[0], "cx"); err != nil {
				return err
			}
			if _, err := parseIntArg(args[1], "cy"); err != nil {
				return err
			}
			rx, err := parseIntArg(args[2], "rx")
			if err != nil {
				return err
			}
			if rx <= 0 {
				return invalidArgsf("rx must be > 0")
			}
			ry, err := parseIntArg(args[3], "ry")
			if err != nil {
				return err
			}
			if ry <= 0 {
				return invalidArgsf("ry must be > 0")
			}
			request := fmt.Sprintf("ellipse %s %s %s %s %s", args[0], args[1], args[2], args[3], args[4])
			if filled {
				request += " fill"
			}
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}

// NewDitherFillCmd creates the dither_fill command.
func NewDitherFillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dither_fill <x> <y> <w> <h> <color1> <color2> [pattern]",
		Short: "Fill a rectangle by alternating two colors (checkerboard, horizontal, or vertical)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 6 && len(args) != 7 {
				return invalidArgsf("expected 6 args (or 7 with a trailing pattern name), got %d", len(args))
			}
			if _, err := parseIntArg(args[0], "x"); err != nil {
				return err
			}
			if _, err := parseIntArg(args[1], "y"); err != nil {
				return err
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
			request := fmt.Sprintf("dither_fill %s %s %s %s %s %s", args[0], args[1], args[2], args[3], args[4], args[5])
			if len(args) == 7 {
				request += " " + args[6]
			}
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}
