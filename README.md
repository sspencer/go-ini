go-ini
======

INI file decoder for Go lang.  Idea is to have an unmarshaller similar to JSON - specify parts of the file you want coded with structs and tags.

For example, for an INI file like this:

    [Pod_mysql]
    cache_size = 2000
    default_socket = /tmp/mysql.sock

    [Mysql]
    default_socket = /tmp/mysql.sock

Decode into a structure like this:

    type MyIni struct {

        PdoMysql struct {
            CacheSize     int `ini:"cache_size"`
            DefaultSocket string `ini:"default_socket"`
        } `ini:"[Pdo_myqsl]"`

        Mysql struct {
            DefaultSocket string `ini:"default_socket"`
        } `ini:"[Myqsl]"`
    }

With code like this:

    var config MyIni
    var b []byte      // config file stored here
    err := ini.Unmarshal(b, &config)


Current Status
==============

Structs with scalar values in [SECTIONS] now parsed.



Todo
=====

Need to parse inner array of structs

    struct {
        Playlists []struct {
            Id int
            Title string
            Programs []struct {
                Id int
                Mix string
                Separation int
            } `ini:"Play Program"`
        } `ini:"[CREATE PLAYLIST]"`
    }

    [CREATE PLAYLIST]
    ID=6524
    Title=Pop
    Start Schedule

    Play Program
    ID=391
    Mix=RAND

    Play Program
    ID=3912
    Separation=10
    End Schedule
