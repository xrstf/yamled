// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"go.xrstf.de/yamled"
	"gopkg.in/yaml.v3"
)

var (
	input = strings.TrimSpace(`
# hello world

bla:
  - fasesl: huhu
    mooh: kuna
    time: for den persjoladatamaskinen

run:
  modules-download-mode: readonly
  skip-files:
    - zz_generated.*.go

linters:
  enable:
    - tagliatelle
  disable-all: true
`)
)

func main() {
	var node yaml.Node

	decoder := yaml.NewDecoder(strings.NewReader(input))

	if err := decoder.Decode(&node); err != nil {
		log.Fatalf("Failed to decode YAML: %v", err)
	}

	fmt.Printf("[--- raw input document ---]\n\n")
	fmt.Println(input)

	fmt.Printf("\n[--- parsed input document ---]\n\n")
	PrintYAMLNode(node)

	fmt.Printf("\n[--- encode(yamlNode) ---]\n\n")
	fmt.Println(encode(&node))

	document, err := yamled.NewDocument(&node)
	if err != nil {
		log.Fatalf("Failed to create document: %v", err)
	}

	fmt.Printf("\n[--- document.Bytes(2) ---]\n\n")
	encoded, err := document.Bytes(2)
	if err != nil {
		log.Fatalf("Failed to encode document: %v", err)
	}
	fmt.Println(string(encoded))

	bla, err := document.Get("linters")
	if err != nil {
		log.Fatalf("Failed to get `bla` node: %+v (%v)", err, yamled.IsNotFound(err))
	}

	fmt.Printf("\n[--- encode(blaNode) ---]\n\n")
	fmt.Println(encode(bla))

	fmt.Printf("\n[--- encode(blaNode.foo = true) ---]\n\n")
	fmt.Println(bla.Set(yamled.Path{"foo"}, struct{}{}))
	bla2, err := bla.Get("foo")
	fmt.Println(bla2.Set(yamled.Path{"bla", 0, "muh"}, true))
	PrintYAMLNode(node)

	fmt.Println(encode(bla))

	// bla0, err := bla.Get(0)
	// if err != nil {
	// 	log.Fatalf("Failed to get `bla.0` node: %+v (%v)", err, yamled.IsNotFound(err))
	// }

	// fmt.Printf("\n[--- encode(bla0Node) ---]\n\n")
	// fmt.Println(encode(bla0))

	// bla0mooh, err := bla0.Get("mooh")
	// if err != nil {
	// 	log.Fatalf("Failed to get `bla.0.mooh` node: %+v (%v)", err, yamled.IsNotFound(err))
	// }

	// fmt.Printf("\n[--- encode(bla0moohNode) ---]\n\n")
	// fmt.Println(encode(bla0mooh))

	// foo, _ := betterNode.Get(yamled.Path{})
	// PrintYAMLNode(*foo.GetNode())
	// fmt.Println(encode(foo))
}

func encode(v interface{}) string {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(v); err != nil {
		panic(err)
	}

	return buf.String()
}

func PrintYAMLNode(node yaml.Node) {
	printYAMLNode(0, node)
}

func printYAMLNode(depth int, node yaml.Node) {
	prefix := strings.Repeat("  ", depth)
	// fmt.Printf("%sDEBUG: node: %+v\n", prefix, node)

	kindName := yamled.KindName(node.Kind)
	styleName := yamled.StyleName(node.Style)

	switch node.Kind {
	case yaml.DocumentNode:
		fmt.Printf("%s%s(%s)\n", prefix, kindName, styleName)
	case yaml.SequenceNode:
		fmt.Printf("%s%s(%s)\n", prefix, kindName, styleName)
	case yaml.MappingNode:
		fmt.Printf("%s%s(%s)\n", prefix, kindName, styleName)
	case yaml.ScalarNode:
		fmt.Printf("%s%s(%s, %s): %#v\n", prefix, kindName, styleName, node.Tag, node.Value)
	case yaml.AliasNode:
		fmt.Printf("%s%s(%s)\n", prefix, kindName, styleName)
	}

	if node.HeadComment != "" {
		fmt.Printf("%s HeadCmt: %q\n", prefix, node.HeadComment)
	}

	if node.LineComment != "" {
		fmt.Printf("%s LineCmt: %q\n", prefix, node.LineComment)
	}

	if node.FootComment != "" {
		fmt.Printf("%s FootCmt: %q\n", prefix, node.FootComment)
	}

	for _, child := range node.Content {
		printYAMLNode(depth+1, *child)
	}
}
