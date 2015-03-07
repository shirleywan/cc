package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/state"
)

type FailoverEndCommand struct {
	NodeId string
}

func (self *FailoverEndCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	node := cs.FindNode(self.NodeId)
	if node == nil {
		return nil, ErrNodeNotExist
	}
	err := node.AdvanceFSM(cs, state.CMD_FAILOVER_END_SIGNAL)
	return nil, err
}