package storage

import (
	"bufio"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

// 用于判断是否是新节点，并且记录上次查询时间
// 0表示新发现节点，-1表示还没查询过的节点
var seenLock sync.RWMutex
var seenNode = make(map[enode.ID]int64)

// 利用保存的文件进行状态恢复
func (l *Logger) restore() {
	fmt.Println("loading nodes")
	scanner := bufio.NewScanner(l.nodes)
	count := 0
	for scanner.Scan() {
		var timestamp int64
		var url string
		str := scanner.Text()
		fmt.Sscanf(str, "%d %s", &timestamp, &url)
		node := enode.MustParseV4(url)
		id := node.ID()
		// 恢复节点记录
		if seenNode[id] == 0 {
			seenNode[id] = -1
			l.AllNodes = append(l.AllNodes, node)
		}
		count++
	}
	fmt.Println("nodes count:", count)
	// relation文件单行过长，不能使用scanner读取
	fmt.Println("loading relations")
	reader := bufio.NewReader(l.relation)
	count = 0
	for {
		lineBytes, isPrefix, err := reader.ReadLine()
		if err != nil {
			break
		}
		str := string(lineBytes)
		var timestamp int64
		var url string
		fmt.Sscanf(str, "%d %s", &timestamp, &url)
		node := enode.MustParseV4(url)
		seenNode[node.ID()] = timestamp
		// 只需要一行的最开始信息，此行剩余内容忽略
		for isPrefix {
			_, isPrefix, err = reader.ReadLine()
			if err != nil {
				break
			}
		}
		count++
	}
	fmt.Println("searched count:", count)
}

func (l *Logger) AddSeen(n *enode.Node) {
	id := n.ID()
	seenLock.RLock()
	old := seenNode[id]
	seenLock.RUnlock()
	// 没见过的节点记录下来
	if old == 0 {
		l.Nodes <- n.URLv4()
		l.AllNodes = append(l.AllNodes, n)
	}
	// 更新节点记录为-1，表示观察到了
	seenLock.Lock()
	seenNode[id] = -1
	seenLock.Unlock()
}

func (l *Logger) AddSeens(ns []*enode.Node) {
	for _, n := range ns {
		l.AddSeen(n)
	}
}

func (l *Logger) AddFinished(n *enode.Node) {
	id := n.ID()
	seenLock.RLock()
	old := seenNode[id]
	seenLock.RUnlock()
	// 没见过的节点记录下来
	if old == 0 {
		l.Nodes <- n.URLv4()
	}
	// 更新节点时间是现在
	now := time.Now().Unix()
	seenLock.Lock()
	seenNode[id] = now
	seenLock.Unlock()
}

func (l *Logger) Seen(id enode.ID) int64 {
	seenLock.RLock()
	defer seenLock.RUnlock()

	return seenNode[id]
}
