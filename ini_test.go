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

func TestArray(t *testing.T) {
	var d struct {
		Playlist struct {
			Id      int
			Name    string
			Titles  []string  `ini:"Add Title"`
			Songs   []int     `ini:"Add Song"`
			Genres  []uint    `ini:"Add Genre"`
			Volumes []float32 `ini:"Add Volume"`
			Videos  []bool    `ini:"Is Video"`
		} `ini:"[CREATE PLAYLIST]"`
	}

	b := []byte(`
[CREATE PLAYLIST]
ID=349
Name=Rock & Roll, D00d
Add Title=Time to Run
Add Title=W H O K I L L
Add Song=71
Add Song=136
Add Genre=17
Add Genre=31
Add Volume=0.75
Add Volume=0.81
Is Video=false
Is Video=true
`)

	err := Unmarshal(b, &d)

	if err != nil {
		t.Fatal(err)
	}

	if d.Playlist.Id != 349 {
		t.Fatal("Playlist Id not set")
	} else if d.Playlist.Name != "Rock & Roll, D00d" {
		t.Fatal("Playlist Name not set")
	} else if len(d.Playlist.Titles) != 2 {
		t.Fatal("Playlist Title length is incorrect")
	} else if d.Playlist.Titles[0] != "Time to Run" {
		t.Fatal("Playlist Title[0] is incorrect")
	} else if d.Playlist.Titles[1] != "W H O K I L L" {
		t.Fatal("Playlist Title[1] is incorrect")
	} else if len(d.Playlist.Songs) != 2 {
		t.Fatal("Playlist Songs length is incorrect")
	} else if d.Playlist.Songs[0] != 71 {
		t.Fatal("Playlist Songs[0] is incorrect")
	} else if d.Playlist.Songs[1] != 136 {
		t.Fatal("Playlist Songs[1] is incorrect")
	} else if len(d.Playlist.Genres) != 2 {
		t.Fatal("Playlist Genre length is incorrect")
	} else if d.Playlist.Genres[0] != 17 {
		t.Fatal("Playlist Genre[0] is incorrect")
	} else if d.Playlist.Genres[1] != 31 {
		t.Fatal("Playlist Genre[1] is incorrect")
	} else if len(d.Playlist.Volumes) != 2 {
		t.Fatal("Playlist Volume length is incorrect")
	} else if d.Playlist.Volumes[0] != 0.75 {
		t.Fatal("Playlist Volume[0] is incorrect")
	} else if d.Playlist.Volumes[1] != 0.81 {
		t.Fatal("Playlist Volume[1] is incorrect")
	} else if len(d.Playlist.Videos) != 2 {
		t.Fatal("Playlist Video length is incorrect")
	} else if d.Playlist.Videos[0] != false {
		t.Fatal("Playlist Video[0] is incorrect")
	} else if d.Playlist.Videos[1] != true {
		t.Fatal("Playlist Video[1] is incorrect")
	}
}

func TestArrayStruct(t *testing.T) {
	var d struct {
		Device struct {
			NumZones  int `ini:"SET OPTION ACTIVE ZONES"`
			MaxVolume int `ini:"SET OPTION ALLOW MAX VOLUME"`
		} `ini:"[ALTER DEVICE]"`

		Channels []struct {
			Id         int
			Title      string
			PlaylistId int `ini:"SET DEFAULT PLAYLIST"`
		} `ini:"[CREATE CHANNEL]"`
	}

	b := []byte(`
[ALTER DEVICE]
SET OPTION ACTIVE ZONES=3
SET OPTION ALLOW MAX VOLUME=11
[CREATE CHANNEL]
ID=1
Title=Lounge
SET DEFAULT PLAYLIST=6502
[CREATE CHANNEL]
ID=2
Title=Acid House
SET DEFAULT PLAYLIST=4004`)

	err := Unmarshal(b, &d)

	if err != nil {
		t.Fatal(err)
	}

	if d.Device.NumZones != 3 {
		t.Fatal("NumZones is incorrect")
	} else if d.Device.MaxVolume != 11 {
		t.Fatal("MaxVolume is incorrect")
	} else if len(d.Channels) != 2 {
		t.Fatal("Incorrect number of channels")
	}
}
