package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	conf := NewConfig()
	conf.items = []configItem{}

	// create and add a layer
	l2 := NewLayer("l2")
	conf.AddLayer(l2, 22)
	l2.SetString("my.dog", "vegetables")
	l2.SetString("my.cat", "mouse")

	// create and add one more layer
	l3 := NewLayer("l3")
	conf.AddLayer(l3, 33)
	l3.SetString("my.dog", "bones")
	l3.SetString("my.rabbit", "carrot")
	l3.SetString("your.cat", "-")

	// create and add one more layer
	l1 := NewLayer("l1")
	conf.AddLayer(l1, 11)
	l1.SetString("my.cat", "fish")
	l1.SetString("my.horse", "hay")
	l1.SetString("his.cat", "-")

	// try to get values from the config
	s, ok := conf.GetString("my.dog")
	assert.True(t, ok)
	assert.Equal(t, "bones", s)

	s, ok = conf.GetString("my.cat")
	assert.True(t, ok)
	assert.Equal(t, "mouse", s)

	s, ok = conf.GetString("my.rabbit")
	assert.True(t, ok)
	assert.Equal(t, "carrot", s)

	s, ok = conf.GetString("my.horse")
	assert.True(t, ok)
	assert.Equal(t, "hay", s)

	s, ok = conf.GetString("my.duck")
	assert.False(t, ok)
	assert.Equal(t, "", s)

	// list keys
	keys := listKeys(conf, "", true)
	assert.ElementsMatch(t, keys, []string{"my", "your", "his"})

	keys = listKeys(conf, "", false)
	assert.ElementsMatch(t, keys, []string{"my.dog", "my.cat", "my.rabbit", "my.horse", "your.cat", "his.cat"})

	keys = listKeys(conf, "my", false)
	assert.ElementsMatch(t, keys, []string{"dog", "cat", "rabbit", "horse"})

	keys = listKeys(conf, "noexist", false)
	assert.ElementsMatch(t, keys, []string{})

	// try to set
	assert.False(t, conf.IsWritable())
	assert.Panics(t, func() {
		conf.SetString("fails", "123")
	})

	// list layers
	layers := conf.Layers()
	assert.Equal(t, 3, len(layers))
	assert.Equal(t, "l3", layers[0].Name())
	assert.Equal(t, "l2", layers[1].Name())
	assert.Equal(t, "l1", layers[2].Name())

	// remove layer 3
	assert.True(t, conf.RemoveLayer(l3))
	layers = conf.Layers()
	assert.Equal(t, 2, len(layers))

	assert.False(t, conf.RemoveLayer(l3))
	layers = conf.Layers()
	assert.Equal(t, 2, len(layers))

	// test if values changed accordingly
	s, ok = conf.GetString("my.dog") // changed
	assert.True(t, ok)
	assert.Equal(t, "vegetables", s)

	s, ok = conf.GetString("my.cat")
	assert.True(t, ok)
	assert.Equal(t, "mouse", s)

	s, ok = conf.GetString("my.rabbit") // changed
	assert.False(t, ok)
	assert.Equal(t, "", s)

	s, ok = conf.GetString("my.horse")
	assert.True(t, ok)
	assert.Equal(t, "hay", s)

	// add writable layer
	l9 := NewLayer("l9")
	conf.AddWritableLayer(l9, 99)

	// write value and read it back
	assert.True(t, conf.IsWritable())
	conf.SetString("test", "asdf")
	s, ok = conf.GetString("test")
	assert.True(t, ok)
	assert.Equal(t, "asdf", s)

	// remove writable layer and read this new value back again
	conf.RemoveLayer(l9)
	s, ok = conf.GetString("test")
	assert.False(t, ok)
	assert.Equal(t, "", s)
}
