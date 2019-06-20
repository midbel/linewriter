package linewriter

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"
)

var defaults = []Option{
	WithPadding([]byte("_")),
	WithSeparator([]byte("|")),
}

func ExampleWriter() {
	w1 := NewWriter(256, WithPadding([]byte("_")), WithSeparator([]byte("|")))
	w1.AppendUint(1, 4, AlignRight)
	w1.AppendUint(1, 4, AlignRight|Hex|WithZero)
	w1.AppendString("playback", 10, AlignLeft)
	w1.AppendUint(44, 2, AlignLeft|Decimal)
	w1.AppendBool(false, 3, AlignCenter|OnOff)

	w2 := NewWriter(256, WithPadding([]byte("_")), WithSeparator([]byte("|")), WithLabel("[label]"))
	w2.AppendUint(1, 4, AlignRight)
	w2.AppendUint(1, 4, AlignRight|Hex|WithZero)
	w2.AppendString("playback", 10, AlignLeft)
	w2.AppendUint(44, 2, AlignLeft|Decimal)
	w2.AppendBool(false, 3, AlignCenter|OnOff)

	fmt.Println(w1.String())
	fmt.Println(w2.String())
	// Output:
	// _   1_|_0001_|_playback  _|_44_|_off_
	// [label]_   1_|_0001_|_playback  _|_44_|_off_
}

func ExampleWriter_AppendSeparator() {
	w1 := NewWriter(256, WithPadding([]byte("_")), WithSeparator([]byte("|")))
	w1.AppendUint(1, 4, AlignRight)
	w1.AppendUint(1, 4, AlignRight|Hex|WithZero)
	w1.AppendSeparator(1)
	w1.AppendString("playback", 10, AlignLeft)
	w1.AppendSeparator(3)
	w1.AppendUint(44, 2, AlignLeft|Decimal)
	w1.AppendBool(false, 3, AlignCenter|OnOff)

	fmt.Println(w1.String())
	// Output:
	// _   1_|_0001_||_playback  _|||_44_|_off_
}

func ExampleAsCSV() {
	w1 := NewWriter(256, AsCSV(false))
	w1.AppendUint(1, 4, AlignRight)
	w1.AppendUint(1, 4, AlignRight|Hex|WithZero)
	w1.AppendString("playback", 10, AlignLeft)
	w1.AppendUint(44, 2, AlignLeft|Decimal)
	w1.AppendBool(false, 3, AlignCenter|OnOff)

	w2 := NewWriter(256, AsCSV(true))
	w2.AppendUint(1, 4, AlignRight)
	w2.AppendUint(1, 4, AlignRight|Hex|WithZero)
	w2.AppendString("playback", 10, AlignLeft)
	w2.AppendUint(44, 2, AlignLeft|Decimal)
	w2.AppendBool(false, 3, AlignCenter|OnOff)

	fmt.Println(w1.String())
	fmt.Println(w2.String())
	// Output:
	// 1,0001,playback,44,off
	// "1","0001","playback","44","off"
}

func BenchmarkAppendString(b *testing.B) {
	w := NewWriter(256, defaults...)
	for i := 0; i < b.N; i++ {
		w.AppendString("hello world", 12, Text|AlignRight)
		w.Reset()
	}
}

func TestRead(t *testing.T) {
	w1 := NewWriter(256, WithPadding([]byte("_")), WithSeparator([]byte("|")))
	w1.AppendUint(1, 4, AlignRight)
	w1.AppendUint(1, 4, AlignRight|Hex|WithZero)
	w1.AppendString("playback", 10, AlignLeft)
	w1.AppendUint(44, 2, AlignLeft|Decimal)
	w1.AppendBool(false, 3, AlignCenter|OnOff)

	str := w1.String() + "\n"

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, w1); err != nil && err != io.EOF {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if str != buf.String() {
		t.Errorf("want: %s (%[1]x), got : %s (%[2]x)", str, buf.String())
	}
}

