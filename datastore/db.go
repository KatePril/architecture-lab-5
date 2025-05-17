package datastore

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

const outFileBase = "current-data-"
const maxFileSize = 10 * 1024 * 1024
const mode = os.O_RDWR | os.O_CREATE

var ErrNotFound = errors.New("record does not exist")

type KeyStorage struct {
	file *os.File
	offset int64
}

type Db struct {
	directory string
	files []*os.File
	offset map[string]KeyStorage
}

func Open(directory string) (*Db, error) {
	database := &Db{
		directory: directory,
		files: make([]*os.File, 0),
		offset: make(map[string]KeyStorage),
	}
	pattern := filepath.Join(directory, outFileBase + "*")
	files, _ := filepath.Glob(pattern)
	for _, filepath := range files {
		if file, err := os.OpenFile(filepath, os.O_RDWR, 0o600); err == nil {
			database.files = append(database.files, file)
			if database.recover(file) == nil {
				continue
			}
		}
		database.Close()
		return nil, errors.New("cannot read file " + filepath)
	}
	if len(database.files) == 0 {
		file, err := database.newFile()
		if err != nil {
			return nil, err
		}
		database.files = append(database.files, file)
	}
	return database, nil
}

func (database *Db) recover(file *os.File) error {
	var offset int64
	for pair := range Iterate(file) {
		database.offset[pair.key] = KeyStorage{ file, offset }
		offset += int64(len(pair.key) + len(pair.value) + 8)
	}
	return nil
}

func (database *Db) Close() error {
	for _, file := range database.files {
		if err := file.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (database *Db) newFile() (*os.File, error) {
	filename := outFileBase + strconv.Itoa(len(database.files))
	filepath := filepath.Join(database.directory, filename)
	file, err := os.OpenFile(filepath, mode, 0o600)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (database *Db) Get(key string) (string, error) {
	keyStorage, exists := database.offset[key]
	if !exists {
		return "", ErrNotFound
	}
	entry, err := ReadEntry(keyStorage.file, keyStorage.offset)
	if err != nil {
		return "", err
	}
	return entry.value, nil
}

func (database *Db) Put(key, value string) error {
	file := database.files[len(database.files) - 1]
	fileStat, err := file.Stat()
	if err != nil {
		return err
	}
	fileSize := fileStat.Size()
	if fileSize >= maxFileSize {
		file, err = database.newFile()
		if err != nil {
			return err
		}
		database.files = append(database.files, file)
		fileSize = 0
	}
	data := Encode(entry{ key, value })
	_, err = file.WriteAt(data, fileSize)
	if err != nil {
		return err
	}
	database.offset[key] = KeyStorage{ file, fileSize }
	return nil
}

func (database *Db) Size() (int64, error) {
	var total int64
	for _, file := range database.files {
		stat, err := file.Stat()
		if err != nil {
			return 0, err
		}
		total += stat.Size()
	}
	return total, nil
}
