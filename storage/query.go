package storage

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"github.com/syndtr/goleveldb/leveldb/util"
)

type Query struct {
	l *Logger
}

func (q *Query) NodesCount(args struct{}, rs *int) error {
	*rs = q.l.Nodes()
	return nil
}

func (q *Query) All(args struct{}, info *DBInfo) error {
	info.Nodes = q.l.Nodes()
	info.Relations = q.l.AllRelations()

	// relation的done和doing都只查今天的
	doingIter := q.l.db.NewIterator(util.BytesPrefix([]byte(todayRelationDoingPrefix)), nil)
	for doingIter.Next() {
		info.RelationDoing++
	}
	doingIter.Release()
	if err := doingIter.Error(); err != nil {
		return err
	}
	info.RelationDone = q.l.TodayRelationDones()

	info.Rlpxs = q.l.AllRlpxs()
	info.Enrs = q.l.AllEnrs()
	return nil
}

func (q *Query) Today(args struct{}, info *DBInfo) error {
	info.Nodes = q.l.Nodes()
	info.Relations = q.l.TodayRelations()
	doingIter := q.l.db.NewIterator(util.BytesPrefix([]byte(todayRelationDoingPrefix)), nil)
	for doingIter.Next() {
		info.RelationDoing++
	}
	doingIter.Release()
	if err := doingIter.Error(); err != nil {
		return err
	}
	info.RelationDone = q.l.TodayRelationDones()
	info.Rlpxs = q.l.TodayRlpxs()
	info.Enrs = q.l.TodayEnrs()
	return nil
}

func startServer(l *Logger) {
	os.Remove(rpcPath)
	// 启动rpc服务
	query := &Query{
		l: l,
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
		l := StartLog(nil, false)
		startServer(l)
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
}

func (i DBInfo) String() string {
	str := `	Nodes: %d
	Relations: %d
	RelationDoing: %d
	RelationDone: %d
	Rlpxs: %d
	ENRs: %d`
	return fmt.Sprintf(str, i.Nodes, i.Relations, i.RelationDoing, i.RelationDone, i.Rlpxs, i.Enrs)
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
