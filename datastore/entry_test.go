package datastore

import (
	"bytes"
	"testing"
)

func TestEntry_Encode(t *testing.T) {
	raw := Encode(entry{"key", "value", 0})
	e, _ := ReadEntry(bytes.NewReader(raw), 0)
	if e.key != "key" {
		t.Error("incorrect key")
	}
	if e.value != "value" {
		t.Error("incorrect value")
	}
	if e.kind != 0 {
		t.Error("incorrect kind")
	}
}
