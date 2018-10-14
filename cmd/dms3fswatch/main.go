package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

    config "github.com/dms3-fs/go-dms3-fs-config"
	idxconfig "github.com/dms3-fs/go-dms3-idx-config"
    commands "github.com/dms3-fs/go-dms3-fs/commands"
    core "github.com/dms3-fs/go-dms3-fs/core"
    corehttp "github.com/dms3-fs/go-dms3-fs/core/corehttp"
    coreunix "github.com/dms3-fs/go-dms3-fs/core/coreunix"
    fsrepo "github.com/dms3-fs/go-dms3-fs/repo/fsrepo"

    homedir "github.com/dms3-fs/go-dms3-fs/Godeps/_workspace/src/github.com/mitchellh/go-homedir"

    fsnotify "github.com/fsnotify/fsnotify"
    process "github.com/jbenet/goprocess"
)

var http = flag.Bool("http", false, "expose DMS3FS HTTP API")
var repoPath = flag.String("repo", os.Getenv("DMS3FS_PATH"), "DMS3FS_PATH to use")
var watchPath = flag.String("path", ".", "the path to watch")

func main() {
	flag.Parse()

	// precedence
	// 1. --repo flag
	// 2. DMS3FS_PATH environment variable
	// 3. default repo path
	var dms3fsPath string
	if *repoPath != "" {
		dms3fsPath = *repoPath
	} else {
		var err error
		dms3fsPath, err = fsrepo.BestKnownPath()
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := run(dms3fsPath, *watchPath); err != nil {
		log.Fatal(err)
	}
}

func run(dms3fsPath, watchPath string) error {

	proc := process.WithParent(process.Background())
	log.Printf("running DMS3FSWatch on '%s' using repo at '%s'...", watchPath, dms3fsPath)

	dms3fsPath, err := homedir.Expand(dms3fsPath)
	if err != nil {
		return err
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := addTree(watcher, watchPath); err != nil {
		return err
	}

	r, err := fsrepo.Open(dms3fsPath)
	if err != nil {
		// TODO handle case: daemon running
		// TODO handle case: repo doesn't exist or isn't initialized
		return err
	}

	node, err := core.NewNode(context.Background(), &core.BuildCfg{
		Online: true,
		Repo:   r,
	})
	if err != nil {
		return err
	}
	defer node.Close()

	if *http {
		addr := "/ip4/127.0.0.1/tcp/5101"
		var opts = []corehttp.ServeOption{
			corehttp.GatewayOption(true, "/dms3fs", "/dms3ns"),
			corehttp.WebUIOption,
			corehttp.CommandsOption(cmdCtx(node, dms3fsPath)),
		}
		proc.Go(func(p process.Process) {
			if err := corehttp.ListenAndServe(node, addr, opts...); err != nil {
				return
			}
		})
	}

	interrupts := make(chan os.Signal)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-interrupts:
			return nil
		case e := <-watcher.Events:
			log.Printf("received event: %s", e)
			isDir, err := IsDirectory(e.Name)
			if err != nil {
				continue
			}
			switch e.Op {
			case fsnotify.Remove:
				if isDir {
					if err := watcher.Remove(e.Name); err != nil {
						return err
					}
				}
			default:
				// all events except for Remove result in an DMS3FS.Add, but only
				// directory creation triggers a new watch
				switch e.Op {
				case fsnotify.Create:
					if isDir {
						addTree(watcher, e.Name)
					}
				}
				proc.Go(func(p process.Process) {
					file, err := os.Open(e.Name)
					if err != nil {
						log.Println(err)
					}
					defer file.Close()
					k, err := coreunix.Add(node, file)
					if err != nil {
						log.Println(err)
					}
					log.Printf("added %s... key: %s", e.Name, k)
				})
			}
		case err := <-watcher.Errors:
			log.Println(err)
		}
	}
}

func addTree(w *fsnotify.Watcher, root string) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		isDir, err := IsDirectory(path)
		if err != nil {
			log.Println(err)
			return nil
		}
		switch {
		case isDir && IsHidden(path):
			log.Println(path)
			return filepath.SkipDir
		case isDir:
			log.Println(path)
			if err := w.Add(path); err != nil {
				return err
			}
		default:
			return nil
		}
		return nil
	})
	return err
}

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	return fileInfo.IsDir(), err
}

func IsHidden(path string) bool {
	path = filepath.Base(path)
	if path == "." || path == "" {
		return false
	}
	if rune(path[0]) == rune('.') {
		return true
	}
	return false
}

func cmdCtx(node *core.Dms3FsNode, repoPath string) commands.Context {
	return commands.Context{
		Online:     true,
		ConfigRoot: repoPath,
		LoadConfig: func(path string) (*config.Config, error) {
			return node.Repo.Config()
		},
		LoadIdxConfig: func(path string) (*idxconfig.IdxConfig, error) {
			return node.Repo.IdxConfig()
		},
		ConstructNode: func() (*core.Dms3FsNode, error) {
			return node, nil
		},
	}
}
