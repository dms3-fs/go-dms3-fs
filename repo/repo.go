package repo

import (
	"errors"
	"io"

	filestore "github.com/dms3-fs/go-dms3-fs/filestore"
	keystore "github.com/dms3-fs/go-dms3-fs/keystore"

	ds "github.com/dms3-fs/go-datastore"
	config "github.com/dms3-fs/go-fs-config"
	idxconfig "github.com/dms3-fs/go-idx-config"
	ma "github.com/dms3-mft/go-multiaddr"
)

var (
	ErrApiNotRunning = errors.New("api not running")
)

// Repo represents all persistent data of a given dms3fs node.
type Repo interface {
	// Config returns the dms3fs configuration file from the repo. Changes made
	// to the returned config are not automatically persisted.
	Config() (*config.Config, error)
	IdxConfig() (*idxconfig.IdxConfig, error)

	// BackupConfig creates a backup of the current configuration file using
	// the given prefix for naming.
	BackupConfig(prefix string) (string, error)
	BackupIdxConfig(prefix string) (string, error)

	// SetConfig persists the given configuration struct to storage.
	SetConfig(*config.Config) error
	SetIdxConfig(*idxconfig.IdxConfig) error

	// SetConfigKey sets the given key-value pair within the config and persists it to storage.
	SetConfigKey(key string, value interface{}) error
	SetIdxConfigKey(key string, value interface{}) error

	// GetConfigKey reads the value for the given key from the configuration in storage.
	GetConfigKey(key string) (interface{}, error)
	GetIdxConfigKey(key string) (interface{}, error)

	// Datastore returns a reference to the configured data storage backend.
	Datastore() Datastore

	// GetStorageUsage returns the number of bytes stored.
	GetStorageUsage() (uint64, error)

	// Keystore returns a reference to the key management interface.
	Keystore() keystore.Keystore

	// FileManager returns a reference to the filestore file manager.
	FileManager() *filestore.FileManager

	// SetAPIAddr sets the API address in the repo.
	SetAPIAddr(addr ma.Multiaddr) error

	// SwarmKey returns the configured shared symmetric key for the private networks feature.
	SwarmKey() ([]byte, error)

	io.Closer
}

// Datastore is the interface required from a datastore to be
// acceptable to FSRepo.
type Datastore interface {
	ds.Batching // should be threadsafe, just be careful
	io.Closer
}
