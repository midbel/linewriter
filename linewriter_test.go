package linewriter

import (
	"fmt"
	"testing"
)

func ExampleWriter() {
	w := NewWriter(256, 1, ' ')
	w.AppendUint(1, 4, AlignRight)
	w.AppendUint(1, 4, AlignRight|Base16|ZeroFill)
	w.AppendString("playback", 10, AlignLeft)
	w.AppendUint(44, 2, AlignLeft|Base10)
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
		{Value: 3, Flags: AlignRight | Base10, Want: "_    3_"},
		{Value: 3, Flags: AlignLeft | Base10, Want: "_3    _"},
		{Value: 3, Flags: AlignLeft | WithSign | Base10, Want: "_+3   _"},
		{Value: 15, Flags: AlignRight | WithPrefix | Base16, Want: "_  0xf_"},
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
	t.SkipNow()
}

func TestAppendUint(t *testing.T) {
	w := New(256, 1)
	data := []struct {
		Value uint64
		Want  string
		Flags Flag
	}{
		{Value: 453721, Flags: Base10 | AlignRight, Want: "     453721 "},
		{Value: 453721, Flags: Base10 | AlignLeft, Want: " 453721     "},
		{Value: 453721, Flags: Base10 | AlignRight | NoPadding, Want: "    453721"},
		{Value: 453721, Flags: Base10 | AlignLeft | NoPadding, Want: "453721    "},
		{Value: 453721, Flags: Base16 | AlignLeft | NoPadding, Want: "6ec59     "},
		{Value: 453721, Flags: Base16 | AlignRight | NoPadding, Want: "     6ec59"},
		{Value: 453721, Flags: Base16 | AlignRight | NoPadding | ZeroFill, Want: "000006ec59"},
		{Value: 453721, Flags: Base16 | AlignRight | ZeroFill, Want: " 000006ec59 "},
		{Value: 453721, Flags: Base16 | AlignLeft, Want: " 6ec59      "},
		{Value: 453721, Flags: Base16 | AlignRight, Want: "      6ec59 "},
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
