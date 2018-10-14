package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	cmds "github.com/dms3-fs/go-dms3-fs/commands"
	e "github.com/dms3-fs/go-dms3-fs/core/commands/e"
	repo "github.com/dms3-fs/go-dms3-fs/repo"
	fsrepo "github.com/dms3-fs/go-dms3-fs/repo/fsrepo"

	"github.com/dms3-fs/go-fs-cmdkit"
	idxconfig "github.com/dms3-fs/go-idx-config"
	logging "github.com/dms3-fs/go-log"
)

// clog is the command logger
var clog = logging.Logger("index/config")

type IdxConfigField struct {
	Key   string
	Value interface{}
}

/*
	Note: we want the index commands packaged in /index subfolder.
	However, its IdxConfigCmd subcommand needs wants to be in commands
	package because of dependency on private function unwrapOutput.
*/
var IndexCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Manage DMS3FS index data repository.",
	},
}

var IdxConfigCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Get and set dms3fs index config values.",
		ShortDescription: `
'dms3fs index config' controls configuration variables. It works like 'git config'.
The configuration values are stored in a config file inside your dms3fs
repository.`,
		LongDescription: `
'dms3fs index config' controls configuration variables. It works
much like 'git config'. The configuration values are stored in a config
file inside your DMS3FS repository.

Examples:

Get the value of the 'Corpus.Path' key:

  $ dms3fs index config Corpus.Path

Set the value of the 'Corpus.Path' key:

  $ dms3fs index config Corpus.Path ~/.dms3-fs/index/repo/corpus
`,
	},

	Arguments: []cmdkit.Argument{
		//cmdkit.StringArg("kind", true, false, "The kind of content (e.g. \"blog\")."),
		cmdkit.StringArg("key", true, false, "The key of the config entry (e.g. \"Corpus.Path\")."),
		cmdkit.StringArg("value", false, false, "The value to set the config entry to."),
	},
	Options: []cmdkit.Option{
		cmdkit.BoolOption("bool", "Set a boolean value."),
		cmdkit.BoolOption("json", "Parse stringified JSON."),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		args := req.Arguments()
		kind := args[0]
		key := args[0]

		clog.Debugf("kind value is %s, key is %s", kind, key)

		var output *IdxConfigField
		defer func() {
			if output != nil {
				res.SetOutput(output)
			} else {
				res.SetOutput(nil)
			}
		}()

		// This is a temporary fix until we move the private key out of the config file
		switch strings.ToLower(key) {
		case "identity", "identity.privkey":
			res.SetError(fmt.Errorf("cannot show or change private key through API"), cmdkit.ErrNormal)
			return
		default:
		}

		r, err := fsrepo.Open(req.InvocContext().ConfigRoot)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
		defer r.Close()
		if len(args) == 2 {
			value := args[1]

			if parseJson, _, _ := req.Option("json").Bool(); parseJson {
				var jsonVal interface{}
				if err := json.Unmarshal([]byte(value), &jsonVal); err != nil {
					err = fmt.Errorf("failed to unmarshal json. %s", err)
					res.SetError(err, cmdkit.ErrNormal)
					return
				}

				output, err = setIdxConfig(r, key, jsonVal)
			} else if isbool, _, _ := req.Option("bool").Bool(); isbool {
				output, err = setIdxConfig(r, key, value == "true")
			} else {
				output, err = setIdxConfig(r, key, value)
			}
		} else {
			output, err = getIdxConfig(r, key)
		}
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
	},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			if len(res.Request().Arguments()) == 2 {
				return nil, nil // dont output anything
			}

			if res.Error() != nil {
				return nil, res.Error()
			}

			v, err := unwrapOutput(res.Output())
			if err != nil {
				return nil, err
			}

			vf, ok := v.(*IdxConfigField)
			if !ok {
				return nil, e.TypeErr(vf, v)
			}

			buf, err := idxconfig.HumanOutput(vf.Value)
			if err != nil {
				return nil, err
			}
			buf = append(buf, byte('\n'))
			return bytes.NewReader(buf), nil
		},
	},
	Type: IdxConfigField{},
	Subcommands: map[string]*cmds.Command{
		"show":    idxconfigShowCmd,
		"edit":    idxconfigEditCmd,
		"replace": idxconfigReplaceCmd,
	},
}

var idxconfigShowCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Output index config file contents.",
		ShortDescription: `
Display index config file contents.
`,
	},
	Type: map[string]interface{}{},
	Run: func(req cmds.Request, res cmds.Response) {

		cfgPath := req.InvocContext().ConfigRoot
		fname, err := idxconfig.Filename(filepath.Join(cfgPath, fsrepo.IdxPath))
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		data, err := ioutil.ReadFile(fname)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		var cfg map[string]interface{}
		err = json.Unmarshal(data, &cfg)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		res.SetOutput(&cfg)
	},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			if res.Error() != nil {
				return nil, res.Error()
			}

			v, err := unwrapOutput(res.Output())
			if err != nil {
				return nil, err
			}

			cfg, ok := v.(*map[string]interface{})
			if !ok {
				return nil, e.TypeErr(cfg, v)
			}

			buf, err := idxconfig.HumanOutput(cfg)
			if err != nil {
				return nil, err
			}
			buf = append(buf, byte('\n'))
			return bytes.NewReader(buf), nil
		},
	},
}

var idxconfigEditCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Open the config file for editing in $EDITOR.",
		ShortDescription: `
To use 'dms3fs index config edit', you must have the $EDITOR environment
variable set to your preferred text editor.
`,
	},

	Run: func(req cmds.Request, res cmds.Response) {
		filename, err := idxconfig.Filename(filepath.Join(req.InvocContext().ConfigRoot, fsrepo.IdxPath))
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		err = editIdxConfig(filename)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
		}
	},
}

var idxconfigReplaceCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Replace the config with <file>.",
		ShortDescription: `
Make sure to back up the index config file first if necessary, as this operation
can't be undone.
`,
	},

	Arguments: []cmdkit.Argument{
		cmdkit.FileArg("file", true, false, "The file to use as the new config."),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		// has to be called
		res.SetOutput(nil)

		r, err := fsrepo.Open(req.InvocContext().ConfigRoot)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
		defer r.Close()

		file, err := req.Files().NextFile()
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
		defer file.Close()

		err = replaceIdxConfig(r, file)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
	},
}

func getIdxConfig(r repo.Repo, key string) (*IdxConfigField, error) {
	value, err := r.GetIdxConfigKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get config value: %q", err)
	}
	return &IdxConfigField{
		Key:   key,
		Value: value,
	}, nil
}

func setIdxConfig(r repo.Repo, key string, value interface{}) (*IdxConfigField, error) {
	err := r.SetIdxConfigKey(key, value)
	if err != nil {
		return nil, fmt.Errorf("failed to set config value: %s (maybe use --json?)", err)
	}
	return getIdxConfig(r, key)
}

func editIdxConfig(filename string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return errors.New("ENV variable $EDITOR not set")
	}

	cmd := exec.Command("sh", "-c", editor+" "+filename)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

func replaceIdxConfig(r repo.Repo, file io.Reader) error {
	var cfg idxconfig.IdxConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return errors.New("failed to decode file as config")
	}

	return r.SetIdxConfig(&cfg)
}
