package linewriter

import (
	"encoding/hex"
	"io"
	"strconv"
)

type Flag uint64

const (
	AlignLeft Flag = 1 << iota
	AlignRight
	AlignCenter
	BasePrefix
	Base2
	Base8
	Base10
	Base16
	ZeroFill
	WithSign
	WithPrefix
	NoPadding
	NoSeparator
	YesNo
	OnOff
	TrueFalse
	Hex
	Text
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
	n := len(w.buffer)
	if w.offset > 0 {
		n = w.offset
	}
	for i := 0; i < n; i++ {
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
	if w.offset == 0 {
		return w.offset, io.EOF
	}
	if len(bs) < w.offset {
		return 0, io.ErrShortBuffer
	}
	n := copy(bs, append(w.buffer[:w.offset], '\n'))
	w.Reset()
	return n, nil
}

func (w *Writer) AppendString(str string, width int, flag Flag) {
	flag = flag &^ Hex
	w.AppendBytes([]byte(str), width, flag|Text)
}

func (w *Writer) AppendBytes(bs []byte, width int, flag Flag) {
	w.appendLeft(flag)

	var data []byte
	if set := flag & Hex; set != 0 {
		data = make([]byte, hex.EncodedLen(len(bs)))
		hex.Encode(data, bs)
	} else {
		data = bs
	}
	w.appendRight(data, width, flag)
}

func (w *Writer) AppendBool(b bool, width int, flag Flag) {
	w.appendLeft(flag)

	var tval, fval []byte
	if set := flag & YesNo; set != 0 {
		tval, fval = []byte("yes"), []byte("no")
	} else if set := flag & OnOff; set != 0 {
		tval, fval = []byte("on"), []byte("off")
	} else {
		tval, fval = []byte("true"), []byte("false")
	}
	var data []byte
	if b {
		data = tval
	} else {
		data = fval
	}
	w.appendRight(data, width, flag)
}

func (w *Writer) AppendInt(v int64, width int, flag Flag) {
	w.appendLeft(flag)

	if set := flag & ZeroFill; set != 0 {
		for i := 0; i < width; i++ {
			w.buffer[w.offset+i] = '0'
		}
	}
	base, data := prepareNumber(flag, v > 0)

	tmp := make([]byte, 0, 16)
	tmp = strconv.AppendInt(tmp, v, base)

	w.appendRight(append(data, tmp...), width, flag)
}

func (w *Writer) AppendUint(v uint64, width int, flag Flag) {
	w.appendLeft(flag)

	if set := flag & ZeroFill; set != 0 {
		for i := 0; i < width; i++ {
			w.buffer[w.offset+i] = '0'
		}
	}
	base, data := prepareNumber(flag, v > 0)

	tmp := make([]byte, 0, 16)
	tmp = strconv.AppendUint(tmp, v, base)

	w.appendRight(append(data, tmp...), width, flag)
}

func (w *Writer) appendRight(data []byte, width int, flag Flag) {
	var offset int
	if set := flag & AlignRight; set != 0 {
		offset = w.offset + (width - len(data))
	} else if set := flag & AlignCenter; set != 0 {
		offset = w.offset + ((width - len(data)) / 2)
	} else {
		offset = w.offset
	}
	copy(w.buffer[offset:], data)
	if len(data) > width {
		w.offset += len(data)
	} else {
		w.offset += width
	}

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

func prepareNumber(flag Flag, positive bool) (int, []byte) {
	base := 10
	if set := flag & Base16; set != 0 {
		base = 16
	} else if set := flag & Base8; set != 0 {
		base = 8
	} else if set := flag & Base2; set != 0 {
		base = 2
	}

	var data []byte
	if set := flag & WithSign; set != 0 && positive {
		data = append(data, '+')
	}
	if set := flag & WithPrefix; set != 0 {
		switch base {
		case 2:
			data = append(data, '0', 'b')
		case 8:
			data = append(data, '0', 'o')
		case 16:
			data = append(data, '0', 'x')
		}
	}
	return base, data
}
