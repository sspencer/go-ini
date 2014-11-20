package ini

import (
	"testing"
)

func TestParse1(t *testing.T) {

	type Parse struct {
		StartCmd struct {
			Foo   string `ini:"FOO"`
			Magic int    `ini:"Magic Number"`
		} `ini:"[START]"`
	}

	b := []byte(`
; ignore me
[START]
FOO=BAR
Magic Number = 42`)

	var d Parse
	err := Unmarshal(b, &d)

	if err != nil {
		t.Fatal(err)
	}

	if d.StartCmd.Foo != "BAR" {
		t.Fatal("Field FOO not set")
	} else if d.StartCmd.Magic != 42 {
		t.Fatal("Field Magic not set")
	}
}
