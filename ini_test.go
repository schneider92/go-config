package config

import (
	"bufio"
	"bytes"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type sortingWriter struct {
	currentLine string
	lines       []string
}

func (w *sortingWriter) Write(data []byte) (int, error) {
	s := w.currentLine + string(data)
	for {
		idx := strings.IndexAny(s, "\r\n")
		if idx < 0 {
			w.currentLine = s
			return len(data), nil
		}
		if idx > 0 {
			w.lines = append(w.lines, s[:idx])
		}
		s = s[idx+1:]
	}
}

func (w *sortingWriter) GetData() string {
	if w.currentLine != "" {
		w.lines = append(w.lines, w.currentLine)
		w.currentLine = ""
	}
	slices.Sort(w.lines)
	return strings.Join(w.lines, "\n")
}

func TestSaveIniReal(t *testing.T) {
	// otherwise, heading is skipped and results are sorted, but this test tests the original output
	l := NewLayer("test")
	l.SetString("test.key", "value")
	l.SetString("my.dummy.stuff", "12345")
	var buf bytes.Buffer
	SaveIni(l, &buf)
	s := buf.String()

	// exp1 and exp2 differs only in order of the two values
	exp1 := `;
; This INI file was autogenerated
;

test.key=value
my.dummy.stuff=12345
`
	exp2 := `;
; This INI file was autogenerated
;

my.dummy.stuff=12345
test.key=value
`
	if s != exp1 && s != exp2 {
		assert.Fail(t, "ini output unexpected: ", s)
	}
}

func TestSaveIni(t *testing.T) {
	// create layer
	l := NewLayer("test")
	l.SetString("key=1", "value1") // test escaping equal sign
	l.SetString(`key\2`, "value2") // test escaping backslash
	l.SetString("xx", `yy
zz`)
	l.SetString("cc", "ww")

	// write it
	var w sortingWriter
	saveIniInternal(l, &w, false)
	assert.Equal(t, `cc=ww
key\=1=value1
key\\2=value2
xx=yy\nzz`, w.GetData())
}

func TestLoadIni(t *testing.T) {
	buf := bytes.NewBufferString(`
my.key.1 = value1
my.key.2 = value2`)

	l := NewLayer("test")
	err := LoadIni(l, bufio.NewReader(buf))
	assert.Nil(t, err)

	// check count
	assert.Equal(t, 2, len(listKeys(l, "", false)))

	// check values
	s, ok := l.GetString("my.key.1")
	assert.True(t, ok)
	assert.Equal(t, "value1", s)

	s, ok = l.GetString("my.key.2")
	assert.True(t, ok)
	assert.Equal(t, "value2", s)

	// try to load ini to a read-only target
	assert.Panics(t, func() {
		LoadIni(NewEmptyView(), bufio.NewReader(buf))
	})
}

func TestLoadIniWithEscapes(t *testing.T) {
	buf := bytes.NewBufferString(`
my\=key\=1=value\\1
;this = is a comment
my\rkey\n2= value\t2\r
this line has no eq sign and is ignored
whitespaces=\ s p 
`)

	l := NewLayer("test")
	err := LoadIni(l, bufio.NewReader(buf))
	assert.Nil(t, err)

	// check count
	assert.Equal(t, 3, len(listKeys(l, "", false)))

	// check values
	s, ok := l.GetString("my=key=1")
	assert.True(t, ok)
	assert.Equal(t, `value\1`, s)

	s, ok = l.GetString("my\rkey\n2")
	assert.True(t, ok)
	assert.Equal(t, "value\t2\r", s)

	s, ok = l.GetString("whitespaces")
	assert.True(t, ok)
	assert.Equal(t, " s p ", s)
}

func TestLoadIniWithSections(t *testing.T) {
	buf := bytes.NewBufferString(`
my.key=root

; declare a section
[my.section]
key=sect
new.key=newsect

; declare a new section
[my]
section.k2=v2

; do an empty section to feed root again
[]
my.root=rootval
`)

	l := NewLayer("test")
	err := LoadIni(l, bufio.NewReader(buf))
	assert.Nil(t, err)

	// check count
	assert.Equal(t, 5, len(listKeys(l, "", false)))

	// check values
	s, ok := l.GetString("my.key")
	assert.True(t, ok)
	assert.Equal(t, "root", s)

	s, ok = l.GetString("my.section.key")
	assert.True(t, ok)
	assert.Equal(t, "sect", s)

	s, ok = l.GetString("my.section.new.key")
	assert.True(t, ok)
	assert.Equal(t, "newsect", s)

	s, ok = l.GetString("my.section.k2")
	assert.True(t, ok)
	assert.Equal(t, "v2", s)

	s, ok = l.GetString("my.root")
	assert.True(t, ok)
	assert.Equal(t, "rootval", s)
}

type fakeIniTestErrorReader struct {
	data []byte
}

func (r *fakeIniTestErrorReader) Read(buf []byte) (int, error) {
	if r.data == nil {
		return 0, io.ErrNoProgress // anything but not EOF
	} else {
		ret := copy(buf, r.data)
		r.data = r.data[ret:]
		if len(r.data) == 0 {
			r.data = nil
		}
		return ret, nil
	}
}

func TestLoadIniIoError(t *testing.T) {
	var rd fakeIniTestErrorReader
	rd.data = []byte("key=value\n")

	l := NewLayer("test")
	err := LoadIni(l, bufio.NewReader(&rd))
	assert.Equal(t, io.ErrNoProgress, err)

	s, ok := l.GetString("key")
	assert.True(t, ok)
	assert.Equal(t, "value", s)
}
