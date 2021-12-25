package storage

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var nodesPrefix = "n"
var relationPrefix = "r"
var rlpxPrefix = "x"
var enrPrefix = "e"
var metaPrefix = "m"

var data = "d"
var meta = "m"

var doing = "i"
var done = "d"

// 关系表分成两类
// 数据表保存所有关系
// 元表保存正在查询的节点和完成的节点
var relationDataPrefix = relationPrefix + data
var relationMetaPrefix = relationPrefix + meta
var relationDoingPrefix = relationMetaPrefix + doing
var relationDonePrefix = relationMetaPrefix + done

var todayRelationPrefix = relationDataPrefix + date
var todayRelationDoingPrefix = relationDoingPrefix + date
var todayRelationDonePrefix = relationDonePrefix + date

var todayRlpxPrefix = rlpxPrefix + date
var todayEnrPrefix = enrPrefix + date

// 保存正在查询的日期
var todayKey = metaPrefix + "today"
var nodeCountKey = metaPrefix + "nodeCount"

// 今天查询到的各个节点的关系个数，需要加上enode链接
var todayNodeRelationCount = metaPrefix + date + "nodeRelationCount"

// 今天查询到的所有关系个数
var todayRelationCount = metaPrefix + date + "relationCount"

// 所有的关系个数
var allRelationCount = metaPrefix + "relationCount"

// 今天有多少个节点已经完成查询关系了
var todayRelationDoneCount = metaPrefix + date + "relationDoneCount"

var todayRlpxDoneCount = metaPrefix + date + "rlpxDoneCount"
var allRlpxDoneCount = metaPrefix + "rlpxDoneCount"

var todayEnrDoneCount = metaPrefix + date + "enrDoneCount"
var allEnrDoneCount = metaPrefix + "enrDoneCount"

// 数据库中键的类型
type KeyType int

const (
	Node KeyType = iota
	RelationData
	RelationDoing
	RelationDone
	Rlpx
	ENR
	Meta
	Unknown
)

func keyType(key []byte) KeyType {
	if bytes.HasPrefix(key, []byte(nodesPrefix)) {
		return Node
	} else if bytes.HasPrefix(key, []byte(relationDataPrefix)) {
		return RelationData
	} else if bytes.HasPrefix(key, []byte(relationDoingPrefix)) {
		return RelationDoing
	} else if bytes.HasPrefix(key, []byte(relationDonePrefix)) {
		return RelationDone
	} else if bytes.HasPrefix(key, []byte(rlpxPrefix)) {
		return Rlpx
	} else if bytes.HasPrefix(key, []byte(enrPrefix)) {
		return ENR
	} else if bytes.HasPrefix(key, []byte(metaPrefix)) {
		return Meta
	} else {
		return Unknown
	}
}

func openDB() *leveldb.DB {
	o := &opt.Options{
		Filter: filter.NewBloomFilter(10),
	}
	db, err := leveldb.OpenFile(dbPath, o)
	if err != nil {
		panic(err)
	}
	return db
}

func OpenDB() *leveldb.DB {
	return openDB()
}

func (l *Logger) queryDate() string {
	today := time.Now().Format("2006-01-02")
	v, err := l.db.Get([]byte(todayKey), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			l.db.Put([]byte(todayKey), []byte(today), nil)
			return today
		} else {
			panic(err)
		}
	}
	if len(string(v)) == 10 {
		return string(v)
	} else {
		panic("wrong date long")
	}
}

func (l *Logger) RemoveDate() {
	l.db.Delete([]byte(todayKey), nil)
}

func (l *Logger) WriteNode(n *enode.Node) bool {
	l.dbLock.Lock()
	defer l.dbLock.Unlock()
	if l.hasNode(n) {
		return false
	}
	l.waitingLock.Lock()
	l.waitingNodes = append(l.waitingNodes, n)
	l.waitingLock.Unlock()

	// 读取之前的个数，并自增
	count := l.nodes()
	count++

	now := time.Now().Unix()
	batch := leveldb.MakeBatch(100)
	// 写入新的个数
	batch.Put([]byte(nodeCountKey), int64ToBytes(int64(count)))
	// 写入节点记录
	batch.Put([]byte(nodesPrefix+n.URLv4()), int64ToBytes(now))
	// 执行这两次写入操作
	err := l.db.Write(batch, nil)
	if err != nil {
		panic(err)
	}
	return true
}

func (l *Logger) HasNode(n *enode.Node) bool {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.hasNode(n)
}
func (l *Logger) hasNode(n *enode.Node) bool {
	ret, err := l.db.Has([]byte(nodesPrefix+n.URLv4()), nil)
	if err != nil {
		panic(err)
	}
	return ret
}

