package coreindex

import (
    "encoding/json"
	"fmt"

    cid "github.com/dms3-fs/go-cid"
)

type rps struct {
    Cid *cid.Cid      // reposet cid
}

type Rps interface {

    GetCid() *cid.Cid

    SetCid(v *cid.Cid)

    Equals(o Rps) bool

    Marshal() ([]byte, error)
    Unmarshal(b []byte) error
}

func NewRps() Rps {
    return &rps{
        Cid: nil,
    }
}

func (c *rps) GetCid() *cid.Cid {
    return c.Cid
}

func (c *rps) SetCid(v *cid.Cid) {
    c.Cid = v
}

func (c *rps) Equals(o Rps) bool {
    return c.Cid.Equals(o.GetCid())
}

func (c *rps) Marshal() ([]byte, error) {

    b, err := json.Marshal(*c)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal reposet properties: %v", err)
    } else {
        return b, nil
    }
}

func (c *rps) Unmarshal(b []byte) error {

	err := json.Unmarshal(b, c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal reposet properties: %v", err)
	}
    return nil
}
