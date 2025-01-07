package config

import (
	"strings"
	"sync"
)

// Config layer storing key-value pairs in memory
type Layer struct {
	mx       sync.Mutex
	name     string
	values   map[string]string
	writable bool
}

// Create a layer with the given name
func NewLayer(name string) *Layer {
	return &Layer{
		name:     name,
		values:   map[string]string{},
		writable: true,
	}
}

// Get the layer name
func (l *Layer) Name() string {
	return l.name
}

// Make the layer read-only. Call once after loading values. After called, the
// layer is not writable anymore
func (l *Layer) LockReadOnly() {
	// lock mutex
	l.mx.Lock()
	defer l.mx.Unlock()

	l.writable = false
}

// Check if the layer is writable
func (l *Layer) IsWritable() bool {
	// lock mutex
	l.mx.Lock()
	defer l.mx.Unlock()

	return l.writable
}

// Get raw string value for the given key
func (l *Layer) GetString(key string) (value string, found bool) {
	// lock mutex
	l.mx.Lock()
	defer l.mx.Unlock()

	ret, found := l.values[key]
	return ret, found
}

// Delete all values from the layer
func (l *Layer) Clear() {
	// lock mutex
	l.mx.Lock()
	defer l.mx.Unlock()

	if !l.writable {
		panic("trying to clear a read-only layer")
	}
	l.values = map[string]string{}
}

// Set raw string value for the given key
func (l *Layer) SetString(key, value string) {
	// lock mutex
	l.mx.Lock()
	defer l.mx.Unlock()

	if !l.writable {
		panic("trying to write a read-only layer")
	}
	l.values[key] = value
}

// Delete value for the given key
func (l *Layer) DeleteValue(key string) {
	// lock mutex
	l.mx.Lock()
	defer l.mx.Unlock()

	if !l.writable {
		panic("trying to delete from read-only layer")
	}
	delete(l.values, key)
}

// List keys, see Viewable for detail
func (l *Layer) ListKeys(prefix string, out *KeyList, direct bool) {
	// ensure trailing dot
	if !strings.HasSuffix(prefix, ".") && prefix != "" {
		prefix += "."
	}
	prefixlen := len(prefix)

	// lock mutex
	l.mx.Lock()
	defer l.mx.Unlock()

	// go through keys
	for k := range l.values {
		if strings.HasPrefix(k, prefix) {
			k = k[prefixlen:]
			if direct {
				idx := strings.IndexRune(k, '.')
				if idx >= 0 {
					k = k[:idx]
				}
			}
			out.addKey(k)
		}
	}
}
