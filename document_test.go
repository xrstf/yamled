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
