package norm

import (
	"bytes"
	"errors"
	"strings"
	"unicode"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Normalizer struct {
	Flag uint8
}

func (n Normalizer) String() string {
	var s string
	if n.Flag & 1 != 0 {
		s = "nfd "
	}
	if n.Flag & 2 != 0 {
		s += "lowercase "
	}
	if n.Flag & 4 != 0 {
		s += "accents "
	}
	if n.Flag & 8 != 0 {
		s += "quotemarks "
	}
	if n.Flag & 16 != 0 {
		s += "collapse "
	}
	if n.Flag & 32 != 0 {
		s += "trim "
	}
	if n.Flag & 64 != 0 {
		s += "leading-space "
	}
	if n.Flag & 128 != 0 {
		s += "lines "
	}
	return strings.TrimSpace(s)
}

func (n Normalizer) SpecifiedNFD() bool {
	return n.Flag & 1 != 0
}

func (n Normalizer) SpecifiedLowercase() bool {
	return n.Flag & 2 != 0
}

func (n Normalizer) SpecifiedAccents() bool {
	return n.Flag & 4 != 0
}

func (n Normalizer) SpecifiedQuotemarks() bool {
	return n.Flag & 8 != 0
}

func (n Normalizer) SpecifiedCollapse() bool {
	return n.Flag & 16 != 0
}

func (n Normalizer) SpecifiedTrim() bool {
	return n.Flag & 32 != 0
}

func (n Normalizer) SpecifiedLeadingSpace() bool {
	return n.Flag & 64 != 0
}

func (n Normalizer) SpecifiedLines() bool {
	return n.Flag & 128 != 0
}

func NewNormalizer(s string) (Normalizer, error) {
	var n Normalizer
	var err error
	for _, each := range strings.Split(strings.ToLower(s), " ") {
		switch each {
			case "nfd":
				n.Flag |= 1
			case "lowercase", "case":
				n.Flag |= 2
			case "accents", "accent":
				n.Flag |= 4
			case "quotemarks", "quotemark", "apostrophes":
				n.Flag |= 8
			case "collapse", "spaces", "space", "doublespace", "doublespaces":
				n.Flag |= 16
			case "trim", "trimspace", "trim-space":
				n.Flag |= 32
			case "leadingspace", "leading-space", "addleadingspace":
				n.Flag |= 64
			case "unixlines", "unix-lines", "newlines", "lines":
				n.Flag |= 128
			case "":
				// do nothing
			default:
				err = errors.New(`Unrecognized normalization parameter: ` + each)
		}
	}
	return n, err
}

func (n Normalizer) Normalize(data []byte) ([]byte, error) {
	if n.Flag == 1 { // likely
		return NFD(data)
	}
	// unixlines
	if n.Flag & 128 != 0 {
		if n.Flag & 16 != 0 && n.Flag & 8 != 0 {
			data = CollapseAndQuotemarksAndUnixLines(data)
			goto skipahead
		} else {
			data = UnixLines(data)
		}
	}
	// collapse & quotemark
	if n.Flag & 8 != 0 {
		if n.Flag & 16 != 0 { // both
			data = CollapseAndQuotemarks(data)
		} else { // quotemarks
			data = Quotemarks(data)
		}
	} else if n.Flag & 16 != 0 { // collapse
		data = Collapse(data)
	}
skipahead:
	// trim, leading-space
	if n.Flag & 32 != 0 {
		if n.Flag & 64 != 0 { // both
			data = TrimAndAddLeadingSpace(data)
		} else { // trim
			data = Trim(data)
		}
	} else if n.Flag & 64 != 0 { // leadingspace
		data = AddLeadingSpace(data)
	}
	// NFD, lowercase & accents
	// Accents includes NFD so ignore NFD if removing accents
	if n.Flag & 4 != 0 { // accents
		if n.Flag & 2 != 0 { // accents + lowercase
			return CaseAndAccents(data)
		}
		return Accents(data)
	}
	if n.Flag & 2 != 0 { // lowercase
		if n.Flag & 1 != 0 { // lowercase + NFD
			return NFDAndCase(data)
		}
		return Case(data)
	}
	if n.Flag & 1 != 0 { // NFD
		return NFD(data)
	}
	return data, nil
}

// Adds a leading space if there isn't one already.
func AddLeadingSpace(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	if b[0] == ' ' {
		return b
	}
	var new []byte
	if cap(b) > len(b) {
		new = b[0 : len(b)+1]
	} else {
		new = make([]byte, len(b)+1)
	}
	copy(new[1:], b)
	new[0] = ' '
	return new
}

// Removes preceding and trailing whitespace (& some non-printable characters).
func Trim(b []byte) []byte {
	var i int
	var c byte
	for i, c = range b {
		if c > 32 {
			break
		}
	}
	for i2 := len(b) - 1; i2 >= 0; i2-- {
		if b[i2] > 32 {
			return b[i : i2+1]
		}
	}
	return []byte{}
}

// Removes preceding and trailing whitespace (& some non-printable characters), then adds space to the front.
func TrimAndAddLeadingSpace(b []byte) []byte {
	var i int
	var c byte
	for i, c = range b {
		if c > 32 {
			break
		}
	}
	var i2 int
	for i2 = len(b) - 1; i2 >= 0; i2-- {
		if b[i2] > 32 {
			break
		}
	}
	if i == 0 {
		return AddLeadingSpace(b[0:i2])
	} else if i2 < 0 {
		return []byte{}
	}
	b[i-1] = ' '
	return b[i-1 : i2+1]
}

// All sequences of 2 or more spaces are converted into single spaces.
func Collapse(input []byte) []byte {
	var on uintptr
	var last byte
	for _, b := range input {
		if b != 32 {
			input[on] = b
			on++
		} else {
			if last != 32 {
				input[on] = 32
				on++
			}
		}
		last = b
	}
	return input[0:on]
}

// Newlines converts /r/n to /n
func UnixLines(input []byte) []byte {
	if len(input) < 2 {
		return input
	}
	var on uintptr
	for i, b := range input[0 : len(input)-1] {
		if b == '\r' {
			if input[i+1] == '\n' {
				continue
			}
		}
		input[on] = b
		on++
	}
	input[on] = input[len(input)-1]
	return input[0 : on+1]
}

// Curly UTF-8 apostrophes and quotes are converted into ASCII.
func Quotemarks(input []byte) []byte {
	var on uintptr
	for i, b := range input {
		if b != 152 && b != 153 && b != 156 && b != 157 {
			input[on] = b
			on++
			continue
		}
		if i > 1 {
			if input[i-1] == 128 && input[i-2] == 226 {
				if b < 156 {
					input[on-2] = '\''
				} else {
					input[on-2] = '"'
				}
				on--
				continue
			}
		}
		input[on] = b
		on++
	}
	return input[0:on]
}

// All sequences of 2 or more spaces are converted into single spaces, and curly UTF-8 apostrophes and quotes are converted into ASCII.
func CollapseAndQuotemarks(input []byte) []byte {
	var on uintptr
	var last byte
	for i, b := range input {
		if b == 32 {
			if last != 32 {
				input[on] = 32
				on++
				last = 32
			}
			continue
		}
		last = b
		if b != 152 && b != 153 && b != 156 && b != 157 {
			input[on] = b
			on++
			continue
		}
		if i > 1 {
			if input[i-1] == 128 && input[i-2] == 226 {
				if b < 156 {
					input[on-2] = '\''
				} else {
					input[on-2] = '"'
				}
				on--
				continue
			}
		}
		input[on] = b
		on++
	}
	return input[0:on]
}

// All sequences of 2 or more spaces are converted into single spaces, and curly UTF-8 apostrophes and quotes are converted into ASCII.
func CollapseAndQuotemarksAndUnixLines(input []byte) []byte {
	var on uintptr
	var last byte
	for i, b := range input {
		if b == 32 {
			if last != 32 {
				input[on] = 32
				on++
				last = 32
			}
			continue
		}
		if b == '\n' {
			if last == '\r' {
				input[on - 1] = '\n'
				last = '\n'
				continue
			}
		}
		last = b
		if b != 152 && b != 153 && b != 156 && b != 157 {
			input[on] = b
			on++
			continue
		}
		if i > 1 {
			if input[i-1] == 128 && input[i-2] == 226 {
				if b < 156 {
					input[on-2] = '\''
				} else {
					input[on-2] = '"'
				}
				on--
				continue
			}
		}
		input[on] = b
		on++
	}
	return input[0:on]
}

// Performs UTF-8 NFD normalization.
func NFD(input []byte) (output []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Convert panic into error
			err = errors.New(`Panicked`)
		}
	}()
	normalized := bytes.NewBuffer(make([]byte, 0, len(input) + (len(input) / 3) + 4))
	normalizer := norm.NFD.Writer(normalized)
	_, err = normalizer.Write(input)
	normalizer.Close()
	output = normalized.Bytes()
	return
}

