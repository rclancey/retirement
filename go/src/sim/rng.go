package sim

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
	"os"
)

type SeedPod struct {
	src *os.File
	n int
	offset int64
	size int64
}

func NewSeedPod(src string, offset int64) (*SeedPod, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return &SeedPod{
		src: f,
		n: -1,
		offset: offset,
		size: info.Size(),
	}, nil
}

func (s *SeedPod) Get(i int) (*rand.Rand, error) {
	pos := (int64(i / 4) * 32 + s.offset) % s.size
	s.src.Seek(pos, os.SEEK_SET)
	buf := make([]byte, 32)
	n, err := s.src.Read(buf)
	if err != nil {
		return nil, err
	}
	if n < 32 {
		s.src.Seek(0, os.SEEK_SET)
		xbuf := make([]byte, 32 - n)
		_, err := s.src.Read(xbuf)
		if err != nil {
			return nil, err
		}
		for j, v := range xbuf {
			buf[n+j] = v
		}
	}
	m := (i % 4) * 8
	sum := sha256.Sum256(buf)
	seedBytes := bytes.NewReader(sum[m:m + 8])
	var seed int64
	err = binary.Read(seedBytes, binary.BigEndian, &seed)
	if err != nil {
		return nil, err
	}
	source := rand.NewSource(seed)
	return rand.New(source), nil
}

func (s *SeedPod) Next() (*rand.Rand, error) {
	s.n += 1
	return s.Get(s.n)
}

func (s *SeedPod) Reset() {
	s.n = -1
}

