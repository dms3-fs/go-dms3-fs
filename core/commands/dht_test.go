package commands

import (
	"testing"

	"github.com/dms3-fs/go-dms3-fs/namesys"

	dms3ns "github.com/dms3-fs/go-dms3ns"
	tu "github.com/dms3-p2p/go-testutil"
)

func TestKeyTranslation(t *testing.T) {
	pid := tu.RandPeerIDFatal(t)
	pkname := namesys.PkKeyForID(pid)
	dms3nsname := dms3ns.RecordKey(pid)

	pkk, err := escapeDhtKey("/pk/" + pid.Pretty())
	if err != nil {
		t.Fatal(err)
	}

	dms3nsk, err := escapeDhtKey("/dms3ns/" + pid.Pretty())
	if err != nil {
		t.Fatal(err)
	}

	if pkk != pkname {
		t.Fatal("keys didnt match!")
	}

	if dms3nsk != dms3nsname {
		t.Fatal("keys didnt match!")
	}
}
