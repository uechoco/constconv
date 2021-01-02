package main

import (
	"bytes"
	"fmt"
	"go/format"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/stoewer/go-strcase"
)

type Generator struct {
	c *Config            // The configure
	t *template.Template // The template instance which set the functions and parsed files
}

func NewGenerator(c *Config) *Generator {
	return &Generator{
		c: c,
	}
}

// LoadTemplate loads the template file.
func (g *Generator) LoadTemplate() error {
	templateFile := g.c.templateFile
	basename := filepath.Base(templateFile)
	t, err := template.New(basename).Funcs(g.funcMap()).ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("can't parse template file %s. err:%w", templateFile, err)
	}
	g.t = t
	return nil
}

func (g *Generator) Generate(p *Parser) ([]byte, error) {
	valuesList := p.ValuesList()

	templateData := map[string]interface{}{
		"_doNotEdit":       fmt.Sprintf("// DO NOT EDIT.; Code generated by \"constconv %s\"", g.c.execArgsStr),
		"_extra":           g.c.extraData,
		"basePackageName":  p.BasePackageName(),
		"values":           valuesList[0], // syntax sugar for first valuesList element.
		"valuesList":       valuesList,
	}

	var buf bytes.Buffer
	if err := g.t.Execute(&buf, templateData); err != nil {
		return nil, fmt.Errorf("template execution failed: %w", err)
	}

	return g.format(buf)
}

// format returns the gofmt-ed contents of the Generator's buffer.
func (g *Generator) format(buf bytes.Buffer) ([]byte, error) {
	body := buf.Bytes()
	src, err := format.Source(body)
	if err != nil {
		return body, fmt.Errorf("internal error. format failed: %w", err)
	}
	return src, nil
}

// funcMap returns the template function map.
func (g *Generator) funcMap() template.FuncMap {
	return template.FuncMap{
		"SnakeCase":      strcase.SnakeCase,
		"KebabCase":      strcase.KebabCase,
		"LowerCamelCase": strcase.LowerCamelCase,
		"UpperSnakeCase": strcase.UpperSnakeCase,
		"UpperKebabCase": strcase.UpperKebabCase,
		"UpperCamelCase": strcase.UpperCamelCase,
		"HasPrefix":      strings.HasPrefix,
		"HasSuffix":      strings.HasSuffix,
		"Contains":       strings.Contains,
		"Title":          strings.Title,
		"ToLower":        strings.ToLower,
		"ToUpper":        strings.ToUpper,
		"TrimSpace":      strings.TrimSpace,
		"TrimPrefix":     strings.TrimPrefix,
		"TrimSuffix":     strings.TrimSuffix,
		"Trim":           strings.Trim,
		"TrimLeft":       strings.TrimLeft,
		"TrimRight":      strings.TrimRight,
		"Quote":          strconv.Quote,
		"Unquote": func(s string) string {
			ret, err := strconv.Unquote(s)
			if err != nil {
				return s
			}
			return ret
		},
	}
}
