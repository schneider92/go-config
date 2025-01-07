package config

import (
	"slices"
	"sync"
)

type configItem struct {
	layer    *Layer
	prio     int
	writable bool
}

// Collection of config layers with priorities
type Config struct {
	mx    sync.Mutex
	items []configItem
}

// Create a config with no layers
func NewConfig() *Config {
	return &Config{
		items: []configItem{},
	}
}

func (c *Config) addLayerImpl(layer *Layer, prio int, writable bool) {
	// lock layer list
	c.mx.Lock()
	defer c.mx.Unlock()

	// find slot for layer
	i := 0
	max := len(c.items)
	for i < max {
		if c.items[i].prio < prio {
			break
		}
		i++
	}

	// insert layer to list
	c.items = slices.Insert(c.items, i, configItem{
		layer,
		prio,
		writable,
	})
}

// Add a read only layer to the config
func (c *Config) AddLayer(layer *Layer, prio int) {
	c.addLayerImpl(layer, prio, false)
}

// Add a writable layer to the config. Adding it this way allows you to call
// function Config.SetString only if the layer itself is also writable
func (c *Config) AddWritableLayer(layer *Layer, prio int) {
	c.addLayerImpl(layer, prio, true)
}

// Remove a layer from the config, return true if such layer exists
func (c *Config) RemoveLayer(layer *Layer) bool {
	// lock layer list
	c.mx.Lock()
	defer c.mx.Unlock()

	// find layer
	idx := slices.IndexFunc(c.items, func(item configItem) bool {
		return item.layer == layer
	})

	// delete layer if found
	if idx < 0 {
		return false
	} else {
		c.items = slices.Delete(c.items, idx, idx+1)
		return true
	}
}

// Get a list of the layers the config currently contains
func (c *Config) Layers() []*Layer {
	// lock layer list
	c.mx.Lock()
	defer c.mx.Unlock()

	// create return list and fill it
	ret := make([]*Layer, 0, len(c.items))
	for _, item := range c.items {
		ret = append(ret, item.layer)
	}
	return ret
}

// Test if the config is writable
func (c *Config) IsWritable() bool {
	// lock layer list
	c.mx.Lock()
	defer c.mx.Unlock()

	// find writable layer
	for _, item := range c.items {
		if item.writable && item.layer.IsWritable() {
			return true
		}
	}
	return false
}

// Get raw string value for the given key
func (c *Config) GetString(key string) (string, bool) {
	// lock layer list
	c.mx.Lock()
	defer c.mx.Unlock()

	// find in all layers
	for _, item := range c.items {
		s, ok := item.layer.GetString(key)
		if ok {
			return s, true
		}
	}

	// not found
	return "", false
}

// Set raw string value for the given key
func (c *Config) SetString(key, value string) {
	// lock layer list
	c.mx.Lock()
	defer c.mx.Unlock()

	// find writable layer
	for _, item := range c.items {
		if item.writable && item.layer.IsWritable() {
			item.layer.SetString(key, value)
			return
		}
	}

	// no writable layer, set not possible
	panic("config is not writable")
}

// Only supplied for interface compatibility
func (c *Config) DeleteValue(key string) {} // not implemented in Config

// List keys, see Viewable for details
func (c *Config) ListKeys(prefix string, out *KeyList, direct bool) {
	// lock layer list
	c.mx.Lock()
	defer c.mx.Unlock()

	// list in all layers
	for _, item := range c.items {
		item.layer.ListKeys(prefix, out, direct)
	}
}
