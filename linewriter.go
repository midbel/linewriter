package linewriter

import (
	"io"
	"strconv"
)

type Flag uint64

const (
	AlignLeft Flag = 1 << iota
	AlignRight
	AlignCenter
	BasePrefix
	Base10
	Base16
	ZeroFill
	NoPadding
	NoSeparator
	YesNo
	OnOff
	TrueFalse
)

type Writer struct {
	buffer []byte
	offset int

	padding   []byte
	separator []byte
}

func NewWriter(size, padsiz int, padchar byte) *Writer {
	w := Writer{
		buffer:    make([]byte, size),
		separator: []byte("|"),
	}
	if padsiz > 0 {
		w.padding = make([]byte, padsiz)
		for i := 0; i < padsiz; i++ {
			w.padding[i] = padchar
		}
	}
	w.Reset()
	return &w
}

func New(size, padsiz int) *Writer {
	return NewWriter(size, padsiz, ' ')
}

func (w *Writer) Reset() {
	for i := 0; i < len(w.buffer); i++ {
		w.buffer[i] = ' '
	}
	w.offset = 0
}

func (w *Writer) Bytes() []byte {
	return w.buffer[:w.offset]
}

func (w *Writer) String() string {
	return string(w.Bytes())
}

func (w *Writer) Read(bs []byte) (int, error) {
	if len(bs) < w.offset {
		return 0, io.ErrShortBuffer
	}
	n := copy(bs, append(w.buffer, '\n'))
	w.Reset()
	return n, nil
}

func (w *Writer) AppendString(str string, width int, flag Flag) {
	w.AppendBytes([]byte(str), width, flag)
}

func (w *Writer) AppendBytes(bs []byte, width int, flag Flag) {
	w.appendLeft(flag)

	var offset int
	if set := flag & AlignRight; set != 0 {
		offset = w.offset + (width - len(bs))
	} else if set := flag & AlignCenter; set != 0 {

	} else {
		offset = w.offset
	}
	copy(w.buffer[offset:], bs)
	w.offset += width

	if set := flag & NoPadding; set == 0 {
		w.offset += copy(w.buffer[w.offset:], w.padding)
	}
}

func (w *Writer) AppendUint(v uint64, width int, flag Flag) {
	w.appendLeft(flag)

	if set := flag & ZeroFill; set != 0 {
		for i := 0; i < width; i++ {
			w.buffer[w.offset+i] = '0'
		}
	}

	base := 10
	if set := flag & Base16; set != 0 {
		base = 16
	}
	tmp := make([]byte, 0, 16)
	tmp = strconv.AppendUint(tmp, v, base)

	var offset int
	if set := flag & AlignRight; set != 0 {
		offset = w.offset + (width - len(tmp))
	} else if set := flag & AlignCenter; set != 0 {

	} else {
		offset = w.offset
	}
	copy(w.buffer[offset:], tmp)
	w.offset += width

	if set := flag & NoPadding; set == 0 {
		w.offset += copy(w.buffer[w.offset:], w.padding)
	}
}

func (w *Writer) appendLeft(flag Flag) {
	if set := flag & NoSeparator; w.offset > 0 && set == 0 {
		n := copy(w.buffer[w.offset:], w.separator)
		w.offset += n
	}
	if set := flag & NoPadding; set == 0 {
		w.offset += copy(w.buffer[w.offset:], w.padding)
	}
}
