package ini

import (
	"testing"
)

type Parse1 struct {
	StartCmd struct {
		Foo string `ini:"FOO"`
	} `ini:"[START]"`
}

func TestParse1(t *testing.T) {
	b := []byte(`
; ignore me
[START]
FOO=BAR`)

	var d Parse1
	err := Unmarshal(b, &d)

	if err != nil {
		t.Fatal(err)
	}

	if d.StartCmd.Foo != "BAR" {
		t.Fatal("Field FOO not set")
	}
}
