package storage

import (
	"fmt"
	"path"
	"testing"
	"time"
)

func TestOutJson(t *testing.T) {
	nodes := ReadNodes(path.Join(GetCurrentAbPath(), "nodes.json"))
	for _, n := range nodes {
		fmt.Println(time.Now().Unix(), n.URLv4())
	}
}
