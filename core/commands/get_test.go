package commands

import (
	"context"
	"fmt"
	"testing"

	cmdkit "github.com/dms3-fs/go-fs-cmdkit"
	cmds "github.com/dms3-fs/go-fs-cmds"
)

func TestGetOutputPath(t *testing.T) {
	cases := []struct {
		args    []string
		opts    cmdkit.OptMap
		outPath string
	}{
		{
			args: []string{"/dms3ns/dms3.io/"},
			opts: map[string]interface{}{
				"output": "takes-precedence",
			},
			outPath: "takes-precedence",
		},
		{
			args: []string{"/dms3ns/dms3.io/", "some-other-arg-to-be-ignored"},
			opts: cmdkit.OptMap{
				"output": "takes-precedence",
			},
			outPath: "takes-precedence",
		},
		{
			args:    []string{"/dms3ns/dms3.io/"},
			outPath: "dms3.io",
			opts:    cmdkit.OptMap{},
		},
		{
			args:    []string{"/dms3ns/dms3.io/logo.svg/"},
			outPath: "logo.svg",
			opts:    cmdkit.OptMap{},
		},
		{
			args:    []string{"/dms3ns/dms3.io", "some-other-arg-to-be-ignored"},
			outPath: "dms3.io",
			opts:    cmdkit.OptMap{},
		},
	}

	_, err := GetCmd.GetOptions([]string{})
	if err != nil {
		t.Fatalf("error getting default command options: %v", err)
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%s-%d", t.Name(), i), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			req, err := cmds.NewRequest(ctx, []string{}, tc.opts, tc.args, nil, GetCmd)
			if err != nil {
				t.Fatalf("error creating a command request: %v", err)
			}

			if outPath := getOutPath(req); outPath != tc.outPath {
				t.Errorf("expected outPath %s to be %s", outPath, tc.outPath)
			}
		})
	}
}
