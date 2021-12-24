package storage

import (
	"node_hunter/config"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

type Logger struct {
	// 记录所有节点
	waitingNodes []*enode.Node
	waitingLock  sync.Mutex
	db           *leveldb.DB
	dbLock       sync.RWMutex
	nodeIter     iterator.Iterator
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
func StartLog(seedNodes []*enode.Node, load bool) *Logger {
	os.MkdirAll(BasePath, 0777)
	l := &Logger{
		db: openDB(),
	}
	if load {
		// 启动rpc服务
		startServer(l)

		// 加载所有节点
		for {
			node := l.NextNode()
			if node == nil {
				break
			}
			// 加载还没完成查询的节点
			if !l.IsRelationDone(node) && !config.Reject(node) {
				// 之前没查询完成的放到等待列表的最前面
				if l.IsRelationDoing(node) {
					l.waitingNodes = append([]*enode.Node{node}, l.waitingNodes...)
				} else {
					l.waitingNodes = append(l.waitingNodes, node)
				}
			}
		}
	}
	for _, seed := range seedNodes {
		l.WriteNode(seed)
	}
	return l
}

func (l *Logger) Close() error {
	if l.nodeIter != nil {
		l.nodeIter.Release()
	}
	return l.db.Close()
}
