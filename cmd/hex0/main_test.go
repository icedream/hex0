package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"io"
	"testing"
	"time"
)

func TestDecodeNibble(t *testing.T) {
	for _, test_ := range []struct {
		Name        string
		Inputs      []byte
		Outputs     []byte
		ExpectNotOK bool
	}{
		{
			Name:    "Numbers",
			Inputs:  []byte("0123456789"),
			Outputs: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			Name:    "Lowercase",
			Inputs:  []byte("abcdef"),
			Outputs: []byte{10, 11, 12, 13, 14, 15},
		},
		{
			Name:    "Uppercase",
			Inputs:  []byte("ABCDEF"),
			Outputs: []byte{10, 11, 12, 13, 14, 15},
		},
		{
			Name:        "Invalid",
			Inputs:      []byte(" _%$ⓢ"),
			ExpectNotOK: true,
		},
	} {
		test := test_
		for i, c := range test.Inputs {
			character := c
			var expectedOutputByte byte
			if !test.ExpectNotOK {
				expectedOutputByte = test.Outputs[i]
			}
			t.Run(test.Name+"/"+string(c), func(t *testing.T) {
				result, ok := decodeNibble(character)
				if test.ExpectNotOK && ok {
					t.Errorf("unexpected success: %X", result)
					return
				}
				if result != expectedOutputByte {
					t.Errorf("unexpected output: %X", result)
				}
			})
		}
	}
}

var expected = []byte{0x7F, 0x45, 0x4C, 0x46}

//go:embed test_onlyhex.hex0
var onlyHex []byte

//go:embed test_normal.hex0
var normalHex []byte

//go:embed test_chaotic.hex0
var chaoticHex []byte

func TestDecodeByte(t *testing.T) {
	for _, test_ := range []struct {
		Name        string
		Input       []byte
		Output      []byte
		ExpectError error
	}{
		{
			Name:   "Single AA",
			Input:  []byte("aa"),
			Output: []byte{0xaa},
		},
		{
			Name:   "Single FE",
			Input:  []byte("fe"),
			Output: []byte{0xfe},
		},
		{
			Name:   "Single EF",
			Input:  []byte("ef"),
			Output: []byte{0xef},
		},
		{
			Name:   "Double EF",
			Input:  []byte("efef"),
			Output: []byte{0xef, 0xef},
		},
		{
			Name:   "EF FE",
			Input:  []byte("effe"),
			Output: []byte{0xef, 0xfe},
		},
		{
			Name:   "Single 0F",
			Input:  []byte("0F"),
			Output: []byte{0x0F},
		},
		{
			Name:   "Single F0",
			Input:  []byte("F0"),
			Output: []byte{0xF0},
		},
		{
			Name:   "Minimal",
			Input:  onlyHex,
			Output: expected,
		},
		{
			Name:   "Normal",
			Input:  normalHex,
			Output: expected,
		},
		{
			Name:   "Chaotic",
			Input:  chaoticHex,
			Output: expected,
		},
		{
			Name:        "Unexpected EOF",
			Input:       []byte("AABBC"),
			Output:      []byte{0xaa, 0xbb, 0},
			ExpectError: io.ErrUnexpectedEOF,
		},
	} {
		test := test_
		t.Run(test.Name, func(t *testing.T) {
			br := bufio.NewReader(bytes.NewReader(test.Input))
			for _, expectedByte := range test.Output {
				peek16, _ := br.Peek(0x20)
				b, err := decodeByte(br)
				if err != nil && err != test.ExpectError {
					t.Errorf("unexpected error value: %q", err)
					return
				}
				if b != expectedByte {
					t.Errorf("unexpected output: %X, input peek:\n%s",
						b, hex.Dump(peek16))
					return
				}
			}
		})
	}
}

func TestCompile(t *testing.T) {
	for _, test_ := range []struct {
		Name        string
		Input       []byte
		Output      []byte
		ExpectError error
	}{
		{
			Name:   "Minimal",
			Input:  onlyHex,
			Output: expected,
		},
		{
			Name:   "Normal",
			Input:  normalHex,
			Output: expected,
		},
		{
			Name:   "Chaotic",
			Input:  chaoticHex,
			Output: expected,
		},
		{
			Name:        "Unexpected EOF",
			Input:       []byte("AABBC"),
			Output:      []byte{0xaa, 0xbb},
			ExpectError: io.ErrUnexpectedEOF,
		},
	} {
		test := test_
		t.Run(test.Name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			r := bytes.NewReader(test.Input)
			err := compile(r, buf)
			if err != nil && err != test.ExpectError {
				t.Errorf("unexpected error value: %q", err)
				return
			}
			if !bytes.Equal(buf.Bytes(), test.Output) {
				t.Errorf("unexpected output:\n%s", hex.Dump(buf.Bytes()))
				return
			}
		})
	}
}

type fakeReader struct {
	bytes     int64
	b         *testing.B
	startTime time.Time
}

// Read implements io.Reader.
func (f *fakeReader) Read(p []byte) (n int, err error) {
	if time.Now().Sub(f.startTime) > time.Second {
		return 0, io.EOF
	}
	n, err = rand.Read(p)
	f.bytes += int64(n)
	f.b.SetBytes(f.bytes)
	return n, err
}

var _ io.Reader = (*fakeReader)(nil)

func BenchmarkDecodeNibble(b *testing.B) {
	for range b.N {
		decodeNibble('a')
	}
}

func BenchmarkDecodeByte(b *testing.B) {
	fr := &fakeReader{
		b: b,
	}
	r := bufio.NewReader(fr)
	for range b.N {
		decodeByte(r)
	}
}

func BenchmarkCompile(b *testing.B) {
	// not really meaningful, more of a rough test
	var previousBytes int64
	for range b.N {
		fr := &fakeReader{
			b:     b,
			bytes: previousBytes,
		}
		r := bufio.NewReader(fr)
		w := bufio.NewWriter(io.Discard)
		compile(r, w)
		previousBytes = fr.bytes
	}
}
