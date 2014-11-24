package ini

import (
	//	"log"
	"bytes"
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

func TestNew(t *testing.T) {
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
UNMATCHED=ME
Magic Number = 42`)

	ini := NewDecoder(bytes.NewReader(b))
	err := ini.Decode(&d)

	if err != nil {
		t.Fatal(err)
	}

	unmatched := ini.Unmatched()
	if d.Start.Foo != "BAR" {
		t.Fatal("FOO not set")
	} else if d.Start.Magic != 42 {
		t.Fatal("Magic not set")
	} else if len(unmatched) != 1 {
		t.Fatal("Wrong number of unmatched lines")
	} else if unmatched[0].line != "UNMATCHED=ME" {
		t.Fatal("Unmatched line does not match")
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

func TestDeep(t *testing.T) {

	var d struct {
		SecurePort      int    `ini:"SET OPTION SECURE AUTH PORT"`
		FallbackAddress string `ini:"SET OPTION SERVER FALLBACK ADDRESS"`
		NetworkId       int    `ini:"SET OPTION NETWORK ID"`
		Download        struct {
			MaxSpeed      int    `ini:"SET OPTION CONTENT DOWNLOAD MAX KBS"`
			DownloadStart string `ini:"SET OPTION NETWORK DOWNLOAD WINDOW START"`
			DownloadEnd   string `ini:"SET OPTION NETWORK DOWNLOAD WINDOW END"`
		} `ini:"-"`
	}
	b := []byte(`
SET OPTION SECURE AUTH PORT=8080
SET OPTION SERVER FALLBACK ADDRESS=127.0.0.1
SET OPTION CONTENT DOWNLOAD MAX KBS=56
SET OPTION NETWORK DOWNLOAD WINDOW START=22:00:00
SET OPTION NETWORK DOWNLOAD WINDOW END=23:59:00
SET OPTION NETWORK ID=53
`)

	err := Unmarshal(b, &d)

	if err != nil {
		t.Fatal(err)
	}

	if d.SecurePort != 8080 {
		t.Fatal("Secure Port does not match")
	} else if d.FallbackAddress != "127.0.0.1" {
		t.Fatal("FallbackAddress does not match")
	} else if d.NetworkId != 53 {
		t.Fatal("NetworkId does not match")
	} else if d.Download.MaxSpeed != 56 {
		t.Fatal("MaxSpeed does not match")
	} else if d.Download.DownloadStart != "22:00:00" {
		t.Fatal("DownloadStart does not match")
	} else if d.Download.DownloadEnd != "23:59:00" {
		t.Fatal("DownloadEnd does not match")
	}
}

func TestUnmatched(t *testing.T) {
	var d struct {
		Start struct {
			Foo  string `ini:"FOO"`
			Host string
		} `ini:"[START]"`
	}

	b := []byte(`
; ignore me
[START]
FOO=BAR
SOMETHING=OTHER
OTHER=NOT
HOST=192.168.1.10
`)

	err := Unmarshal(b, &d)

	if err != nil {
		t.Fatal(err)
	}

	if d.Start.Foo != "BAR" {
		t.Fatal("Field FOO not set")
	}

	// need to add NewDecoder to have indirection access
	// to decode state to query unmatched lines ... TBD
}

func TestArrayInt(t *testing.T) {
	var d struct {
		Playlist struct {
			Id    int
			Title string
			Songs []int `ini:"Add Song"`
		} `ini:"[CREATE PLAYLIST]"`
	}

	b := []byte(`
[CREATE PLAYLIST]
ID=349
Title=Rock & Roll, D00d
Add Song=71
Add Song=136
Add Song=252`)

	err := Unmarshal(b, &d)

	if err != nil {
		t.Fatal(err)
	}

	if d.Playlist.Id != 349 {
		t.Fatal("Playlist Id not set")
	} else if d.Playlist.Title != "Rock & Roll, D00d" {
		t.Fatal("Playlist Title not set")
	} else if len(d.Playlist.Songs) != 3 {
		t.Fatal("Playlist Songs length is incorrect")
	} else if d.Playlist.Songs[0] != 71 {
		t.Fatal("Playlist Songs[0] is incorrect")
	} else if d.Playlist.Songs[1] != 136 {
		t.Fatal("Playlist Songs[1] is incorrect")
	} else if d.Playlist.Songs[2] != 252 {
		t.Fatal("Playlist Songs[2] is incorrect")
	}
}

func TestArrayString(t *testing.T) {
	var d struct {
		Playlist struct {
			Id    int
			Title string
			Songs []string `ini:"Add Song"`
		} `ini:"[CREATE PLAYLIST]"`
	}

	b := []byte(`
[CREATE PLAYLIST]
ID=349
Title=Rock & Roll, D00d
Add Song=Time to Run
Add Song=W H O K I L L
Add Song=Mandinka`)

	err := Unmarshal(b, &d)

	if err != nil {
		t.Fatal(err)
	}

	if d.Playlist.Id != 349 {
		t.Fatal("Playlist Id not set")
	} else if d.Playlist.Title != "Rock & Roll, D00d" {
		t.Fatal("Playlist Title not set")
	} else if len(d.Playlist.Songs) != 3 {
		t.Fatal("Playlist Songs length is incorrect")
	} else if d.Playlist.Songs[0] != "Tim to Run" {
		t.Fatal("Playlist Songs[0] is incorrect")
	} else if d.Playlist.Songs[1] != "W H O K I L L" {
		t.Fatal("Playlist Songs[1] is incorrect")
	} else if d.Playlist.Songs[2] != "Mandinka" {
		t.Fatal("Playlist Songs[2] is incorrect")
	}
}
