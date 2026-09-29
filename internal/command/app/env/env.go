// Package env implements `app env` — environment variable management.
package env

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v3"
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
		Use:         "list [app id]",
		Short:       squareCli.I18n().T("metadata.commands.app.env.list.short"),
		Annotations: map[string]string{"json": "true"},
		Args:        cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			envs, err := squareCli.API().Apps.Envs.Get(cmd.Context(), appID)
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
			envs, err := collectEnvs(squareCli, args, fromFile)
			if err != nil {
				return err
			}
			if len(envs) == 0 {
				return cmd.Help()
			}

			id, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, argsFromID(appID))
			if err != nil {
				return err
			}

			if _, err := squareCli.API().Apps.Envs.Set(cmd.Context(), id, envs); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.env.set.success", map[string]any{"Count": len(envs)}))
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", squareCli.I18n().T("metadata.flags.app"))
	cmd.Flags().StringVar(&fromFile, "from-file", "", squareCli.I18n().T("metadata.flags.from_file"))
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
			envs, err := collectEnvs(squareCli, args, fromFile)
			if err != nil {
				return err
			}
			// An empty set would delete every variable; that is remove --all.
			if len(envs) == 0 {
				return errors.New(squareCli.I18n().T("commands.app.env.replace.empty"))
			}

			id, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, argsFromID(appID))
			if err != nil {
				return err
			}

			message := squareCli.I18n().T("commands.app.env.replace.confirm", map[string]any{"Appid": id})
			if err := cmdutil.Confirm(squareCli, message, yes); err != nil {
				return err
			}

			if _, err := squareCli.API().Apps.Envs.Replace(cmd.Context(), id, envs); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.env.replace.success"))
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", squareCli.I18n().T("metadata.flags.app"))
	cmd.Flags().StringVar(&fromFile, "from-file", "", squareCli.I18n().T("metadata.flags.from_file"))
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes"))
	return cmd
}

func newRemoveCommand(squareCli cli.SquareCLI) *cobra.Command {
	var appID string
	var all, yes bool

	cmd := &cobra.Command{
		Use:     "remove KEY... | --all",
		Aliases: []string{"delete", "unset"},
		Short:   squareCli.I18n().T("metadata.commands.app.env.remove.short"),
		Args: func(cmd *cobra.Command, args []string) error {
			if all == (len(args) > 0) { // keys or --all, exactly one
				return errors.New(squareCli.I18n().T("commands.app.env.remove.usage"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, argsFromID(appID))
			if err != nil {
				return err
			}

			if all {
				message := squareCli.I18n().T("commands.app.env.remove.confirm_all", map[string]any{"Appid": id})
				if err := cmdutil.Confirm(squareCli, message, yes); err != nil {
					return err
				}
				if _, err := squareCli.API().Apps.Envs.Replace(cmd.Context(), id, squarecloud.EnvVars{}); err != nil {
					return err
				}
				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.env.remove.success_all"))
				return nil
			}

			if _, err := squareCli.API().Apps.Envs.Delete(cmd.Context(), id, args...); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.env.remove.success", map[string]any{"Count": len(args)}))
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", squareCli.I18n().T("metadata.flags.app"))
	cmd.Flags().BoolVar(&all, "all", false, squareCli.I18n().T("metadata.flags.remove_all_envs"))
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, squareCli.I18n().T("metadata.flags.yes"))
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
func collectEnvs(squareCli cli.SquareCLI, args []string, fromFile string) (squarecloud.EnvVars, error) {
	envs := squarecloud.EnvVars{}

	if fromFile != "" {
		content, err := os.ReadFile(fromFile)
		if err != nil {
			return nil, err
		}
		parseDotEnv(string(content), envs)
	}

	for _, arg := range args {
		key, value, found := strings.Cut(arg, "=")
		if !found {
			return nil, errors.New(squareCli.I18n().T("commands.app.env.invalid", map[string]any{"Variable": arg}))
		}

		envs[key] = value
	}

	return envs, nil
}

// parseDotEnv adds the KEY=VALUE lines of a .env file to envs. Blank lines,
// # comments and lines without "=" are skipped; "export " is allowed. A value
// in matching quotes is taken as is between them; an unquoted value ends at
// " #" (a comment) and is otherwise kept verbatim, stray quotes included.
// ponytail: no escape sequences or multi-line values; add if asked.
func parseDotEnv(content string, envs squarecloud.EnvVars) {
	// A UTF-8 BOM (Windows editors) would otherwise prefix the first key.
	for _, line := range strings.Split(strings.TrimPrefix(content, "\xef\xbb\xbf"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(strings.TrimPrefix(line, "export "), "=")
		if !found {
			continue
		}
		envs[strings.TrimSpace(key)] = dotEnvValue(strings.TrimSpace(value))
	}
}

func dotEnvValue(value string) string {
	if value != "" && (value[0] == '"' || value[0] == '\'') {
		if end := strings.IndexByte(value[1:], value[0]); end >= 0 {
			rest := strings.TrimSpace(value[end+2:])
			if rest == "" || rest[0] == '#' {
				return value[1 : end+1]
			}
		}
	}

	for i := 1; i < len(value); i++ {
		if value[i] == '#' && (value[i-1] == ' ' || value[i-1] == '\t') {
			return strings.TrimSpace(value[:i])
		}
	}
	return value
}
