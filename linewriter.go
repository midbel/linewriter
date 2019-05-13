package linewriter

import (
	// "fmt"
	"encoding/hex"
	"io"
	"strconv"
	"time"
	"unicode/utf8"
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
	NoSpace
	NoPadding
	NoSeparator
	YesNo
	OnOff
	TrueFalse
	OneZero
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
	SizeSI
	SizeIEC
)

type Option func(*Writer)

const DefaultFlags = AlignRight | Text | Second | TrueFalse | Decimal | Float | SizeIEC

type Writer struct {
	buffer []byte
	tmp    []byte
	base   int
	offset int

	dontaddsep bool

	padding   []byte
	separator []byte
	newline   []byte

	flags Flag
}

func NewWriter(size int, options ...Option) *Writer {
	w := Writer{
		buffer: make([]byte, size),
		tmp:    make([]byte, 0, 512),
		flags:  DefaultFlags,
	}
	for i := 0; i < len(options); i++ {
		options[i](&w)
	}
	if len(w.newline) == 0 {
		w.newline = []byte("\n")
	}
	w.Reset()
	return &w
}

func AsCSV(quoted bool) Option {
	return func(w *Writer) {
		w.separator = append(w.separator, ',')
		w.newline = append(w.newline, '\r', '\n')
		w.flags |= NoPadding | NoSpace // | WithPrefix
		if quoted {
			w.flags |= WithQuote
		}
	}
}

func WithLabel(p string) Option {
	return func(w *Writer) {
		str := []byte(p)
		if n := len(str) - 1; str[n] == ' ' {
			str = str[:n]
		}
		w.base = copy(w.buffer, str)
	}
}

func WithFlag(flag Flag) Option {
	return func(w *Writer) {
		w.flags = flag
	}
}

func WithPadding(pad []byte) Option {
	return func(w *Writer) {
		w.padding = append(w.padding, pad...)
	}
}

func WithSeparator(seq []byte) Option {
	return func(w *Writer) {
		w.separator = append(w.separator, seq...)
	}
}

func WithCRLF() Option {
	return func(w *Writer) {
		w.newline = append(w.newline, '\r', '\n')
	}
}

func (w *Writer) Reset() {
	n := len(w.buffer)
	if w.offset > 0 {
		n = w.offset
	}
	for i := w.base; i < n; i++ {
		w.buffer[i] = ' '
	}
	w.offset = w.base
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
	n := copy(bs, append(w.buffer[:w.offset], w.newline...))
	w.Reset()
	return n, nil
}

func (w *Writer) WriteTo(ws io.Writer) (int64, error) {
	defer w.Reset()
	if w.offset == 0 || w.offset == w.base {
		return 0, io.EOF
	}
	n, err := ws.Write(append(w.buffer[:w.offset], w.newline...))
	if err == nil {
		err = io.EOF
	}
	return int64(n), err
}

func (w *Writer) AppendSeparator(n int) {
	for i := 0; i < n; i++ {
		w.offset += copy(w.buffer[w.offset:], w.separator)
	}
	w.dontaddsep = n > 1
}

// func (w *Writer) AppendDatum(bs []byte, width int, flag Flag) {
// }

func (w *Writer) AppendString(str string, width int, flag Flag) {
	flag = flag &^ Hex
	w.AppendBytes([]byte(str), width, flag|Text)
}

func (w *Writer) AppendBytes(bs []byte, width int, flag Flag) {
	w.appendLeft(flag)

	if set := flag & Hex; set != 0 {
		data := make([]byte, hex.EncodedLen(len(bs)))
		hex.Encode(data, bs)
		w.tmp = append(w.tmp, data...)
	} else {
		w.tmp = append(w.tmp, bs...)
	}
	w.appendRight(w.tmp, width, flag)
	w.tmp = w.tmp[:0]
}

func (w *Writer) AppendTime(t time.Time, format string, flag Flag) {
	w.appendLeft(flag)

	w.tmp = t.AppendFormat(w.tmp, format)

	w.appendRight(w.tmp, len(w.tmp), flag)
	w.tmp = w.tmp[:0]
}