// 查询现在有多少节点记录
func (l *Logger) nodes() int {
	return l.readCount(nodeCountKey, nodesPrefix)
}
func (l *Logger) Nodes() int {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.nodes()
}

func (l *Logger) WriteRelation(from *enode.Node, to *enode.Node) bool {
	l.dbLock.Lock()
	defer l.dbLock.Unlock()
	if l.hasRelation(from, to) {
		return false
	}
	// 自增from的关系条数
	count := l.nodeRelations(from)
	count++
	batch := leveldb.MakeBatch(100)
	batch.Put([]byte(todayNodeRelationCount+from.URLv4()), int64ToBytes(int64(count)))

	// 自增今天的关系条数
	count = l.todayRelations()
	count++
	batch.Put([]byte(todayRelationCount), int64ToBytes(int64(count)))

	// 自增总关系条数
	count = l.allRelations()
	count++
	batch.Put([]byte(allRelationCount), int64ToBytes(int64(count)))

	// 再写入具体的关系记录
	key := todayRelationPrefix + from.URLv4() + to.URLv4()
	now := time.Now().Unix()
	batch.Put([]byte(key), int64ToBytes(now))
	err := l.db.Write(batch, nil)
	if err != nil {
		panic(err)
	}
	return true
}

func (l *Logger) HasRelation(from *enode.Node, to *enode.Node) bool {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.hasRelation(from, to)
}

func (l *Logger) hasRelation(from *enode.Node, to *enode.Node) bool {
	key := todayRelationPrefix + from.URLv4() + to.URLv4()
	ret, err := l.db.Has([]byte(key), nil)
	if err != nil {
		panic(err)
	}
	return ret
}

// 统计某个节点认识的节点个数
func (l *Logger) nodeRelations(from *enode.Node) int {
	url := from.URLv4()
	return l.readCount(todayNodeRelationCount+url, todayRelationPrefix+url)
}
func (l *Logger) NodeRelations(from *enode.Node) int {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.nodeRelations(from)
}

// 统计今天总共记录了多少条关系
func (l *Logger) todayRelations() int {
	return l.readCount(todayRelationCount, todayRelationPrefix)
}
func (l *Logger) TodayRelations() int {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.todayRelations()
}

// 统计总共记录了多少条关系
func (l *Logger) allRelations() int {
	return l.readCount(allRelationCount, relationDataPrefix)
}
func (l *Logger) AllRelations() int {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.allRelations()
}

func (l *Logger) RelationDoing(from *enode.Node) {
	l.dbLock.Lock()
	defer l.dbLock.Unlock()
	key := todayRelationDoingPrefix + from.URLv4()
	now := time.Now().Unix()
	err := l.db.Put([]byte(key), int64ToBytes(now), nil)
	if err != nil {
		panic(err)
	}
}

func (l *Logger) IsRelationDoing(from *enode.Node) bool {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.isRelationDoing(from)
}
func (l *Logger) isRelationDoing(from *enode.Node) bool {
	key := todayRelationDoingPrefix + from.URLv4()
	ret, err := l.db.Has([]byte(key), nil)
	if err != nil {
		panic(err)
	}
	return ret
}

// 记录一个节点查询关系完成
// 1. 删除doing标记
// 2. 自增关系查询完成个数
// 3. 记录查询完成
func (l *Logger) RelationDone(from *enode.Node) {
	l.dbLock.Lock()
	defer l.dbLock.Unlock()
	if l.isRelationDone(from) {
		panic("duplicate done")
	}

	batch := leveldb.MakeBatch(100)
	// 删除之前的doing标记
	batch.Delete([]byte(todayRelationDoingPrefix + from.URLv4()))

	// 自增查询完成的个数
	count := l.todayRelationDones()
	count++
	batch.Put([]byte(todayRelationDoneCount), int64ToBytes(int64(count)))

	// 记录done标记
	key := todayRelationDonePrefix + from.URLv4()
	now := time.Now().Unix()
	batch.Put([]byte(key), int64ToBytes(now))
	err := l.db.Write(batch, nil)
	if err != nil {
		panic(err)
	}
}

func (l *Logger) IsRelationDone(from *enode.Node) bool {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.isRelationDone(from)
}

func (l *Logger) isRelationDone(from *enode.Node) bool {
	key := todayRelationDonePrefix + from.URLv4()
	ret, err := l.db.Has([]byte(key), nil)
	if err != nil {
		panic(err)
	}
	return ret
}

