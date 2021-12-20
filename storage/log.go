package storage

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var BasePath string = GetCurrentAbPath() + "/data"

type Logger struct {
	Relation chan string
	relation *os.File
	Nodes    chan string
	nodes    *os.File
	wg       sync.WaitGroup
}

func createOrOpen(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
}

func StartLog() *Logger {
	os.MkdirAll(BasePath, 0777)
	relation, err := createOrOpen(BasePath + "/relation")
	nodes, err := createOrOpen(BasePath + "/nodes")
	if err != nil {
		panic(err)
	}
	l := &Logger{
		Relation: make(chan string, 10),
		relation: relation,
		Nodes:    make(chan string, 10),
		nodes:    nodes,
	}
	l.wg.Add(1)
	go l.loop()
	return l
}

func (l *Logger) loop() {
	defer l.wg.Done()
	for {
		now := time.Now().Unix()
		select {
		case r, ok := <-l.Relation:
			if !ok {
				return
			}
			str := fmt.Sprintf("%d %s\n", now, r)
			if _, err := l.relation.WriteString(str); err != nil {
				panic(err)
			}
		case r, ok := <-l.Nodes:
			if !ok {
				return
			}
			str := fmt.Sprintf("%d %s\n", now, r)
			if _, err := l.nodes.WriteString(str); err != nil {
				panic(err)
			}
		}
	}
}

func (l *Logger) Close() error {
	close(l.Relation)
	l.wg.Wait()
	return l.relation.Close()
}
