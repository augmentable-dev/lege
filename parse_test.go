package lege

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestSingleCollection(t *testing.T) {
	const src = `ABCDEFGHIJKLMNOPQRSTUVWXYZ`
	p, err := NewParser(&ParseOptions{
		BoundaryOptions: []BoundaryOption{
			{Starts: []string{"ABC"}, Ends: []string{"G"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	collections, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"DEF"}
	got := collections.Strings()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want: %v, got: %v", want, got)
	}
	startLocation := collections[0].StartLocation
	if startLocation.Line != 1 {
		t.Fatal("expected collection to start on line 1")
	}
	if want := 4; startLocation.Pos != want {
		t.Fatalf("expected collection to start at position: %d, got: %d", want, startLocation.Pos)
	}
	endLocation := collections[0].EndLocation
	if endLocation.Line != 1 {
		t.Fatal("expected collection to end on line 1")
	}
	if want := 6; endLocation.Pos != want {
		t.Fatalf("expected collection to end at position: %d, got: %d", want, endLocation.Pos)
	}
}

func TestMultipleCollections(t *testing.T) {
	const src = `<ABCD><EFGH><><IHJKLMNO><hello`
	p, err := NewParser(&ParseOptions{
		BoundaryOptions: []BoundaryOption{
			{Starts: []string{"<"}, Ends: []string{">"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	collections, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"ABCD", "EFGH", "", "IHJKLMNO"}
	got := collections.Strings()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want: %v, got: %v", want, got)
	}
	fmt.Println(collections[1], collections[1].StartLocation.Pos, collections[1].EndLocation.Pos)
}

func TestEmojiOptions(t *testing.T) {
	const src = `ABCDE✅FGHIJKLMNOP✅QRSTUVWXYZ`
	p, err := NewParser(&ParseOptions{
		BoundaryOptions: []BoundaryOption{
			{Starts: []string{"✅"}, Ends: []string{"✅"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	collections, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"FGHIJKLMNOP"}
	got := collections.Strings()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want: %v, got: %v", want, got)
	}
	fmt.Println(collections[0], collections[0].StartLocation.Pos, collections[0].EndLocation.Pos)
}

func TestCStyleCodeComments(t *testing.T) {
	const src = `
        // A COMMENT
        i_am = "some pseudo code"
        log(i_am)
        /* A MULTI
        LINE COMMENT */
        `
	p, err := NewParser(&ParseOptions{
		BoundaryOptions: []BoundaryOption{
			{Starts: []string{"//"}, Ends: []string{"\n"}},
			{Starts: []string{"/*"}, Ends: []string{"*/"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	collections, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{" A COMMENT", " A MULTI\n        LINE COMMENT "}
	got := collections.Strings()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want: %q, got: %q", want, got)
	}
	fmt.Println(collections[1], collections[1].StartLocation.Pos, collections[1].EndLocation.Pos)
}

func RubyStyleCodeComment(t *testing.T) {
	const src = `
        # A COMMENT
        i_am = "some pseudo code"
        log(i_am)

        `
	p, err := NewParser(&ParseOptions{
		BoundaryOptions: []BoundaryOption{
			{Starts: []string{"#"}, Ends: []string{"\n"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	collections, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{" A COMMENT"}
	got := collections.Strings()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want: %v, got: %v", want, got)
	}
}
