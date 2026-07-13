// Package env implements `app env` — environment variable management.
package env

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v2/squarecloud"
)

func NewCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: squareCli.I18n().T("metadata.commands.app.env.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newListCommand(squareCli),
		newSetCommand(squareCli),
		newRemoveCommand(squareCli),
		newReplaceCommand(squareCli),
	)

	return cmd
}

func newListCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "list [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.env.list.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(squareCli, args)
			if err != nil {
				return err
			}

			envs, err := squareCli.Rest().GetApplicationEnvs(appID)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), envs)
			}

			if len(envs) == 0 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.env.list.empty"))
				return nil
			}

			keys := make([]string, 0, len(envs))
			for key := range envs {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			for _, key := range keys {
				fmt.Fprintf(squareCli.Out(), "%s=%s\n", key, envs[key])
			}

			return nil
		},
	}
}

func newSetCommand(squareCli cli.SquareCLI) *cobra.Command {
	var appID string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "set KEY=VALUE...",
		Short: squareCli.I18n().T("metadata.commands.app.env.set.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			envs, err := collectEnvs(args, fromFile)
			if err != nil {
				return err
			}
			if len(envs) == 0 {
				return cmd.Help()
			}

			id, err := cmdutil.ResolveAppID(squareCli, argsFromID(appID))
			if err != nil {
				return err
			}

			if _, err := squareCli.Rest().SetApplicationEnvs(id, envs); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.env.set.success", map[string]any{"Count": len(envs)}))
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", "Application ID")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Read variables from a .env style file")
	return cmd
}

func newReplaceCommand(squareCli cli.SquareCLI) *cobra.Command {
	var appID string
	var fromFile string
	var yes bool

	cmd := &cobra.Command{
		Use:   "replace KEY=VALUE...",
		Short: squareCli.I18n().T("metadata.commands.app.env.replace.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			envs, err := collectEnvs(args, fromFile)
			if err != nil {
				return err
			}

			id, err := cmdutil.ResolveAppID(squareCli, argsFromID(appID))
			if err != nil {
				return err
			}

			message := squareCli.I18n().T("commands.app.env.replace.confirm", map[string]any{"Appid": id})
			if !cmdutil.Confirm(squareCli, message, yes) {
				return nil
			}

			if _, err := squareCli.Rest().ReplaceApplicationEnvs(id, envs); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.env.replace.success"))
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", "Application ID")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Read variables from a .env style file")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
	return cmd
}

func newRemoveCommand(squareCli cli.SquareCLI) *cobra.Command {
	var appID string

	cmd := &cobra.Command{
		Use:     "remove KEY...",
		Aliases: []string{"delete", "unset"},
		Short:   squareCli.I18n().T("metadata.commands.app.env.remove.short"),
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := cmdutil.ResolveAppID(squareCli, argsFromID(appID))
			if err != nil {
				return err
			}

			if err := squareCli.Rest().DeleteApplicationEnvs(id, args); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.env.remove.success", map[string]any{"Count": len(args)}))
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", "Application ID")
	return cmd
}

// argsFromID adapts an optional --app flag to the args slice shape
// ResolveAppID expects.
func argsFromID(appID string) []string {
	if appID == "" {
		return nil
	}
	return []string{appID}
}

// collectEnvs merges KEY=VALUE arguments with an optional .env style file.
func collectEnvs(args []string, fromFile string) (squarecloud.EnvVars, error) {
	envs := squarecloud.EnvVars{}

	if fromFile != "" {
		content, err := os.ReadFile(fromFile)
		if err != nil {
			return nil, err
		}

		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			key, value, found := strings.Cut(line, "=")
			if !found {
				continue
			}

			envs[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
		}
	}

	for _, arg := range args {
		key, value, found := strings.Cut(arg, "=")
		if !found {
			return nil, fmt.Errorf("invalid variable %q, expected KEY=VALUE", arg)
		}

		envs[key] = value
	}

	return envs, nil
}
