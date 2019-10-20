package lege

import (
	"bufio"
	"io"
	"unicode/utf8"
)

// ParseOptions are options passed to a parser
type ParseOptions struct {
	Start           []string
	End             []string
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

type Collection struct {
	string
	Boundary      BoundaryOption
	StartLocation Location
	EndLocation   Location
}

func (options *ParseOptions) maxStartLength() (max int) {
	for _, s := range options.Start {
		if utf8.RuneCountInString(s) > max {
			max = utf8.RuneCountInString(s)
		}
	}
	return max
}

func (options *ParseOptions) maxEndLength() (max int) {
	for _, s := range options.End {
		if utf8.RuneCountInString(s) > max {
			max = utf8.RuneCountInString(s)
		}
	}
	return max
}

// NewParser creates a *Parser
func NewParser(options *ParseOptions) (*Parser, error) {
	parser := &Parser{options: options}
	return parser, nil
}

// ParseReader takes a reader
func (p *Parser) ParseReader(reader io.Reader) ([]string, error) {
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
	collections := make([][]rune, 0)
	collecting := false

	for {
		if c, _, err := r.ReadRune(); err != nil {
			if err == io.EOF {
				if collecting { // if we're still collecting at the EOF, drop the last collection
					collections = collections[:len(collections)-1]
				}
				break
			} else {
				return nil, err
			}
		} else {
			// fmt.Printf("[%v],%q,\n", collecting, string(window))
			// fmt.Printf("%q : %d : %v\n", string(window), len(window), collecting)

			if index < windowSize { // the window needs to be initially populated
				window[index] = c
				index++
				continue
			}

			if !collecting { // if we're not collecting, we're looking for a start match
				for _, startOption := range p.options.Start {
					var winString string
					runesInOption := utf8.RuneCountInString(startOption)
					if runesInOption < len(window) {
						winString = string(window[len(window)-runesInOption:])
					} else {
						winString = string(window)
					}
					if startOption == winString { // if the window matches a start option
						collecting = true                            // go into collecting mode
						collections = append(collections, []rune{c}) // create a new collection, starting with this rune
						break
					}
				}
			} else { // if we're collecting, we're looking for an end match and storing runes along the way
				currentCollection := collections[len(collections)-1]
				for _, endOption := range p.options.End {
					var winString string
					runesInOption := utf8.RuneCountInString(endOption)
					if runesInOption < len(window) {
						winString = string(window[len(window)-runesInOption:])
					} else {
						winString = string(window)
					}
					if endOption == winString { // if the window matches an end option
						collecting = false // leave collecting mode
						// if we're stopping collection, since the window trails the current index, we need to reslice the current collection to take off
						// the runes we just matched
						collections[len(collections)-1] = currentCollection[:len(currentCollection)-utf8.RuneCountInString(endOption)]
						break
					}
				}
				if collecting {
					currentCollection = append(currentCollection, c)
					collections[len(collections)-1] = currentCollection
				}
			}

			// shift the window by one rune
			for i := range window {
				if i == len(window)-1 {
					window[i] = c
				} else {
					window[i] = window[i+1]
				}
			}
			index++
		}
	}

	output := make([]string, len(collections))
	for c, collection := range collections {
		output[c] = string(collection)
	}
	return output, nil
}
