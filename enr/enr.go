package enr

import (
	"bufio"
	"fmt"
	"node_hunter/search"
	"node_hunter/storage"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

func UpdateENR(threads int) {
	fmt.Printf("updating enr threads=%d\n", threads)
	udpv4 := search.InitV4()
	searched := make(map[enode.ID]bool)

	nodesF, err := storage.CreateOrOpen(storage.NodesPath)
	if err != nil {
		panic(err)
	}
	enrF := storage.NewSyncWriter(storage.ENRPath)

	enrRead, err := storage.CreateOrOpen(storage.ENRPath)
	if err != nil {
		panic(err)
	}
	fmt.Println("loading searched enr")
	enrScanner := bufio.NewScanner(enrRead)
	for enrScanner.Scan() {
		str := enrScanner.Text()
		var timestamp int64
		var url string
		fmt.Sscanf(str, "%d %s", &timestamp, &url)
		n := enode.MustParseV4(url)
		searched[n.ID()] = true
	}
	enrRead.Close()
	fmt.Println("loaded", len(searched))

	token := make(chan struct{}, threads)
	for i := 0; i < threads; i++ {
		token <- struct{}{}
	}

	scanner := bufio.NewScanner(nodesF)
	var count int64
	var wg sync.WaitGroup
	for scanner.Scan() {
		wg.Add(1)
		<-token
		scanStr := scanner.Text()
		go func(str string) {
			defer wg.Done()
			defer func() { token <- struct{}{} }()
			var timestamp int64
			var url string
			fmt.Sscanf(str, "%d %s", &timestamp, &url)
			node := enode.MustParseV4(url)
			// 缓存了结果，跳过
			if searched[node.ID()] {
				return
			}
			fmt.Println("requesting", node.URLv4())
			nn, err := udpv4.RequestENR(node)
			now := time.Now().Unix()
			str = fmt.Sprintf("%d %s", now, url)
			if err != nil {
				str += fmt.Sprintf(" error %s", err.Error())
			} else {
				seq := nn.Seq()
				str += fmt.Sprintf(" info %d %s", seq, nn.String())
			}
			fmt.Println(str)
			str += "\n"
			enrF.Write([]byte(str))
			atomic.AddInt64(&count, 1)
			if atomic.LoadInt64(&count)%1000 == 0 {
				fmt.Printf("done %d nodes", count)
			}
		}(scanStr)
	}
	wg.Wait()
}
