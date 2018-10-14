package repo

import (
	"errors"

	filestore "github.com/dms3-fs/go-dms3-fs/filestore"
	keystore "github.com/dms3-fs/go-dms3-fs/keystore"

	config "github.com/dms3-fs/go-fs-config"
	idxconfig "github.com/dms3-fs/go-idx-config"
	ma "github.com/dms3-mft/go-multiaddr"
)

var errTODO = errors.New("TODO: mock repo")

// Mock is not thread-safe
type Mock struct {
	C config.Config
	I idxconfig.IdxConfig
	D Datastore
	K keystore.Keystore
}

func (m *Mock) Config() (*config.Config, error) {
	return &m.C, nil // FIXME threadsafety
}
func (m *Mock) IdxConfig() (*idxconfig.IdxConfig, error) {
	return &m.I, nil // FIXME threadsafety
}

func (m *Mock) SetConfig(updated *config.Config) error {
	m.C = *updated // FIXME threadsafety
	return nil
}
func (m *Mock) SetIdxConfig(updated *idxconfig.IdxConfig) error {
	m.I = *updated // FIXME threadsafety
	return nil
}

func (m *Mock) BackupConfig(prefix string) (string, error) {
	return "", errTODO
}
func (m *Mock) BackupIdxConfig(prefix string) (string, error) {
	return "", errTODO
}

func (m *Mock) SetConfigKey(key string, value interface{}) error {
	return errTODO
}
func (m *Mock) SetIdxConfigKey(key string, value interface{}) error {
	return errTODO
}

func (m *Mock) GetConfigKey(key string) (interface{}, error) {
	return nil, errTODO
}
func (m *Mock) GetIdxConfigKey(key string) (interface{}, error) {
	return nil, errTODO
}

func (m *Mock) Datastore() Datastore { return m.D }

func (m *Mock) GetStorageUsage() (uint64, error) { return 0, nil }

func (m *Mock) Close() error { return errTODO }

func (m *Mock) SetAPIAddr(addr ma.Multiaddr) error { return errTODO }

func (m *Mock) Keystore() keystore.Keystore { return m.K }

func (m *Mock) SwarmKey() ([]byte, error) {
	return nil, nil
}

func (m *Mock) FileManager() *filestore.FileManager { return nil }
