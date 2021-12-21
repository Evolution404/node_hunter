package storage

import (
	"io"
	"sync"
)

// 同步写文件
type SyncWriter struct {
	m sync.Mutex
	io.ReadWriter
}

func NewSyncWriter(path string) *SyncWriter {
	f, err := createOrOpen(path)
	if err != nil {
		panic(err)
	}
	f.Seek(0, io.SeekEnd)
	return &SyncWriter{
		m:          sync.Mutex{},
		ReadWriter: f,
	}
}

func (w *SyncWriter) Write(b []byte) (n int, err error) {
	w.m.Lock()
	defer w.m.Unlock()
	return w.ReadWriter.Write(b)
}
