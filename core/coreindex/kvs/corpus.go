package coreindex

import (
    "encoding/json"
	"fmt"

    cid "github.com/dms3-fs/go-cid"
    logging "github.com/dms3-fs/go-log"
)

// log is the command logger
var log = logging.Logger("coreindex")

type corpusProps struct {
    Rclass  string          // repo class
    Rkind   string          // repo kind
    Rindex  int64           // repo index in reposet
    Rcid    *cid.Cid        // corpus document cid
}

// Corpus provides an abstraction for corpus document cid tracking.
type CorpusProps interface {

    GetRclass() string
    GetRkind() string
    GetRindex() int64
    GetRcid() *cid.Cid

    SetRclass(rc string)
    SetRkind(rk string)
    SetRindex(ri int64)
    SetRcid(id *cid.Cid)

    Equals(o CorpusProps) bool

    Marshal() ([]byte, error)
    Unmarshal(b []byte) error
}

func NewCorpusProps(rc string, rk string, ri int64, id *cid.Cid) CorpusProps {
    return &corpusProps{
        Rclass: rc,
        Rkind:  rk,
        Rindex: ri,
        Rcid:   id,
    }
}

func (c *corpusProps) GetRclass() string {
    return c.Rclass
}

func (c *corpusProps) GetRkind() string {
    return c.Rkind
}

func (c *corpusProps) GetRindex() int64 {
    return c.Rindex
}

func (c *corpusProps) GetRcid() *cid.Cid {
    return c.Rcid
}

func (c *corpusProps) SetRclass(rc string) {
    c.Rclass = rc
}

func (c *corpusProps) SetRkind(rk string) {
    c.Rkind = rk
}

func (c *corpusProps) SetRindex(ri int64) {
    c.Rindex = ri
}

func (c *corpusProps) SetRcid(id *cid.Cid) {
    c.Rcid = id
}

func (c *corpusProps) Equals(o CorpusProps) bool {
    return c.Rclass == o.GetRclass() &&
            c.Rkind == o.GetRkind() &&
            c.Rindex == o.GetRindex() &&
            c.Rcid.Equals(o.GetRcid())
}

func (c *corpusProps) Marshal() ([]byte, error) {

    b, err := json.Marshal(*c)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal corpus properties: %v", err)
    } else {
        return b, nil
    }
}

func (c *corpusProps) Unmarshal(b []byte) error {

	err := json.Unmarshal(b, c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal corpus doc properties: %v", err)
	}
    return nil
}
