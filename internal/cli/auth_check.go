package cli

import (
	"reflect"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// thank you https://github.com/cli/cli/blob/trunk/pkg/cmdutil/auth_check.go#L41
func ShouldCheckAuth(squareCli *SquareCli, cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "help", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		return false
	}

	for c := cmd; c.Parent() != nil; c = c.Parent() {
		if c.Annotations != nil && c.Annotations["skipAuthCheck"] == "true" {
			return false
		}

		var skipAuthCheck bool
		c.Flags().Visit(func(f *pflag.Flag) {
			if f.Annotations != nil && reflect.DeepEqual(f.Annotations["skipAuthCheck"], []string{"true"}) {
				skipAuthCheck = true
			}
		})
		if skipAuthCheck {
			return false
		}
	}
	return true
}

func CheckAuth(squareCli *SquareCli) bool {
	return squareCli.Config.AuthToken != ""
}
