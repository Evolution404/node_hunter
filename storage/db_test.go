package storage

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func TestWriteNode(t *testing.T) {
	node := enode.MustParseV4("enode://6da566ba5f4e82cf07969915fc6c0f8e33783ccd07561e68de51ec761606c648cb139f6f3142138707902224261cae4b4f4126141792f4250cb1d39aa7c73fce@77.170.227.84:30303")
	l := StartLog(nil, true)
	fmt.Println(l.HasNode(node))
	l.WriteNode(node)
	fmt.Println(l.HasNode(node))
}

func TestShowRE(t *testing.T) {
	db := openDB()
	iter := db.NewIterator(util.BytesPrefix([]byte(rlpxPrefix)), nil)
	for iter.Next() {
		key := string(iter.Key())
		val := string(iter.Value())
		fmt.Println(key, val)
	}
	iter = db.NewIterator(util.BytesPrefix([]byte(enrPrefix)), nil)
	for iter.Next() {
		key := string(iter.Key())
		val := string(iter.Value())
		fmt.Println(key, val)
	}

	// iter = db.NewIterator(util.BytesPrefix([]byte(nodesPrefix)), nil)
	// for iter.Next() {
	// 	key := string(iter.Key())
	// 	val := string(iter.Value())
	// 	fmt.Println(key, val)
	// }
}

func TestCheck(t *testing.T) {
	date = "2021-12-24"
	updateDate()
	db := openDB()
	v, _ := db.Get([]byte(nodeCountKey), nil)
	nodes := bytesToInt64(v)
	nodesIter := db.NewIterator(util.BytesPrefix([]byte(nodesPrefix)), nil)
	count := 0
	for nodesIter.Next() {
		count++
	}
	fmt.Println("nodes")
	fmt.Println(nodes == int64(count))

	v, _ = db.Get([]byte(allRelationCount), nil)
	v2, _ := db.Get([]byte(allRelationCount), nil)
	relations := bytesToInt64(v)
	todayRelations := bytesToInt64(v2)
	relationIter := db.NewIterator(util.BytesPrefix([]byte(relationDataPrefix)), nil)
	count = 0
	for relationIter.Next() {
		count++
	}

	fmt.Println("relations")
	// fmt.Println(relations)
	// fmt.Println(todayRelations)
	// fmt.Println(count)
	fmt.Println(relations == todayRelations)
	fmt.Println(relations == int64(count))

	v, _ = db.Get([]byte(todayRelationDoneCount), nil)
	done := bytesToInt64(v)
	doneIter := db.NewIterator(util.BytesPrefix([]byte(todayRelationDonePrefix)), nil)
	count = 0
	for doneIter.Next() {
		count++
	}
	fmt.Println("done")
	fmt.Println(done)
	fmt.Println(count)
	fmt.Println(done == int64(count))

	v, _ = db.Get([]byte(todayEnrDoneCount), nil)
	enrs := bytesToInt64(v)
	enrIter := db.NewIterator(util.BytesPrefix([]byte(enrPrefix)), nil)
	count = 0
	for enrIter.Next() {
		count++
	}
	fmt.Println("enrs")
	fmt.Println(enrs)
	fmt.Println(count)
	fmt.Println(enrs == int64(count))

	v, _ = db.Get([]byte(todayRlpxDoneCount), nil)
	rlpxs := bytesToInt64(v)
	rlpxIter := db.NewIterator(util.BytesPrefix([]byte(rlpxPrefix)), nil)
	count = 0
	for rlpxIter.Next() {
		count++
	}
	fmt.Println("rlpxs")
	fmt.Println(rlpxs)
	fmt.Println(count)
	fmt.Println(rlpxs == int64(count))
}

func TestParseFrom(t *testing.T) {
	n := enode.MustParseV4("enode://40468e55b635e9513ed4cc54434b34c0f3866c4ac11d7d0827643e9184689a3325c55a00ddc6a8901fadfd018c646192e002544962410cb8ddce8ba6c2b9d350@168.119.18.20:13580?discport=30303")
	fmt.Println(parseFrom(n))
}

func TestRelations(t *testing.T) {
	db := openDB()
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		if len(iter.Value()) == 8 {
			fmt.Println(string(iter.Key()), bytesToInt64(iter.Value()))
		} else {
			fmt.Println(string(iter.Key()), string(iter.Value()))
		}
	}
}
