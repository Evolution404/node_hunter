package main

import (
	"encoding/binary"
	"fmt"
	"node_hunter/discover"
	"node_hunter/enr"
	"node_hunter/query"
	"node_hunter/rlpx"
	"node_hunter/storage"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/jessevdk/go-flags"
	"github.com/syndtr/goleveldb/leveldb"
)

type DiscoverCommand struct {
	NoRlpx      bool     `long:"norlpx" default:"false" description:"disable rlpx"`
	NoEnr       bool     `long:"noenr" default:"false" description:"disable enr"`
	Remove      bool     `short:"r" long:"remove" default:"false" description:"remove all done sign"`
	Threads     int      `short:"t" long:"threads" default:"30" description:"threads to execute node discover"`
	NodeThreads int      `short:"n" long:"nodethreads" default:"10" description:"threads to execute node discover"`
	SeedNodes   []string `short:"s" long:"seeds" description:"initial seed nodes"`
}

func (d *DiscoverCommand) Execute(args []string) error {
	if d.Remove {
		l := storage.StartLog(nil, false)
		l.RemoveDone()
		return nil
	}
	var seed []*enode.Node
	for _, s := range d.SeedNodes {
		n := enode.MustParseV4(s)
		seed = append(seed, n)
	}
	discover.StartDiscover(seed, d.Threads, d.NodeThreads, d.NoEnr, d.NoRlpx)
	return nil
}

type RlpxCommand struct {
	Threads int `short:"t" long:"threads" default:"30" description:"threads to query node meta data"`
}

func (r *RlpxCommand) Execute(args []string) error {
	q := rlpx.NewQuery()
	l := storage.StartLog(nil, false)
	q.Query(l, r.Threads)
	return nil
}

type ENRCommand struct {
	Threads int `short:"t" long:"threads" default:"30" description:"threads to query node enr record"`
}

func (e *ENRCommand) Execute(args []string) error {
	enr.UpdateENR(e.Threads)
	return nil
}

type QueryCommand struct {
	Today      bool `short:"t" long:"today" default:"false" description:"show today's data"`
	All        bool `short:"a" long:"all" default:"false" description:"show all data"`
	Nodes      bool `short:"n" long:"nodes" default:"false" description:"show the number of node records"`
	Active     bool `short:"i" long:"active" default:"false" description:"show the number of active nodes"`
	ActiveInfo bool `short:"v" long:"activeinfo" default:"false" description:"show the info of active nodes"`
}

func (q *QueryCommand) Execute(args []string) error {
	query := query.NewQueryer()
	if q.Today {
		fmt.Println(query.Today())
	} else if q.All {
		fmt.Println(query.All())
	} else if q.Nodes {
		fmt.Println(query.Nodes())
	} else if q.Active {
		fmt.Println(query.Active())
	} else if q.ActiveInfo {
		actives := query.ActiveInfo()
		for _, n := range actives.Nodes {
			fmt.Println(n.Url, n.Number)
		}
	}
	return query.Close()
}

type DBCommand struct {
	Read   bool `short:"r" long:"read" default:"false" description:"read key"`
	Write  bool `short:"w" long:"write" default:"false" description:"write key value"`
	Delete bool `short:"d" long:"delete" default:"false" description:"delete key"`
}

func (d *DBCommand) Execute(args []string) error {
	db := storage.OpenDB()
	if d.Read {
		if len(args) != 1 {
			fmt.Println("wrong key")
		}
		key := args[0]
		v, err := db.Get([]byte(key), nil)
		if err != nil {
			if err == leveldb.ErrNotFound {
				fmt.Println(key, "not found")
				return nil
			} else {
				panic(err)
			}
		}
		if len(v) == 8 {
			fmt.Println(key, binary.BigEndian.Uint64(v))
		} else {
			fmt.Println(key, string(v))
			fmt.Printf("%s %x\n", key, string(v))
		}
	}
	if d.Write {
		if len(args) != 2 {
			fmt.Println("wrong key value")
			return nil
		}
		key := args[0]
		value := args[1]
		err := db.Put([]byte(key), []byte(value), nil)
		if err != nil {
			panic(err)
		}
	}
	if d.Delete {
		if len(args) != 1 {
			fmt.Println("wrong key")
		}
		err := db.Delete([]byte(args[0]), nil)
		if err != nil {
			panic(err)
		}
	}
	return db.Close()
}

// 定义的所有子命令
type Option struct {
	Discover DiscoverCommand `command:"disc"`
	Rlpx     RlpxCommand     `command:"rlpx"`
	ENR      ENRCommand      `command:"enr"`
	Query    QueryCommand    `command:"query" alias:"q"`
	DB       DBCommand       `command:"db"`
}

func main() {
	var opt Option
	_, err := flags.Parse(&opt)
	if err != nil {
		fmt.Println(err)
	}
}
