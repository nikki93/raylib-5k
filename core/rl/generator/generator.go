// Generates ../generated.gx.go from Raylib's raylib_api.json

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"unicode"
)

func typeToGo(typ string) string {
	if typ == "const char *" { // Strings
		return "string"
	}
	ptrs := 0
	for typ[len(typ)-1] == '*' { // Count and remove `*`s
		typ = typ[:len(typ)-1]
		ptrs++
	}
	for typ[len(typ)-1] == ' ' { // Remove trailing spaces
		typ = typ[:len(typ)-1]
	}
	if strings.HasPrefix(typ, "const ") { // Remove const
		typ = typ[6:]
	}
	switch typ { // Substitute basic types
	case "float":
		typ = "float64"
	case "double":
		typ = "float64"
	case "unsigned int":
		typ = "uint"
	case "unsigned short":
		typ = "uint16"
	case "long":
		typ = "int64"
	case "char":
		typ = "byte"
	case "unsigned char":
		typ = "byte"
	case "void":
		typ = "byte"
	case "Vector2":
		typ = "Vec2"
	case "Vector3":
		typ = "Vec3"
	case "Vector4":
		typ = "Vec4"
	case "Texture2D":
		typ = "Texture"
	case "rAudioBuffer": // Skipped
		return ""
	}
	for i := 0; i < ptrs; i++ { // Put `*`s at front
		typ = "*" + typ
	}
	return typ
}

type Param struct {
	Name string
	Type string
}

var arrayRe = regexp.MustCompile(`^(\w*)\[(.*)\]$`)

func (p Param) toGo() string {
	typ := typeToGo(p.Type)
	if typ == "" {
		return ""
	}
	name := p.Name
	if name == "type" { // `type` is a keyword
		name = "typ"
	}
	arrayMatch := arrayRe.FindStringSubmatch(name) // Put array size at front
	if len(arrayMatch) > 0 {
		name = arrayMatch[1]
		typ = "[" + arrayMatch[2] + "]" + typ
	}
	return name + " " + typ
}

type Parsed struct {
	Structs []struct {
		Name        string
		Description string
		Fields      []struct {
			Name        string
			Type        string
			Description string
		}
	}
	Enums []struct {
		Name        string
		Description string
		Values      []struct {
			Name        string
			Value       int
			Description string
		}
	}
	Functions []struct {
		Name        string
		Description string
		ReturnType  string
		Params      []Param
	}
}

func main() {
	// Read JSON
	read, err := ioutil.ReadFile("vendor/raylib/parser/raylib_api.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	parsed := &Parsed{}
	err = json.Unmarshal(read, parsed)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Start buffer
	buf := &bytes.Buffer{}

	// Write header comment and `//gx:externs` directive
	fmt.Fprintf(buf, "// Generated by ./generator/generator.go\n//gx:externs rl::\n\n")

	// Write package name
	fmt.Fprintf(buf, "package rl\n\n")

	// Write imports
	fmt.Fprintf(buf, "import (\n\t. \"github.com/nikki93/raylib-5k/core/geom\"\n)\n\n")

	// Write structs
	for _, s := range parsed.Structs {
		switch s.Name {
		case "Vector2", "Vector3", "Vector4", "Matrix":
			continue
		}
		if s.Description != "" {
			fmt.Fprintf(buf, "// %s\n", s.Description)
		}
		fmt.Fprintf(buf, "type %s struct {\n", s.Name)
		for _, field := range s.Fields {
			name := []rune(field.Name)
			name[0] = unicode.ToUpper(name[0])
			f := Param{Name: string(name), Type: field.Type}.toGo()
			if f == "" {
				continue
			}
			fmt.Fprintf(buf, "\t%s", f)
			if field.Description != "" {
				fmt.Fprintf(buf, " // %s", field.Description)
			}
			fmt.Fprintf(buf, "\n")
		}
		fmt.Fprintf(buf, "}\n\n")
	}

	// Write enums
	for _, e := range parsed.Enums {
		if e.Description != "" {
			fmt.Fprintf(buf, "// %s\n", e.Description)
		}
		fmt.Fprintf(buf, "const (\n")
		for _, value := range e.Values {
			fmt.Fprintf(buf, "\t%s = %d\n", value.Name, value.Value)
		}
		fmt.Fprintf(buf, ")\n\n")
	}

	// Write functions
	for _, f := range parsed.Functions {
		if strings.HasSuffix(f.Name, "Callback") {
			continue
		}
		if f.Description != "" {
			fmt.Fprintf(buf, "// %s\n", f.Description)
		}
		fmt.Fprintf(buf, "func %s(", f.Name)
		for i, param := range f.Params {
			if i > 0 {
				fmt.Fprintf(buf, ", ")
			}
			if param.Name == "" || param.Type == "" {
				fmt.Fprint(buf, "args ...interface{}")
				break
			}
			if f.Name == "SetShaderValue" && param.Name == "value" {
				fmt.Fprintf(buf, "value *float64")
			} else {
				fmt.Fprint(buf, param.toGo())
			}
		}
		fmt.Fprintf(buf, ")")
		if f.ReturnType != "" && f.ReturnType != "void" {
			fmt.Fprintf(buf, typeToGo(f.ReturnType))
		}
		fmt.Fprintf(buf, "\n\n")
	}

	// Format buffer
	unformatted := buf.Bytes()
	formatted, err := format.Source(unformatted)
	if err != nil {
		fmt.Println(err)
		formatted = unformatted
	}

	// Write buffer to file
	file, err := os.Create("core/rl/generated.gx.go")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	file.Write(formatted)
}