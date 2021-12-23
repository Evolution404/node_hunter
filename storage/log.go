package storage

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

var PrivateKey = genPriv()

// 用于初始化各种map的初始大小
const NodeCount = 800000

var date = time.Now().Format("2006-01-02")
var BasePath string = path.Join(GetCurrentAbPath(), "data")
var nodesPath string = path.Join(BasePath, "nodes")
var NodesPath string = nodesPath

var relationPath string = path.Join(BasePath, "relation-"+date)
var rlpxPath string = path.Join(BasePath, "rlpx-"+date)
var ENRPath string = path.Join(BasePath, "/enr-"+date)

func genPriv() *ecdsa.PrivateKey {
	priv, err := crypto.ToECDSA(common.FromHex("51e00445e18afc55f9c76a2640538204940abebc9704823052aac6c7275923db"))
	// priv, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	return priv
}

type Logger struct {
	// 记录所有节点
	waitingNodes []*enode.Node
	waitingLock  sync.Mutex
	Relation     chan string
	relation     *os.File
	Nodes        chan string
	nodes        *os.File
	Rlpx         chan string
	rlpx         *os.File
	wg           sync.WaitGroup
}

func createOrOpen(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
}

func CreateOrOpen(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
}

// 输入若干种子节点，作为初始化节点
// 如果输入nil，说明全部使用data文件夹内记录的节点
func StartLog(seedNodes []*enode.Node) *Logger {
	os.MkdirAll(BasePath, 0777)
	relation, err := createOrOpen(relationPath)
	if err != nil {
		panic(err)
	}
	nodes, err := createOrOpen(nodesPath)
	if err != nil {
		panic(err)
	}

	l := &Logger{
		Relation: make(chan string, 10),
		relation: relation,
		Nodes:    make(chan string, 10),
		nodes:    nodes,
	}
	// 恢复数据
	l.restore()

	// 如果种子节点之前没有被保存下来，记录种子节点
	if seedNodes != nil {
		for _, seedNode := range seedNodes {
			if seenNode[seedNode.URLv4()] == 0 {
				seenNode[seedNode.URLv4()] = -1
				l.waitingNodes = append(l.waitingNodes, seedNode)
				l.Nodes <- seedNode.URLv4()
			}
		}
	}

	fmt.Println(len(seenNode))
	l.wg.Add(1)
	go l.loop()
	return l
}
func StartRlpxLog() *Logger {
	rlpx, err := createOrOpen(rlpxPath)
	if err != nil {
		panic(err)
	}
	l := &Logger{
		Rlpx: make(chan string, 10),
		rlpx: rlpx,
	}
	l.restoreRlpx()
	l.wg.Add(1)
	go l.loop()
	return l
}

func (l *Logger) loop() {
	defer l.wg.Done()
	// 定位到文件末尾，开始记录新内容
	l.nodes.Seek(0, io.SeekEnd)
	l.relation.Seek(0, io.SeekEnd)
	l.rlpx.Seek(0, io.SeekEnd)
	for {
		now := time.Now().Unix()
		select {
		case r, ok := <-l.Relation:
			if !ok {
				return
			}
			str := fmt.Sprintf("%d %s\n", now, r)
			if _, err := l.relation.WriteString(str); err != nil {
				panic(err)
			}
		case r, ok := <-l.Nodes:
			if !ok {
				return
			}
			str := fmt.Sprintf("%d %s\n", now, r)
			if _, err := l.nodes.WriteString(str); err != nil {
				panic(err)
			}
		case r, ok := <-l.Rlpx:
			if !ok {
				return
			}
			str := fmt.Sprintf("%d %s\n", now, r)
			if _, err := l.rlpx.WriteString(str); err != nil {
				panic(err)
			}
		}
	}
}

func (l *Logger) Close() error {
	if l.Relation != nil {
		close(l.Relation)
	}
	if l.Rlpx != nil {
		close(l.Rlpx)
	}
	if l.Nodes != nil {
		close(l.Nodes)
	}
	l.wg.Wait()
	if l.relation != nil {
		if err := l.relation.Close(); err != nil {
			return err
		}
	}
	if l.rlpx != nil {
		if err := l.rlpx.Close(); err != nil {
			return err
		}
	}
	if l.nodes != nil {
		if err := l.nodes.Close(); err != nil {
			return err
		}
	}
	return nil
}
