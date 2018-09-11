package name

import (
	"errors"
	"fmt"
	"io"
	"strings"

    "github.com/dms3-fs/go-dms3-fs/core/commands/cmdenv"
    "github.com/dms3-fs/go-dms3-fs/core/commands/e"

    "github.com/dms3-fs/go-fs-cmdkit"
    "github.com/dms3-fs/go-fs-cmds"
    "github.com/dms3-p2p/go-p2p-peer"
    "github.com/dms3-p2p/go-p2p-record"
)

type dms3nsPubsubState struct {
	Enabled bool
}

type dms3nsPubsubCancel struct {
	Canceled bool
}

type stringList struct {
	Strings []string
}

// Dms3NsPubsubCmd is the subcommand that allows us to manage the DMS3NS pubsub system
var Dms3NsPubsubCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "DMS3NS pubsub management",
		ShortDescription: `
Manage and inspect the state of the DMS3NS pubsub resolver.

Note: this command is experimental and subject to change as the system is refined
`,
	},
	Subcommands: map[string]*cmds.Command{
		"state":  dms3nspsStateCmd,
		"subs":   dms3nspsSubsCmd,
		"cancel": dms3nspsCancelCmd,
	},
}

var dms3nspsStateCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Query the state of DMS3NS pubsub",
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		cmds.EmitOnce(res, &dms3nsPubsubState{n.PSRouter != nil})
	},
	Type: dms3nsPubsubState{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeEncoder(func(req *cmds.Request, w io.Writer, v interface{}) error {
			output, ok := v.(*dms3nsPubsubState)
			if !ok {
				return e.TypeErr(output, v)
			}

			var state string
			if output.Enabled {
				state = "enabled"
			} else {
				state = "disabled"
			}

			_, err := fmt.Fprintln(w, state)
			return err
		}),
	},
}

var dms3nspsSubsCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Show current name subscriptions",
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		if n.PSRouter == nil {
			res.SetError(errors.New("DMS3NS pubsub subsystem is not enabled"), cmdkit.ErrClient)
			return
		}
		var paths []string
		for _, key := range n.PSRouter.GetSubscriptions() {
			ns, k, err := record.SplitKey(key)
			if err != nil || ns != "dms3ns" {
				// Not necessarily an error.
				continue
			}
			pid, err := peer.IDFromString(k)
			if err != nil {
				log.Errorf("dms3ns key not a valid peer ID: %s", err)
				continue
			}
			paths = append(paths, "/dms3ns/"+peer.IDB58Encode(pid))
		}

		cmds.EmitOnce(res, &stringList{paths})
	},
	Type: stringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: stringListMarshaler(),
	},
}

var dms3nspsCancelCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Cancel a name subscription",
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		if n.PSRouter == nil {
			res.SetError(errors.New("DMS3NS pubsub subsystem is not enabled"), cmdkit.ErrClient)
			return
		}

		name := req.Arguments[0]
		name = strings.TrimPrefix(name, "/dms3ns/")
		pid, err := peer.IDB58Decode(name)
		if err != nil {
			res.SetError(err, cmdkit.ErrClient)
			return
		}

		ok := n.PSRouter.Cancel("/dms3ns/" + string(pid))
		cmds.EmitOnce(res, &dms3nsPubsubCancel{ok})
	},
	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("name", true, false, "Name to cancel the subscription for."),
	},
	Type: dms3nsPubsubCancel{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeEncoder(func(req *cmds.Request, w io.Writer, v interface{}) error {
			output, ok := v.(*dms3nsPubsubCancel)
			if !ok {
				return e.TypeErr(output, v)
			}

			var state string
			if output.Canceled {
				state = "canceled"
			} else {
				state = "no subscription"
			}

			_, err := fmt.Fprintln(w, state)
			return err
		}),
	},
}

func stringListMarshaler() cmds.EncoderFunc {
	return cmds.MakeEncoder(func(req *cmds.Request, w io.Writer, v interface{}) error {
		list, ok := v.(*stringList)
		if !ok {
			return e.TypeErr(list, v)
		}

		for _, s := range list.Strings {
			_, err := fmt.Fprintln(w, s)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
