package datastore

import (
	"bytes"
	"testing"
)

func TestEntry_Encode(t *testing.T) {
	raw := Encode(entry{"key", "value", 1})
	e, _ := ReadEntry(bytes.NewReader(raw), 0)
	if e.key != "key" {
		t.Error("incorrect key")
	}
	if e.value != "value" {
		t.Error("incorrect value")
	}
	if e.isDeleted != 1 {
		t.Error("incorrect isDeleted flag")
	}
}
