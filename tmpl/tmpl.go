package tmpl

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
	"text/template/parse"
)

type TemplateItem int

const (
	Unknown TemplateItem = iota
	RequestID
	Email
	Date
	Time
	DateTime
	Integer
	Float
	String
)

var (
	ErrNoTmpl       = errors.New("no templates in request")
	ErrUnknownsTmpl = errors.New("unknowns templates in request")
)

var stringToTemplateItem = map[string]TemplateItem{
	"RequestID": RequestID, // Random requestID
	"Email":     Email,     // Random email
	"Date":      Date,      // Random date
	"Time":      Time,      // Random time
	"DtTm":      DateTime,  // Date and Time
	"Integer":   Integer,   // Random integer value
	"Float":     Float,     // Random float value
	"String":    String,    // Random string
}

func (tmplItem TemplateItem) String() string {
	for key, item := range stringToTemplateItem {
		if item == tmplItem {
			return key
		}
	}

	return ""
}

// Generate randdov value based on tipe TemplateItem
func (tmplItem TemplateItem) Generate() string {
	switch tmplItem {
	case RequestID:
		return RandomRequest()
	case Email:
		return RandomEmail()
	case String:
		return RandomString()
	case Time:
		return RandomTime()
	case Date:
		return RandomDate()
	case DateTime:
		return RandomDateTime()
	case Integer:
		return RandomInteger()
	case Float:
		return RandomFloat()
	}

	return ""
}

// Convernt node form template to TemplateItem
func nodeToTemplateItem(node string) TemplateItem {
	switch {
	case strings.Contains(node, RequestID.String()):
		return RequestID
	case strings.Contains(node, Email.String()):
		return Email
	case strings.Contains(node, Integer.String()):
		return Integer
	case strings.Contains(node, Date.String()):
		return Date
	case strings.Contains(node, Time.String()):
		return Time
	case strings.Contains(node, DateTime.String()):
		return DateTime
	case strings.Contains(node, Float.String()):
		return Float
	case strings.Contains(node, String.String()):
		return String
	}

	return Unknown
}

func ListTemplFields(t *template.Template) []string {
	return listNodeFields(t.Tree.Root, nil)
}

func listNodeFields(node parse.Node, res []string) []string {
	if node.Type() == parse.NodeAction {
		res = append(res, node.String())
	}

	if ln, ok := node.(*parse.ListNode); ok {
		for _, n := range ln.Nodes {
			res = listNodeFields(n, res)
		}
	}
	return res
}

// Build template from string
func templateString(value string) (*template.Template, error) {
	if !(strings.Contains(value, "{{.") && strings.Contains(value, "}}")) {
		return nil, ErrNoTmpl
	}
	tmpl, err := template.New("").Parse(value)
	if err != nil {
		return nil, ErrNoTmpl
	}

	i := 0
	for _, node := range ListTemplFields(tmpl) {
		if nodeToTemplateItem(node) == Unknown {
			i++
		}
	}

	if i > 0 {
		return nil, ErrUnknownsTmpl
	}

	return tmpl, nil
}

// Build template from []byte value
func templateByte(body []byte) (*template.Template, error) {
	return templateString(string(body))
}

// Validate headers and body of request. If they not consist any template
// return ErrNoTmpl.
func Validate(req *http.Request, body []byte) error {
	i := 0

	_, err := templateByte(body)
	if err != ErrUnknownsTmpl {
		i++
	}

	for key := range req.Header {
		value := req.Header.Get(key)

		_, err1 := templateString(value)
		if err1 != ErrUnknownsTmpl {
			i++
		}
	}

	if i == len(req.Header) {
		return ErrUnknownsTmpl
	}

	return nil
}

func fillTemplate(tmpl *template.Template) (string, error) {
	tmplItems := make(map[string]string)
	for _, node := range ListTemplFields(tmpl) {
		tmplItem := nodeToTemplateItem(node)
		if tmplItem == Unknown {
			continue
		}

		key := strings.TrimLeft(node, "{{.")
		key = strings.TrimRight(key, "}}")
		tmplItems[key] = tmplItem.Generate()
	}

	var b bytes.Buffer
	err := tmpl.Execute(&b, &tmplItems)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// FillOut fill out templates in headers and body of request with random values
func FillOut(r *http.Request, body []byte) (req *http.Request) {
	r2 := new(http.Request)
	*r2 = *r

	tmplBody, err := templateByte(body)
	if err == nil {
		bodyString, err1 := fillTemplate(tmplBody)
		if err1 == nil {
			r2.ContentLength = int64(len(bodyString))
			r2.Body = ioutil.NopCloser(bytes.NewReader([]byte(bodyString)))
		}
	} else {
		r2.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	r2.Header = make(http.Header, len(r.Header))

	for k, s := range r.Header {
		value := r.Header.Get(k)

		tmplHeader, err := templateString(value)
		if err != nil {
			r2.Header[k] = append([]string(nil), s...)
			continue
		}

		headerString, err := fillTemplate(tmplHeader)
		if err == nil {
			r2.Header[k] = append([]string(nil), []string{headerString}...)
		}
	}

	return r2
}
