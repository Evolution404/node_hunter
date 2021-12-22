package discover

import (
	"fmt"
	"net"
	"node_hunter/storage"
	"os"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func InitV4(port int) *discover.UDPv4 {
	// 构造UDP连接，要使用ListenUDP不能使用DialUDP
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   []byte{},
		Port: port,
	})
	if err != nil {
		panic(err)
	}

	// 准备enode.DB对象
	db, err := enode.OpenDB("")
	if err != nil {
		panic(err)
	}

	// 准备节点私钥
	priv := storage.PrivateKey
	ln := enode.NewLocalNode(db, priv)

	logger := log.New()
	logger.SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stderr, log.LogfmtFormat())))

	// 启动节点发现协议
	udpv4, err := discover.ListenV4(conn, ln, discover.Config{
		PrivateKey: priv,
		// Log:        logger,
	})
	if err != nil {
		panic(err)
	}
	return udpv4
}

// 利用输入的启动节点和nodes文件中的节点开始探测网络
func DumpRelation(l *storage.Logger, udpv4 *discover.UDPv4, initial *enode.Node) error {
	nodeMap := make(map[enode.ID]string)
	last := 0
	count := 0
	// 查询出错超过3次停止查询
	errCount := 0
	var err error
	for {
		var rs []*enode.Node
		rs, err = udpv4.FindRandomNode(initial)
		if err != nil {
			if errCount >= 3 {
				break
			}
			errCount++
		} else {
			errCount = 0
		}
		for _, r := range rs {
			nodeMap[r.ID()] = r.URLv4()
			l.AddSeen(r)
		}
		if last == len(nodeMap) {
			if count >= 20 {
				break
			}
			count++
		} else {
			count = 0
		}
		last = len(nodeMap)
		fmt.Printf("searching: %d, %s\n", len(nodeMap), initial.URLv4())
	}
	fmt.Printf("relation: %d, %s\n", len(nodeMap), initial.URLv4())

	// 记录查询完成的时间
	l.AddFinished(initial)

	// 写入节点关系记录
	relation := fmt.Sprintf("%s %d", initial.URLv4(), len(nodeMap))
	for _, url := range nodeMap {
		relation += " " + url
	}
	l.Relation <- relation
	return err
}

func StartDiscover(nodes []*enode.Node, threads int) {
	fmt.Printf("start discover: threads=%d\n", threads)
	l := storage.StartLog(nodes)
	defer l.Close()

	udpv4 := InitV4(30303)

	// 控制同时查询的协程数
	token := make(chan struct{}, threads)
	for i := 0; i < threads; i++ {
		token <- struct{}{}
	}
	var running int32 = 0
	// 不断循环所有节点进行搜索
	for {
		for _, node := range l.AllNodes {
			// 查询过不再查询
			if l.Seen(node.ID()) > 0 {
				continue
			}
			<-token
			fmt.Println("start search:", node.URLv4())
			// 避免重复查询，在开始查询的时候就记录一下时间
			l.AddFinished(node)
			atomic.AddInt32(&running, 1)
			go func(n *enode.Node) {
				err := DumpRelation(l, udpv4, n)
				if err != nil {
					fmt.Println(err)
				}
				token <- struct{}{}
				atomic.AddInt32(&running, -1)
			}(node)
		}
		time.Sleep(time.Second * 3)
		if atomic.LoadInt32(&running) > 0 {
			fmt.Printf("all nodes finished, running goroutine=%d\n", atomic.LoadInt32(&running))
		} else {
			fmt.Printf("all nodes finished, stop")
			break
		}
	}
}
