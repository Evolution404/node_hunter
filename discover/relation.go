package discover

import (
	"fmt"
	"node_hunter/config"
	"node_hunter/rlpx"
	"node_hunter/storage"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// 控制不论正在运行多少查询线程，都是每秒打印一个节点的信息
var defaultSessionLogInterval = time.Second

// 每个正在查询的节点的日志时间间隔
var sessionLogInterval = time.Second

type session struct {
	initial    *enode.Node // 要查询的节点
	udpv4      *discover.UDPv4
	l          *storage.Logger
	rtt        time.Duration // 查询这个节点的rtt时间
	threads    int
	maxThreads int
	errCount   int32 // 出现错误的次数，一旦查询成功就归零
	err        error // 最后的错误
	nodes      int32 // 这个节点认识的节点个数

	noEnr  bool
	noRlpx bool
}

func newSession(l *storage.Logger, udpv4 *discover.UDPv4, initial *enode.Node, maxThreads int, noEnr, noRlpx bool) *session {
	return &session{
		initial:    initial,
		udpv4:      udpv4,
		l:          l,
		rtt:        time.Millisecond * 100,
		nodes:      int32(l.NodeRelations(initial)),
		maxThreads: maxThreads,
		noEnr:      noEnr,
		noRlpx:     noRlpx,
	}
}

// 执行在一个RTT时间内的查询
// 根据之前的RTT时间来确定要查询的线程数
func (s *session) doRTT() int {
	udpv4 := s.udpv4
	start := time.Now()
	// rtt时间是100ms的多少倍，就使用多少线程查询
	threads := int(s.rtt / (time.Millisecond * 100))
	// 最少使用一个线程查询，有错误也使用一个线程
	if threads == 0 || s.err != nil {
		threads = 1
	}
	// 最多10个线程
	if threads > s.maxThreads {
		threads = s.maxThreads
	}
	s.threads = threads

	var wg sync.WaitGroup
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rs, err := udpv4.FindRandomNode(s.initial)
			if err != nil {
				atomic.AddInt32(&s.errCount, 1)
				s.err = err
			} else {
				atomic.StoreInt32(&s.errCount, 0)
				s.err = nil
			}
			for _, r := range rs {
				s.l.WriteNode(r)
				// 新写入了认识节点，增加计数
				if s.l.WriteRelation(s.initial, r) {
					atomic.AddInt32(&s.nodes, 1)
				}
			}
		}()
	}
	wg.Wait()
	end := time.Now()
	// 新的rtt时间是此次的与之前的取平均
	s.rtt = (end.Sub(start) + s.rtt) / 2
	// 返回此次实际执行了多少次查询
	return threads
}

func (s *session) do() error {
	fmt.Println("start search:", s.initial.URLv4())
	done := make(chan struct{})
	// 等待enr和rlpx执行完成
	var wg sync.WaitGroup

	// 每五秒打印一次
	go func() {
		for {
			select {
			case <-done:
				return
			case <-time.Tick(time.Second * 5):
				count := s.nodes
				// 节点数超过0，或者报错了才打印
				if count != 0 || (s.err != nil && s.err.Error() != "RPC timeout") {
					if s.err != nil {
						fmt.Printf("count: %d, rtt: %v, threads: %d, err: %v %s\n", count, s.rtt/time.Millisecond*time.Millisecond, s.threads, s.err, s.initial.URLv4())
					} else {
						fmt.Printf("count: %d, rtt: %v, threads: %d %s\n", count, s.rtt/time.Millisecond*time.Millisecond, s.threads, s.initial.URLv4())
					}
				}
			}
		}
	}()

	// 查询enr记录
	if !s.noEnr {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if s.l.HasEnr(s.initial) {
				return
			}
			for i := 0; i < 3; i++ {
				nn, err := s.udpv4.RequestENR(s.initial)
				if err == nil {
					s.l.WriteEnr(s.initial, nn, err)
					fmt.Println("enr done:", nn.URLv4(), "seq:", nn.Seq())
					break
				}
				if i == 2 {
					s.l.WriteEnr(s.initial, nil, err)
					fmt.Println("error enr:", s.initial.URLv4(), err)
				}
			}
		}()
	}

	// 查询rlpx记录
	if !s.noRlpx {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if s.l.HasRlpx(s.initial) {
				return
			}
			q := rlpx.NewQuery()
			err := q.QueryNode(s.l, s.initial)
			if err != nil {
				if err.Error() == "too many open files" {
					panic(err)
				}
			}
		}()
	}

	// 查询了多少次后没有增加
	stopCount := 0
	for {
		if s.errCount >= 5 {
			break
		}
		lastCount := s.nodes
		s.doRTT()
		// threads := s.doRTT()
		newCount := s.nodes
		if newCount == lastCount {
			stopCount++
		} else {
			stopCount = 0
		}
		if stopCount >= 20 {
			break
		}
	}
	wg.Wait()
	fmt.Printf("search node done, count=%d %s\n", s.nodes, s.initial.URLv4())
	close(done)
	return s.err
}

// 查询指定的节点认识的所有节点，并导出到relation文件中
func DumpRelation(l *storage.Logger, udpv4 *discover.UDPv4, initial *enode.Node, nodeThreads int, noEnr, noRlpx bool) error {
	// 启动与对方节点的会话，并进行查询
	s := newSession(l, udpv4, initial, nodeThreads, noEnr, noRlpx)
	err := s.do()

	return err
}

func StartDiscover(nodes []*enode.Node, threads int, nodeThreads int, noEnr, noRlpx bool) {
	fmt.Printf("start discover: threads=%d\n", threads)
	l := storage.StartLog(nodes, true)
	defer l.Close()

	udpv4 := InitV4(30303)

	// 控制同时查询的线程数
	token := make(chan struct{}, threads)
	for i := 0; i < threads; i++ {
		token <- struct{}{}
	}
	var running int32 = 0
	// 每秒打印一次当前运行查询线程个数
	go func() {
		for {
			running := atomic.LoadInt32(&running)
			fmt.Printf("running search goroutine=%d\n", running)
			c := running
			if running == 0 {
				c = 1
			}
			sessionLogInterval = time.Duration(c) * defaultSessionLogInterval
			time.Sleep(time.Second)
		}
	}()
	// 不断循环所有节点进行搜索
	for {
		// 每秒打印一次当前正在有多少节点被查询
		for node := l.GetWaiting(); node != nil; node = l.GetWaiting() {
			// 不查询被拒绝的节点
			if config.Reject(node) {
				continue
			}
			<-token
			// 开始查询
			l.RelationDoing(node)
			atomic.AddInt32(&running, 1)
			go func(n *enode.Node) {
				err := DumpRelation(l, udpv4, n, nodeThreads, noEnr, noRlpx)
				if err != nil {
					fmt.Println("error", n.URLv4(), err)
				}
				l.RelationDone(n)
				token <- struct{}{}
				atomic.AddInt32(&running, -1)
			}(node)
		}
		if atomic.LoadInt32(&running) > 0 {
			fmt.Println("waiting potential new nodes")
			time.Sleep(time.Second * 3)
			fmt.Printf("all nodes finished, running goroutine=%d\n", atomic.LoadInt32(&running))
		} else {
			fmt.Println("all nodes finished, stop")
			break
		}
	}
	// 结束后删除今天的日期
	l.RemoveDate()
}
