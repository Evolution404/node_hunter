package storage

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"node_hunter/config"
	"os"
)

type DBInfo struct {
	Nodes         int // 节点记录的条数
	Relations     int // 关系条数
	RelationDoing int
	RelationDone  int
	Rlpxs         int // rlpx记录条数
	Enrs          int // enr记录条数
}

type ActiveNode struct {
	Url    string
	Number int
}

type Actives struct {
	Nodes []ActiveNode
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
	info.RelationDoing = q.l.TodayRelationDoings()
	info.RelationDone = q.l.TodayRelationDones()

	info.Rlpxs = q.l.AllRlpxs()
	info.Enrs = q.l.AllEnrs()
	return nil
}

func (q *Query) Today(args struct{}, info *DBInfo) error {
	info.Nodes = q.l.Nodes()
	info.Relations = q.l.TodayRelations()
	info.RelationDoing = q.l.TodayRelationDoings()
	info.RelationDone = q.l.TodayRelationDones()
	info.Rlpxs = q.l.TodayRlpxs()
	info.Enrs = q.l.TodayEnrs()
	return nil
}

func (q *Query) Active(args struct{}, number *int) error {
	*number = q.l.TodayActives()
	return nil
}

func (q *Query) ActiveInfo(args struct{}, actives *Actives) error {
	rs := q.l.TodayActivesInfo()
	*actives = *rs
	return nil
}

func startServer(l *Logger) {
	os.Remove(config.RpcPath)
	// 启动rpc服务
	query := &Query{
		l: l,
	}
	rpc.Register(query)
	rpc.HandleHTTP()
	listener, err := net.Listen("unix", config.RpcPath)
	if err != nil {
		panic(err)
	}
	go http.Serve(listener, nil)
}

type Queryer struct {
	r         *rpc.Client
	runServer bool
}
