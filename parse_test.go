package lege

import (
	"reflect"
	"strings"
	"testing"
)

func TestInvalidOptions(t *testing.T) {
	invalidOptions := map[string]*ParseOptions{
		"NilOptions":    nil,
		"NilBoundaries": &ParseOptions{},
		"EmptyBoundary": &ParseOptions{
			Boundaries: []Boundary{},
		},
		"EmptyStart": &ParseOptions{
			Boundaries: []Boundary{
				Boundary{Start: "", End: "\n"},
			},
		},
		"EmptyEnd": &ParseOptions{
			Boundaries: []Boundary{
				Boundary{Start: "//", End: ""},
			},
		},
	}

	for name, opt := range invalidOptions {
		t.Run(name, func(t *testing.T) {
			_, err := NewParser(opt)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestSingleCollection(t *testing.T) {
	const src = `ABCDEFGHIJKLMNOPQRSTUVWXYZ`
	p, err := NewParser(&ParseOptions{
		Boundaries: []Boundary{
			{Start: "ABC", End: "G"},
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
	boundaryOptions := []Boundary{{Start: "<", End: ">"}}
	p, err := NewParser(&ParseOptions{
		Boundaries: boundaryOptions,
	})
	if err != nil {
		t.Fatal(err)
	}
	collections, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	want := Collections{
		&Collection{
			runes:         []rune("ABCD"),
			Boundary:      boundaryOptions[0],
			StartLocation: Location{Line: 1, Pos: 2},
			EndLocation:   Location{Line: 1, Pos: 5},
		},
		&Collection{
			runes:         []rune("EFGH"),
			Boundary:      boundaryOptions[0],
			StartLocation: Location{Line: 1, Pos: 8},
			EndLocation:   Location{Line: 1, Pos: 11},
		},
		// TODO fix the following
		&Collection{
			runes:         []rune(""),
			Boundary:      boundaryOptions[0],
			StartLocation: Location{Line: 1, Pos: 14},
			EndLocation:   Location{Line: 1, Pos: 13},
		},
		&Collection{
			runes:         []rune("IHJKLMNO"),
			Boundary:      boundaryOptions[0],
			StartLocation: Location{Line: 1, Pos: 16},
			EndLocation:   Location{Line: 1, Pos: 23},
		},
	}
	got := collections
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want: %v, got: %v", want, got)
	}
}

func TestEmojiOptions(t *testing.T) {
	const src = `ABCDE✅FGHIJKLMNOP✅QRSTUVWXYZ`
	boundaryOptions := []Boundary{{Start: "✅", End: "✅"}}
	p, err := NewParser(&ParseOptions{
		Boundaries: boundaryOptions,
	})
	if err != nil {
		t.Fatal(err)
	}
	collections, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}

	want := Collections{
		&Collection{
			runes:         []rune("FGHIJKLMNOP"),
			Boundary:      boundaryOptions[0],
			StartLocation: Location{Line: 1, Pos: 7},
			EndLocation:   Location{Line: 1, Pos: 17},
		},
	}
	got := collections
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want: %v, got: %v", want, got)
	}
}

func TestCStyleCodeComments(t *testing.T) {
	const src = `
        // A COMMENT
        i_am = "some pseudo code"
        log(i_am)
        /* A MULTI
        LINE COMMENT */
		`
	boundaryOptions := []Boundary{
		{Start: "//", End: "\n"},
		{Start: "/*", End: "*/"},
	}
	p, err := NewParser(&ParseOptions{
		Boundaries: boundaryOptions,
	})
	if err != nil {
		t.Fatal(err)
	}
	collections, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	want := Collections{
		// TODO address the -1
		&Collection{
			runes:         []rune(" A COMMENT"),
			Boundary:      boundaryOptions[0],
			StartLocation: Location{Line: 1, Pos: 12},
			EndLocation:   Location{Line: 2, Pos: -1},
		},
		&Collection{
			runes:         []rune(" A MULTI\n        LINE COMMENT "),
			Boundary:      boundaryOptions[1],
			StartLocation: Location{Line: 4, Pos: 11},
			EndLocation:   Location{Line: 5, Pos: 21},
		},
	}
	got := collections
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want: %v, got: %v", want, got)
	}

}

func TestRubyStyleCodeComment(t *testing.T) {
	const src = `
        # A COMMENT
        i_am = "some pseudo code"
        log(i_am)

		`
	boundaryOptions := []Boundary{
		{Start: "#", End: "\n"},
	}
	p, err := NewParser(&ParseOptions{
		Boundaries: boundaryOptions,
	})
	if err != nil {
		t.Fatal(err)
	}
	collections, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	want := Collections{
		// TODO address the -1
		&Collection{
			runes:         []rune(" A COMMENT"),
			Boundary:      boundaryOptions[0],
			StartLocation: Location{Line: 1, Pos: 11},
			EndLocation:   Location{Line: 2, Pos: -1},
		},
	}
	got := collections
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want: %v, got: %v", want, got)
	}
}
