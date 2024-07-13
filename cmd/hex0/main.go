/*
 * hex0 (boot0) compiling tool written by Carl Kittelberger <icedream@icedream.pw>
 */
package main

import (
	"bufio"
	"errors"
	"io"
	"os"
	"slices"
)

/*
minimal set of hex support
*/

var (
	hexUpper = []rune("0123456789abcdef")
	hexLower = []rune("0123456789ABCDEF")
)

func decodeNibble(c rune) (byte, bool) {
	v := slices.Index(hexUpper, c)
	if v < 0 {
		v = slices.Index(hexLower, c)
	}
	if v < 0 {
		return 0, false
	}
	return byte(v), true
}

type textReader interface {
	io.RuneReader
	ReadString(delim byte) (s string, err error)
}

/*
hex0 tokens
*/

const (
	tokenCommentSemicolon = ';'
	tokenCommentHash      = '#'
	tokenLineBreak        = '\n'
)

func decodeByte(r textReader) (byte, error) {
	var digitIndex byte
	var b byte
	for {
		// TODO - handle encoding error (sz=1 with 0xFFFD)?
		c, sz, err := r.ReadRune()
		if err != nil {
			// Is this actually the end of the file?
			if errors.Is(err, io.EOF) {
				// Did we expect EOF?
				if digitIndex != 0 {
					// No, we were in the middle of a byte already!
					err = io.ErrUnexpectedEOF
				}
			}
			return 0, err
		}
		if sz != 1 {
			// We only work with ASCII here really, everything else is safe to
			// assume not to be part of hex0/boot0 syntax.
			continue
		}
		if c == tokenCommentSemicolon || c == tokenCommentHash {
			// Read comment line until line end
			_, err := r.ReadString(tokenLineBreak)
			if err != nil {
				return 0, err
			}
			continue
		}
		nbl, ok := decodeNibble(c)
		if !ok {
			// We got some non-nibble character, ignore
			continue
		}
		switch digitIndex {
		case 0:
			b |= nbl << 4
		case 1:
			b |= nbl
		}
		digitIndex++
		if digitIndex == 2 {
			// We got two nibbles, time to parse and write out
			return b, nil
		}
	}
}

func compile(r io.Reader, w io.ByteWriter) error {
	br := bufio.NewReader(r)
	for {
		// TODO - handle encoding error (sz=1 with 0xFFFD)?
		b, err := decodeByte(br)
		if err != nil {
			// Is this actually the end of the file?
			if errors.Is(err, io.EOF) {
				return nil
			}
			// Some other error occurred
			return err
		}
		if err := w.WriteByte(b); err != nil {
			return err
		}
	}
}

func run() error {
	var in io.Reader = os.Stdin
	var out io.Writer = os.Stdout

	if len(os.Args) > 1 {
		// read argument 1 as stdin
		f, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0o400)
		if err != nil {
			return errors.Join(
				errors.New("failed to open input file "+os.Args[1]),
				err)
		}
		defer func() {
			_ = f.Close()
		}()
		in = f
	}

	if len(os.Args) > 2 {
		// read argument 2 as stdin
		f, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_WRONLY, 0o755)
		if err != nil {
			return errors.Join(
				errors.New("failed to create output file "+os.Args[2]),
				err)
		}
		defer func() {
			_ = f.Close()
		}()
		out = f
	}

	// allow efficiently writing individual bytes
	outByteWriter := bufio.NewWriter(out)
	defer outByteWriter.Flush()

	return compile(in, outByteWriter)
}

func main() {
	if err := run(); err != nil {
		os.Stderr.WriteString("ERROR: " + err.Error() + "\n")
	}
}
