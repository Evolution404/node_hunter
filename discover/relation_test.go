package discover

import (
	"fmt"
	"sync"
	"testing"

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
