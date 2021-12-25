package discover

import (
	"net"
	"node_hunter/config"
	"os"
	"path"

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
	db, err := enode.OpenDB(path.Join(config.BasePath, "db"))
	if err != nil {
		panic(err)
	}

	// 准备节点私钥
	priv := config.PrivateKey
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
