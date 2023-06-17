// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package yamled

import (
	"strings"
	"testing"
)

func TestKeyNodeComments(t *testing.T) {
	input := strings.TrimSpace(`
foo:
  # this is a comment
  hello: world
`)

	node, doc, err := yamlLoad(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	keyNode, ok := doc.GetKey("foo", "hello")
	if !ok {
		t.Fatal("Failed to get key node")
	}

	keyNode.
		SetHeadComment("new head comment").
		SetLineComment("new line comment").
		SetFootComment("new foot comment")

	expectYAML(t, node, `
foo:
  # new head comment
  hello: world # new line comment
  # new foot comment
`)
}
