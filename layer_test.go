package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLayerSetGet(t *testing.T) {
	l := NewLayer("testlayer")
	assert.Equal(t, "testlayer", l.Name())

	// get values
	s, ok := l.GetString("qwer")
	assert.False(t, ok)
	assert.Equal(t, "", s)

	s, ok = l.GetString("asdf")
	assert.False(t, ok)
	assert.Equal(t, "", s)

	// set values
	assert.True(t, l.IsWritable())
	l.SetString("qwer", "123")
	l.SetString("zxcv", "456")

	// get values
	s, ok = l.GetString("qwer")
	assert.True(t, ok)
	assert.Equal(t, "123", s)

	s, ok = l.GetString("asdf")
	assert.False(t, ok)
	assert.Equal(t, "", s)

	s, ok = l.GetString("zxcv")
	assert.True(t, ok)
	assert.Equal(t, "456", s)

	// lock to read-only and try to set value
	l.LockReadOnly()
	assert.False(t, l.IsWritable())
	assert.Panics(t, func() {
		l.SetString("panics", "death")
	})

	s, ok = l.GetString("panics")
	assert.False(t, ok)
	assert.Equal(t, "", s)
}

func listKeys(l Viewable, prefix string, direct bool) []string {
	var keys KeyList
	l.ListKeys(prefix, &keys, direct)
	return keys.ToSlice()
}

func TestLayerListKeys(t *testing.T) {
	l := NewLayer("testlayer")

	// list keys -> empty
	keys := listKeys(l, "", false)
	assert.Zero(t, len(keys))

	// add some keys
	l.SetString("x.my.qwer", "123")
	l.SetString("x.my.asdf", "456")
	l.SetString("x.your.asdf", "789")

	// list keys
	keys = listKeys(l, "", false)
	assert.ElementsMatch(t, keys, []string{"x.my.asdf", "x.my.qwer", "x.your.asdf"})

	// list keys under x
	keys = listKeys(l, "x", false)
	assert.ElementsMatch(t, keys, []string{"my.asdf", "my.qwer", "your.asdf"})

	// list keys under x.my
	keys = listKeys(l, "x.my", false)
	assert.ElementsMatch(t, keys, []string{"asdf", "qwer"})

	// list direct keys
	keys = listKeys(l, "", true)
	assert.ElementsMatch(t, keys, []string{"x"})

	// list direct keys under x
	keys = listKeys(l, "x", true)
	assert.ElementsMatch(t, keys, []string{"my", "your"})

	// list direct keys under x.my
	keys = listKeys(l, "x.my", true)
	assert.ElementsMatch(t, keys, []string{"asdf", "qwer"})

	// list direct keys under noexist
	keys = listKeys(l, "noexist", true)
	assert.ElementsMatch(t, keys, []string{})
}

func TestLayerDelete(t *testing.T) {
	l := NewLayer("testlayer")

	// add some keys
	l.SetString("qwer", "123")
	l.SetString("asdf", "456")
	l.SetString("zxcv", "789")
	assert.Equal(t, 3, len(l.values))

	// delete a key
	l.DeleteValue("qwer")
	assert.Equal(t, 2, len(l.values))
	l.DeleteValue("qwer")
	assert.Equal(t, 2, len(l.values))
	l.DeleteValue("xxx")
	assert.Equal(t, 2, len(l.values))

	// clear
	l.Clear()
	assert.Equal(t, 0, len(l.values))

	// add some keys
	l.SetString("ppp", "999")
	l.SetString("qqq", "888")
	assert.Equal(t, 2, len(l.values))

	// make read-only
	l.LockReadOnly()

	// try to clear and remove
	assert.Panics(t, func() {
		l.Clear()
	})
	assert.Panics(t, func() {
		l.DeleteValue("ppp")
	})
	assert.Equal(t, 2, len(l.values))
}
