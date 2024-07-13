/*
 * hex0 (boot0) compiling tool written by Carl Kittelberger <icedream@icedream.pw>
 */
package main

import (
	"bytes"
	"errors"
	"io"
	"os"
)

/*
minimal set of hex support
*/

var (
	hexLower = []byte("0123456789abcdef")
	hexUpper = []byte("ABCDEF")
)

func decodeNibble(c byte) (byte, bool) {
	v := bytes.IndexByte(hexLower, c)
	var vo byte
	if v < 0 {
		v = bytes.IndexByte(hexUpper, c)
		vo = 0xA
	}
	if v < 0 {
		return 0, false
	}
	return byte(v) + vo, true
}

/*
hex0 tokens
*/

const (
	tokenCommentSemicolon = ';'
	tokenCommentHash      = '#'
	tokenLineBreak        = '\n'
)

func decodeByte(r io.ByteReader) (byte, error) {
	var digitIndex byte
	var b byte
	for {
		// TODO - handle encoding error (sz=1 with 0xFFFD)?
		rb, err := r.ReadByte()
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
		if rb&0b10000000 != 0 {
			// UTF-8 multibyte.
			//
			// This ain't going to be hex, so skip all of it.
			continue
		}
		if rb == tokenCommentSemicolon || rb == tokenCommentHash {
			// Read comment line until line end
			for {
				rb, err = r.ReadByte()
				if err != nil {
					return 0, err
				}
				if rb == tokenLineBreak {
					// End of line = end of comment
					break
				}
			}
			continue
		}
		nbl, ok := decodeNibble(rb)
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

func compile(r io.ByteReader, w io.Writer) error {
	for {
		// TODO - handle encoding error (sz=1 with 0xFFFD)?
		b, err := decodeByte(r)
		if err != nil {
			// Is this actually the end of the file?
			if errors.Is(err, io.EOF) {
				return nil
			}
			// Some other error occurred
			return err
		}
		if _, err := w.Write([]byte{b}); err != nil {
			return err
		}
	}
}

type byteReaderWrapper struct {
	r io.Reader
	b [1]byte
}

// ReadByte implements io.ByteReader.
func (b *byteReaderWrapper) ReadByte() (byte, error) {
	_, err := b.r.Read(b.b[:])
	return b.b[0], err
}

var _ io.ByteReader = (*byteReaderWrapper)(nil)

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

	inbr := &byteReaderWrapper{r: in}

	return compile(inbr, out)
}

func main() {
	if err := run(); err != nil {
		os.Stderr.WriteString("ERROR: " + err.Error() + "\n")
	}
}
