package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var nodesPrefix = "n"
var relationPrefix = "r"
var rlpxPrefix = "x"
var enrPrefix = "e"

var doing = "i"
var done = "d"

var todayRelationPrefix = date + relationPrefix
var todayRelationDoingPrefix = date + relationPrefix + doing
var todayRelationDonePrefix = date + relationPrefix + done

var todayRlpxPrefix = date + rlpxPrefix
var todayEnrPrefix = date + enrPrefix

// 数据库中键的类型
type KeyType int

const (
	Node KeyType = iota
	Relation
	RelationDoing
	RelationDone
	Rlpx
	ENR
	Unknown
)

func keyType(key []byte) KeyType {
	if bytes.HasPrefix(key, []byte(nodesPrefix)) {
		return Node
	}
	// 去除日期前缀
	key = key[len(date):]
	if bytes.HasPrefix(key, []byte(relationPrefix+doing)) {
		return RelationDoing
	} else if bytes.HasPrefix(key, []byte(relationPrefix+done)) {
		return RelationDone
	} else if bytes.HasPrefix(key, []byte(relationPrefix)) {
		return Relation
	} else if bytes.HasPrefix(key, []byte(rlpxPrefix)) {
		return Rlpx
	} else if bytes.HasPrefix(key, []byte(enrPrefix)) {
		return ENR
	} else {
		return Unknown
	}
}

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
	ret, err := l.db.Has([]byte(nodesPrefix+n.URLv4()), nil)
	if err != nil {
		panic(err)
	}
	return ret
}

func (l *Logger) WriteRelation(from *enode.Node, to *enode.Node) bool {
	if l.HasRelation(from, to) {
		return false
	}
	key := todayRelationPrefix + from.URLv4() + to.URLv4()
	now := time.Now().Unix()
	err := l.db.Put([]byte(key), int64ToBytes(now), nil)
	if err != nil {
		panic(err)
	}
	return true
}

func (l *Logger) HasRelation(from *enode.Node, to *enode.Node) bool {
	key := todayRelationPrefix + from.URLv4() + to.URLv4()
	ret, err := l.db.Has([]byte(key), nil)
	if err != nil {
		panic(err)
	}
	return ret
}

// 统计某个节点认识的节点个数
func (l *Logger) Relations(from *enode.Node) int {
	count := 0
	key := todayRelationPrefix + from.URLv4()
	iter := l.db.NewIterator(util.BytesPrefix([]byte(key)), nil)
	for iter.Next() {
		count++
	}
	return count
}

func (l *Logger) RelationDoing(from *enode.Node) {
	key := todayRelationDoingPrefix + from.URLv4()
	now := time.Now().Unix()
	err := l.db.Put([]byte(key), int64ToBytes(now), nil)
	if err != nil {
		panic(err)
	}
}

func (l *Logger) IsRelationDoing(from *enode.Node) bool {
	key := todayRelationDoingPrefix + from.URLv4()
	ret, err := l.db.Has([]byte(key), nil)
	if err != nil {
		panic(err)
	}
	return ret
}

func (l *Logger) RelationDone(from *enode.Node) {
	// 删除不能失败
	if err := l.db.Delete([]byte(todayRelationDoingPrefix+from.URLv4()), nil); err != nil {
		panic(err)
	}
	key := todayRelationDonePrefix + from.URLv4()
	now := time.Now().Unix()
	err := l.db.Put([]byte(key), int64ToBytes(now), nil)
	if err != nil {
		panic(err)
	}
}

func (l *Logger) IsRelationDone(from *enode.Node) bool {
	key := todayRelationDonePrefix + from.URLv4()
	ret, err := l.db.Has([]byte(key), nil)
	if err != nil {
		panic(err)
	}
	return ret
}

func (l *Logger) shouldRelation(url string) bool {
	doingKey := todayRelationDoingPrefix + url
	doneKey := todayRelationDonePrefix + url
	has1, err1 := l.db.Has([]byte(doingKey), nil)
	has2, err2 := l.db.Has([]byte(doneKey), nil)

	if err1 != nil {
		panic(err1)
	}
	if err2 != nil {
		panic(err2)
	}

	// 有两者之一都不应该再进行查询
	if has1 || has2 {
		return false
	}
	return true
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
	// 在前方追加时间戳
	now := time.Now().Unix()
	info = fmt.Sprintf("%d", now) + info

	err := l.db.Put([]byte(todayRlpxPrefix+n.URLv4()), []byte(info), nil)
	if err != nil {
		panic(err)
	}
	return true
}

func (l *Logger) HasRlpx(n *enode.Node) bool {
	ret, err := l.db.Has([]byte(todayRlpxPrefix+n.URLv4()), nil)
	if err != nil {
		panic(err)
	}
	return ret
}

func (l *Logger) WriteEnr(n *enode.Node, err error) bool {
	if l.HasEnr(n) {
		return false
	}
	now := time.Now().Unix()
	str := fmt.Sprintf("%d", now)
	if err != nil {
		str += "e" + err.Error()
	} else {
		str += "i" + n.String()
	}
	err = l.db.Put([]byte(todayEnrPrefix+n.URLv4()), []byte(str), nil)
	if err != nil {
		panic(err)
	}
	return true
}

func (l *Logger) HasEnr(n *enode.Node) bool {
	ret, err := l.db.Has([]byte(todayEnrPrefix+n.URLv4()), nil)
	if err != nil {
		panic(err)
	}
	return ret
}

func int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func bytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}