// 已经有多少节点查询完成了
func (l *Logger) todayRelationDones() int {
	return l.readCount(todayRelationDoneCount, todayRelationDonePrefix)
}
func (l *Logger) TodayRelationDones() int {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.todayRelationDones()
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
	if l.nodeIter == nil {
		l.nodeIter = l.db.NewIterator(util.BytesPrefix([]byte(nodesPrefix)), nil)
	}
	if !l.nodeIter.Next() {
		l.nodeIter.Release()
		if err := l.nodeIter.Error(); err != nil {
			panic(err)
		}
		l.nodeIter = nil
		return nil
	}
	url := string(l.nodeIter.Key()[len(nodesPrefix):])
	return enode.MustParseV4(url)
}

func (l *Logger) WriteRlpx(n *enode.Node, info string) bool {
	l.dbLock.Lock()
	defer l.dbLock.Unlock()

	if l.hasRlpx(n) {
		return false
	}

	count := l.todayRlpxs()
	count++
	batch := leveldb.MakeBatch(100)
	batch.Put([]byte(todayRlpxDoneCount), int64ToBytes(int64(count)))

	// 在前方追加时间戳
	now := int64ToBytes(time.Now().Unix())
	info = string(now) + info

	batch.Put([]byte(todayRlpxPrefix+n.URLv4()), []byte(info))

	err := l.db.Write(batch, nil)
	if err != nil {
		panic(err)
	}
	return true
}

func (l *Logger) HasRlpx(n *enode.Node) bool {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.hasRlpx(n)
}

func (l *Logger) hasRlpx(n *enode.Node) bool {
	ret, err := l.db.Has([]byte(todayRlpxPrefix+n.URLv4()), nil)
	if err != nil {
		panic(err)
	}
	return ret
}

func (l *Logger) todayRlpxs() int {
	return l.readCount(todayRlpxDoneCount, todayRlpxPrefix)
}
func (l *Logger) TodayRlpxs() int {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.todayRlpxs()
}

func (l *Logger) allRlpxs() int {
	return l.readCount(allRlpxDoneCount, rlpxPrefix)
}
func (l *Logger) AllRlpxs() int {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.allRlpxs()
}

func (l *Logger) WriteEnr(n *enode.Node, err error) bool {
	l.dbLock.Lock()
	defer l.dbLock.Unlock()
	if l.hasEnr(n) {
		return false
	}

	count := l.todayEnrs()
	count++
	batch := leveldb.MakeBatch(100)
	batch.Put([]byte(todayEnrDoneCount), int64ToBytes(int64(count)))

	now := int64ToBytes(time.Now().Unix())
	str := string(now)
	if err != nil {
		str += "e" + err.Error()
	} else {
		str += "i" + n.String()
	}
	batch.Put([]byte(todayEnrPrefix+n.URLv4()), []byte(str))
	err = l.db.Write(batch, nil)
	if err != nil {
		panic(err)
	}
	return true
}

func (l *Logger) HasEnr(n *enode.Node) bool {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.hasEnr(n)
}

func (l *Logger) hasEnr(n *enode.Node) bool {
	ret, err := l.db.Has([]byte(todayEnrPrefix+n.URLv4()), nil)
	if err != nil {
		panic(err)
	}
	return ret
}

func (l *Logger) todayEnrs() int {
	return l.readCount(todayEnrDoneCount, todayEnrPrefix)
}
func (l *Logger) TodayEnrs() int {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.todayEnrs()
}

func (l *Logger) allEnrs() int {
	return l.readCount(allEnrDoneCount, enrPrefix)
}
func (l *Logger) AllEnrs() int {
	l.dbLock.RLock()
	defer l.dbLock.RUnlock()
	return l.allEnrs()
}

// countKey代表可以直接查询到当前数量的key
// 如果不存在countKey就遍历以prefix开头的内容，获取条数
func (l *Logger) readCount(countKey, prefix string) int {
	count := 0
	v, err := l.db.Get([]byte(countKey), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			iter := l.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
			for iter.Next() {
				count++
			}
			iter.Release()
			if err := iter.Error(); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		count = int(bytesToInt64(v))
	}
	return count
}

func (l *Logger) RemoveDone() {
	dones := l.todayRelationDones()
	iter := l.db.NewIterator(util.BytesPrefix([]byte(relationDonePrefix)), nil)
	for iter.Next() {
		key := iter.Key()
		if bytes.HasPrefix(key, []byte(todayRelationDonePrefix)) {
			dones--
		}
		l.db.Delete(key, nil)
	}
	if dones != 0 {
		panic("wrong done number")
	}
	l.db.Delete([]byte(todayRelationDoneCount), nil)
}

func int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func bytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}
