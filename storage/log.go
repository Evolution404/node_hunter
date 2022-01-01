package storage

import (
	"fmt"
	"node_hunter/config"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/redmask-hb/GoSimplePrint/goPrint"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// 运行使用的日期
// 如果之前的执行完了使用今天
// 否则继续按照之前的日期查询
var date string = ""

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
	// 结束后删除今天的日期
	os.MkdirAll(config.BasePath, 0777)
	l := &Logger{
		db: openDB(),
	}
	date = l.queryDate()
	updateDate()
	// 启动rpc服务
	startServer(l)

	if load {

		l.waitingNodes = make([]*enode.Node, 0, 500000)
		// 数据库中保存的节点记录总数
		nodes := l.Nodes()

		// 生成进度条
		i := 0
		bar := goPrint.NewBar(nodes)
		bar.SetNotice("loading nodes")

		iter := l.db.NewIterator(util.BytesPrefix([]byte(nodesPrefix)), nil)
		// 加载所有节点
		for iter.Next() {
			if i%1000 == 0 {
				bar.PrintBar(i)
			}
			url := string(iter.Key()[len(nodesPrefix):])
			node := enode.MustParseV4(url)
			// 加载还没完成查询的节点
			if !l.IsRelationDone(node) && !config.Reject(node) {
				// 之前没查询完成的放到等待列表的最前面
				if l.IsRelationDoing(node) {
					l.waitingNodes = append([]*enode.Node{node}, l.waitingNodes...)
				} else {
					l.waitingNodes = append(l.waitingNodes, node)
				}
			}
			i++
		}
		if err := iter.Error(); err != nil {
			panic(err)
		}
		bar.PrintBar(nodes)
		fmt.Println()
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
