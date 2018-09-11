package fsrepo

import (
	"os"

	homedir "github.com/dms3-fs/go-dms3-fs/Godeps/_workspace/src/github.com/mitchellh/go-homedir"

	config "github.com/dms3-fs/go-fs-config"
)

// BestKnownPath returns the best known fsrepo path. If the ENV override is
// present, this function returns that value. Otherwise, it returns the default
// repo path.
func BestKnownPath() (string, error) {
	dms3fsPath := config.DefaultPathRoot
	if os.Getenv(config.EnvDir) != "" {
		dms3fsPath = os.Getenv(config.EnvDir)
	}
	dms3fsPath, err := homedir.Expand(dms3fsPath)
	if err != nil {
		return "", err
	}
	return dms3fsPath, nil
}
