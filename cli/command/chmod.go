package command

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"

	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
)

var ChmodCommand = cli.Command{
	Name:   "chmod",
	Usage:  "change node read write state",
	Action: chmodAction,
	Flags: []cli.Flag{
		cli.BoolFlag{"r", "read state"},
		cli.BoolFlag{"w", "write state"},
	},
}

func chmodAction(c *cli.Context) {
	r := c.Bool("r")
	w := c.Bool("w")

	addr := context.GetLeaderAddr()

	url := "http://" + addr + api.NodePermPath
	var act string
	var nodeid string
	var action string
	var perm string

	//-r -w
	if r || w {
		if len(c.Args()) != 1 || r == w {
			fmt.Println(ErrInvalidParameter)
			return
		}
		action = "disable"
		nodeid = context.GetId(c.Args()[0])

		if r {
			perm = "read"
		} else {
			perm = "write"
		}
	} else {
		//+r +w
		if len(c.Args()) != 2 {
			fmt.Println(ErrInvalidParameter)
			return
		}
		act = c.Args()[0]
		if string(act[0]) == "+" {
			action = "enable"
			nodeid = context.GetId(c.Args()[1])

			if string(act[1]) == "r" {
				perm = "read"
			} else if string(act[1]) == "w" {
				perm = "write"
			} else {
				fmt.Println(ErrInvalidParameter)
				return
			}
		} else {
			fmt.Println(ErrInvalidParameter)
			return
		}
	}

	req := api.ToggleModeParams{
		NodeId: nodeid,
		Action: action,
		Perm:   perm,
	}
	resp, err := utils.HttpPost(url, req, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	ShowResponse(resp)
}