func TestAppendDuration(t *testing.T) {
	w := NewWriter(256, defaults...)
	data := []struct {
		Value string
		Want  string
		Flags Flag
	}{
		{Value: "9m47.8791231s", Flags: AlignRight | Second | WithZero, Want: "_       9m47s_"},
		{Value: "9m47.8791231s", Flags: AlignRight | Millisecond | WithZero, Want: "_   9m47.879s_"},
		{Value: "9m47.8791231s", Flags: AlignRight | Microsecond | WithZero, Want: "_9m47.879123s_"},
		{Value: "17h31m10.100s", Flags: AlignRight | Second | WithZero, Want: "_   17h31m10s_"},
		{Value: "17h31m10.100s", Flags: AlignRight | Millisecond | WithZero, Want: "_ 17h31m10.1s_"},
		{Value: "17h31m10.100s", Flags: AlignRight | Microsecond | WithZero, Want: "_ 17h31m10.1s_"},
		{Value: "246h18m17.012387s", Flags: AlignRight | Second | WithZero, Want: "_10d06h18m17s_"},
		{Value: "246h18m17.012387s", Flags: AlignRight | Millisecond | WithZero, Want: "_10d06h18m17.012s_"},
		{Value: "246h18m17.012387s", Flags: AlignRight | Microsecond | WithZero, Want: "_10d06h18m17.012387s_"},
		{Value: "1452.32µs", Flags: AlignRight | WithZero, Want: "_   1.45232ms_"},
		{Value: "452.32µs", Flags: AlignRight | WithZero, Want: "_    452.32µs_"},
		{Value: "452ns", Flags: AlignRight | WithZero, Want: "_       452ns_"},
		{Value: "-1m12s", Flags: AlignRight | Second, Want: "_      -1m12s_"},
		{Value: "-1h12m37s", Flags: AlignRight | Second, Want: "_   -1h12m37s_"},
		{Value: "-34s", Flags: AlignRight | Second, Want: "_        -34s_"},
		{Value: "-3s", Flags: AlignRight | Second, Want: "_         -3s_"},
	}
	for i, d := range data {
		v, _ := time.ParseDuration(d.Value)
		w.AppendDuration(v, 12, d.Flags)
		got := w.String()

		w.Reset()
		if got != d.Want {
			t.Errorf("%d: failed: want %q (%d), got: %q (%d)", i+1, d.Want, len(d.Want), got, len(got))
		}
	}
}

func TestAppendFloat(t *testing.T) {
	w := NewWriter(256, defaults...)
	data := []struct {
		Value float64
		Want  string
		Flags Flag
	}{
		{Value: 0.9845, Flags: Float | AlignRight, Want: "_      0.98_"},
		{Value: 0.9845, Flags: Float | Percent | AlignRight, Want: "_    98.45%_"},
		{Value: 0, Flags: Float | AlignRight, Want: "_         0_"},
		{Value: 0.00, Flags: Float | AlignRight, Want: "_         0_"},
		{Value: 0.01, Flags: Float | AlignRight, Want: "_      0.01_"},
		{Value: 0.1230, Flags: Float | AlignRight, Want: "_      0.12_"},
	}
	for i, d := range data {
		if set := d.Flags & Percent; set == 0 {
			w.AppendFloat(d.Value, 10, 2, d.Flags)
		} else {
			w.AppendPercent(d.Value, 10, 2, d.Flags)
		}
		got := w.String()

		w.Reset()
		if got != d.Want {
			t.Errorf("%d: failed: want %q (%d), got: %q (%d)", i+1, d.Want, len(d.Want), got, len(got))
		}
	}
}

func TestAppendString(t *testing.T) {
	w := NewWriter(256, defaults...)
	data := []struct {
		Value string
		Want  string
		Flags Flag
	}{
		{Value: "hello", Flags: AlignRight, Want: "_     hello_"},
		{Value: "hello", Flags: AlignRight | WithQuote, Want: "_     hello_"},
		{Value: "hello", Flags: AlignRight | WithQuote | NoSpace, Want: "_\"hello\"_"},
		{Value: "hello", Flags: AlignRight | NoSpace, Want: "_hello_"},
		{Value: "hello", Flags: AlignRight | NoSpace | NoPadding, Want: "hello"},
		{Value: "hello", Flags: AlignLeft, Want: "_hello     _"},
		{Value: "hello", Flags: AlignRight | NoPadding, Want: "     hello"},
		{Value: "hello", Flags: AlignLeft | NoPadding, Want: "hello     "},
		{Value: "hello", Flags: AlignCenter | NoPadding, Want: "  hello   "},
	}
	for i, d := range data {
		w.AppendString(d.Value, 10, d.Flags)
		got := w.String()

		w.Reset()
		if got != d.Want {
			t.Errorf("%d: failed: want %q (%d), got: %q (%d)", i+1, d.Want, len(d.Want), got, len(got))
		}
	}
}

func TestAppendInt(t *testing.T) {
	w := NewWriter(256, defaults...)
	data := []struct {
		Value int64
		Want  string
		Flags Flag
	}{
		{Value: 3, Flags: AlignRight | Decimal, Want: "_    3_"},
		{Value: 3, Flags: AlignLeft | Decimal, Want: "_3    _"},
		{Value: 3, Flags: AlignLeft | WithSign | Decimal, Want: "_+3   _"},
		{Value: 15, Flags: AlignRight | WithPrefix | Hex, Want: "_  0xf_"},
	}
	for i, d := range data {
		w.AppendInt(d.Value, 5, d.Flags)
		got := w.String()

		w.Reset()
		if got != d.Want {
			t.Errorf("%d: failed: want %q (%d), got: %q (%d)", i+1, d.Want, len(d.Want), got, len(got))
		}
	}
}

