// Decode INI files with a syntax similar to JSON decoding
package ini

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"strings"
)

// Unmarshal parses the INI-encoded data and stores the result
// in the value pointed to by v.
func Unmarshal(data []byte, v interface{}) error {
	var d decodeState
	d.init(data)
	return d.unmarshal(v)
}

type Unmatched struct {
	lineNum int
	line    string
}

func (u Unmatched) String() string {
	return fmt.Sprintf("%d %s", u.lineNum, u.line)
}

// decodeState represents the state while decoding a INI value.
type decodeState struct {
	currentPath    string
	lineNum        int
	scanner        *bufio.Scanner
	savedError     error
	unmatchedLines []Unmatched
}

type sectionTag struct {
	tag      string
	value    reflect.Value
	children map[string]sectionTag
}

func (t sectionTag) String() string {
	return fmt.Sprintf("<section %s>", t.tag)
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

			tag := sf.Tag.Get("ini")
			if len(tag) == 0 {
				tag = sf.Name
			}
			tag = strings.TrimSpace(strings.ToLower(tag))

			st := sectionTag{tag, f, make(map[string]sectionTag)}

			// some structures are just for organizing data
			if tag != "-" {
				m[tag] = st
			}

			if f.Type().Kind() == reflect.Struct {
				if tag == "-" {
					generateMap(m, f)
				} else {
					// little namespacing here so property names can
					// be the same under different sections
					generateMap(st.children, f)
				}
			}
		}
	} else {
		panic(fmt.Sprintf("Don't handle this type yet: %s", v.Kind()))
	}

}

func (d *decodeState) unmarshal(x interface{}) error {

	var parentMap map[string]sectionTag = make(map[string]sectionTag)

	generateMap(parentMap, reflect.ValueOf(x))

	//log.Printf("%#v\n", parentMap)

	var parentSection sectionTag
	var hasParent bool = false

	for d.scanner.Scan() {
		line := strings.TrimSpace(d.scanner.Text())
		log.Printf("Scanned (%d): %s\n", d.lineNum, line)
		d.lineNum = d.lineNum + 1

		if len(line) < 1 || line[0] == ';' || line[0] == '#' {
			continue // skip comments
		}

		if line[0] == '[' && line[len(line)-1] == ']' {
			parentSection, hasParent = parentMap[strings.ToLower(line)]

			continue
		}

		matches := strings.SplitN(line, "=", 2)
		matched := false
		if len(matches) == 2 {
			n := strings.ToLower(strings.TrimSpace(matches[0]))
			s := strings.TrimSpace(matches[1])

			if hasParent {

				childSection, hasChild := parentSection.children[n]
				if hasChild {
					setValue(childSection.value, s, d.lineNum)
					matched = true
				} // else look for wildcard??
			} else {
				propSection, hasProp := parentMap[n]
				if hasProp {
					setValue(propSection.value, s, d.lineNum)
					matched = true
				}
			}
		}

		if !matched {
			d.unmatchedLines = append(d.unmatchedLines, Unmatched{d.lineNum, line})
		}
	}

	// temp - print out unmatch lines to verify they're being kept
	if len(d.unmatchedLines) > 0 {
		log.Println("==== Unmatched Lines ====")
		for _, line := range d.unmatchedLines {
			log.Println("    ", line)
		}
	}

	return nil
}

// Set Value with given string
func setValue(v reflect.Value, s string, lineNum int) {
	log.Printf("SET(%s, %s)", v.Kind(), s)

	switch v.Kind() {

	case reflect.String:
		v.SetString(s)

	case reflect.Bool:
		v.SetBool(getBoolValue(s))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || v.OverflowInt(n) {
			panic(fmt.Sprintf("Invalid int '%s' specified on line %d", s, lineNum))
		}
		v.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil || v.OverflowUint(n) {
			panic(fmt.Sprintf("Invalid uint '%s' specified on line %d", s, lineNum))
		}
		v.SetUint(n)

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil || v.OverflowFloat(n) {
			panic(fmt.Sprintf("Invalid float '%s' specified on line %d", s, lineNum))
		}
		v.SetFloat(n)

	case reflect.Slice:

		// Hardcoding of []int temporarily
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Invalid int '%s' specified on line %d", s, lineNum))
		}

		n1 := reflect.ValueOf(n)
		n2 := n1.Convert(v.Type().Elem())

		v.Set(reflect.Append(v, n2))

	default:
		log.Println("Can't set that kind yet!")
	}

}

// Returns true for truthy values like t/true/y/yes/1, false otherwise
func getBoolValue(s string) bool {
	v := false
	switch strings.ToLower(s) {
	case "t", "true", "y", "yes", "1":
		v = true
	}

	return v
}

// A Decoder reads and decodes INI object from an input stream.
type Decoder struct {
	r io.Reader
	d decodeState
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may
// read data from r beyond the JSON values requested.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the INI file and stores it in the value pointed to by v.
//
// See the documentation for Unmarshal for details about
// the conversion of an INI into a Go value.
func (dec *Decoder) Decode(v interface{}) error {

	buf, readErr := ioutil.ReadAll(dec.r)
	if readErr != nil {
		return readErr
	}
	// Don't save err from unmarshal into dec.err:
	// the connection is still usable since we read a complete JSON
	// object from it before the error happened.
	dec.d.init(buf)
	err := dec.d.unmarshal(v)

	return err
}

// UnparsedLines returns an array of strings where each string is an
// unparsed line from the file.
func (dec *Decoder) Unmatched() []Unmatched {
	return dec.d.unmatchedLines
}
