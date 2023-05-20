// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package yamled

import (
	"bytes"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Document interface {
	// Documents *cannot* be marshalled by just throwing them into
	// a YAML encoder. The yaml.v3 encoder is hard-wired to expect
	// a specific type for the document node and since yaml.Document
	// is a wrapper, yaml.v3 will ignore the HeadComment as it is
	// not recognized as a document.
	// Document implements this interface to issue a panic whenever
	// an accidental attempt is made to encode it.
	// Use Bytes() or Encode() instead.
	// This limitation does not apply to yamled.Node objects.
	yaml.Marshaler

	Bytes(indent int) ([]byte, error)
	Encode(encoder *yaml.Encoder) error

	Get(path ...Step) (Node, error)

	// debugging only
	GetNode() *yaml.Node
}

type document struct {
	node *yaml.Node
}

func NewDocument(n *yaml.Node) (Document, error) {
	if n == nil {
		return nil, errors.New("node cannot be nil")
	}

	if n.Kind != yaml.DocumentNode {
		return nil, fmt.Errorf("expected document node, but got %v", KindName(n.Kind))
	}

	return &document{
		node: n,
	}, nil
}

func (d *document) GetNode() *yaml.Node {
	return d.node
}

func (d *document) Bytes(indent int) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(d.node); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (d *document) Encode(encoder *yaml.Encoder) error {
	return encoder.Encode(d.node)
}

func (*document) MarshalYAML() (interface{}, error) {
	panic("yamled.Document objects cannot be marshalled indirectly with a YAML encoder. Instead, use Bytes() or Encode() to get the desired results.")
}

func (d *document) Get(path ...Step) (Node, error) {
	n, err := NewNode(d.node.Content[0])
	if err != nil {
		return nil, err
	}

	return n.Get(path...)
}
