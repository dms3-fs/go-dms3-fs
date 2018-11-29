
package coreindex
import (
	"errors"
	//"fmt"
	"path/filepath"

	idxconfig "github.com/dms3-fs/go-idx-config"
    util "github.com/dms3-fs/go-fs-util"

)

func ReposetExists(kind, name string) (found bool, path string, err error) {

	if kind == "" {
		err = errors.New("reposet kind must not be null.")
		return
	}

	if name == "" {
		err = errors.New("reposet name must not be null.")
		return
	}

	if path, err = ReposetLocalPath(kind, name); err != nil {
		return
	}

	found = PathExists(path)

	return
}


func PathExists(p string) bool {
	if !util.FileExists(p) {
			return false
	}
	return true
}

// ReposetLocalPath returns the index reposet path for a specified reposet.
func ReposetLocalPath(kind, name string) (path string, err error) {

	//
	// return local filesystem reposet path for a named reposet of a kind
	//
	// index <repo root>, cfg parameter Indexer.Path, must be relative path
	// 	  - <index>
	// reposet root
	// 	  - <index>/reposet
	// reposet kind root
	//	  - <index>/reposet/<kind>
	// reposet root folder
	// 	  - <index>/reposet/<kind>/<name>
	//
	var rootpath string

	if rootpath, err = idxconfig.PathRoot(); err != nil {
		return
	}

	path = filepath.Join(rootpath, "reposet", kind, name)
	//fmt.Printf("rootpath: %v\n", rootpath)
	//fmt.Printf("kind: %v\n", kind)
	//fmt.Printf("name: %v\n", name)
	//fmt.Printf("path: %v\n", path)

	return
}
