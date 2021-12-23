package storage

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var date = time.Now().Format("2006-01-02")
var BasePath string = path.Join(GetCurrentAbPath(), "data")

var dbPath string = path.Join(BasePath, "storagedb")

var nodesPath string = path.Join(BasePath, "nodes")
var NodesPath string = nodesPath

var relationPath string = path.Join(BasePath, "relation-"+date)
var rlpxPath string = path.Join(BasePath, "rlpx-"+date)
var ENRPath string = path.Join(BasePath, "/enr-"+date)

// 最终方案-全兼容
func GetCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		return getCurrentAbPathByCaller()
	}
	return dir
}

// 获取当前执行文件绝对路径
func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

// 获取当前执行文件绝对路径（go run）
func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath[:len(abPath)-8]
}
