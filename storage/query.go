package storage

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Query struct {
	db *leveldb.DB
}

func (q *Query) NodesCount(args struct{}, rs *int) error {
	iter := q.db.NewIterator(util.BytesPrefix([]byte(nodesPrefix)), nil)
	count := 0
	for iter.Next() {
		count++
	}
	*rs = count
	iter.Release()
	return iter.Error()
}

func (q *Query) All(args struct{}, info *DBInfo) error {
	iter := q.db.NewIterator(nil, nil)
	for iter.Next() {
		switch keyType(iter.Key()) {
		case Node:
			info.Nodes++
		case Relation:
			info.Relations++
		case RelationDoing:
			info.RelationDoing++
		case RelationDone:
			info.RelationDone++
		case Rlpx:
			info.Rlpxs++
		case ENR:
			info.Enrs++
		case Unknown:
			info.Unknowns++
		}
	}
	iter.Release()
	return iter.Error()
}

func (q *Query) Today(args struct{}, info *DBInfo) error {
	// 先计算节点记录的条数
	nodesIter := q.db.NewIterator(util.BytesPrefix([]byte(nodesPrefix)), nil)
	for nodesIter.Next() {
		info.Nodes++
	}
	nodesIter.Release()
	if err := nodesIter.Error(); err != nil {
		return err
	}
	// 再计算带有日期前缀的数据条数
	iter := q.db.NewIterator(util.BytesPrefix([]byte(date)), nil)
	for iter.Next() {
		switch keyType(iter.Key()) {
		case Node:
			info.Nodes++
		case Relation:
			info.Relations++
		case RelationDoing:
			info.RelationDoing++
		case RelationDone:
			info.RelationDone++
		case Rlpx:
			info.Rlpxs++
		case ENR:
			info.Enrs++
		case Unknown:
			info.Unknowns++
		}
	}
	nodesIter.Release()
	return iter.Error()
}

func startServer(db *leveldb.DB) {
	os.Remove(rpcPath)
	// 启动rpc服务
	query := &Query{
		db: db,
	}
	rpc.Register(query)
	rpc.HandleHTTP()
	listener, err := net.Listen("unix", rpcPath)
	if err != nil {
		panic(err)
	}
	go http.Serve(listener, nil)
}

type Queryer struct {
	r         *rpc.Client
	runServer bool
}

func NewQueryer() *Queryer {
	server := false
	rc, err := rpc.DialHTTP("unix", rpcPath)
	if err != nil {
		db := openDB()
		startServer(db)
		server = true
		rc, err = rpc.DialHTTP("unix", rpcPath)
		if err != nil {
			panic(err)
		}
	}
	return &Queryer{
		r:         rc,
		runServer: server,
	}
}

type DBInfo struct {
	Nodes         int // 节点记录的条数
	Relations     int // 关系条数
	RelationDoing int
	RelationDone  int
	Rlpxs         int // rlpx记录条数
	Enrs          int // enr记录条数
	Unknowns      int // 未知类型记录条数
}

func (i DBInfo) String() string {
	str := `	Nodes: %d
	Relations: %d
	RelationDoing: %d
	RelationDone: %d
	Rlpxs: %d
	ENRs: %d
	Unknowns: %d`
	return fmt.Sprintf(str, i.Nodes, i.Relations, i.RelationDoing, i.RelationDone, i.Rlpxs, i.Enrs, i.Unknowns)
}

// 查询节点记录条数
func (q *Queryer) Nodes() int {
	rs := 0
	err := q.r.Call("Query.NodesCount", struct{}{}, &rs)
	if err != nil {
		panic(err)
	}
	return rs
}

func (q *Queryer) Today() DBInfo {
	info := DBInfo{}
	err := q.r.Call("Query.Today", struct{}{}, &info)
	if err != nil {
		panic(err)
	}
	return info
}

func (q *Queryer) All() DBInfo {
	info := DBInfo{}
	err := q.r.Call("Query.All", struct{}{}, &info)
	if err != nil {
		panic(err)
	}
	return info
}

func (q *Queryer) Close() error {
	if q.runServer {
		return os.Remove(rpcPath)
	}
	return nil
}
