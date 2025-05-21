package datastore

import (
	"encoding/binary"
	"io"
	"iter"
	"os"
)

type entry struct {
	key, value string
	isDeleted  byte
}

// 0    4     4+kl  4+kl+4   4+kl+4+vl   <-- offset
// (kl) (key) (vl)  (value)  (isDeleted)
// 4    ..... 4     ......   1           <-- length

func Encode(e entry) []byte {
	kl, vl := len(e.key), len(e.value)
	size := kl + vl + 9
	res := make([]byte, size)
	binary.LittleEndian.PutUint32(res, uint32(kl))
	copy(res[4:], e.key)
	binary.LittleEndian.PutUint32(res[kl+4:], uint32(vl))
	copy(res[kl+8:], e.value)
	res[kl+vl+8] = e.isDeleted
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

func readIsDeletedFlag(file io.ReaderAt, offset int64) (byte, error) {
	flag := make([]byte, 1)
	_, err := file.ReadAt(flag, offset)
	if err != nil {
		return 0, err
	}
	return flag[0], nil
}

func ReadEntry(file io.ReaderAt, offset int64) (entry, error) {
	key, err := readValue(file, offset)
	if err != nil {
		return entry{}, err
	}
	value, err := readValue(file, offset+int64(len(key))+4)
	if err != nil {
		return entry{}, err
	}
	isDeleted, err := readIsDeletedFlag(file, offset+int64(len(key)+len(value)+8))
	if err != nil {
		return entry{}, err
	}
	return entry{key, value, isDeleted}, nil
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
