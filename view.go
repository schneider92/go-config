package config

import (
	"strconv"
	"strings"
)

// Key list type where the Viewable.ListKeys function collects keys
type KeyList struct {
	v map[string]bool
}

func (list *KeyList) addKey(key string) {
	if list.v == nil {
		list.v = map[string]bool{}
	}
	list.v[key] = true
}

// Convert KeyList to a slice of strings
func (list *KeyList) ToSlice() []string {
	ret := make([]string, 0, len(list.v))
	for k := range list.v {
		ret = append(ret, k)
	}
	return ret
}

// Viewable config entity
type Viewable interface {
	// Test if the config entity is writable
	IsWritable() bool
	// Get raw string value for the given key. On fail, the second return value is false
	GetString(key string) (string, bool)
	// Set raw string value for the given key
	SetString(key, value string)
	// Delete value for the given key
	DeleteValue(key string)
	// List keys under the given prefix and collect them into out. If direct is
	// true, only the immediate direct subkeys are collected, otherwise all
	// nested children.
	//
	// For example, we have the following keys and calling ListKeys with prefix="test":
	// - test.key.1
	// - test.key.2
	// - test.value.3
	// If direct is true, ["key", "value"] is returned, otherwise ["key.1", "key.2", "value.3"]
	ListKeys(prefix string, out *KeyList, direct bool)
}

// View on a config entity
type View struct {
	prefix   string
	viewable Viewable
	writable bool
}

func newViewImpl(viewable Viewable, prefix string, writable bool) View {
	if !strings.HasSuffix(prefix, ".") && prefix != "" {
		prefix += "."
	}

	// test if viewable is a View, then we can refer to its viewable instead
	view, ok := viewable.(View)
	if ok {
		prefix = view.deriveKey(prefix)
		viewable = view.viewable
	}

	return View{
		prefix,
		viewable,
		writable,
	}
}

type emptyViewable struct{}

func (v emptyViewable) IsWritable() bool {
	return false
}

func (v emptyViewable) GetString(key string) (string, bool) {
	return "", false
}

func (v emptyViewable) SetString(key, value string) {
	panic("empty viewable is not writable")
}

func (v emptyViewable) DeleteValue(key string) {
	panic("empty viewable is not writable")
}

func (v emptyViewable) ListKeys(prefix string, out *KeyList, direct bool) {}

// Create an empty read-only view
func NewEmptyView() View {
	return newViewImpl(emptyViewable{}, "", false)
}

// Create view that wraps the given viewable. Prefix is prepended to all get/set
// calls, so that when prefix is "a.b" and calling GetString for "c.d", the
// received value of the wrapped viewable is the one at a.b.c.d
func NewView(viewable Viewable, prefix string) View {
	return newViewImpl(viewable, prefix, true)
}

// Test if the View is writable
func (view View) IsWritable() bool {
	return view.writable && view.viewable.IsWritable()
}

// Get raw string value for the given key
func (view View) GetString(key string) (value string, found bool) {
	return view.viewable.GetString(view.deriveKey(key))
}

// Set raw string value for the given key
func (view View) SetString(key, value string) {
	if !view.writable {
		panic("trying to write a read-only view")
	}
	view.viewable.SetString(view.deriveKey(key), value)
}

// Get int value for the given key. Reports not found either if value is not
// convertible to integer
func (view View) GetInt(key string) (result int64, found bool) {
	sv, ok := view.GetString(key)
	if ok {
		ret, err := strconv.ParseInt(sv, 0, 64)
		if err == nil {
			return ret, true
		}
	}
	return 0, false
}

// Set int value for the given key
func (view View) SetInt(key string, value int64) {
	view.SetString(key, strconv.FormatInt(value, 10))
}

// Get int value for the given key. Reports not found either if value is not
// convertible to bool
func (view View) GetBool(key string) (result bool, found bool) {
	sv, ok := view.GetString(key)
	if ok {
		ret, err := strconv.ParseBool(sv)
		if err == nil {
			return ret, true
		}
	}
	return false, false
}

// Set bool value for the given key
func (view View) SetBool(key string, value bool) {
	sval := "false"
	if value {
		sval = "true"
	}
	view.SetString(key, sval)
}

// Delete value for the given key
func (view View) DeleteValue(key string) {
	if !view.writable {
		panic("trying to delete from a read-only view")
	}
	view.viewable.DeleteValue(view.deriveKey(key))
}

// List keys, see Viewable for details
func (view View) ListKeys(prefix string, out *KeyList, direct bool) {
	view.viewable.ListKeys(view.deriveKey(prefix), out, direct)
}

func (view View) deriveKey(key string) string {
	return view.prefix + key
}

// Create a writable subview
func (view View) SubView(prefix string) View {
	return newViewImpl(view, prefix, true)
}

// Create a read-only subview
func (view View) SubViewReadOnly(prefix string) View {
	return newViewImpl(view, prefix, false)
}
