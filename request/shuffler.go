package request

import (
	"github.com/lucasjones/reggen"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var /* const */ variablePattern = regexp.MustCompile("{{.*}}")

type generator interface {
	Generate(limit int) string
}

type Shuffler struct {
	generatorMap map[string]generator
}

func NewShuffler(baseRequest *http.Request) *Shuffler {
	return newShuffler(baseRequest, func(match string) generator {
		gen, err := reggen.NewGenerator(match)
		if err != nil {
			log.Panicf("Invalid RegeExp: %s", match)
		}
		return gen
	})
}

func newShuffler(baseRequest *http.Request, genFunc func(match string) generator) *Shuffler {
	shuffler := &Shuffler{}
	shuffler.generatorMap = make(map[string]generator)
	for _, outerMatch := range variablePattern.FindStringSubmatch(baseRequest.URL.Path) {
		innerMatch := strings.Replace(outerMatch, "{{", "", 1)
		innerMatch = strings.Replace(innerMatch, "}}", "", 1)
		if innerMatch != "" {
			shuffler.generatorMap[outerMatch] = genFunc(outerMatch)
		}
	}

	return shuffler
}

func (s *Shuffler) Shuffle(r *http.Request) {
	for match, gen := range s.generatorMap {
		r.URL, _ = r.URL.Parse(strings.Replace(r.URL.Path, match, gen.Generate(1), -1))
	}
}
