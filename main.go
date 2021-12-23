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
	NodeThreads int      `short:"nt" long:"nodethreads" default:"100" description:"threads to execute node discover"`
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
	l := storage.StartRlpxLog()
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

// 定义的所有子命令
type Option struct {
	Discover DiscoverCommand `command:"disc"`
	Rlpx     RlpxCommand     `command:"rlpx"`
	ENR      ENRCommand      `command:"enr"`
}

func main() {
	var opt Option
	_, err := flags.Parse(&opt)
	if err != nil {
		fmt.Println(err)
	}
}
