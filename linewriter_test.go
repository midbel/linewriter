package linewriter

import (
	"testing"
)

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
