// Decode INI files with a syntax similar to JSON decoding
package ini

// Unmarshal parses the INI-encoded data and stores the result
// in the value pointed to by v.
func Unmarshal(data []byte, v interface{}) error {
	var d decodeState
	d.init(data)
	return d.unmarshal(v)
}

// decodeState represents the state while decoding a INI value.
type decodeState struct {
	data       []byte
	off        int // read offset in data
	savedError error
	tempstr    string // scratch space to avoid some allocations
}

func (d *decodeState) init(data []byte) *decodeState {
	d.data = data
	d.off = 0
	d.savedError = nil
	return d
}

func (d *decodeState) unmarshal(interface{}) error {
	return nil
}
