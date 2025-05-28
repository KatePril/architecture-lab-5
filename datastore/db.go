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
	file   *os.File
	offset int64
}

type Db struct {
	directory string
	files     []*os.File
	offset    map[string]KeyStorage
}

func Open(directory string) (*Db, error) {
	database := &Db{
		directory: directory,
		files:     make([]*os.File, 0),
		offset:    make(map[string]KeyStorage),
	}
	pattern := filepath.Join(directory, outFileBase+"*")
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
	for value := range Iterate(file) {
		key := value.data.getId()
		database.offset[key] = KeyStorage{file, value.offset}
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

	err := os.MkdirAll(database.directory, 0o700)
	if err != nil {
		return nil, err
	}

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
	data, _, err := ReadRecord(keyStorage.file, keyStorage.offset)
	if err != nil {
		return "", err
	}
	switch record := data.(type) {
	case entryRecord:
		return record.value, nil
	case deleteRecord:
		return "", ErrNotFound
	default:
		return "", nil
	}
}

func (database *Db) Put(key, value string) error {
	return database.putEntry(entryRecord{key, value})
}

func (database *Db) Delete(key string) error {
	_, exists := database.offset[key]
	if !exists {
		return nil
	}

	return database.putEntry(deleteRecord(key))
}

func (database *Db) putEntry(entry record) error {
	file := database.files[len(database.files)-1]
	fileStat, err := file.Stat()
	if err != nil {
		return err
	}
	fileSize := fileStat.Size()
	if fileSize >= maxFileSize {
		if len(database.files) >= 3 {
			if err := database.mergeFiles(); err != nil {
				return err
			}
		}
		file, err = database.newFile()
		if err != nil {
			return err
		}
		database.files = append(database.files, file)
		fileSize = 0
	}
	data := Encode(entry)
	_, err = file.WriteAt(data, fileSize)
	if err != nil {
		return err
	}
	database.offset[entry.getId()] = KeyStorage{file, fileSize}
	return nil
}

func (database *Db) mergeFiles() error {
	records := make(map[string]record)

	for _, file := range database.files {
		for it := range Iterate(file) {
			key := it.data.getId()
			records[key] = it.data
		}
	}

	var currentFile *os.File
	var currentSize int64
	var newFiles []*os.File
	fileIndex := 0

	createNewFile := func() (*os.File, error) {
		filename := outFileBase + strconv.Itoa(fileIndex)
		fileIndex++
		filepath := filepath.Join(database.directory, filename)
		return os.OpenFile(filepath, mode, 0o600)
	}

	var err error
	currentFile, err = createNewFile()
	if err != nil {
		return err
	}
	newFiles = append(newFiles, currentFile)
	currentSize = 0
	newOffset := make(map[string]KeyStorage)

	for _, rec := range records {
		data := Encode(rec)
		if currentSize+int64(len(data)) > maxFileSize {
			currentFile, err = createNewFile()
			if err != nil {
				return err
			}
			newFiles = append(newFiles, currentFile)
			currentSize = 0
		}
		if data[0] != DELETE_TYPE {
			_, err := currentFile.WriteAt(data, currentSize)
			if err != nil {
				return err
			}

			newOffset[rec.getId()] = KeyStorage{currentFile, currentSize}
			currentSize += int64(len(data))
		}
	}

	for _, file := range database.files {
		file.Close()
		os.Remove(file.Name())
	}

	database.files = newFiles
	database.offset = newOffset
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
