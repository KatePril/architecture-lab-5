package datastore

import (
	"bytes"
	"testing"
)

func TestEntry_Encode(t *testing.T) {
	raw := Encode(entryRecord{"key", "value"})
	record, _, _ := ReadRecord(bytes.NewReader(raw), 0)
	entry, ok := record.(entryRecord)
	if !ok {
		t.Fatal("invalid convertation")
	}
	if entry.key != "key" {
		t.Error("incorrect key")
	}
	if entry.value != "value" {
		t.Error("incorrect value")
	}
}
