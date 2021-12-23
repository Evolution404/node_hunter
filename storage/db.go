package storage

import (
	"encoding/binary"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var nodesPrefix = "n"
var relationPrefix = "r" + date
var rlpxPrefix = "x" + date
var enrPrefix = "e" + date

var doing = "i"
var done = "d"

func openDB() *leveldb.DB {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		panic(err)
	}
	return db
}

func (l *Logger) WriteNode(n *enode.Node) bool {
	if l.HasNode(n) {
		return false
	}
	l.waitingLock.Lock()
	l.waitingNodes = append(l.waitingNodes, n)
	l.waitingLock.Unlock()
	now := time.Now().Unix()
	err := l.db.Put([]byte(nodesPrefix+n.URLv4()), int64ToBytes(now), nil)
	if err != nil {
		panic(err)
	}
	return true
}

func (l *Logger) HasNode(n *enode.Node) bool {
	_, err := l.db.Get([]byte(nodesPrefix+n.URLv4()), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

func (l *Logger) WriteRelation(from *enode.Node, to *enode.Node) bool {
	if l.HasRelation(from, to) {
		return false
	}
	key := relationPrefix + from.URLv4() + to.URLv4()
	now := time.Now().Unix()
	err := l.db.Put([]byte(key), int64ToBytes(now), nil)
	if err != nil {
		panic(err)
	}
	return true
}

func (l *Logger) HasRelation(from *enode.Node, to *enode.Node) bool {
	key := relationPrefix + from.URLv4() + to.URLv4()
	_, err := l.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

// 统计某个节点认识的节点个数
func (l *Logger) Relations(from *enode.Node) int {
	count := 0
	key := relationPrefix + from.URLv4()
	iter := l.db.NewIterator(util.BytesPrefix([]byte(key)), nil)
	for iter.Next() {
		count++
	}
	return count
}

func (l *Logger) RelationDoing(from *enode.Node) {
	key := relationPrefix + doing + from.URLv4()
	now := time.Now().Unix()
	err := l.db.Put([]byte(key), int64ToBytes(now), nil)
	if err != nil {
		panic(err)
	}
}

func (l *Logger) IsRelationDoing(from *enode.Node) bool {
	key := relationPrefix + doing + from.URLv4()
	_, err := l.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

func (l *Logger) RelationDone(from *enode.Node) {
	l.db.Delete([]byte(relationPrefix+doing+from.URLv4()), nil)
	key := relationPrefix + done + from.URLv4()
	now := time.Now().Unix()
	err := l.db.Put([]byte(key), int64ToBytes(now), nil)
	if err != nil {
		panic(err)
	}
}

func (l *Logger) IsRelationDone(from *enode.Node) bool {
	key := relationPrefix + "d" + from.URLv4()
	_, err := l.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

func (l *Logger) shouldRelation(url string) bool {
	doingKey := relationPrefix + doing + url
	doneKey := relationPrefix + done + url
	_, err1 := l.db.Get([]byte(doingKey), nil)
	_, err2 := l.db.Get([]byte(doneKey), nil)
	if err1 == leveldb.ErrNotFound && err2 == leveldb.ErrNotFound {
		return true
	}
	if err1 != nil {
		panic(err1)
	}
	if err2 != nil {
		panic(err2)
	}
	return false
}

func (l *Logger) GetWaiting() *enode.Node {
	l.waitingLock.Lock()
	defer l.waitingLock.Unlock()
	if len(l.waitingNodes) == 0 {
		return nil
	}
	first := l.waitingNodes[0]
	l.waitingNodes = l.waitingNodes[1:]
	return first
}

func (l *Logger) NextNode() *enode.Node {
	if !l.nodeIter.Next() {
		return nil
	}
	url := string(l.nodeIter.Key()[len(nodesPrefix):])
	return enode.MustParseV4(url)
}

func (l *Logger) WriteRlpx(n *enode.Node, info string) bool {
	if l.HasRlpx(n) {
		return false
	}
	err := l.db.Put([]byte(rlpxPrefix+n.URLv4()), []byte(info), nil)
	if err != nil {
		panic(err)
	}
	return true
}

func (l *Logger) HasRlpx(n *enode.Node) bool {
	_, err := l.db.Get([]byte(rlpxPrefix+n.URLv4()), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

func (l *Logger) WriteEnr(n *enode.Node, enr string) bool {
	if l.HasRlpx(n) {
		return false
	}
	err := l.db.Put([]byte(enrPrefix+n.URLv4()), []byte(enr), nil)
	if err != nil {
		panic(err)
	}
	return true
}

func (l *Logger) HasEnr(n *enode.Node) bool {
	_, err := l.db.Get([]byte(enrPrefix+n.URLv4()), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

func int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func bytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}
