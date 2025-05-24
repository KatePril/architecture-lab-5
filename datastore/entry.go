package datastore

import (
	"encoding/binary"
	"fmt"
	"io"
	"iter"
	"os"
	"slices"
)

const (
	ENTRY_TYPE = iota
	DELETE_TYPE
)

var types = []uint8{ENTRY_TYPE, DELETE_TYPE}

type record interface {
	getId() string
}

type entryRecord struct {
	key, value string
}

func (entry entryRecord) getId() string {
	return entry.key
}

type deleteRecord string

func (entry deleteRecord) getId() string {
	return string(entry)
}

var parsers = map[uint8]func(io.ReaderAt, int64) (record, uint32, error){
	ENTRY_TYPE: func(reader io.ReaderAt, offset int64) (record, uint32, error) {
		lengthBuffer := make([]byte, 4)
		// it is dangerous to check only first ReadAt, but if offsets are valid it is ok
		_, err := reader.ReadAt(lengthBuffer, offset)
		if err != nil {
			return nil, 0, err
		}
		keyLength := binary.LittleEndian.Uint32(lengthBuffer)
		reader.ReadAt(lengthBuffer, offset+int64(keyLength)+4)
		valLength := binary.LittleEndian.Uint32(lengthBuffer)
		keyBuffer := make([]byte, keyLength)
		valBuffer := make([]byte, valLength)
		reader.ReadAt(keyBuffer, offset+4)
		reader.ReadAt(valBuffer, offset+8+int64(keyLength))
		length := 8 + keyLength + valLength
		return entryRecord{string(keyBuffer), string(valBuffer)}, length, nil
	},
	DELETE_TYPE: func(reader io.ReaderAt, offset int64) (record, uint32, error) {
		lengthBuffer := make([]byte, 4)
		_, err := reader.ReadAt(lengthBuffer, offset)
		if err != nil {
			return nil, 0, err
		}
		keyLength := binary.LittleEndian.Uint32(lengthBuffer)
		keyBuffer := make([]byte, keyLength)
		reader.ReadAt(keyBuffer, offset+4)
		return deleteRecord(keyBuffer), 4 + keyLength, nil
	},
}

var encoders = map[uint8]func(data record) []byte{
	ENTRY_TYPE: func(data record) []byte {
		entry, _ := data.(entryRecord)
		kl, vl := len(entry.key), len(entry.value)
		size := kl + vl + 8
		buffer := make([]byte, size)
		binary.LittleEndian.PutUint32(buffer, uint32(kl))
		copy(buffer[4:], entry.key)
		binary.LittleEndian.PutUint32(buffer[kl+4:], uint32(vl))
		copy(buffer[kl+8:], entry.value)
		return buffer
	},
	DELETE_TYPE: func(data record) []byte {
		record, _ := data.(deleteRecord)
		length := len(record)
		buffer := make([]byte, 4+length)
		binary.LittleEndian.PutUint32(buffer, uint32(length))
		copy(buffer[4:], []byte(record))
		return buffer
	},
}

func Encode(data record) []byte {
	// bad style, switch to map
	var kind uint8
	switch data.(type) {
	case entryRecord:
		kind = ENTRY_TYPE
	case deleteRecord:
		kind = DELETE_TYPE
	default:
		return nil
	}
	encoded := encoders[kind](data)
	record := make([]byte, len(encoded)+1)
	record[0] = kind
	copy(record[1:], encoded)
	return record
}

func ReadRecord(file io.ReaderAt, offset int64) (record, uint32, error) {
	kindBuffer := make([]byte, 1)
	_, err := file.ReadAt(kindBuffer, offset)
	if err != nil {
		return nil, 0, err
	}
	kind := kindBuffer[0]
	if !slices.Contains(types, kind) {
		return nil, 0, fmt.Errorf("unknown record type: %d", kind)
	}
	data, size, err := parsers[kind](file, offset+1)
	return data, size + 1, err
}

// bad name
type iterator struct {
	offset int64
	data   record
}

func Iterate(file *os.File) iter.Seq[iterator] {
	return func(yield func(iterator) bool) {
		var offset int64
		for {
			data, size, err := ReadRecord(file, offset)
			if err != nil {
				return
			}
			if !yield(iterator{offset, data}) {
				return
			}
			offset += int64(size)
		}
	}
}
