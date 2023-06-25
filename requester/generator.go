package requester

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var tmplRegex = regexp.MustCompile(`\{(i\d*|f\d*|s\d*)(?::(\d+):(\d+))?\}`)
var charPool = []byte("abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 0123456789")
var pathEnc = strings.NewReplacer("%7B", "{", "%7D", "}")
var poolLen = len(charPool) - 1

// RequestGenerator yields a func that gives request with dynamic Body or URI path/params.
// OR, nil if the request payload and path/params are not dynamic (based on placeholder).
func RequestGenerator(req *http.Request, body []byte) func() *http.Request {
	// Check if URL path has dynamic placeholder tokens
	fullPath := []byte(pathEnc.Replace(req.URL.String()))
	umatches := tmplRegex.FindAllSubmatch(fullPath, -1)

	// Check if body has dynamic placeholder tokens
	matches := tmplRegex.FindAllSubmatch(body, -1)
	dynaBody, dynaURI := len(matches) > 0, len(umatches) > 0

	if !dynaBody && !dynaURI {
		return nil
	}

	toInt := func(b []byte) int {
		i, _ := strconv.ParseInt(string(b), 10, 0)
		return int(i)
	}

	replacers := map[string]func() []byte{}
	nameTokens := map[string]string{}
	holders := map[string][][]byte{}

	matches = append(matches, umatches...)
	for _, match := range matches {
		token, name := string(match[0]), string(match[1])
		if _, ok := replacers[token]; ok {
			continue
		}

		if tok, ok := nameTokens[name]; ok {
			holders[tok] = append(holders[tok], match[0])
			continue
		}

		nameTokens[name] = token
		if _, ok := holders[token]; !ok {
			holders[token] = [][]byte{match[0]}
		}

		min, max := 1, 10
		if len(match[2]) > 0 {
			min = toInt(match[2])
		}
		if len(match[3]) > 0 {
			max = toInt(match[3])
		}
		if max < min {
			max = min
		}

		switch match[1][0] {
		case 'i': // int
			replacers[token] = func() []byte {
				if max == min {
					val := fmt.Sprintf("%d", min)
					return []byte(val)
				}
				rand.Seed(time.Now().UnixNano())
				val := fmt.Sprintf("%d", rand.Intn(max-min)+min)
				return []byte(val)
			}
		case 'f': // float
			replacers[token] = func() []byte {
				rand.Seed(time.Now().UnixNano())
				if max == min {
					val := fmt.Sprintf("%d.%d", min, rand.Intn(99))
					return []byte(val)
				}
				val := fmt.Sprintf("%d.%d", rand.Intn(max-min)+min, rand.Intn(99))
				return []byte(val)
			}
		case 's': // string
			replacers[token] = func() []byte {
				rand.Seed(time.Now().UnixNano())
				sLen := min
				if min != max {
					sLen = rand.Intn(max-min) + min
				}
				val := make([]byte, sLen)
				for i := 0; i < sLen; i++ {
					val[i] = charPool[rand.Intn(poolLen)]
				}
				return val
			}
		}
	}

	// All the preparations are already done,
	// So the final function just replaces the token placeholders
	return func() *http.Request {
		newBody, newPath := body, fullPath
		for token, replacer := range replacers {
			val := replacer()
			for _, holder := range holders[token] {
				if dynaBody {
					newBody = bytes.Replace(newBody, holder, val, -1)
				}
				if dynaURI {
					newPath = bytes.Replace(newPath, holder, val, -1)
				}
			}
		}
		newReq := cloneRequest(req, newBody)
		if dynaBody {
			newReq.ContentLength = int64(len(newBody))
		}
		if dynaURI {
			newReq.URL, _ = url.Parse(string(newPath))
		}
		return newReq
	}
}
