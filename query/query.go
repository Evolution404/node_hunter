package query

import (
	"net/rpc"
	"node_hunter/config"
	"node_hunter/storage"
	"os"
)

type Queryer struct {
	r         *rpc.Client
	runServer bool
}

func NewQueryer() *Queryer {
	server := false
	rc, err := rpc.DialHTTP("unix", config.RpcPath)
	if err != nil {
		storage.StartLog(nil, false)
		server = true
		rc, err = rpc.DialHTTP("unix", config.RpcPath)
		if err != nil {
			panic(err)
		}
	}
	return &Queryer{
		r:         rc,
		runServer: server,
	}
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

func (q *Queryer) Today() storage.DBInfo {
	info := storage.DBInfo{}
	err := q.r.Call("Query.Today", struct{}{}, &info)
	if err != nil {
		panic(err)
	}
	return info
}

func (q *Queryer) All() storage.DBInfo {
	info := storage.DBInfo{}
	err := q.r.Call("Query.All", struct{}{}, &info)
	if err != nil {
		panic(err)
	}
	return info
}

func (q *Queryer) Active() int {
	number := 0
	err := q.r.Call("Query.Active", struct{}{}, &number)
	if err != nil {
		panic(err)
	}
	return number
}

func (q *Queryer) ActiveInfo() *storage.Actives {
	rs := new(storage.Actives)
	err := q.r.Call("Query.ActiveInfo", struct{}{}, &rs)
	if err != nil {
		panic(err)
	}
	return rs
}

func (q *Queryer) Close() error {
	if q.runServer {
		return os.Remove(config.RpcPath)
	}
	return nil
}
