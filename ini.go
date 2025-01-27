package config

import (
	"bufio"
	"io"
	"strings"
)

func indexOfUnescapedEq(s string) int {
	// start at 1 (0 length key is not valid)
	st := 1
	s = s[1:]

	for {
		idx := strings.IndexByte(s, '=')
		if idx < 0 {
			return -1
		}
		if idx == 0 || s[idx-1] != '\\' {
			// found
			return st + idx
		}
		idx++
		s = s[idx:]
		st += idx
	}
}

var unescapes = strings.NewReplacer(
	`\:`, ":",
	`\;`, ";",
	`\=`, "=",
	`\r`, "\r",
	`\n`, "\n",
	`\t`, "\t",
	`\f`, "\f",
	`\0`, "\000",
	`\ `, " ",
	`\\`, `\`,
)

// Load INI config from reader and store the values in viewable (must be writable)
func LoadIni(viewable Viewable, reader *bufio.Reader) error {
	// check if viewable is writable
	if !viewable.IsWritable() {
		panic("cannot load ini config to read-only target")
	}

	var section = ""
	var line string
	var err error = nil
	for err != io.EOF {
		// read a line and handle error
		line, err = reader.ReadString('\n')
		if err != nil && err != io.EOF {
			// IO error
			return err
		}
		line, _ = strings.CutSuffix(line, "\n")
		line, _ = strings.CutSuffix(line, "\r") // also handle windows line encoding

		// check if line is empty or is a comment line
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		// check if line contains a section definition
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			newsection := line[1 : len(line)-1]
			if !strings.HasSuffix(newsection, ".") && newsection != "" {
				newsection += "."
			}
			section = newsection
			continue
		}

		// find the first non-escaped equal sign
		pos := indexOfUnescapedEq(line)
		if pos < 0 {
			// invalid line
			continue
		}

		// extract key and value
		key := line[:pos]
		val := line[pos+1:]

		// trim key and value
		key = strings.Trim(key, " \t\f")
		val = strings.TrimLeft(val, " \t\f")

		// unescape key and value and prepend key with the section
		key = section + unescapes.Replace(key)
		val = unescapes.Replace(val)

		// save value
		viewable.SetString(key, val)
	}
	return nil
}

func escapeIniString(s string, equal bool) string {
	// escape backslashes
	s = strings.ReplaceAll(s, "\\", "\\\\")

	// escape newlines
	s = strings.ReplaceAll(s, "\n", "\\n")

	// escape equal signs
	if equal {
		s = strings.ReplaceAll(s, "=", "\\=")
	}

	return s
}

func saveIniInternal(viewable Viewable, writer io.Writer, head bool) {
	if head {
		writer.Write([]byte(";\n; This INI file was autogenerated\n;\n\n"))
	}
	var keylist KeyList
	viewable.ListKeys("", &keylist, false)
	for _, key := range keylist.ToSlice() {
		val, ok := viewable.GetString(key)
		if ok {
			// likely
			writer.Write([]byte(escapeIniString(key, true)))
			writer.Write([]byte("="))
			writer.Write([]byte(escapeIniString(val, false)))
			writer.Write([]byte("\n"))
		}
	}
}

// Serialize viewable to writable in INI format
func SaveIni(viewable Viewable, writer io.Writer) {
	saveIniInternal(viewable, writer, true)
}
