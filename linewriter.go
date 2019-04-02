package linewriter

import (
	"encoding/hex"
	"io"
	"strconv"
	"time"
)

type Flag uint64

const (
	AlignLeft Flag = 1 << iota
	AlignRight
	AlignCenter
	WithZero
	WithSign
	WithPrefix
	WithQuote
	NoPadding
	NoSeparator
	YesNo
	OnOff
	TrueFalse
	Hex
	Octal
	Binary
	Decimal
	Percent
	Float
	Scientific
	Text
	Bytes
	Second
	Millisecond
	Microsecond
)

type Writer struct {
	buffer []byte
	tmp    []byte
	offset int

	padding   []byte
	separator []byte
}

func NewWriter(size, padsiz int, padchar byte) *Writer {
	w := Writer{
		buffer:    make([]byte, size),
		tmp:       make([]byte, 0, 512),
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

func (w *Writer) AppendTime(t time.Time, format string, flag Flag) {
	w.appendLeft(flag)

	// tmp := make([]byte, 0, 64)
	w.tmp = t.AppendFormat(w.tmp, format)

	w.appendRight(w.tmp, len(w.tmp), flag)
	w.tmp = w.tmp[:0]
}

func (w *Writer) AppendDuration(d time.Duration, width int, flag Flag) {
	w.appendLeft(flag)

	ns := d.Nanoseconds()
	if d := ns / (int64(time.Hour) * 24); d > 0 {
		w.tmp = strconv.AppendInt(w.tmp, int64(d), 10)
		w.tmp = append(w.tmp, 'd')
	}
	if d := (ns / int64(time.Hour)) % 24; d > 0 {
		if set := flag & WithZero; set != 0 && d < 10 {
			w.tmp = append(w.tmp, '0')
		}
		w.tmp = strconv.AppendInt(w.tmp, int64(d), 10)
		w.tmp = append(w.tmp, 'h')
	}
	if d := (ns / int64(time.Minute)) % 60; d > 0 {
		if set := flag & WithZero; set != 0 && d < 10 {
			w.tmp = append(w.tmp, '0')
		}
		w.tmp = strconv.AppendInt(w.tmp, int64(d), 10)
		w.tmp = append(w.tmp, 'm')
	}
	var v int64
	if set := flag & Microsecond; set != 0 {
		ms := (ns / int64(time.Millisecond)) % 1000
		v = (ms * 1000) + ((ns / int64(time.Microsecond)) % 1000)
	} else if set := flag & Millisecond; set != 0 {
		v = (ns / int64(time.Millisecond)) % 1000
	} else {
		v = (ns / int64(time.Second)) % 60
		if set := flag & WithZero; set != 0 && v < 10 {
			w.tmp = append(w.tmp, '0')
		}
	}
	w.tmp = strconv.AppendInt(w.tmp, v, 10)
	w.tmp = append(w.tmp, 's')

	w.appendRight(w.tmp, width, flag)
	w.tmp = w.tmp[:0]
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

func (w *Writer) AppendPercent(v float64, width, prec int, flag Flag) {
	w.AppendFloat(v*100.0, width, prec, flag|Percent|Float)
}

func (w *Writer) AppendFloat(v float64, width, prec int, flag Flag) {
	w.appendLeft(flag)

	var format byte = 'g'
	if set := flag & Scientific; set != 0 {
		format = 'e'
	} else if set := flag & Float; set != 0 {
		format = 'f'
	}
	w.tmp = strconv.AppendFloat(w.tmp, v, format, prec, 64)
	if set := flag & Percent; set != 0 {
		w.tmp = append(w.tmp, '%')
	}
	w.appendRight(w.tmp, width, flag)
	w.tmp = w.tmp[:0]
}

func (w *Writer) AppendInt(v int64, width int, flag Flag) {
	w.appendLeft(flag)

	if set := flag & WithZero; set != 0 {
		for i := 0; i < width; i++ {
			w.buffer[w.offset+i] = '0'
		}
	}
	var base int
	base, w.tmp = prepareNumber(w.tmp, flag, v > 0)

	w.tmp = strconv.AppendInt(w.tmp, v, base)

	w.appendRight(w.tmp, width, flag)
	w.tmp = w.tmp[:0]
}

func (w *Writer) AppendUint(v uint64, width int, flag Flag) {
	w.appendLeft(flag)

	if set := flag & WithZero; set != 0 {
		for i := 0; i < width; i++ {
			w.buffer[w.offset+i] = '0'
		}
	}
	var base int
	base, w.tmp = prepareNumber(w.tmp, flag, v > 0)

	// tmp := make([]byte, 0, 16)
	w.tmp = strconv.AppendUint(w.tmp, v, base)

	w.appendRight(w.tmp, width, flag)
	w.tmp = w.tmp[:0]
}

func (w *Writer) appendRight(data []byte, width int, flag Flag) {
	var offset int
	if len(data) > width {
		width = len(data)
	}
	if set := flag & AlignRight; set != 0 {
		offset = w.offset + (width - len(data))
	} else if set := flag & AlignCenter; set != 0 {
		offset = w.offset + ((width - len(data)) / 2)
	} else {
		offset = w.offset
	}
	copy(w.buffer[offset:], data)
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

func prepareNumber(data []byte, flag Flag, positive bool) (int, []byte) {
	base := 10
	if set := flag & Hex; set != 0 {
		base = 16
	} else if set := flag & Octal; set != 0 {
		base = 8
	} else if set := flag & Binary; set != 0 {
		base = 2
	}

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