func (w *Writer) AppendDuration(d time.Duration, width int, flag Flag) {
	w.appendLeft(flag)

	if d == 0 {
		w.tmp = append(w.tmp, '0')
	} else {
		if d < 0 {
			w.tmp = append(w.tmp, '-')
			d = -d
		}
		ns := d.Nanoseconds()
		if d >= time.Minute {
			w.appendDHM(ns, flag)
		}
		if d >= time.Second {
			w.appendSeconds(ns, flag)
		} else {
			w.appendMillis(ns, flag)
		}
	}

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
	} else if set := flag & OneZero; set != 0 {
		tval, fval = []byte("1"), []byte("0")
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

func (w *Writer) AppendSize(v int64, width int, flag Flag) {
	w.appendLeft(flag)

	var (
		size int64
		prec int64
		unit byte
	)
	if isSizeIEC(w.flags, flag) {
		size, prec, unit = prepareSize(v, kibi, mebi, gibi, tebi, pebi, exbi)
	} else {
		size, prec, unit = prepareSize(v, kilo, mega, giga, tera, peta, exa)
	}
	w.tmp = strconv.AppendInt(w.tmp, size, 10)
	if prec > 0 {
		w.tmp = append(w.tmp, '.')
		n := len(w.tmp)
		w.tmp = strconv.AppendInt(w.tmp, prec, 10)
		if len(w.tmp)-n > 2 {
			w.tmp = w.tmp[:n+2]
		}
	}
	if unit != 0 {
		w.tmp = append(w.tmp, unit)
	}

	w.appendRight(w.tmp, width, flag)
	w.tmp = w.tmp[:0]
}

func (w *Writer) AppendInt(v int64, width int, flag Flag) {
	w.appendLeft(flag)

	base := w.prepareNumber(flag, v > 0)
	if set := flag & WithZero; set != 0 {
		for i := len(w.tmp); i < width; i++ {
			w.tmp = append(w.tmp, '0')
		}
	}
	tmp := make([]byte, 0, 8)
	tmp = strconv.AppendInt(tmp, v, base)

	if n := len(w.tmp); n == 0 || n < len(tmp) || flag&WithZero == 0 {
		w.tmp = append(w.tmp, tmp...)
	} else {
		copy(w.tmp[n-len(tmp):], tmp)
	}

	w.appendRight(w.tmp, width, flag)
	w.tmp = w.tmp[:0]
}

func (w *Writer) AppendUint(v uint64, width int, flag Flag) {
	w.appendLeft(flag)

	base := w.prepareNumber(flag, v > 0)
	if set := flag & WithZero; set != 0 {
		for i := len(w.tmp); i < width; i++ {
			w.tmp = append(w.tmp, '0')
		}
	}
	tmp := make([]byte, 0, 8)
	tmp = strconv.AppendUint(tmp, v, base)

	if n := len(w.tmp); n == 0 || n < len(tmp) || flag&WithZero == 0 {
		w.tmp = append(w.tmp, tmp...)
	} else {
		copy(w.tmp[n-len(tmp):], tmp)
	}

	w.appendRight(w.tmp, width, flag)
	w.tmp = w.tmp[:0]
}

func (w *Writer) appendMillis(ns int64, flag Flag) {
	const (
		micros = 1000
		millis = micros * 1000
	)
	var unit []byte
	if ns >= millis {
		w.tmp = strconv.AppendInt(w.tmp, ns/millis, 10)
		w.tmp = append(w.tmp, '.')
		if µs := ns % millis; µs > 0 && (flag & Millisecond) == 0 {
			w.tmp = strconv.AppendInt(w.tmp, µs, 10)
		}
		unit = []byte("ms")
	} else if ns >= micros {
		w.tmp = strconv.AppendInt(w.tmp, ns/micros, 10)
		w.tmp = append(w.tmp, '.')
		if ns := ns % micros; ns > 0 {
			w.tmp = strconv.AppendInt(w.tmp, ns, 10)
		}
		unit = []byte("µs")
	} else {
		w.tmp = strconv.AppendInt(w.tmp, ns, 10)
		unit = []byte("ns")
	}
	n := skipZeros(w.tmp)
	w.tmp = append(w.tmp[:n], unit...)
}

func (w *Writer) appendDHM(ns int64, flag Flag) {
	if d := ns / (int64(time.Hour) * 24); d > 0 {
		w.tmp = strconv.AppendInt(w.tmp, int64(d), 10)
		w.tmp = append(w.tmp, 'd')
	}
	if d := (ns / int64(time.Hour)) % 24; d > 0 {
		if d < 10 && len(w.tmp) > 0 && w.tmp[0] != '-' {
			w.tmp = append(w.tmp, '0')
		}
		w.tmp = strconv.AppendInt(w.tmp, int64(d), 10)
		w.tmp = append(w.tmp, 'h')
	}
	if d := (ns / int64(time.Minute)) % 60; d > 0 {
		if d < 10 && len(w.tmp) > 0 && w.tmp[0] != '-' {
			w.tmp = append(w.tmp, '0')
		}
		w.tmp = strconv.AppendInt(w.tmp, int64(d), 10)
		w.tmp = append(w.tmp, 'm')
	}
}

func (w *Writer) appendSeconds(ns int64, flag Flag) {
	v := (ns / int64(time.Second)) % 60
	if v < 10 && len(w.tmp) > 0 && w.tmp[0] != '-' {
		w.tmp = append(w.tmp, '0')
	}
	w.tmp, v = strconv.AppendInt(w.tmp, v, 10), -1
	if s1, s2 := flag&Millisecond, flag&Microsecond; s1 > 0 || s2 > 0 {
		w.tmp = append(w.tmp, '.')
	}

	if set := flag & Microsecond; set != 0 {
		ms := (ns / int64(time.Millisecond)) % 1000
		v = (ms * 1000) + ((ns / int64(time.Microsecond)) % 1000)
	} else if set := flag & Millisecond; set != 0 {
		v = (ns / int64(time.Millisecond)) % 1000
	} else {
		w.tmp = append(w.tmp, 's')
		return
	}
	n := len(w.tmp)
	if s1, s2 := flag&Millisecond, flag&Microsecond; s1 > 0 || s2 > 0 {
		if v < 10 {
			w.tmp = append(w.tmp, '0')
		}
		if v < 100 {
			w.tmp = append(w.tmp, '0')
		}
		if s2 > 0 && v < 100000 {
			w.tmp = append(w.tmp, '0')
		}
	}
	w.tmp = strconv.AppendInt(w.tmp, v, 10)

	n = skipZeros(w.tmp)
	w.tmp = append(w.tmp[:n], 's')
}

func skipZeros(tmp []byte) int {
	var n int
	for i := len(tmp) - 1; i >= 0; i-- {
		if tmp[i] != '0' {
			if i == len(tmp)-1 {
				n = len(tmp)
			} else {
				n = i + 1
			}
			break
		}
	}
	if n == 0 {
		n = len(tmp)
	}
	if tmp[n-1] == '.' {
		n--
	}
	return n
}

func (w *Writer) appendRight(data []byte, width int, flag Flag) {
	size := len(data)
	if size > width {
		width = size
	}

	var padleft, padright int
	if isWithSpace(w.flags, flag) {
		if set := flag & AlignRight; set != 0 {
			padleft = width - utf8.RuneCount(data)
		} else if set := flag & AlignCenter; set != 0 {
			padleft = (width - utf8.RuneCount(data)) / 2
			padright = padleft

			if c := padleft + padright + size; c < width {
				padright += width - c
			}
		} else {
			padright = width - utf8.RuneCount(data)
		}
	} else {
		if isWithQuote(w.flags, flag) {
			w.buffer[w.offset] = '"'
			w.offset++
		}
	}

	n := copy(w.buffer[w.offset+padleft:], data)
	w.offset += n
	if isWithSpace(w.flags, flag) {
		w.offset += padleft + padright
	} else {
		if isWithQuote(w.flags, flag) {
			w.buffer[w.offset] = '"'
			w.offset++
		}
	}

	if isWithPadding(w.flags, flag) {
		w.offset += copy(w.buffer[w.offset:], w.padding)
	}
}

func isSizeIEC(def, giv Flag) bool {
	d := def & SizeIEC
	g := giv & SizeIEC
	return g > 0 || d > 0
}

func isWithPrefix(def, giv Flag) bool {
	d := def & WithPrefix
	g := giv & WithPrefix
	return d > 0 || g > 0
}

func isWithQuote(def, giv Flag) bool {
	d := def & WithQuote
	g := giv & WithQuote
	return d > 0 || g > 0
}

func isWithSpace(def, giv Flag) bool {
	d := def & NoSpace
	g := giv & NoSpace
	return d == 0 && g == 0
}

func isWithPadding(def, giv Flag) bool {
	d := def & NoPadding
	g := giv & NoPadding
	return d == 0 && g == 0
}

func (w *Writer) appendLeft(flag Flag) {
	if set := flag & NoSeparator; w.offset > w.base && set == 0 {
		if !w.dontaddsep {
			n := copy(w.buffer[w.offset:], w.separator)
			w.offset += n
		} else {
			w.dontaddsep = !w.dontaddsep
		}
	}
	if isWithPadding(w.flags, flag) {
		w.offset += copy(w.buffer[w.offset:], w.padding)
	}
}

func (w *Writer) prepareNumber(flag Flag, positive bool) int {
	base := 10
	if set := flag & Hex; set != 0 {
		base = 16
	} else if set := flag & Octal; set != 0 {
		base = 8
	} else if set := flag & Binary; set != 0 {
		base = 2
	}

	if set := flag & WithSign; set != 0 && positive {
		w.tmp = append(w.tmp, '+')
	}
	if isWithPrefix(w.flags, flag) {
		switch base {
		case 2:
			w.tmp = append(w.tmp, '0', 'b')
		case 8:
			w.tmp = append(w.tmp, '0', 'o')
		case 16:
			w.tmp = append(w.tmp, '0', 'x')
		}
	}
	return base
}

const (
	kilo = 1000
	mega = kilo * kilo
	giga = kilo * mega
	tera = kilo * giga
	peta = kilo * tera
	exa  = kilo * peta
)

const (
	kibi = 1 << 10
	mebi = 1 << 20
	gibi = 1 << 30
	tebi = 1 << 40
	pebi = 1 << 50
	exbi = 1 << 60
)

func prepareSize(v, kb, mb, gb, tb, pb, eb int64) (int64, int64, byte) {
	var (
		unit byte
		mod  int64
	)
	switch {
	default:
		mod = 1
	case v >= kb && v < mb:
		mod, unit = kb, 'K'
	case v >= mb && v < gb:
		mod, unit = mb, 'M'
	case v >= gb && v < tb:
		mod, unit = gb, 'G'
	case v >= tb && v < pb:
		mod, unit = pb, 'P'
	case v >= eb:
		mod, unit = eb, 'E'
	}
	return v / mod, v % mod, unit
}
