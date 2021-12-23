package storage

import (
	"encoding/json"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

type NodeRecord struct {
	Record string `json:"record"`
}

// 读取以太坊官方维护的节点列表
func ReadNodes(path string) []*enode.Node {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	nodes := make(map[string]NodeRecord)
	r := io.Reader(f)
	if err := json.NewDecoder(r).Decode(&nodes); err != nil {
		panic(err)
	}
	nodeList := []*enode.Node{}
	for _, v := range nodes {
		n := enode.MustParse(v.Record)
		nodeList = append(nodeList, n)
	}
	return nodeList
}
