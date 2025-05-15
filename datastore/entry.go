package datastore

import (
	"encoding/binary"
	"io"
	"iter"
	"os"
)

type entry struct {
	key, value string
}

// 0    4     4+kl  4+kl+4     <-- offset
// (kl) (key) (vl)  (value)
// 4    ..... 4     ......     <-- length

func Encode(e entry) []byte {
	kl, vl := len(e.key), len(e.value)
	size := kl + vl + 8
	res := make([]byte, size)
	binary.LittleEndian.PutUint32(res, uint32(kl))
	copy(res[4:], e.key)
	binary.LittleEndian.PutUint32(res[kl + 4:], uint32(vl))
	copy(res[kl + 8:], e.value)
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
	_, err = file.ReadAt(data, offset + 4)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ReadEntry(file io.ReaderAt, offset int64) (entry, error) {	
	key, err := readValue(file, offset)
	if err != nil {
		return entry{ }, err
	}
	value, err := readValue(file, offset + int64(len(key)) + 4)
	if err != nil {
		return entry{ }, err
	}
	return entry{ key, value }, nil
}

func Iterate(file *os.File) iter.Seq[entry] {
	return func(yield func(entry) bool) {
		var offset int64
		for {
			e, err := ReadEntry(file, offset)
			if err != nil {
				return
			}
			offset += int64(len(e.key) + len(e.value) + 8)
			if !yield(e) {
				return
			}
		}
	}
}
