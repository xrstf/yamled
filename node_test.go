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
list: [1, 2, 3]
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

	if m := node.ToMap(); len(m) != 1 {
		t.Fatalf("ToMap() should have returned a map with 1 element, but has %d.", len(m))
	}

	if doc.MustGet("nonexisting").ToString() != "" {
		t.Fatal("Expected to be able to MustGet() a dummy item but failed.")
	}

	if s := doc.MustGet("list").ToSlice(); len(s) != 3 {
		t.Fatalf("ToSlice() should have returned a slice with 3 elements, but has %d.", len(s))
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

func TestNodeAllowSetAnyNodeToNull(t *testing.T) {
	input := strings.TrimSpace(`
foo: bar
hello: world
list: [1, 2, 3]
obj: {key: value}
`)

	node, doc, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	if _, err := doc.SetKey("foo", nil); err != nil {
		t.Errorf("Failed to set scalar key to null: %v", err)
	}

	if _, err := doc.SetKey("list", nil); err != nil {
		t.Errorf("Failed to set sequence key to null: %v", err)
	}

	if _, err := doc.SetKey("obj", nil); err != nil {
		t.Errorf("Failed to set mapping key to null: %v", err)
	}

	expectYAML(t, node, `
foo: null
hello: world
list: null
obj: null
`)
}

func TestNodeAllowSetNilNodesToAnything(t *testing.T) {
	input := strings.TrimSpace(`
str: null
list: null
obj: null
`)

	node, doc, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	if _, err := doc.SetKey("str", "foo"); err != nil {
		t.Errorf("Failed to set null node to scalar: %v", err)
	}

	if _, err := doc.SetKey("list", []int{1, 2, 3}); err != nil {
		t.Errorf("Failed to set null node to sequence: %v", err)
	}

	if _, err := doc.SetKey("obj", map[string]int{"foo": 1}); err != nil {
		t.Errorf("Failed to set null node to mapping: %v", err)
	}

	expectYAML(t, node, `
str: foo
list:
  - 1
  - 2
  - 3
obj:
  foo: 1
`)
}

func TestNodeForbidKindChange(t *testing.T) {
	input := strings.TrimSpace(`
foo: bar
hello: world
list: [1, 2, 3]
obj: {key: value}
`)

	node, doc, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	if _, err := doc.SetKey("foo", []string{"foo", "bar"}); err == nil {
		t.Error("Succeeded in changing scalar to sequence.")
	}

	if _, err := doc.SetKey("list", "foo"); err == nil {
		t.Error("Succeeded in changing sequence to scalar.")
	}

	if _, err := doc.SetKey("obj", "foo"); err == nil {
		t.Error("Succeeded in changing mapping to scalar.")
	}

	if _, err := doc.SetKey("list", []string{"foo", "bar"}); err != nil {
		t.Errorf("Failed to replace list with list: %v", err)
	}

	expectYAML(t, node, `
foo: bar
hello: world
list:
  - foo
  - bar
obj: {key: value}
`)
}

func TestNodeAllowReplacingKinds(t *testing.T) {
	input := strings.TrimSpace(`
str: foo
list: [1, 2, 3]
obj: {key: value}
`)

	node, doc, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	if _, err := doc.ReplaceKey("str", []string{"foo", "bar"}); err != nil {
		t.Errorf("Failed to replace scalar node with sequence: %v", err)
	}

	if _, err := doc.ReplaceKey("list", "foo"); err != nil {
		t.Errorf("Failed to replace sequence node with scalar: %v", err)
	}

	if _, err := doc.ReplaceKey("obj", "foo"); err != nil {
		t.Errorf("Failed to replace mapping node with scalar: %v", err)
	}

	if err := doc.MustGet("obj").Replace([]int{1, 2}); err != nil {
		t.Errorf("Failed to replace new scalar node with sequence: %v", err)
	}

	expectYAML(t, node, `
str:
  - foo
  - bar
list: foo
obj:
  - 1
  - 2
`)
}

func TestNodeComments(t *testing.T) {
	input := strings.TrimSpace(`
foo:
  # this is a comment
  hello: world
`)

	node, doc, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	// As this sets the comments on the _value node_, the
	// result is not what you might expect. See the KeyNode
	// tests.
	doc.MustGet("foo", "hello").
		SetHeadComment("new head comment").
		SetLineComment("new line comment").
		SetFootComment("new foot comment")

	expectYAML(t, node, `
foo:
  # this is a comment
  hello: world # new line comment
  # new foot comment

  # new head comment
`)
}
