package enr

import (
	"fmt"
	"node_hunter/config"
	"node_hunter/discover"
	"node_hunter/storage"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

func UpdateENR(threads int) {
	fmt.Printf("updating enr threads=%d\n", threads)
	udpv4 := discover.InitV4(30304)
	l := storage.StartLog(nil, false)

	token := make(chan struct{}, threads)
	for i := 0; i < threads; i++ {
		token <- struct{}{}
	}

	// scanner := bufio.NewScanner(nodesF)
	var count int64
	var wg sync.WaitGroup
	for {
		node := l.NextNode()
		// 遍历所有节点到末尾了，结束
		if node == nil {
			break
		}
		// 拒绝的节点跳过
		if config.Reject(node) {
			return
		}
		// 查询过的节点跳过
		if l.HasEnr(node) {
			continue
		}
		wg.Add(1)
		<-token
		go func(n *enode.Node) {
			defer wg.Done()
			defer func() { token <- struct{}{} }()

			// 判断是否要拒绝此节点
			if config.Reject(node) {
				return
			}
			fmt.Println("requesting", n.URLv4())
			nn, err := udpv4.RequestENR(n)
			l.WriteEnr(n, nn, err)
			str := n.URLv4()
			if err != nil {
				str += fmt.Sprintf(" error %s", err.Error())
			} else {
				seq := nn.Seq()
				str += fmt.Sprintf(" info %d %s", seq, nn.String())
			}
			fmt.Println(str)
			atomic.AddInt64(&count, 1)
			if atomic.LoadInt64(&count)%1000 == 0 {
				fmt.Printf("done %d nodes", count)
			}
		}(node)
	}
	wg.Wait()
}
