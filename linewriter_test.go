package linewriter

import (
	"fmt"
	"testing"
	"time"
)

func ExampleWriter() {
	w := NewWriter(256, 1, ' ')
	w.AppendUint(1, 4, AlignRight)
	w.AppendUint(1, 4, AlignRight|Hex|WithZero)
	w.AppendString("playback", 10, AlignLeft)
	w.AppendUint(44, 2, AlignLeft|Decimal)
	w.AppendBool(false, 3, AlignCenter|OnOff)

	fmt.Println(w.String())
	// Output:
	//     1 | 0001 | playback   | 44 | off
}

func BenchmarkAppendString(b *testing.B) {
	w := NewWriter(256, 1, '_')
	for i := 0; i < b.N; i++ {
		w.AppendString("hello world", 12, Text|AlignRight)
		w.Reset()
	}
}

func TestAppendDuration(t *testing.T) {
	w := NewWriter(256, 1, '_')
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
	}
	for i, d := range data {
		v, _ := time.ParseDuration(d.Value)
		w.AppendDuration(v, 12, d.Flags)
		got := w.String()

		w.Reset()
		if got != d.Want {
			t.Logf("want: %x - got: %x", d.Want, got)
			t.Errorf("%d: failed: want %q (%d), got: %q (%d)", i+1, d.Want, len(d.Want), got, len(got))
		}
	}
}

func TestAppendFloat(t *testing.T) {
	w := NewWriter(256, 0, '_')
	data := []struct {
		Value float64
		Want  string
		Flags Flag
	}{
		{Value: 0.9845, Flags: Float | AlignRight, Want: "      0.98"},
		{Value: 0.9845, Flags: Float | Percent | AlignRight, Want: "    98.45%"},
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
	w := NewWriter(256, 1, '_')
	data := []struct {
		Value string
		Want  string
		Flags Flag
	}{
		{Value: "hello", Flags: AlignRight, Want: "_     hello_"},
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
	w := NewWriter(250, 1, '_')
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

func TestAppendBool(t *testing.T) {
	w := NewWriter(250, 1, '_')
	data := []struct {
		Value bool
		Want  string
		Flags Flag
	}{
		{Value: true, Flags: AlignRight | YesNo, Want: "_  yes_"},
		{Value: true, Flags: AlignRight | OnOff, Want: "_   on_"},
		{Value: true, Flags: AlignRight | TrueFalse, Want: "_ true_"},
		{Value: false, Flags: AlignRight | YesNo, Want: "_   no_"},
		{Value: false, Flags: AlignRight | OnOff, Want: "_  off_"},
		{Value: false, Flags: AlignRight | TrueFalse, Want: "_false_"},
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

	w := NewWriter(256, 1, '_')

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

func TestAppendUint(t *testing.T) {
	w := New(256, 1)
	data := []struct {
		Value uint64
		Want  string
		Flags Flag
	}{
		{Value: 453721, Flags: Decimal | AlignRight, Want: "     453721 "},
		{Value: 453721, Flags: Decimal | AlignLeft, Want: " 453721     "},
		{Value: 453721, Flags: Decimal | AlignRight | NoPadding, Want: "    453721"},
		{Value: 453721, Flags: Decimal | AlignLeft | NoPadding, Want: "453721    "},
		{Value: 453721, Flags: Hex | AlignLeft | NoPadding, Want: "6ec59     "},
		{Value: 453721, Flags: Hex | AlignRight | NoPadding, Want: "     6ec59"},
		{Value: 453721, Flags: Hex | AlignRight | NoPadding | WithZero, Want: "000006ec59"},
		{Value: 453721, Flags: Hex | AlignRight | WithZero, Want: " 000006ec59 "},
		{Value: 453721, Flags: Hex | AlignLeft, Want: " 6ec59      "},
		{Value: 453721, Flags: Hex | AlignRight, Want: "      6ec59 "},
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
