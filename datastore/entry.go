package datastore

import (
	"encoding/binary"
	"io"
	"iter"
	"os"
)

type entry struct {
	key, value string
	kind       uint8
}

// 0      1    5     5+kl  9+kl     <-- offset
// (flag) (kl) (key) (vl)  (value)
// 1      4    ..... 4     ......   <-- length

func Encode(e entry) []byte {
	kl, vl := len(e.key), len(e.value)
	size := kl + vl + 9
	res := make([]byte, size)
	res[0] = e.kind
	binary.LittleEndian.PutUint32(res[1:], uint32(kl))
	copy(res[5:], e.key)
	binary.LittleEndian.PutUint32(res[kl+5:], uint32(vl))
	copy(res[kl+9:], e.value)
	return res
}

func readValue(file io.ReaderAt, offset int64) (string, error) {
	lengthBuffer := make([]byte, 4)
	_, err := file.ReadAt(lengthBuffer, offset)
	if err != nil {
		return "", err
	}
	length := int64(binary.LittleEndian.Uint32(lengthBuffer))
	data := make([]byte, length)
	_, err = file.ReadAt(data, offset+4)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ReadEntry(file io.ReaderAt, offset int64) (entry, error) {
	kindBuffer := make([]byte, 1)
	file.ReadAt(kindBuffer, offset)
	kind := kindBuffer[0]
	key, err := readValue(file, offset+1)
	if err != nil {
		return entry{}, err
	}
	value, err := readValue(file, offset+int64(len(key))+5)
	if err != nil {
		return entry{}, err
	}

	return entry{key, value, kind}, nil
}

func Iterate(file *os.File) iter.Seq[entry] {
	return func(yield func(entry) bool) {
		var offset int64
		for {
			e, err := ReadEntry(file, offset)
			if err != nil {
				return
			}
			offset += int64(len(e.key) + len(e.value) + 9)
			if !yield(e) {
				return
			}
		}
	}
}
