package snowflake

import (
	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func NewSnowflake(nodeID int64) (*snowflake.Node, error) {
	var err error
	node, err = snowflake.NewNode(nodeID)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func GenerateID() int64 {
	return node.Generate().Int64()
}
