package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestView(t *testing.T) {
	// create layer and set some values
	l := NewLayer("test")
	l.SetString("my.test.key.first", "1st")
	l.SetString("my.test.key.onlymy", "2nd")
	l.SetString("my.test.stuff", "mystuff")
	l.SetString("your.test.key.first", "3rd")
	l.SetString("your.test.key.onlyyour", "4th")
	l.SetString("your.test.stuff", "yourstuff")

	// create view with empty prefix
	v := NewView(l, "")

	// create view with empty prefix that wraps the previous one
	v = NewView(v, "")

	// get some values
	s, ok := v.GetString("my.test.stuff")
	assert.True(t, ok)
	assert.Equal(t, "mystuff", s)

	s, ok = v.GetString("your.test.key.first")
	assert.True(t, ok)
	assert.Equal(t, "3rd", s)

	// create view with "my" prefix
	v = NewView(v, "my")

	// get some values
	s, ok = v.GetString("my.test.key.first")
	assert.False(t, ok)
	assert.Equal(t, "", s)

	s, ok = v.GetString("test.key.first")
	assert.True(t, ok)
	assert.Equal(t, "1st", s)

	s, ok = v.GetString("test.key.onlymy")
	assert.True(t, ok)
	assert.Equal(t, "2nd", s)

	s, ok = v.GetString("test.key.onlyyour")
	assert.False(t, ok)
	assert.Equal(t, "", s)

	s, ok = v.GetString("test.stuff")
	assert.True(t, ok)
	assert.Equal(t, "mystuff", s)

	// create new view
	v = NewView(l, "my")

	// create subview
	v2 := v.SubView("test")

	// set a value
	assert.True(t, v2.IsWritable())
	v2.SetString("stuff", "123")

	// create a read-only subview
	v2 = v.SubViewReadOnly("test")

	// cannot set value
	assert.False(t, v2.IsWritable())
	assert.Panics(t, func() {
		v2.SetString("stuff", "456")
	})

	// get stuff value
	s, ok = v2.GetString("stuff")
	assert.True(t, ok)
	assert.Equal(t, "123", s)

	// get keys
	keys := listKeys(v2, "", true)
	assert.ElementsMatch(t, keys, []string{"stuff", "key"})

	// try to delete through a read-only view
	assert.Panics(t, func() {
		v2.DeleteValue("stuff")
	})

	// delete key.first through a writable view
	v2 = v.SubView("test")
	v2.DeleteValue("key.first")
	keys = listKeys(v2, "", true)
	assert.ElementsMatch(t, keys, []string{"stuff", "key"})

	// delete key.onlymy
	v2.DeleteValue("key.onlymy")
	keys = listKeys(v2, "", true)
	assert.ElementsMatch(t, keys, []string{"stuff"})
}

func TestViewIntBool(t *testing.T) {
	// create layer and set some values
	l := NewLayer("test")
	v := NewView(l, "")
	v.SetInt("year", 2024)
	v.SetInt("numtrue", 1)
	v.SetBool("altrue", true)
	v.SetString("numstart", "123text")
	v.SetString("hex", "0xff")

	// check strings
	s, ok := l.GetString("year")
	assert.True(t, ok)
	assert.Equal(t, "2024", s)

	s, ok = l.GetString("numtrue")
	assert.True(t, ok)
	assert.Equal(t, "1", s)

	s, ok = l.GetString("altrue")
	assert.True(t, ok)
	assert.Equal(t, "true", s)

	s, ok = l.GetString("numstart")
	assert.True(t, ok)
	assert.Equal(t, "123text", s)

	s, ok = l.GetString("hex")
	assert.True(t, ok)
	assert.Equal(t, "0xff", s)

	// check numbers
	i, ok := v.GetInt("year")
	assert.True(t, ok)
	assert.EqualValues(t, 2024, i)

	i, ok = v.GetInt("numtrue")
	assert.True(t, ok)
	assert.EqualValues(t, 1, i)

	i, ok = v.GetInt("altrue")
	assert.False(t, ok)

	i, ok = v.GetInt("numstart")
	assert.False(t, ok)

	i, ok = v.GetInt("hex")
	assert.True(t, ok)
	assert.EqualValues(t, 255, i)

	// check bools
	b, ok := v.GetBool("year")
	assert.False(t, ok)

	b, ok = v.GetBool("numtrue")
	assert.True(t, ok)
	assert.True(t, b)

	b, ok = v.GetBool("altrue")
	assert.True(t, ok)
	assert.True(t, b)

	b, ok = v.GetBool("numstart")
	assert.False(t, ok)

	b, ok = v.GetBool("hex")
	assert.False(t, ok)
}

func TestEmptyView(t *testing.T) {
	// create empty view, test if read-only
	view := NewEmptyView()
	assert.False(t, view.writable)

	// set writable flag to test and cover the emptyViewable methods
	view.writable = true

	// get
	_, found := view.GetString("")
	assert.False(t, found)
	_, found = view.GetString("test")
	assert.False(t, found)
	assert.False(t, view.IsWritable())

	// try to set and delete
	assert.Panics(t, func() {
		view.SetString("key", "value")
	})
	_, found = view.GetString("key")
	assert.False(t, found)
	assert.Panics(t, func() {
		view.DeleteValue("key")
	})

	// list keys
	keys := KeyList{}
	view.ListKeys("", &keys, true)
	assert.Nil(t, keys.v)
}
