// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package yamled

import (
	"strings"
	"testing"
)

func TestNodeGetObjectKey(t *testing.T) {
	input := strings.TrimSpace(`
string: bar
number: 12
object:
  key: value
`)

	_, doc, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	node, ok := doc.Get("string")
	if !ok {
		t.Fatal("Expected to find string key, but did not.")
	}

	if node.ToString() != "bar" {
		t.Fatalf("Expected string node to have value \"bar\", but got %q.", node.ToString())
	}

	node, ok = doc.Get("number")
	if !ok {
		t.Fatal("Expected to find number key, but did not.")
	}

	if node.ToInt() != 12 {
		t.Fatalf("Expected number node to have value 12, but got %d.", node.ToInt())
	}

	node, ok = doc.Get("object")
	if !ok {
		t.Fatal("Expected to find object key, but did not.")
	}

	type dummyStruct struct {
		Key string `yaml:"key"`
	}

	var i dummyStruct
	if err := node.To(&i); err != nil {
		t.Fatalf("Should have been able to cast node to int, but: %v", err)
	}

	if doc.MustGet("nonexisting").ToString() != "" {
		t.Fatal("Expected to be able to MustGet() a dummy item but failed.")
	}
}

func TestNodeGetArrayItem(t *testing.T) {
	input := strings.TrimSpace(`
list: [1, foo, 2]
`)

	_, doc, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	list, ok := doc.Get("list")
	if !ok {
		t.Fatal("Expected to find list key, but did not.")
	}

	first, ok := list.Get(0)
	if !ok {
		t.Fatal("Expected to get first list item, but found none.")
	}

	if first.ToInt() != 1 {
		t.Fatalf("Expected first item in the list to be 1, but got %d.", first.ToInt())
	}

	if list.MustGet(1).ToString() != "foo" {
		t.Fatalf("Expected first item in the list to have value \"foo\", but got %q.", list.MustGet(1).ToString())
	}

	if list.MustGet(3).ToString() != "" {
		t.Fatal("Expected to be able to MustGet() a dummy item beyond the list capacity, but failed.")
	}
}

func TestNodeGetDeep(t *testing.T) {
	input := strings.TrimSpace(`
foo:
  bar:
    - hello
    - key: value
      anotherkey:
        - first
        - second
        - hello: world
`)

	_, doc, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	item, ok := doc.Get("foo", "bar", 1, "anotherkey", 1)
	if !ok {
		t.Fatal("Expected to find list key, but did not.")
	}

	if item.ToString() != "second" {
		t.Fatal("Expected to be able to get the second item, but got not.")
	}
}

func TestNodeScalarCannotGetKeys(t *testing.T) {
	input := strings.TrimSpace(`"hello world"`)

	_, doc, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	_, ok := doc.Get("list")
	if ok {
		t.Fatal("Should not have been able to get a key subitem from a scalar value.")
	}

	dummy := doc.MustGet("list")
	if dummy == nil {
		t.Fatal("MustGet should still return a usable, empty node.")
	}

	input = strings.TrimSpace(`
foo: "bar"
`)

	_, doc, err = yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load second YAML: %v", err)
	}

	_, ok = doc.Get("foo", "sub", "bar")
	if ok {
		t.Fatal("Should not have been able to get a subkey subitem from a scalar value.")
	}
}

func TestNodeDeleteKey(t *testing.T) {
	input := strings.TrimSpace(`
foo:
  bar:
    - hello
    - key: value
      anotherkey:
        - first
        - second
        - hello: world
`)

	_, doc, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	if err := doc.DeleteKey("foo", "bar", 1, "anotherkey", 1); err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}
}
