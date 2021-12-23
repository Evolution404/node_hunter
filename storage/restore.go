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
var seenNode = make(map[string]int64, NodeCount)

var rlpxLock sync.RWMutex

// 0代表没查询，大于零记录查询完成的时间戳
var doRlpxNode = make(map[enode.ID]int64, NodeCount)

func (l *Logger) restoreRlpx() {
	fmt.Println("loading rlpx records")
	scanner := bufio.NewScanner(l.rlpx)
	count := 0
	for scanner.Scan() {
		var timestamp int64
		var url string
		str := scanner.Text()
		fmt.Sscanf(str, "%d %s", &timestamp, &url)
		node := enode.MustParseV4(url)
		id := node.ID()
		// 从文件中得到上次请求rlpx的时间
		doRlpxNode[id] = timestamp
		count++
	}
	if scanner.Err() != nil {
		panic(scanner.Err())
	}
	fmt.Println("rlpx nodes count:", count)
}

func (l *Logger) ShouldRlpx(node *enode.Node) bool {
	rlpxLock.RLock()
	last := doRlpxNode[node.ID()]
	rlpxLock.RUnlock()
	return last <= 0
}

func (l *Logger) RlpxDone(node *enode.Node) {
	now := time.Now().Unix()
	rlpxLock.Lock()
	doRlpxNode[node.ID()] = now
	rlpxLock.Unlock()
}

// 利用保存的文件进行状态恢复
func (l *Logger) restore() {
	// relation文件单行过长，不能使用scanner读取
	fmt.Println("loading relations")
	reader := bufio.NewReader(l.relation)
	count := 0
	for {
		lineBytes, isPrefix, err := reader.ReadLine()
		if err != nil {
			break
		}
		str := string(lineBytes)
		var timestamp, relations int64
		var url string
		fmt.Sscanf(str, "%d %s %d", &timestamp, &url, &relations)
		node := enode.MustParseV4(url)

		seenNode[node.URLv4()] = timestamp

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

	count = 0
	fmt.Println("loading nodes")
	scanner := bufio.NewScanner(l.nodes)
	for scanner.Scan() {
		var timestamp int64
		var url string
		str := scanner.Text()
		fmt.Sscanf(str, "%d %s", &timestamp, &url)
		node := enode.MustParseV4(url)
		// 时间戳为0的说明是还没搜索过的节点，加入到等待列表中
		if seenNode[url] == 0 {
			seenNode[url] = -1
			l.waitingNodes = append(l.waitingNodes, node)
		}
		count++
	}
	fmt.Println("nodes count:", count)

}

func (l *Logger) AddSeen(n *enode.Node) bool {
	url := n.URLv4()
	seenLock.RLock()
	old := seenNode[url]
	seenLock.RUnlock()
	// 没见过的节点记录下来
	if old == 0 {
		l.Nodes <- n.URLv4()
		l.waitingLock.Lock()
		l.waitingNodes = append(l.waitingNodes, n)
		l.waitingLock.Unlock()
		seenLock.Lock()
		// 更新节点记录为-1，表示观察到了
		seenNode[url] = -1
		seenLock.Unlock()
		return true
	}
	return false
}

func (l *Logger) AddSeens(ns []*enode.Node) {
	for _, n := range ns {
		l.AddSeen(n)
	}
}

func (l *Logger) AddFinished(n *enode.Node) {
	url := n.URLv4()
	seenLock.RLock()
	old := seenNode[url]
	seenLock.RUnlock()
	// 没见过的节点记录下来
	if old == 0 {
		l.Nodes <- n.URLv4()
	}
	// 更新节点时间是现在
	now := time.Now().Unix()
	seenLock.Lock()
	seenNode[url] = now
	seenLock.Unlock()
}

func (l *Logger) Seen(n *enode.Node) int64 {
	seenLock.RLock()
	defer seenLock.RUnlock()
	return seenNode[n.URLv4()]
}

func (l *Logger) GetWaiting() *enode.Node {
	l.waitingLock.Lock()
	defer l.waitingLock.Unlock()
	if len(l.waitingNodes) == 0 {
		return nil
	}
	first := l.waitingNodes[0]
	l.waitingNodes = l.waitingNodes[1:]
	return first
}
