package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

    assets "github.com/dms3-fs/go-dms3-fs/assets"
    oldcmds "github.com/dms3-fs/go-dms3-fs/commands"
    core "github.com/dms3-fs/go-dms3-fs/core"
    namesys "github.com/dms3-fs/go-dms3-fs/namesys"
    fsrepo "github.com/dms3-fs/go-dms3-fs/repo/fsrepo"

    "github.com/dms3-fs/go-fs-cmdkit"
    "github.com/dms3-fs/go-fs-cmds"
    "github.com/dms3-fs/go-fs-config"
	idxconfig "github.com/dms3-fs/go-idx-config"
)

const (
	nBitsForKeypairDefault = 2048
)

var initCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Initializes dms3fs config file.",
		ShortDescription: `
Initializes dms3fs configuration files and generates a new keypair.

If you are going to run DMS3FS in server environment, you may want to
initialize it using 'server' profile.

For the list of available profiles see 'dms3fs config profile --help'

dms3fs uses a repository in the local file system. By default, the repo is
located at ~/.dms3-fs. To change the repo location, set the $DMS3FS_PATH
environment variable:

    export DMS3FS_PATH=/path/to/dms3fsrepo
`,
	},
	Arguments: []cmdkit.Argument{
		cmdkit.FileArg("default-config", false, false, "Initialize with the given configuration.").EnableStdin(),
	},
	Options: []cmdkit.Option{
		cmdkit.IntOption("bits", "b", "Number of bits to use in the generated RSA private key.").WithDefault(nBitsForKeypairDefault),
		cmdkit.BoolOption("empty-repo", "e", "Don't add and pin help files to the local storage."),
		cmdkit.StringOption("profile", "p", "Apply profile settings to config. Multiple profiles can be separated by ','"),

		// TODO need to decide whether to expose the override as a file or a
		// directory. That is: should we allow the user to also specify the
		// name of the file?
		// TODO cmdkit.StringOption("event-logs", "l", "Location for machine-readable event logs."),
	},
	PreRun: func(req *cmds.Request, env cmds.Environment) error {
		cctx := env.(*oldcmds.Context)
		daemonLocked, err := fsrepo.LockedByOtherProcess(cctx.ConfigRoot)
		if err != nil {
			return err
		}

		log.Info("checking if daemon is running...")
		if daemonLocked {
			log.Debug("dms3fs daemon is running")
			e := "dms3fs daemon is running. please stop it to run this command"
			return cmds.ClientError(e)
		}

		return nil
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		cctx := env.(*oldcmds.Context)
		if cctx.Online {
			res.SetError(errors.New("init must be run offline only"), cmdkit.ErrNormal)
			return
		}

		empty, _ := req.Options["empty-repo"].(bool)
		nBitsForKeypair, _ := req.Options["bits"].(int)

		var conf *config.Config
		var iconf *idxconfig.IdxConfig

		f := req.Files
		if f != nil {
			confFile, err := f.NextFile()
			if err != nil {
				res.SetError(err, cmdkit.ErrNormal)
				return
			}

			conf = &config.Config{}
			if err := json.NewDecoder(confFile).Decode(conf); err != nil {
				res.SetError(err, cmdkit.ErrNormal)
				return
			}

			iconfFile, err := f.NextFile()
			if err != nil {
				res.SetError(err, cmdkit.ErrNormal)
				return
			}

			iconf = &idxconfig.IdxConfig{}
			if err := json.NewDecoder(iconfFile).Decode(iconf); err != nil {
				res.SetError(err, cmdkit.ErrNormal)
				return
			}
		}

		profile, _ := req.Options["profile"].(string)

		var profiles []string
		if profile != "" {
			profiles = strings.Split(profile, ",")
		}

		if err := doInit(os.Stdout, cctx.ConfigRoot, empty, nBitsForKeypair, profiles, conf, iconf); err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
	},
}

var errRepoExists = errors.New(`dms3fs configuration file already exists!
Reinitializing would overwrite your keys.
`)

func initWithDefaults(out io.Writer, repoRoot string, profile string) error {
	var profiles []string
	if profile != "" {
		profiles = strings.Split(profile, ",")
	}

	return doInit(out, repoRoot, false, nBitsForKeypairDefault, profiles, nil, nil)
}

func doInit(out io.Writer, repoRoot string, empty bool, nBitsForKeypair int, confProfiles []string, conf *config.Config, iconf *idxconfig.IdxConfig) error {
	if _, err := fmt.Fprintf(out, "initializing DMS3FS node at %s\n", repoRoot); err != nil {
		return err
	}

	if err := checkWritable(repoRoot); err != nil {
		return err
	}

	if fsrepo.IsInitialized(repoRoot) {
		return errRepoExists
	}

	if conf == nil {
		var err error
		conf, err = config.Init(out, nBitsForKeypair)
		if err != nil {
			return err
		}
	}

	if iconf == nil {
		var err error
		iconf, err = idxconfig.Init(out)
		if err != nil {
			return err
		}
	}

	for _, profile := range confProfiles {
		transformer, ok := config.Profiles[profile]
		if !ok {
			return fmt.Errorf("invalid configuration profile: %s", profile)
		}

		if err := transformer.Transform(conf); err != nil {
			return err
		}
	}

	if err := fsrepo.Init(repoRoot, conf, iconf); err != nil {
		return err
	}

	if !empty {
		if err := addDefaultAssets(out, repoRoot); err != nil {
			return err
		}
	}

	return initializeDms3NsKeyspace(repoRoot)
}

func checkWritable(dir string) error {
	_, err := os.Stat(dir)
	if err == nil {
		// dir exists, make sure we can write to it
		testfile := path.Join(dir, "test")
		fi, err := os.Create(testfile)
		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("%s is not writeable by the current user", dir)
			}
			return fmt.Errorf("unexpected error while checking writeablility of repo root: %s", err)
		}
		fi.Close()
		return os.Remove(testfile)
	}

	if os.IsNotExist(err) {
		// dir doesn't exist, check that we can create it
		return os.Mkdir(dir, 0775)
	}

	if os.IsPermission(err) {
		return fmt.Errorf("cannot write to %s, incorrect permissions", err)
	}

	return err
}

func addDefaultAssets(out io.Writer, repoRoot string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := fsrepo.Open(repoRoot)
	if err != nil { // NB: repo is owned by the node
		return err
	}

	nd, err := core.NewNode(ctx, &core.BuildCfg{Repo: r})
	if err != nil {
		return err
	}
	defer nd.Close()

	dkey, err := assets.SeedInitDocs(nd)
	if err != nil {
		return fmt.Errorf("init: seeding init docs failed: %s", err)
	}
	log.Debugf("init: seeded init docs %s", dkey)

	if _, err = fmt.Fprintf(out, "to get started, enter:\n"); err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "\n\tdms3fs cat /dms3fs/%s/readme\n\n", dkey)
	return err
}

func initializeDms3NsKeyspace(repoRoot string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := fsrepo.Open(repoRoot)
	if err != nil { // NB: repo is owned by the node
		return err
	}

	nd, err := core.NewNode(ctx, &core.BuildCfg{Repo: r})
	if err != nil {
		return err
	}
	defer nd.Close()

	err = nd.SetupOfflineRouting()
	if err != nil {
		return err
	}

	return namesys.InitializeKeyspace(ctx, nd.Namesys, nd.Pinning, nd.PrivateKey)
}
