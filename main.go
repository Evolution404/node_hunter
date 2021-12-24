package main

import (
	"fmt"
	"node_hunter/discover"
	"node_hunter/enr"
	"node_hunter/rlpx"
	"node_hunter/storage"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/jessevdk/go-flags"
)

type DiscoverCommand struct {
	Threads     int      `short:"t" long:"threads" default:"30" description:"threads to execute node discover"`
	NodeThreads int      `short:"n" long:"nodethreads" default:"10" description:"threads to execute node discover"`
	SeedNodes   []string `short:"s" long:"seeds" description:"initial seed nodes"`
}

func (d *DiscoverCommand) Execute(args []string) error {
	var seed []*enode.Node
	for _, s := range d.SeedNodes {
		n := enode.MustParseV4(s)
		seed = append(seed, n)
	}
	discover.StartDiscover(seed, d.Threads, d.NodeThreads)
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
	Threads int `short:"t" long:"threads" default:"30" description:"threads to query node meta data"`
}

func (e *ENRCommand) Execute(args []string) error {
	enr.UpdateENR(e.Threads)
	return nil
}

type QueryCommand struct {
	Today bool `short:"t" long:"today" default:"false" description:"show today's data"`
	All   bool `short:"a" long:"all" default:"false" description:"show all data"`
	Nodes bool `short:"n" long:"nodes" default:"false" description:"show the number of node records"`
}

func (q *QueryCommand) Execute(args []string) error {
	query := storage.NewQueryer()
	if q.Today {
		fmt.Println(query.Today())
	}
	if q.All {
		fmt.Println(query.All())
		query.All()
	}
	if q.Nodes {
		fmt.Println(query.Nodes())
	}
	return query.Close()
}

// 定义的所有子命令
type Option struct {
	Discover DiscoverCommand `command:"disc"`
	Rlpx     RlpxCommand     `command:"rlpx"`
	ENR      ENRCommand      `command:"enr"`
	Query    QueryCommand    `command:"query"`
}

func main() {
	var opt Option
	_, err := flags.Parse(&opt)
	if err != nil {
		fmt.Println(err)
	}
}
