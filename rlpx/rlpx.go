package rlpx

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"node_hunter/config"
	"node_hunter/storage"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type Query struct {
	priv *ecdsa.PrivateKey
}

func NewQuery() *Query {
	priv := config.PrivateKey
	return &Query{
		priv: priv,
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
	for {
		// 遍历完成，结束循环
		node := l.NextNode()
		if node == nil {
			break
		}
		// 跳过拒绝节点
		if config.Reject(node) {
			continue
		}
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
func (q *Query) QueryNode(l *storage.Logger, node *enode.Node) error {
	// 最近查询过rlpx元数据了，跳过查询
	if l.HasRlpx(node) {
		return nil
	}
	endpoint := fmt.Sprintf("%s:%d", node.IP().String(), node.TCP())
	fmt.Println("querying", node.URLv4())
	conn, err := net.DialTimeout("tcp4", endpoint, time.Second*3)
	if err != nil {
		str := fmt.Sprintf("e%s", err.Error())
		fmt.Println("rlpx:", str)
		l.WriteRlpx(node, str)
		return err
	}
	t := p2p.NewRLPX(conn, node.Pubkey())
	_, err = t.DoEncHandshake(q.priv)
	if err != nil {
		str := fmt.Sprintf("e%s", err.Error())
		fmt.Println("rlpx:", str)
		l.WriteRlpx(node, str)
		return err
	}
	their, err := t.DoProtoHandshake()
	if err != nil {
		str := fmt.Sprintf("e%s", err.Error())
		fmt.Println("rlpx:", str)
		l.WriteRlpx(node, str)
		return err
	}
	conn.Close()
	str := fmt.Sprintf("i%s ", their.Name)
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
	fmt.Println("rlpx:", str)
	l.WriteRlpx(node, str)
	return nil
}
