package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestShouldCheckAuth(t *testing.T) {
	root := &cobra.Command{Use: "squarecloud"}
	app := &cobra.Command{Use: "app", RunE: func(*cobra.Command, []string) error { return nil }}
	list := &cobra.Command{Use: "list", RunE: func(*cobra.Command, []string) error { return nil }}
	login := &cobra.Command{Use: "login", Annotations: map[string]string{"skipAuthCheck": "true"}}
	app.AddCommand(list)
	root.AddCommand(app, login)

	f := &fakeCLI{}
	for cmd, want := range map[*cobra.Command]bool{app: false, list: true, login: false} {
		if got := ShouldCheckAuth(f, cmd); got != want {
			t.Errorf("ShouldCheckAuth(%s) = %v, want %v", cmd.CommandPath(), got, want)
		}
	}
}