// Removes accents from UTF-8 text.
func Accents(b []byte) (output []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Convert panic into error
			err = errors.New(`Panicked`)
		}
	}()
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn))
	output, _, err = transform.Bytes(t, b)
	return
}

// Lowercases UTF-8 text.
func Case(b []byte) (output []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Convert panic into error
			err = errors.New(`Panicked`)
		}
	}()
	buf := bytes.NewBuffer(make([]byte, 0, len(b) + (len(b) / 3) + 4))
	writer := transform.NewWriter(buf, cases.Lower(language.Und))
	_, err = writer.Write(b)
	writer.Close()
	output = buf.Bytes()
	return
}

// Lowercase and remove accents from UTF-8 text.
func CaseAndAccents(b []byte) (output []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Convert panic into error
			err = errors.New(`Panicked`)
		}
	}()
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), cases.Lower(language.Und))
	buf := bytes.NewBuffer(make([]byte, 0, len(b) + (len(b) / 3) + 4))
	writer := transform.NewWriter(buf, t)
	_, err = writer.Write(b)
	writer.Close()
	output = buf.Bytes()
	return
}

// Lowercases and performs UTF-8 NFD normalization.
func NFDAndCase(b []byte) (output []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Convert panic into error
			err = errors.New(`Panicked`)
		}
	}()
	t := transform.Chain(norm.NFD, cases.Lower(language.Und))
	buf := bytes.NewBuffer(make([]byte, 0, len(b) + (len(b) / 3) + 4))
	writer := transform.NewWriter(buf, t)
	_, err = writer.Write(b)
	writer.Close()
	output = buf.Bytes()
	return
}

func isMn(r rune) bool {
    return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}
