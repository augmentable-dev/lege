package main

import (
	"strings"

	"github.com/augmentable-dev/lege"
)

func main() {
	const src = `ABCDEFGHIJKLMNOPQRSTUVXYZ`

	p, err := lege.NewParser(&lege.ParseOptions{
		Start: []string{"EF"},
		End:   []string{"NO"},
	})
	if err != nil {
		panic(err)
	}
	p.ParseReader(strings.NewReader(src))
}
