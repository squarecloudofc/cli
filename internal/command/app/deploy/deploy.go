// Package deploy implements `app deploy` — deployments and Git integrations.
package deploy

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
)

func NewCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: squareCli.I18n().T("metadata.commands.app.deploy.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newListCommand(squareCli),
		newCurrentCommand(squareCli),
		newWebhookCommand(squareCli),
		newGithubCommand(squareCli),
	)

	return cmd
}

func newListCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "list [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.deploy.list.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(squareCli, args)
			if err != nil {
				return err
			}

			deploys, err := squareCli.Rest().GetApplicationDeployments(appID)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), deploys)
			}

			if len(deploys) == 0 {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.deploy.list.empty"))
				return nil
			}

			w := tabwriter.NewWriter(squareCli.Out(), 0, 0, 2, ' ', tabwriter.TabIndent)
			defer w.Flush()

			fmt.Fprintln(w, strings.Join([]string{"COMMIT", "DATE", "STATE", "BRANCH"}, " \t "))

			// Each outer element is one deploy; its inner timeline shares the
			// commit SHA. Show the latest event of each deploy.
			for _, timeline := range deploys {
				if len(timeline) == 0 {
					continue
				}

				last := timeline[len(timeline)-1]
				branch := "-"
				for _, event := range timeline {
					if event.Branch != "" {
						branch = event.Branch
					}
				}

				sha := last.ID
				if len(sha) > 8 {
					sha = sha[:8]
				}

				fmt.Fprintf(w, "%s \t %s \t %s \t %s \t\n",
					sha, last.Date.Local().Format(time.DateTime), last.State, branch)
			}

			return nil
		},
	}
}

func newCurrentCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "current [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.deploy.current.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(squareCli, args)
			if err != nil {
				return err
			}

			current, err := squareCli.Rest().GetApplicationCurrentDeployment(appID)
			if err != nil {
				return err
			}

			if cmdutil.WantsJSON(cmd) {
				return cmdutil.PrintJSON(squareCli.Out(), current)
			}

			if current.App == nil && current.Webhook == "" {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.deploy.current.empty"))
				return nil
			}

			if current.App != nil {
				name, branch := "-", "-"
				if current.App.Name != nil {
					name = *current.App.Name
				}
				if current.App.Branch != nil {
					branch = *current.App.Branch
				}

				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.deploy.current.github_app", map[string]any{
					"Repository": name,
					"Branch":     branch,
				}))
			}

			if current.Webhook != "" {
				fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.deploy.current.webhook", map[string]any{
					"Webhook": current.Webhook,
				}))
			}

			return nil
		},
	}
}

func newWebhookCommand(squareCli cli.SquareCLI) *cobra.Command {
	var appID string
	var remove bool

	cmd := &cobra.Command{
		Use:   "webhook <github token>",
		Short: squareCli.I18n().T("metadata.commands.app.deploy.webhook.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := cmdutil.ResolveAppID(squareCli, argsFromID(appID))
			if err != nil {
				return err
			}

			token := "@"
			if !remove {
				if len(args) == 0 {
					return cmd.Help()
				}
				token = args[0]
			}

			result, err := squareCli.Rest().PostApplicationDeployWebhook(id, token)
			if err != nil {
				return err
			}

			if remove {
				fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.deploy.webhook.removed"))
				return nil
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.deploy.webhook.success"))
			fmt.Fprintf(squareCli.Out(), "  %s\n", result.Webhook)
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", "Application ID")
	cmd.Flags().BoolVar(&remove, "remove", false, "Remove the configured webhook")
	return cmd
}

func newGithubCommand(squareCli cli.SquareCLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "github",
		Short: squareCli.I18n().T("metadata.commands.app.deploy.github.root.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	var linkApp string
	link := &cobra.Command{
		Use:   "link <owner/repository> <branch>",
		Short: squareCli.I18n().T("metadata.commands.app.deploy.github.link.short"),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := cmdutil.ResolveAppID(squareCli, argsFromID(linkApp))
			if err != nil {
				return err
			}

			result, err := squareCli.Rest().LinkApplicationGithubApp(id, args[0], args[1])
			if err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.deploy.github.link.success", map[string]any{
				"Repository": result.Repository.FullName,
				"Branch":     result.Repository.Branch,
			}))
			return nil
		},
	}
	link.Flags().StringVar(&linkApp, "app", "", "Application ID")

	var unlinkApp string
	unlink := &cobra.Command{
		Use:   "unlink",
		Short: squareCli.I18n().T("metadata.commands.app.deploy.github.unlink.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := cmdutil.ResolveAppID(squareCli, argsFromID(unlinkApp))
			if err != nil {
				return err
			}

			if err := squareCli.Rest().UnlinkApplicationGithubApp(id); err != nil {
				return err
			}

			fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.CheckMark, squareCli.I18n().T("commands.app.deploy.github.unlink.success"))
			return nil
		},
	}
	unlink.Flags().StringVar(&unlinkApp, "app", "", "Application ID")

	cmd.AddCommand(link, unlink)
	return cmd
}

func argsFromID(appID string) []string {
	if appID == "" {
		return nil
	}
	return []string{appID}
}
