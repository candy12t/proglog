package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8 // uint64 size
)

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}

	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

// 書き込まれるデータ構造
// 0  1  2  3  4  5  6  7  8  9  10  11  12  13  ... n        ...     n+8       ...     n+8+m   ...   (byte)
// +-----------------------+-------------------------+-----------------+------------------+---------
// |      Byte Length      |       Byte Content      |   Byte Length   |   Byte Content   |   Byte Length
// +-----------------------+-------------------------+-----------------+------------------+---------
func (s *store) Append(p []byte) (uint64, uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pos := s.size

	// レコードを読み出すときに何バイト読めばよいかわかるようにするために、レコードを長さを書き込む
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}

	// ファイルに直接書き込まないで、バッファに書き込むことで、システムコール数を減らす
	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}

	w += lenWidth
	s.size += uint64(w)
	return uint64(w), pos, nil
}

func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return nil, err
	}

	// 8 byte読み込んで、レコードの長さを取得
	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}

	// 取得した長さ分読み込んで、レコードの内容を取得
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return err
	}
	return s.File.Close()
}
