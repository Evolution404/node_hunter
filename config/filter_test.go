package config

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

func TestBlack(t *testing.T) {
	node := enode.MustParseV4("enode://6f04d3be3ccc7fabc1e216d6f85be945e991ee9948204e2597b29c74ca334993ccf6303e9209ce52d1b73b0b7a168efb9c11284c281c75aa852b1f73895556d8@94.79.55.28:30000")
	fmt.Println(Reject(node))
}
