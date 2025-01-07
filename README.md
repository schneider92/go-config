# go-config

Simple config library written in golang, supporting multiple config layers and reading/writing .ini files

## Example

```go
file, _ := os.Open("filepath")
inireader := bufio.NewReader(file)
var err error

// create a layer for default values
defaultlayer := config.NewLayer("defaults")
httpview := config.NewView(defaultlayer, "http")
httpview.SetInt("server.port", 8080)
httpview.SetString("server.servername", "MyTestServer v0.9")
httpview.SetBool("server.multiconnections", false)

// make the layer read-only and add it to config with a low priority
defaultlayer.LockReadOnly()
conf := config.NewConfig()
conf.AddLayer(defaultlayer, 0)

// create a layer loaded from ini files
inilayer := config.NewLayer("ini")
err = config.LoadIni(inilayer, inireader) // handle error

// add this layer to the config with a higher priority
conf.AddLayer(inilayer, 10)

// get values from the config
view := config.NewView(conf, "http.server")
port, ok := view.GetInt("port")
// ok is true here and port contains the value defined in the ini file, or if not defined, 8080
```
