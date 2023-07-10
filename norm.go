package norm

import (
	"bytes"
	"errors"
	"unicode"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)


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
func Space(input []byte) []byte {
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

// Curly UTF-8 apostrophes and quotes are converted into ASCII.
func Quotemark(input []byte) []byte {
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
func SpaceAndQuotemark(input []byte) []byte {
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
