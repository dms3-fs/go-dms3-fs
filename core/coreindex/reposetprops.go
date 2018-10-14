package coreindex

import (
    "encoding/json"
	"fmt"

    cid "github.com/dms3-fs/go-cid"
)

type reposetProps struct {
    Kind string       // reposet kind
    CreatedAt uint64  // creation (Unix) time (seconds since 1970 epoch)
    MaxAreas uint8    // max tag2 shards, default: 64
    MaxCats uint8     // max tag3 shards, default: 64
    MaxDocs uint64    // max # of documents in repo kind, DEFAULT: 50m
    Params *cid.Cid   // cid of reposet paramaters file
    RepoKey []string  // key list of repos in reposet
}

type ReposetProps interface {

    GetKind() string
    GetCreatedAt() uint64
    GetMaxAreas() uint8
    GetMaxCats() uint8
    GetMaxDocs() uint64
    GetParams() *cid.Cid
    GetRepoKey() []string

    SetKind(v string)
    SetCreatedAt(v uint64)
    SetMaxAreas(v uint8)
    SetMaxCats(v uint8)
    SetMaxDocs(v uint64)
    SetParams(v *cid.Cid)
    SetRepoKey(v []string)

    Equals(o ReposetProps) bool

    Marshal() ([]byte, error)
    Unmarshal(b []byte) error
}

func NewReposetProps() ReposetProps {
    return &reposetProps{
        Kind: "",
        CreatedAt: 0,
        MaxAreas: 0,
        MaxCats: 0,
        MaxDocs: 0,
        Params: nil,
        RepoKey: nil,
    }
}

func equals(c, o []string) bool {

    if len(c) != len(o) {
        return false
    }

    for i := 0; i < len(c); i++ {
        if c[i] != o[i] {
            return false
        }
    }

    return true

}

func (c *reposetProps) GetKind() string {
    return c.Kind
}

func (c *reposetProps) GetCreatedAt() uint64 {
    return c.CreatedAt
}

func (c *reposetProps) GetMaxAreas() uint8 {
    return c.MaxAreas
}

func (c *reposetProps) GetMaxCats() uint8 {
    return c.MaxCats
}

func (c *reposetProps) GetMaxDocs() uint64 {
    return c.MaxDocs
}

func (c *reposetProps) GetParams() *cid.Cid {
    return c.Params
}

func (c *reposetProps) GetRepoKey() []string {
    return c.RepoKey
}

func (c *reposetProps) SetKind(v string) {
    c.Kind = v
}

func (c *reposetProps) SetCreatedAt(v uint64) {
    c.CreatedAt = v
}

func (c *reposetProps) SetMaxAreas(v uint8) {
    c.MaxAreas = v
}

func (c *reposetProps) SetMaxCats(v uint8) {
    c.MaxCats = v
}

func (c *reposetProps) SetMaxDocs(v uint64) {
    c.MaxDocs = v
}

func (c *reposetProps) SetParams(v *cid.Cid) {
    c.Params = v
}

func (c *reposetProps) SetRepoKey(v []string) {
    c.RepoKey = v
}

func (c *reposetProps) Equals(o ReposetProps) bool {

    return c.Kind == o.GetKind() &&
            c.CreatedAt == o.GetCreatedAt() &&
            c.MaxAreas == o.GetMaxAreas() &&
            c.MaxCats == o.GetMaxCats() &&
            c.MaxDocs == o.GetMaxDocs() &&
            c.Params.Equals(o.GetParams()) &&
            equals(c.RepoKey, o.GetRepoKey())
}

func (c *reposetProps) Marshal() ([]byte, error) {

    b, err := json.Marshal(*c)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal reposet properties: %v", err)
    } else {
        return b, nil
    }
}

func (c *reposetProps) Unmarshal(b []byte) error {

	err := json.Unmarshal(b, c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal reposet properties: %v", err)
	}
    return nil
}
