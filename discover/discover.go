package discover

import (
	"fmt"
	"node_hunter/search"
	"node_hunter/storage"
	"sync"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

func StartDiscover(nodes []*enode.Node, threads int) {
	fmt.Printf("start discover: threads=%d\n", threads)
	l := storage.StartLog(nodes)
	defer l.Close()

	udpv4 := search.InitV4(30303)

	// 控制同时查询的协程数
	token := make(chan struct{}, threads)
	for i := 0; i < threads; i++ {
		token <- struct{}{}
	}
	var wg sync.WaitGroup
	// 不断循环所有节点进行搜索
	for {
		noNew := true
		for _, node := range l.AllNodes {
			// 查询过不再查询
			if l.Seen(node.ID()) > 0 {
				continue
			}
			noNew = false
			<-token
			fmt.Println("start search:", node.URLv4())
			// 避免重复查询，在开始查询的时候就记录一下时间
			l.AddFinished(node)
			wg.Add(1)
			go func(n *enode.Node) {
				defer wg.Done()
				nodeMap, err := search.DumpRelation(udpv4, n)
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
				token <- struct{}{}
			}(node)
		}
		wg.Wait()
		if noNew {
			break
		}
	}
}
