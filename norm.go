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

func AddLeadingSpace(b []byte) []byte {
	var new []byte
	if cap(b) > len(b) {
		new = b[0 : len(b)+1]
	} else {
		new = make([]byte, len(b)+1)
	}
	new[0] = ' '
	copy(new[1:], b)
	return new
}


func Trim(b []byte) []byte {
	var i int
	var c byte
	for i, c = range b {
		if c > 32 {
			break
		}
	}
	for ii := len(b) - 1; ii >= 0; ii-- {
		if b[ii] > 32 {
			return b[i : ii+1]
		}
	}
	return []byte{}
}

// All sequences of 2 or more spaces are converted to single spaces
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

func Apos(input []byte) []byte {
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

func SpaceAndApos(input []byte) []byte {
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
