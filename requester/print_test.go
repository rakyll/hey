// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package requester

import (
	"testing"
)

var req = `GET / HTTP/1.1
Host: example.com
Content-Type: text/html
User-Agent: hey/0.0.1`
var expectedReq = `> GET / HTTP/1.1
> Host: example.com
> Content-Type: text/html
> User-Agent: hey/0.0.1`

var res = `HTTP/1.1 404 Not Found
Content-Length: 54
Connection: keep-alive
Content-Type: application/json
Date: Tue, 11 Dec 2012 00:00:00 GMT

{
  "statusCode" : 404,
  "description": "Not Found"
}`
var expectedRes = `< HTTP/1.1 404 Not Found
< Content-Length: 54
< Connection: keep-alive
< Content-Type: application/json
< Date: Tue, 11 Dec 2012 00:00:00 GMT
<` + " \n" +
	`< {
<   "statusCode" : 404,
<   "description": "Not Found"
< }`

func TestAppendPrefixToEachLine(t *testing.T) {
	var tests = []struct {
		input    string
		prefix   string
		expected string
	}{
		{req, "> ", expectedReq},
		{res, "< ", expectedRes},
	}

	for _, test := range tests {
		if output := appendPrefixToEachLine(test.input, test.prefix); output != test.expected {
			t.Errorf("input:\n%s\nprefix:\n\"%s\"\n\nexpected:\n%s \n\nactual:\n%s", test.input, test.prefix, test.expected, output)
		}
	}
}
