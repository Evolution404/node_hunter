package discover

import (
	"fmt"
	"sync"
	"testing"

	crand "crypto/rand"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

func TestRelation1(t *testing.T) {
	node := enode.MustParseV4("enode://8935c9600d925fd46bdf9d1d155ae682c420d75e4546bfd1de4f9cd18c13aab8edd12a1222d34b10113b091f7d95e85e6c985db93086806535a28efd52002109@175.214.58.105:30303")
	v4 := InitV4(30303)
	rss := make(map[enode.ID]bool)
	for i := 0; i < 3; i++ {
		rs, err := v4.FindRandomNode(node)
		fmt.Println(rs, err)
		for _, r := range rs {
			rss[r.ID()] = true
		}
	}
	fmt.Println(len(rss))
	v4.Close()
}

func TestRelation2(t *testing.T) {
	node := enode.MustParseV4("enode://8935c9600d925fd46bdf9d1d155ae682c420d75e4546bfd1de4f9cd18c13aab8edd12a1222d34b10113b091f7d95e85e6c985db93086806535a28efd52002109@175.214.58.105:30303")
	v4 := InitV4(30303)
	var wg sync.WaitGroup
	var lock sync.Mutex
	rss := make(map[enode.ID]bool)
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rs, err := v4.FindRandomNode(node)
			fmt.Println(rs, err)
			lock.Lock()
			for _, r := range rs {
				rss[r.ID()] = true
			}
			lock.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println(len(rss))
	v4.Close()
}

func TestFind(t *testing.T) {
	node := enode.MustParseV4("enode://8935c9600d925fd46bdf9d1d155ae682c420d75e4546bfd1de4f9cd18c13aab8edd12a1222d34b10113b091f7d95e85e6c985db93086806535a28efd52002109@175.214.58.105:30303")
	v4 := InitV4(30303)
	fmt.Println(v4.FindRandomNode(node))
}

func TestPing(t *testing.T) {
	node := enode.MustParseV4("enode://8935c9600d925fd46bdf9d1d155ae682c420d75e4546bfd1de4f9cd18c13aab8edd12a1222d34b10113b091f7d95e85e6c985db93086806535a28efd52002109@175.214.58.105:30303")
	v4 := InitV4(30303)
	fmt.Println(v4.Ping(node))
}

func TestFindDist(t *testing.T) {
	node := enode.MustParseV4("enode://f645a111bf38373a17d769bb919aa5e04efa967fa2c291c743d6e3ad9a4a6c5a19d74acdc4ebf89bf278d302208d886954fb28ad324e0e28981ba42b5a20834b@216.62.161.238:30303?discport=30305")
	v4 := InitV4(30303)
	rs, err := v4.FindNode(node, node.Pubkey())
	if err != nil {
		panic(err)
	}
	for _, r := range rs {
		fmt.Println(r)
		fmt.Println(enode.LogDist(node.ID(), r.ID()))
	}
}

func TestRand(t *testing.T) {
	buf := make([]byte, 64)
	for i := 0; i < 100000; i++ {
		crand.Read(buf)
	}
	fmt.Println(buf)
}
