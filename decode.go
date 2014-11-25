// Decode INI files with a syntax similar to JSON decoding
package ini

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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

type IniError struct {
	lineNum  int
	line     string
	iniError string
}

// conform to Error Interfacer
func (e *IniError) Error() string {
	return fmt.Sprintf("%s on line %d: \"%s\"", e.iniError, e.lineNum, e.line)
}

// decodeState represents the state while decoding a INI value.
type decodeState struct {
	lineNum    int
	line       string
	scanner    *bufio.Scanner
	savedError error
	unmatched  []Unmatched
}

type sectionTag struct {
	tag      string
	value    reflect.Value
	children map[string]sectionTag
	isArray  bool
	array    []interface{}
}

func (t sectionTag) String() string {
	return fmt.Sprintf("<section %s, isArray:%t>", t.tag, t.isArray)
}

func (d *decodeState) init(data []byte) *decodeState {

	d.lineNum = 0
	d.line = ""
	d.scanner = bufio.NewScanner(bytes.NewReader(data))
	d.savedError = nil

	return d
}

// saveError saves the first err it is called with,
// for reporting at the end of the unmarshal.
func (d *decodeState) saveError(err error) {
	if d.savedError == nil {
		d.savedError = err
	}
}

func (d *decodeState) generateMap(m map[string]sectionTag, v reflect.Value) {

	if v.Type().Kind() == reflect.Ptr {
		d.generateMap(m, v.Elem())
	} else if v.Kind() == reflect.Struct {
		typ := v.Type()
		for i := 0; i < typ.NumField(); i++ {

			sf := typ.Field(i)
			f := v.Field(i)
			kind := f.Type().Kind()

			tag := sf.Tag.Get("ini")
			if len(tag) == 0 {
				tag = sf.Name
			}
			tag = strings.TrimSpace(strings.ToLower(tag))

			st := sectionTag{tag, f, make(map[string]sectionTag), kind == reflect.Slice, nil}

			// some structures are just for organizing data
			if tag != "-" {
				m[tag] = st
			}

			if kind == reflect.Struct {
				if tag == "-" {
					d.generateMap(m, f)
				} else {
					// little namespacing here so property names can
					// be the same under different sections
					//fmt.Printf("Struct tag: %s, type: %s\n", tag, f.Type())
					d.generateMap(st.children, f)
				}
			} /*else if kind == reflect.Slice {
				fmt.Printf("Slice tag: %s, type: %s\n", tag, f.Type().Elem())
				d.generateMap(st.children, reflect.New(f.Type().Elem()))
			}*/
		}
	} else {
		d.saveError(&IniError{d.lineNum, d.line, fmt.Sprintf("Can't map into type %s", v.Kind())})
	}

}

func (d *decodeState) unmarshal(x interface{}) error {

	var sectionMap map[string]sectionTag = make(map[string]sectionTag)
	var section, nextSection sectionTag
	var inSection, nextHasSection bool = false, false

	d.generateMap(sectionMap, reflect.ValueOf(x))
	//fmt.Println(sectionMap)

	for d.scanner.Scan() {
		if d.savedError != nil {
			break
		}

		d.line = d.scanner.Text()
		d.lineNum++

		//fmt.Printf("Scanned (%d): %s\n", d.lineNum, d.line)

		line := strings.ToLower(strings.TrimSpace(d.line))

		if len(line) < 1 || line[0] == ';' || line[0] == '#' {
			continue // skip comments
		}

		// [Sections] could appear at any time (square brackets not required)
		// When in a section, also look in children map
		nextSection, nextHasSection = sectionMap[line]
		if nextHasSection {
			section = nextSection
			inSection = true
			continue
		}

		/*
			if hasParent {
				fmt.Printf("PARENT: %s\n", parentSection.tag)
				//fmt.Printf("  CHIL: %s\n", parentSection.children)
				fmt.Printf("  VALU: %s\n", parentSection.value)
				if parentSection.isArray {
					tv := reflect.New(parentSection.value.Type().Elem())
					d.unmarshal(tv)
					fmt.Printf("  SET IT: %s\n", tv)
				}
			}
		*/

		matches := strings.SplitN(d.line, "=", 2)
		matched := false

		// potential property=value
		if len(matches) == 2 {
			n := strings.ToLower(strings.TrimSpace(matches[0]))
			s := strings.TrimSpace(matches[1])

			if inSection {
				// child property, within a section
				childProperty, hasProp := section.children[n]

				if hasProp {
					//fmt.Println("CHILD:", childProperty)
					d.setValue(childProperty.value, s)
					//fmt.Printf("  Partial? %v\n", childProperty.value)
					matched = true
				}
			}

			if !matched {
				// top level property
				topLevelProperty, hasProp := sectionMap[n]
				if hasProp {
					// just encountered a top level property - switch out of section mode
					inSection = false
					matched = true
					d.setValue(topLevelProperty.value, s)
				}
			}
		}

		if !matched {
			d.unmatched = append(d.unmatched, Unmatched{d.lineNum, d.line})
		}
	}

	return d.savedError
}

// Set Value with given string
func (d *decodeState) setValue(v reflect.Value, s string) {
	//fmt.Printf("SET(%s, %s)\n", v.Kind(), s)

	switch v.Kind() {

	case reflect.String:
		v.SetString(s)

	case reflect.Bool:
		v.SetBool(boolValue(s))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || v.OverflowInt(n) {
			d.saveError(&IniError{d.lineNum, d.line, "Invalid int"})
			return
		}
		v.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil || v.OverflowUint(n) {
			d.saveError(&IniError{d.lineNum, d.line, "Invalid uint"})
			return
		}
		v.SetUint(n)

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil || v.OverflowFloat(n) {
			d.saveError(&IniError{d.lineNum, d.line, "Invalid float"})
			return
		}
		v.SetFloat(n)

	case reflect.Slice:
		d.sliceValue(v, s)

	default:
		d.saveError(&IniError{d.lineNum, d.line, fmt.Sprintf("Can't set value of type %s", v.Kind())})
	}

}

func (d *decodeState) sliceValue(v reflect.Value, s string) {

	switch v.Type().Elem().Kind() {

	case reflect.String:
		v.Set(reflect.Append(v, reflect.ValueOf(s)))

	case reflect.Bool:
		v.Set(reflect.Append(v, reflect.ValueOf(boolValue(s))))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Hardcoding of []int temporarily
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			d.saveError(&IniError{d.lineNum, d.line, "Invalid int"})
			return
		}

		n1 := reflect.ValueOf(n)
		n2 := n1.Convert(v.Type().Elem())

		v.Set(reflect.Append(v, n2))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			d.saveError(&IniError{d.lineNum, d.line, "Invalid uint"})
			return
		}

		n1 := reflect.ValueOf(n)
		n2 := n1.Convert(v.Type().Elem())

		v.Set(reflect.Append(v, n2))

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			d.saveError(&IniError{d.lineNum, d.line, "Invalid float"})
			return
		}

		n1 := reflect.ValueOf(n)
		n2 := n1.Convert(v.Type().Elem())

		v.Set(reflect.Append(v, n2))

	default:
		d.saveError(&IniError{d.lineNum, d.line, fmt.Sprintf("Can't set value in array of type %s",
			v.Type().Elem().Kind())})
	}

}

// Returns true for truthy values like t/true/y/yes/1, false otherwise
func boolValue(s string) bool {
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
	return dec.d.unmatched
}
