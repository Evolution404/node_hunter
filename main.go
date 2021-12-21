package main

import (
	"encoding/json"
	"fmt"
	"io"
	"node_hunter/search"
	"node_hunter/storage"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

type NodeRecord struct {
	Record string `json:"record"`
}

// 读取以太坊官方维护的节点列表
func ReadNodes() []*enode.Node {
	f, err := os.Open("nodes.json")
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

func main() {
	nodes := []*enode.Node{enode.MustParse("enr:-Je4QN9cEF4RMRF8zG_Bng1ZWG5VSH98w0H4U1FIcIRIuOFIMTh_QQeD390aKb0hPibD6__EYhC7b1RZHpO5P5ayEggbg2V0aMfGhOAp6ZGAgmlkgnY0gmlwhMOwtZSJc2VjcDI1NmsxoQP1j8zSY7oyJBL_NyRGa713TTAYt_oAyIdQtZwn5geYhYN0Y3CCdl-DdWRwgnZf")}
	l := storage.StartLog(nodes)

	udpv4 := search.InitV4()

	// 最多同时查询100个节点
	token := make(chan struct{}, 100)
	for i := 0; i < 100; i++ {
		token <- struct{}{}
	}

	// 不断循环所有节点进行搜索
	for i := 0; i < 100; i++ {
		var wg sync.WaitGroup
		for _, node := range l.AllNodes {
			// 24小时内不重复查询
			if time.Now().Unix()-l.Seen(node.ID()) < 24*3600 {
				continue
			}
			<-token
			fmt.Println("start search:", node.URLv4())
			wg.Add(1)
			go func(n *enode.Node) {
				nodeMap, err := search.RelationNodes(udpv4, n)
				if err != nil {
					fmt.Println(err)
				}
				// 记录查询完成时间
				l.AddFinished(n)
				// 写入节点关系记录
				relation := fmt.Sprintf("%s %d", n.URLv4(), len(nodeMap))
				for _, n := range nodeMap {
					url := n.URLv4()
					relation += " " + url
					// 如果发现了新节点，加入到数组并记录到文件中
					l.AddSeen(n)
				}
				l.Relation <- relation
				wg.Done()
				token <- struct{}{}
			}(node)
		}
		wg.Wait()
	}
	l.Close()
}
