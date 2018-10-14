package options

import (
//	"time"
)

const (
	RepoWindowStart = iota
)

type IndexListSettings struct {
	Meta      string
}

type IndexListOption func(*IndexListSettings) error

func IndexListOptions(opts ...IndexListOption) (*IndexListSettings, error) {
	options := &IndexListSettings{
		Meta:       "",
	}

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}
