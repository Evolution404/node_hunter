package storage

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

func TestWriteNode(t *testing.T) {
	node := enode.MustParseV4("enode://6da566ba5f4e82cf07969915fc6c0f8e33783ccd07561e68de51ec761606c648cb139f6f3142138707902224261cae4b4f4126141792f4250cb1d39aa7c73fce@77.170.227.84:30303")
	l := StartLog(nil, true)
	fmt.Println(l.HasNode(node))
	l.WriteNode(node)
	fmt.Println(l.HasNode(node))
}
