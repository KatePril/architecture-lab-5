package datastore

import (
	"errors"
	"fmt"
	"math"
	"os"
	"testing"
)

func TestDb(t *testing.T) {
	tmp := t.TempDir()
	db, err := Open(tmp)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	pairs := [][]string{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k3", "v3"},
		{"k2", "v2.1"},
	}

	t.Run("put/get", func(t *testing.T) {
		for _, pair := range pairs {
			err := db.Put(pair[0], pair[1])
			if err != nil {
				t.Errorf("Cannot put %s: %s", pairs[0], err)
			}
			value, err := db.Get(pair[0])
			if err != nil {
				t.Errorf("Cannot get %s: %s", pairs[0], err)
			}
			if value != pair[1] {
				t.Errorf("Bad value returned expected %s, got %s", pair[1], value)
			}
		}
	})

	t.Run("delete", func(t *testing.T) {
		key, firstValue, secondValue, thirdValue := "k5", "v1", "v2", "v3"
		db.Put(key, firstValue)
		db.Put(key, secondValue)
		err := db.Delete(key)
		if err != nil {
			t.Errorf("Cannot delete %s: %s", key, err)
		}
		_, err = db.Get(key)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Data wasn`t deleted successfully")
		}
		db.Put(key, thirdValue)
		value, err := db.Get(key)
		if err != nil {
			t.Errorf("Cannot get %s: %s", key, err)
		}
		if value != thirdValue {
			t.Errorf("Bad value returned expected %s, got %s", thirdValue, value)
		}
	})

	t.Run("file growth", func(t *testing.T) {
		sizeBefore, err := db.Size()
		if err != nil {
			t.Fatal(err)
		}
		for _, pair := range pairs {
			err := db.Put(pair[0], pair[1])
			if err != nil {
				t.Errorf("Cannot put %s: %s", pairs[0], err)
			}
		}
		sizeAfter, err := db.Size()
		if err != nil {
			t.Fatal(err)
		}
		if sizeAfter <= sizeBefore {
			t.Errorf("Size does not grow after put (before %d, after %d)", sizeBefore, sizeAfter)
		}
	})

	t.Run("new db process", func(t *testing.T) {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		db, err = Open(tmp)
		if err != nil {
			t.Fatal(err)
		}

		uniquePairs := make(map[string]string)
		for _, pair := range pairs {
			uniquePairs[pair[0]] = pair[1]
		}

		for key, expectedValue := range uniquePairs {
			value, err := db.Get(key)
			if err != nil {
				t.Errorf("Cannot get %s: %s", key, err)
			}
			if value != expectedValue {
				t.Errorf("Get(%q) = %q, wanted %q", key, value, expectedValue)
			}
		}

	})
	t.Run("Merging", func(t *testing.T) {
		tmp := t.TempDir()
		mergeDb, err := Open(tmp)
		if err != nil {
			t.Fatal(err)
		}
		for i := range 10000000 {
			kv_pair_index := int(math.Mod(float64(i), 4.0))
			key := pairs[kv_pair_index][0]
			value := pairs[kv_pair_index][1]
			mergeDb.Put(key, value)
		}
		dirTree, err := os.ReadDir(tmp)
		if err != nil {
			t.Fatal(err)
		}

		for _, e := range dirTree {
			fmt.Println(e.Name())
		}
	})
}
