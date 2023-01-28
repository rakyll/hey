package requester

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var tmplRegex = regexp.MustCompile(`\{(i\d*|f\d*|s\d*)(?::(\d+):(\d+))?\}`)
var charPool = []byte("abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 0123456789")
var poolLen = len(charPool) - 1

// RequestGenerator yields a func that gives request with dynamic Body
// OR, nil if the request payload is not dynamic (based on placeholder).
func RequestGenerator(req *http.Request, body []byte) func() *http.Request {
	// Check if body has dynamic placeholder tokens
	matches := tmplRegex.FindAllSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}

	toInt := func(b []byte) int {
		i, _ := strconv.ParseInt(string(b), 10, 0)
		return int(i)
	}

	replacers := map[string]func() []byte{}
	nameTokens := map[string]string{}
	holders := map[string][][]byte{}

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

		min, max := toInt(match[2]), toInt(match[3])

		switch match[1][0] {
		case 'i': // int
			replacers[token] = func() []byte {
				rand.Seed(time.Now().UnixNano())
				val := fmt.Sprintf("%d", rand.Intn(max-min)+min)
				return []byte(val)
			}
		case 'f': // float
			replacers[token] = func() []byte {
				rand.Seed(time.Now().UnixNano())
				val := fmt.Sprintf("%d.%d", rand.Intn(max-min)+min, rand.Intn(99))
				return []byte(val)
			}
		case 's': // string
			replacers[token] = func() []byte {
				rand.Seed(time.Now().UnixNano())
				sLen := rand.Intn(max-min) + min
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
		newBody := body
		for token, replacer := range replacers {
			val := replacer()
			for _, holder := range holders[token] {
				newBody = bytes.Replace(newBody, holder, val, -1)
			}
		}
		newReq := cloneRequest(req, newBody)
		newReq.ContentLength = int64(len(newBody))
		return newReq
	}
}
