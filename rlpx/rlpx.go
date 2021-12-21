package rlpx

import (
	"bufio"
	"crypto/ecdsa"
	"fmt"
	"net"
	"node_hunter/storage"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type Query struct {
	priv *ecdsa.PrivateKey
	f    *os.File
}

func NewQuery() *Query {
	priv := storage.PrivateKey
	// 打开所有节点记录文件
	f, err := os.Open(storage.NodesPath)
	if err != nil {
		panic(err)
	}
	return &Query{
		priv: priv,
		f:    f,
	}
}

func (q *Query) Query(l *storage.Logger, threads int) {
	fmt.Printf("starting rlpx query threads=%d\n", threads)
	// 控制同时查询的协程数
	var wg sync.WaitGroup
	token := make(chan struct{}, threads)
	for i := 0; i < threads; i++ {
		token <- struct{}{}
	}
	scanner := bufio.NewScanner(q.f)
	for scanner.Scan() {
		var timestamp int64
		var url string
		str := scanner.Text()
		fmt.Sscanf(str, "%d %s", &timestamp, &url)
		node := enode.MustParseV4(url)
		wg.Add(1)
		<-token
		go func(n *enode.Node) {
			defer wg.Done()
			q.QueryNode(l, n)
			token <- struct{}{}
		}(node)
	}
	wg.Wait()
}

// 查询一个节点的版本，操作系统，支持的协议
func (q *Query) QueryNode(l *storage.Logger, node *enode.Node) {
	// 最近查询过rlpx元数据了，跳过查询
	if !l.ShouldRlpx(node) {
		return
	}
	endpoint := fmt.Sprintf("%s:%d", node.IP().String(), node.TCP())
	fmt.Println("querying", node.URLv4())
	conn, err := net.DialTimeout("tcp4", endpoint, time.Second*3)
	if err != nil {
		str := fmt.Sprintf("%s error %s", node.URLv4(), err.Error())
		fmt.Println(str)
		l.Rlpx <- str
		l.RlpxDone(node)
		return
	}
	t := p2p.NewRLPX(conn, node.Pubkey())
	_, err = t.DoEncHandshake(q.priv)
	if err != nil {
		str := fmt.Sprintf("%s error %s", node.URLv4(), err.Error())
		fmt.Println(str)
		l.Rlpx <- str
		l.RlpxDone(node)
		return
	}
	their, err := t.DoProtoHandshake()
	if err != nil {
		str := fmt.Sprintf("%s error %s", node.URLv4(), err.Error())
		fmt.Println(str)
		l.Rlpx <- str
		l.RlpxDone(node)
		return
	}
	str := fmt.Sprintf("%s info %s", node.URLv4(), their.Name)
	caps := their.Caps
	// 格式化各个子协议
	// 第一项前面有个空格，后面使用逗号分隔
	if len(caps) > 0 {
		str += " " + caps[0].String()
		caps = caps[1:]
	}
	for _, cap := range caps {
		str += "," + cap.String()
	}
	fmt.Println(str)
	l.Rlpx <- str
	l.RlpxDone(node)
	conn.Close()
}
