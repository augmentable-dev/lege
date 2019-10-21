package lege

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"
)

// ParseOptions are options passed to a parser
type ParseOptions struct {
	BoundaryOptions []BoundaryOption
}

// BoundaryOption are boundaries to use when collecting strings
type BoundaryOption struct {
	Starts []string
	Ends   []string
}

// Parser is used to...
type Parser struct {
	options *ParseOptions
}

// Location represents location in an input string
type Location struct {
	Line int
	Pos  int
}

// Collection ...
type Collection struct {
	runes         []rune
	Boundary      BoundaryOption
	StartLocation Location
	EndLocation   Location
}

// Collections ...
type Collections []*Collection

func (collections Collections) getLast() *Collection {
	return collections[len(collections)-1]
}

// Strings returns each collection as a string, in a list of strings
func (collections Collections) Strings() (s []string) {
	for _, collection := range collections {
		s = append(s, collection.String())
	}
	return s
}

func (collection *Collection) addRune(r rune) {
	collection.runes = append(collection.runes, r)
}

func (collection *Collection) trimLeftRunes(num int) {
	collection.runes = collection.runes[num:]
}

func (collection *Collection) trimRightRunes(num int) {
	if num <= len(collection.runes) {
		collection.runes = collection.runes[:len(collection.runes)-num]
	}
}

func (collection *Collection) String() string {
	return string(collection.runes)
}

func (options *ParseOptions) maxStartLength() (max int) {
	for _, s := range options.getAllStarts() {
		if l := utf8.RuneCountInString(s); l > max {
			max = l
		}
	}
	return max
}

func (options *ParseOptions) maxEndLength() (max int) {
	for _, s := range options.getAllEnds() {
		if l := utf8.RuneCountInString(s); l > max {
			max = l
		}
	}
	return max
}

func (options *ParseOptions) getAllStarts() []string {
	starts := make([]string, 0)
	for _, boundary := range options.BoundaryOptions {
		for _, start := range boundary.Starts {
			starts = append(starts, start)
		}
	}
	return starts
}

func (options *ParseOptions) getAllEnds() []string {
	ends := make([]string, 0)
	for _, boundary := range options.BoundaryOptions {
		for _, end := range boundary.Ends {
			ends = append(ends, end)
		}
	}
	return ends
}

func (options *ParseOptions) getCorrespondingBoundary(start string) *BoundaryOption {
	for _, boundary := range options.BoundaryOptions {
		for _, s := range boundary.Starts {
			if s == start {
				return &boundary
			}
		}
	}
	return nil
}

// NewParser creates a *Parser
func NewParser(options *ParseOptions) (*Parser, error) {
	parser := &Parser{options: options}
	return parser, nil
}

// ParseReader takes a reader
func (p *Parser) ParseReader(reader io.Reader) (Collections, error) {
	r := bufio.NewReader(reader)
	maxStartLen := p.options.maxStartLength()
	maxEndLen := p.options.maxEndLength()
	windowSize := 0
	if maxStartLen > maxEndLen {
		windowSize = maxStartLen
	} else {
		windowSize = maxEndLen
	}
	window := make([]rune, windowSize)
	index := 0
	lineCounter := 1
	positionCounter := 1
	collections := make(Collections, 0)
	collecting := false

	windowMatchesString := func(window []rune, compareTo string) (bool, string) {
		var winString string
		runesInOption := utf8.RuneCountInString(compareTo)
		if runesInOption < len(window) {
			winString = string(window[len(window)-runesInOption:])
		} else {
			winString = string(window)
		}
		return compareTo == winString, winString
	}

	for {
		c, _, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				if collecting { // if we're still collecting at the EOF, drop the last collection
					collections = collections[:len(collections)-1]
				}
				break
			} else {
				return nil, err
			}
		}

		// fmt.Printf("%q : %q : %v : %d : %d\n", string(window), c, collecting, lineCounter, positionCounter)

		if index < windowSize { // the window needs to be initially populated
			window[index] = c
			index++
			positionCounter++
			continue
		}

		if !collecting { // if we're not collecting, we're looking for a start match
			for _, startOption := range p.options.getAllStarts() { // find a match with any of the possible starts
				match, _ := windowMatchesString(window, startOption)
				if match { // if the window matches a start option
					collecting = true // go into collecting mode
					boundary := p.options.getCorrespondingBoundary(startOption)
					if boundary == nil {
						panic(fmt.Sprintf("boundary not found for start: %s", startOption))
					}
					collections = append(collections, &Collection{
						runes:    []rune{c},
						Boundary: *boundary,
						StartLocation: Location{
							Line: lineCounter,
							Pos:  positionCounter,
						},
					}) // create a new collection, starting with this rune
					break
				}
			}
		} else { // if we're collecting, we're looking for an end match and storing runes along the way
			currentCollection := collections.getLast()
			for _, endOption := range currentCollection.Boundary.Ends {
				match, _ := windowMatchesString(window, endOption)
				if match { // if the window matches an end option
					collecting = false // leave collecting mode
					// if we're stopping collection, since the window trails the current index, we need to reslice the current collection to take off
					// the runes we just matched
					runeCount := utf8.RuneCountInString(endOption)
					currentCollection.trimRightRunes(runeCount)
					currentCollection.EndLocation = Location{
						Line: lineCounter,
						Pos:  positionCounter - runeCount - 1,
					}
					break
				}
			}
			if collecting {
				currentCollection.addRune(c)
			}
		}

		// shift the window by one rune
		for i := range window {
			if i == len(window)-1 { // if we're at the last spot in the window
				window[i] = c // assign it to the current rune
			} else { // otherwise, assign the current spot in the window to what's in the next spot
				window[i] = window[i+1]
			}
		}
		index++
		positionCounter++

		if string(c) == "\n" {
			lineCounter++
			positionCounter = 1
		}
	}

	return collections, nil
}
