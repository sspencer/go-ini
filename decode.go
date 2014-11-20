// Decode INI files with a syntax similar to JSON decoding
package ini

import (
	"bufio"
	"bytes"
	"log"
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
	lineNum    int
	scanner    *bufio.Scanner
	savedError error
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

func (d *decodeState) unmarshal(interface{}) error {
	for d.scanner.Scan() {
		line := strings.TrimSpace(d.scanner.Text())
		log.Printf("Scanned (%d): %s\n", d.lineNum, line)
		d.lineNum = d.lineNum + 1

		if len(line) < 1 || line[0] == ';' || line[0] == '#' {
			continue // skip comments
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
