package storage

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

var BasePath string = GetCurrentAbPath() + "/data"
var relationPath string = path.Join(BasePath + "/relation")
var nodesPath string = path.Join(BasePath + "/nodes")

type Logger struct {
	// 记录所有节点
	AllNodes []*enode.Node
	Relation chan string
	relation *os.File
	Nodes    chan string
	nodes    *os.File
	wg       sync.WaitGroup
}

func createOrOpen(path string) (*os.File, error) {
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
	l.wg.Add(1)
	// 恢复数据
	l.restore()

	// 如果种子节点之前没有被保存下来，记录种子节点
	if seedNodes != nil {
		for _, seedNode := range seedNodes {
			if seenNode[seedNode.ID()] == 0 {
				seenNode[seedNode.ID()] = -1
				l.AllNodes = append(l.AllNodes, seedNode)
				l.Nodes <- seedNode.URLv4()
			}
		}
	}

	fmt.Println(len(seenNode))
	go l.loop()
	return l
}

func (l *Logger) loop() {
	defer l.wg.Done()
	// 定位到文件末尾，开始记录新内容
	l.nodes.Seek(0, io.SeekEnd)
	l.relation.Seek(0, io.SeekEnd)
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
		}
	}
}

func (l *Logger) Close() error {
	close(l.Relation)
	l.wg.Wait()
	return l.relation.Close()
}
