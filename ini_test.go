package ini

import (
	"testing"
)

func TestSimple(t *testing.T) {

	var d struct {
		Start struct {
			Foo   string `ini:"FOO"`
			Magic int    `ini:"Magic Number"`
		} `ini:"[START]"`
	}

	b := []byte(`
; ignore me
[START]
FOO=BAR
Magic Number = 42`)

	err := Unmarshal(b, &d)

	if err != nil {
		t.Fatal(err)
	}

	if d.Start.Foo != "BAR" {
		t.Fatal("Field FOO not set")
	} else if d.Start.Magic != 42 {
		t.Fatal("Field Magic not set")
	}
}

func TestScalar(t *testing.T) {

	var d struct {
		Start struct {
			MyString string  `ini:"MYSTRING"`
			MyInt    int     // ini tag not required if name is same (case ignored)
			MyFloat  float64 `ini:"MYFLOAT"`
			MyBool   bool    `ini:"MYBOOL"`
		} `ini:"[START]"`
	}

	b := []byte(`
[START]
MYSTRING=hello
MYINT=234
MYFLOAT=91.4
MYBOOL=yes`)

	err := Unmarshal(b, &d)

	if err != nil {
		t.Fatal(err)
	}

	if d.Start.MyString != "hello" {
		t.Fatal("Field MyString not set")
	} else if d.Start.MyInt != 234 {
		t.Fatal("Field MyInt not set")
	} else if d.Start.MyFloat != 91.4 {
		t.Fatal("Field MyFloat not set")
	} else if d.Start.MyBool != true {
		t.Fatal("Field MyBool not set")
	}

}

func TestNoSections(t *testing.T) {

	var d struct {
		Title   string
		Version string
	}

	b := []byte(`
TITLE=Go Compiler
VERSION=1.3.3
`)

	err := Unmarshal(b, &d)

	if err != nil {
		t.Fatal(err)
	}

	if d.Title != "Go Compiler" {
		t.Fatal("Field Title not set")
	} else if d.Version != "1.3.3" {
		t.Fatal("Field Version not set")
	}
}

func TestTwoSections(t *testing.T) {

	var d struct {
		Mysql struct {
			Host string
		} `ini:"[MYSQL]"`

		PdoMysql struct {
			Host string
		} `ini:"[PDOMYSQL]"`
	}

	b := []byte(`
[MYSQL]
HOST=mysql:localhost

[PDOMYSQL]
HOST=pdo:127.0.0.1
`)

	err := Unmarshal(b, &d)

	if err != nil {
		t.Fatal(err)
	}

	if d.Mysql.Host != "mysql:localhost" {
		t.Fatal("Field Mysql Host not set")
	} else if d.PdoMysql.Host != "pdo:127.0.0.1" {
		t.Fatal("Field Pdo Host not set")
	}
}

func TestMixed(t *testing.T) {

	var d struct {
		Title   string
		Version string
		Mysql   struct {
			Host string
		} `ini:"[MYSQL]"`
	}

	b := []byte(`
TITLE=Go Compiler
VERSION=1.3.3

[MYSQL]
HOST=localhost
`)

	err := Unmarshal(b, &d)

	if err != nil {
		t.Fatal(err)
	}

	if d.Title != "Go Compiler" {
		t.Fatal("Field Title not set")
	} else if d.Version != "1.3.3" {
		t.Fatal("Field Version not set")
	} else if d.Mysql.Host != "localhost" {
		t.Fatal("Field Host not set")
	}

}
