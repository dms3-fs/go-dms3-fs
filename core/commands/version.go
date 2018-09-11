package commands

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	version "github.com/dms3-fs/go-dms3-fs"
	cmds "github.com/dms3-fs/go-dms3-fs/commands"
	e "github.com/dms3-fs/go-dms3-fs/core/commands/e"
	fsrepo "github.com/dms3-fs/go-dms3-fs/repo/fsrepo"

	"github.com/dms3-fs/go-fs-cmdkit"
)

type VersionOutput struct {
	Version string
	Commit  string
	Repo    string
	System  string
	Golang  string
}

var VersionCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline:          "Show dms3fs version information.",
		ShortDescription: "Returns the current version of dms3fs and exits.",
	},

	Options: []cmdkit.Option{
		cmdkit.BoolOption("number", "n", "Only show the version number."),
		cmdkit.BoolOption("commit", "Show the commit hash."),
		cmdkit.BoolOption("repo", "Show repo version."),
		cmdkit.BoolOption("all", "Show all version information"),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		res.SetOutput(&VersionOutput{
			Version: version.CurrentVersionNumber,
			Commit:  version.CurrentCommit,
			Repo:    fmt.Sprint(fsrepo.RepoVersion),
			System:  runtime.GOARCH + "/" + runtime.GOOS, //TODO: Precise version here
			Golang:  runtime.Version(),
		})
	},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			v, err := unwrapOutput(res.Output())
			if err != nil {
				return nil, err
			}

			version, ok := v.(*VersionOutput)
			if !ok {
				return nil, e.TypeErr(version, v)
			}

			repo, _, err := res.Request().Option("repo").Bool()
			if err != nil {
				return nil, err
			}

			if repo {
				return strings.NewReader(version.Repo + "\n"), nil
			}

			commit, _, err := res.Request().Option("commit").Bool()
			commitTxt := ""
			if err != nil {
				return nil, err
			}
			if commit {
				commitTxt = "-" + version.Commit
			}

			number, _, err := res.Request().Option("number").Bool()
			if err != nil {
				return nil, err
			}
			if number {
				return strings.NewReader(fmt.Sprintln(version.Version + commitTxt)), nil
			}

			all, _, err := res.Request().Option("all").Bool()
			if err != nil {
				return nil, err
			}
			if all {
				out := fmt.Sprintf("go-dms3-fs version: %s-%s\n"+
					"Repo version: %s\nSystem version: %s\nGolang version: %s\n",
					version.Version, version.Commit, version.Repo, version.System, version.Golang)
				return strings.NewReader(out), nil
			}

			return strings.NewReader(fmt.Sprintf("dms3-fs version %s%s\n", version.Version, commitTxt)), nil
		},
	},
	Type: VersionOutput{},
}
