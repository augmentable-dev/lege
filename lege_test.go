package lege

import (
	"strings"
	"testing"
)

func TestSingleCollection(t *testing.T) {
	const src = `ABCDEFGHIJKLMNOPQRSTUVWXYZ`
	p, err := NewParser(&ParseOptions{
		Start: []string{"ABC"},
		End:   []string{"GHI"},
	})
	if err != nil {
		t.Fatal(err)
	}
	collections, err := p.ParseReader(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	{
		want := 1
		got := len(collections)
		if got != want {
			t.Fatalf("want: %d, got: %d collections", want, got)
		}
	}
	{
		want := "DEF"
		got := collections[0]
		if got != want {
			t.Fatalf("want: %s got: %s as collection", want, got)
		}
	}
}

func TestMultipleCollections(t *testing.T) {
	const src = `<ABCD><EFGH><><IHJKLMNO><hello`
	p, err := NewParser(&ParseOptions{
		Start: []string{"<"},
		End:   []string{">"},
	})
	if err != nil {
		t.Fatal(err)
	}
	collections, err := p.ParseReader(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	{
		want := 4
		got := len(collections)
		if got != want {
			t.Fatalf("want: %d, got: %d collections", want, got)
		}
	}
	{
		want := []string{"ABCD", "EFGH", "", "IHJKLMNO"}
		got := collections
		for i, w := range want {
			if g := got[i]; g != w {
				t.Fatalf("want: %s, got: %s", w, g)
			}
		}
	}
}

func TestEmojiOptions(t *testing.T) {
	const src = `ABCDE✅FGHIJKLMNOP✅QRSTUVWXYZ`
	p, err := NewParser(&ParseOptions{
		Start: []string{"✅"},
		End:   []string{"✅"},
	})
	if err != nil {
		t.Fatal(err)
	}
	collections, err := p.ParseReader(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	{
		want := 1
		got := len(collections)
		if got != want {
			t.Fatalf("want: %d, got: %d collections", want, got)
		}
	}
	{
		want := []string{"FGHIJKLMNOP"}
		got := collections
		for i, w := range want {
			if g := got[i]; g != w {
				t.Fatalf("want: %s, got: %s", w, g)
			}
		}
	}
}

func TestCStyleCodeComments(t *testing.T) {
	const src = `
	// A COMMENT
	i_am = "some pseudo code"
	log(i_am)
	`
	p, err := NewParser(&ParseOptions{
		Start: []string{"//"},
		End:   []string{"\n"},
	})
	if err != nil {
		t.Fatal(err)
	}
	collections, err := p.ParseReader(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	{
		want := 1
		got := len(collections)
		if got != want {
			t.Fatalf("want: %d, got: %d collections", want, got)
		}
	}
	{
		want := []string{" A COMMENT"}
		got := collections
		for i, w := range want {
			if g := got[i]; g != w {
				t.Fatalf("want: %s, got: %s", w, g)
			}
		}
	}
}
