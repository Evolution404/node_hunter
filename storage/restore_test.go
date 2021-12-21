package storage

import (
	"bufio"
	"fmt"
	"path"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

// 去除relation文件中的重复内容
func TestDuplicateRemoval(t *testing.T) {
	var seenNode = make(map[enode.ID]int64)
	f, err := createOrOpen(relationPath)
	if err != nil {
		panic(err)
	}
	df, err := createOrOpen(path.Join(GetCurrentAbPath(), "data", "d_relation"))
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(f)
	count := 0
	for {
		lineBytes, isPrefix, err := reader.ReadLine()
		if err != nil {
			break
		}
		str := string(lineBytes)
		var timestamp, relations int64
		var url string
		fmt.Sscanf(str, "%d %s %d", &timestamp, &url, &relations)
		node := enode.MustParseV4(url)

		if seenNode[node.ID()] == 0 {
			df.Write(lineBytes)
		}

		// 只需要一行的最开始信息，此行剩余内容忽略
		for isPrefix {
			var extra []byte
			extra, isPrefix, err = reader.ReadLine()
			if err != nil {
				break
			}
			if seenNode[node.ID()] == 0 {
				df.Write(extra)
			}
		}
		if seenNode[node.ID()] == 0 {
			df.Write([]byte{'\n'})
		}
		seenNode[node.ID()] = timestamp
		count++
	}
	fmt.Println("searched count:", count)
}
