package storage

import (
	"fmt"
	"node_hunter/config"
	"path"
	"testing"
	"time"
)

func TestOutJson(t *testing.T) {
	nodes := ReadNodes(path.Join(config.GetCurrentAbPath(), "nodes.json"))
	for _, n := range nodes {
		fmt.Println(time.Now().Unix(), n.URLv4())
	}
}
