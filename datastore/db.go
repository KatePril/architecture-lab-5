package datastore

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const outFileName = "current-data"
const maxFileSize = 10 * 1024 * 1024 // 10 МБ

var ErrNotFound = fmt.Errorf("record does not exist")

type fileRef struct {
	file   *os.File
	offset int64
	name   string
}

type Db struct {
	dir   string
	files []*fileRef // всі відкриті файли
	curr  *fileRef   // поточний файл для запису
	index map[string]entryRef
}

type entryRef struct {
	fileIndex int
	offset    int64
}

func Open(dir string) (*Db, error) {
	db := &Db{
		dir:   dir,
		files: []*fileRef{},
		index: make(map[string]entryRef),
	}

	// Знайти всі файли формату current-data-*
	files, err := filepath.Glob(filepath.Join(dir, outFileName+"-*"))
	if err != nil {
		return nil, err
	}

	for i, path := range files {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		ref := &fileRef{file: f, offset: 0, name: filepath.Base(path)}
		db.files = append(db.files, ref)

		if err := db.recoverFile(f, i); err != nil && err != io.EOF {
			return nil, err
		}
		f.Close()
	}

	// Відкрити останній файл для дозапису або створити новий
	if err := db.newFile(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *Db) recoverFile(f *os.File, fileIndex int) error {
	in := bufio.NewReader(f)
	var offset int64 = 0

	for {
		var record entry
		n, err := record.DecodeFromReader(in)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		db.index[record.key] = entryRef{
			fileIndex: fileIndex,
			offset:    offset,
		}
		offset += int64(n)
	}
	return nil
}

func (db *Db) Close() error {
	var err error
	for _, f := range db.files {
		if f.file.Close() != nil {
			err = f.file.Close()
			return err
		}
	}
	return err
}

func (db *Db) Get(key string) (string, error) {
	ref, ok := db.index[key]
	if !ok {
		return "", ErrNotFound
	}

	fileRef := db.files[ref.fileIndex]
	file, err := os.Open(filepath.Join(db.dir, fileRef.name))
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Seek(ref.offset, io.SeekStart)
	if err != nil {
		return "", err
	}

	var record entry
	if _, err = record.DecodeFromReader(bufio.NewReader(file)); err != nil {
		return "", err
	}
	return record.value, nil
}

func (db *Db) newFile() error {
	index := len(db.files)
	filename := fmt.Sprintf("current-data-%d", index)
	path := filepath.Join(db.dir, filename)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	ref := &fileRef{file: f, offset: 0, name: filename}
	db.files = append(db.files, ref)
	db.curr = ref
	return nil
}

func (db *Db) Put(key, value string) error {
	e := entry{key: key, value: value}
	data := e.Encode()

	// Ротація файлу, якщо перевищено розмір
	if db.curr.offset+int64(len(data)) > maxFileSize {
		db.curr.file.Close()
		if err := db.newFile(); err != nil {
			return err
		}
	}

	n, err := db.curr.file.Write(data)
	if err != nil {
		return err
	}

	ref := entryRef{
		fileIndex: len(db.files) - 1,
		offset:    db.curr.offset,
	}
	db.index[key] = ref
	db.curr.offset += int64(n)
	return nil
}

func (db *Db) Size() (int64, error) {
	var size int64 = 0
	for _, f := range db.files {
		fileStat, err := f.file.Stat()
		if err == nil {
			size += int64(fileStat.Size())
		}
	}
	return size, nil
}