func TestAppendUint(t *testing.T) {
	w := NewWriter(256, defaults...)
	data := []struct {
		Value uint64
		Want  string
		Flags Flag
	}{
		{Value: 453721, Flags: Decimal | AlignRight, Want: "_    453721_"},
		{Value: 453721, Flags: Decimal | AlignLeft, Want: "_453721    _"},
		{Value: 453721, Flags: Decimal | AlignRight | NoPadding, Want: "    453721"},
		{Value: 453721, Flags: Decimal | AlignLeft | NoPadding, Want: "453721    "},
		{Value: 453721, Flags: Hex | AlignLeft | NoPadding, Want: "6ec59     "},
		{Value: 453721, Flags: Hex | AlignRight | NoPadding, Want: "     6ec59"},
		{Value: 453721, Flags: Hex | AlignRight | NoPadding | WithZero, Want: "000006ec59"},
		{Value: 453721, Flags: Hex | AlignRight | WithZero, Want: "_000006ec59_"},
		{Value: 453721, Flags: Hex | AlignLeft, Want: "_6ec59     _"},
		{Value: 453721, Flags: Hex | AlignRight, Want: "_     6ec59_"},
		{Value: 453721, Flags: Decimal | WithSign | AlignLeft, Want: "_+453721   _"},
		{Value: 453721, Flags: Hex | WithPrefix | AlignLeft, Want: "_0x6ec59   _"},
		{Value: 453721, Flags: Hex | WithPrefix | AlignRight, Want: "_   0x6ec59_"},
		{Value: 453721, Flags: Hex | WithZero | WithPrefix | AlignLeft, Want: "_0x0006ec59_"},
		{Value: 5, Flags: Binary | AlignRight, Want: "_       101_"},
		{Value: 5, Flags: Binary | WithPrefix | AlignRight, Want: "_     0b101_"},
		{Value: 5, Flags: Octal | AlignRight, Want: "_         5_"},
		{Value: 5, Flags: Octal | WithPrefix | AlignRight, Want: "_       0o5_"},
	}
	for i, d := range data {
		w.AppendUint(d.Value, 10, d.Flags)
		got := w.String()

		w.Reset()
		if got != d.Want {
			t.Errorf("%d: failed: want %q (%d), got: %q (%d)", i+1, d.Want, len(d.Want), got, len(got))
		}
	}
}

func TestAppendBool(t *testing.T) {
	w := NewWriter(256, defaults...)
	data := []struct {
		Value bool
		Want  string
		Flags Flag
	}{
		{Value: true, Flags: AlignRight | YesNo, Want: "_  yes_"},
		{Value: true, Flags: AlignRight | OnOff, Want: "_   on_"},
		{Value: true, Flags: AlignRight | TrueFalse, Want: "_ true_"},
		{Value: true, Flags: AlignRight | OneZero, Want: "_    1_"},
		{Value: false, Flags: AlignRight | YesNo, Want: "_   no_"},
		{Value: false, Flags: AlignRight | OnOff, Want: "_  off_"},
		{Value: false, Flags: AlignRight | TrueFalse, Want: "_false_"},
		{Value: false, Flags: AlignRight | OneZero, Want: "_    0_"},
	}
	for i, d := range data {
		w.AppendBool(d.Value, 5, d.Flags)
		got := w.String()

		w.Reset()
		if got != d.Want {
			t.Errorf("%d: failed: want %q (%d), got: %q (%d)", i+1, d.Want, len(d.Want), got, len(got))
		}
	}
}

func TestAppendTime(t *testing.T) {
	const timeFormat = "2006-01-02 15:04:05.000"

	w := NewWriter(256, defaults...)

	d := time.Date(2019, 6, 11, 12, 25, 43, 0, time.UTC)
	data := []struct {
		Value  time.Time
		Want   string
		Format string
		Flags  Flag
	}{
		{Value: d, Format: timeFormat, Want: "_2019-06-11 12:25:43.000_", Flags: AlignRight},
		{Value: d, Format: timeFormat, Want: "_2019-06-11 12:25:43.000_", Flags: AlignCenter},
		{Value: d, Format: timeFormat, Want: "2019-06-11 12:25:43.000", Flags: NoPadding},
	}
	for i, d := range data {
		w.AppendTime(d.Value, d.Format, d.Flags)
		got := w.String()

		w.Reset()
		if got != d.Want {
			t.Errorf("%d: failed: want %q (%d), got: %q (%d)", i+1, d.Want, len(d.Want), got, len(got))
		}
	}
}
