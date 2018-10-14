package coreindex

import (
    "fmt"
    "encoding/json"
)

type suffix struct {// repository folder name suffix
    Offset int64   // shard tag1, repo create relative time,
                    // seconds since reposet creation
    Area int64     // shard tag2
    Category int64 // shard tag3
}

type repoProps struct {
    Suffix suffix
    Docs int64     // # of documents in the repo
}

type RepoProps interface {

    GetOffset() int64
    GetArea() int64
    GetCategory() int64
    GetDocs() int64

    SetOffset(v int64)
    SetArea(v int64)
    SetCategory(v int64)
    SetDocs(v int64)

    Equals(o RepoProps) bool

    Marshal() ([]byte, error)
    Unmarshal(b []byte) error
}

func NewRepoProps() RepoProps {
    return &repoProps{
        Suffix: suffix{
            Offset: 0,
            Area: 0,
            Category: 0,
        },
        Docs: 0,
    }
}

func (c *repoProps) GetOffset() int64 {
    return c.Suffix.Offset
}

func (c *repoProps) GetArea() int64 {
    return c.Suffix.Area
}

func (c *repoProps) GetCategory() int64 {
    return c.Suffix.Category
}

func (c *repoProps) GetDocs() int64 {
    return c.Docs
}

func (c *repoProps) SetOffset(v int64) {
    c.Suffix.Offset = v
}

func (c *repoProps) SetArea(v int64) {
    c.Suffix.Area = v
}

func (c *repoProps) SetCategory(v int64) {
    c.Suffix.Category = v
}

func (c *repoProps) SetDocs(v int64) {
    c.Docs = v
}

func (c *repoProps) Equals(o RepoProps) bool {

    return c.Suffix.Offset == o.GetOffset() &&
            c.Suffix.Area == o.GetArea() &&
            c.Suffix.Category == o.GetCategory() &&
            c.Docs == o.GetDocs()
}

func (c *repoProps) Marshal() ([]byte, error) {

    b, err := json.Marshal(*c)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal repo properties: %v", err)
    } else {
        return b, nil
    }
}

func (c *repoProps) Unmarshal(b []byte) error {

	err := json.Unmarshal(b, c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal repo properties: %v", err)
	}
    return nil
}
