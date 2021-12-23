package config

import "github.com/ethereum/go-ethereum/p2p/enode"

// 返回true说明不查询这个节点
func Reject(n *enode.Node) bool {
	return IsBlack(n.IP())
}
