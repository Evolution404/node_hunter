package search

import (
	"fmt"
	"net"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type node struct {
	n         enode.Node
	timestamp int64
}
type nodes []node

func (n nodes) Len() int           { return len(n) }
func (n nodes) Less(i, j int) bool { return n[i].timestamp < n[j].timestamp } // 大顶堆，返回值决定是否交换元素
func (n nodes) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }

func (n *nodes) Push(x interface{}) {
	*n = append(*n, x.(node))
}

func (h *nodes) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

var nodeHeap nodes

func InitV4() *discover.UDPv4 {
	// 构造UDP连接，要使用ListenUDP不能使用DialUDP
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   []byte{},
		Port: 30303,
	})
	if err != nil {
		panic(err)
	}

	// 准备enode.DB对象
	db, err := enode.OpenDB("")
	if err != nil {
		panic(err)
	}

	// 准备节点私钥
	priv, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	ln := enode.NewLocalNode(db, priv)

	logger := log.New()
	logger.SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stderr, log.LogfmtFormat())))

	// 启动节点发现协议
	udpv4, err := discover.ListenV4(conn, ln, discover.Config{
		PrivateKey: priv,
		// Log:        logger,
	})
	if err != nil {
		panic(err)
	}
	return udpv4
}

func RelationNodes(udpv4 *discover.UDPv4, initial *enode.Node) (map[enode.ID]*enode.Node, error) {
	nodeMap := make(map[enode.ID]*enode.Node)
	last := 0
	count := 0
	// 查询出错超过10次停止查询
	errCount := 0
	for {
		rs, err := udpv4.FindRandomNode(initial)
		if err != nil {
			if errCount >= 3 {
				return nodeMap, err
			}
			errCount++
		} else {
			errCount = 0
		}
		for _, r := range rs {
			nodeMap[r.ID()] = r
		}
		if last == len(nodeMap) {
			if count >= 20 {
				break
			}
			count++
		} else {
			count = 0
		}
		last = len(nodeMap)
	}
	fmt.Printf("relation: %d, %s\n", len(nodeMap), initial.URLv4())
	return nodeMap, nil
}
