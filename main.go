package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discover"
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
	nodes := ReadNodes()
	fmt.Println(len(nodes))
	nodes = nodes[:100]
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
		Log:        logger,
	})
	if err != nil {
		panic(err)
	}
	// 创建远程节点对象并向其发送Ping包
	for _, node := range nodes {
		if err := udpv4.Ping(node); err != nil {
			fmt.Println(err)
			fmt.Println(node.URLv4())
		}
	}
}
