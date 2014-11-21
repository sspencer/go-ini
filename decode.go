// Decode INI files with a syntax similar to JSON decoding
package ini

import (
	"bufio"
	"bytes"

	"log"
	"reflect"
	"strings"
)

// Unmarshal parses the INI-encoded data and stores the result
// in the value pointed to by v.
func Unmarshal(data []byte, v interface{}) error {
	var d decodeState
	d.init(data)
	return d.unmarshal(v)
}

// decodeState represents the state while decoding a INI value.
type decodeState struct {
	currentPath string
	lineNum     int
	scanner     *bufio.Scanner
	savedError  error
}

type sectionTag struct {
	wildcard bool
	value    reflect.Value
	children map[string]sectionTag
}

func (d *decodeState) init(data []byte) *decodeState {

	d.lineNum = 1
	d.scanner = bufio.NewScanner(bytes.NewReader(data))
	d.savedError = nil
	return d
}

// error aborts the decoding by panicking with err.
func (d *decodeState) error(err error) {
	panic(err)
}

// saveError saves the first err it is called with,
// for reporting at the end of the unmarshal.
func (d *decodeState) saveError(err error) {
	if d.savedError == nil {
		d.savedError = err
	}
}

func generateMap(m map[string]sectionTag, v reflect.Value) {

	if v.Type().Kind() == reflect.Ptr {
		generateMap(m, v.Elem())
	} else if v.Kind() == reflect.Struct {
		typ := v.Type()
		for i := 0; i < typ.NumField(); i++ {

			sf := typ.Field(i)
			f := v.Field(i)

			var st sectionTag = sectionTag{false, f, nil}

			log.Println("==== IN ", sf.Tag.Get("ini"))
			m[sf.Tag.Get("ini")] = st

			if f.Type().Kind() == reflect.Struct {
				log.Println("    STRUCT")
				st.children = make(map[string]sectionTag)
				generateMap(st.children, f)
			}

			log.Println(m)
			log.Println("==== OUT ", sf.Tag.Get("ini"))
		}
	} else {
		log.Println("Unhandled type:", v.Kind())
		panic("Don't handle this type yet!")
	}

}

func (d *decodeState) unmarshal(x interface{}) error {

	var parentMap map[string]sectionTag = make(map[string]sectionTag)

	generateMap(parentMap, reflect.ValueOf(x))

	var parentSection sectionTag
	var hasParent bool = false

	log.Println("-----")
	log.Println(parentMap)
	log.Println("-----")
	for d.scanner.Scan() {
		line := strings.TrimSpace(d.scanner.Text())
		log.Printf("Scanned (%d): %s\n", d.lineNum, line)
		d.lineNum = d.lineNum + 1

		if len(line) < 1 || line[0] == ';' || line[0] == '#' {
			continue // skip comments
		}

		if line[0] == '[' && line[len(line)-1] == ']' {
			log.Println("IN PARENT")
			parentSection, hasParent = parentMap[line]
			log.Println(parentSection)
			continue
		}

		if hasParent {
			matches := strings.SplitN(line, "=", 2)
			log.Printf("MATCHES %q\n", matches)
			if len(matches) == 2 {
				childSection, hasChild := parentSection.children[matches[0]]
				if hasChild {
					log.Println("HAS Child", childSection)
				} else {
					log.Println("NO NO NO")
					log.Println(parentSection.children)
				}
			}
		}
	}

	return nil
}

/*
// A Decoder reads and decodes JSON objects from an input stream.
type Decoder struct {
	d    decodeState
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may
// read data from r beyond the JSON values requested.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the next JSON-encoded value from its
// input and stores it in the value pointed to by v.
//
// See the documentation for Unmarshal for details about
// the conversion of JSON into a Go value.
func (dec *Decoder) Decode(v interface{}) error {
	if dec.err != nil {
		return dec.err
	}

	n, err := dec.readValue()
	if err != nil {
		return err
	}

	// Don't save err from unmarshal into dec.err:
	// the connection is still usable since we read a complete JSON
	// object from it before the error happened.
	dec.d.init(dec.buf[0:n])
	err = dec.d.unmarshal(v)

	// Slide rest of data down.
	rest := copy(dec.buf, dec.buf[n:])
	dec.buf = dec.buf[0:rest]

	return err
}
*/
