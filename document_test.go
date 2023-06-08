// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package yamled

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func yamlLoad(input string) (*yaml.Node, Document, error) {
	var node yaml.Node
	decoder := yaml.NewDecoder(strings.NewReader(input))

	if err := decoder.Decode(&node); err != nil {
		return nil, nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	document, err := NewDocument(&node)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create document: %w", err)
	}

	return &node, document, nil
}

func yamlEncode(t *testing.T, v interface{}) string {
	var buf bytes.Buffer

	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(v); err != nil {
		t.Fatalf("Failed to encode data structure as YAML: %v", err)
	}

	return buf.String()
}

func expectYAML(t *testing.T, v interface{}, expectedYAML string) {
	encoded := yamlEncode(t, v)
	encoded = strings.TrimSpace(encoded)

	expectedYAML = strings.TrimSpace(expectedYAML)

	if encoded != expectedYAML {
		t.Fatalf("Expected\n---\n%s\n---\n\nbut got\n\n---\n%s\n---", expectedYAML, encoded)
	}
}

func TestIdempotent(t *testing.T) {
	input := strings.TrimSpace(`
# hello world

foo: bar
list: [1, 2, 3]
`)

	node, _, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	expectYAML(t, node, input)
}

func TestSetKey(t *testing.T) {
	input := strings.TrimSpace(`
foo:
  hello: world
`)

	node, document, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	document.SetAt(Path{"foo", "hello"}, 12)

	// list, found := document.Get("foo")
	// if !found {
	// 	t.Fatalf("Failed to get list")
	// }

	// if err := list.Set(Path{"hello"}, 12); err != nil {
	// 	t.Fatalf("Failed to set foo.hello = 12: %v", err)
	// }

	expectYAML(t, node, `
foo:
  hello: 12
	`)
}

func TestAddNewKey(t *testing.T) {
	input := strings.TrimSpace(`
foo:
  hello: world
`)

	node, document, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	document.SetAt(Path{"foo", "newhello"}, 12)

	// list, found := document.Get("foo")
	// if !found {
	// 	t.Fatalf("Failed to get list")
	// }

	// if err := list.Set(Path{"newhello"}, 12); err != nil {
	// 	t.Fatalf("Failed to set foo.newhello = 12: %v", err)
	// }

	expectYAML(t, node, `
foo:
  hello: world
  newhello: 12
	`)
}

func TestAddMultipleNewKey(t *testing.T) {
	input := strings.TrimSpace(`
foo:
  hello: world
`)

	node, document, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	document.SetAt(Path{"foo", "newhello"}, 12)
	document.SetAt(Path{"foo", "newhello2"}, 13)

	expectYAML(t, node, `
foo:
  hello: world
  newhello: 12
  newhello2: 13
	`)
}

func TestSetListItem(t *testing.T) {
	input := strings.TrimSpace(`
foo:
  hello: [1, 2, 3]
`)

	node, document, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	document.SetAt(Path{"foo", "hello", 0}, 7)

	// list, found := document.Get("foo")
	// if !found {
	// 	t.Fatalf("Failed to get list")
	// }

	// if err := list.Set(Path{"hello", 0}, 7); err != nil {
	// 	t.Fatalf("Failed to set foo.newhello = 12: %v", err)
	// }

	expectYAML(t, node, `
foo:
  hello: [7, 2, 3]
	`)
}

func TestSetRandomListItem(t *testing.T) {
	input := strings.TrimSpace(`
foo:
  hello: [1, 2, 3]
`)

	node, document, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	document.SetAt(Path{"foo", "hello", 1}, 7)

	// list, found := document.Get("foo")
	// if !found {
	// 	t.Fatalf("Failed to get list")
	// }

	// if err := list.Set(Path{"hello", 1}, 7); err != nil {
	// 	t.Fatalf("Failed to set foo.newhello = 12: %v", err)
	// }

	expectYAML(t, node, `
foo:
  hello: [1, 7, 3]
	`)
}
