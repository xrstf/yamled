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

	Get(steps ...Step) (Node, bool)
	GetKey(steps ...Step) (KeyNode, bool)
	MustGet(steps ...Step) Node
	Set(value interface{}) error
	SetKey(key Step, value interface{}) (Node, error)
	SetAt(path Path, value interface{}) (Node, error)

	Replace(value interface{}) error
	ReplaceKey(key Step, value interface{}) (Node, error)
	ReplaceAt(path Path, value interface{}) (Node, error)

	DeleteKey(steps ...Step) error

	HeadComment() string
	LineComment() string
	FootComment() string

	SetHeadComment(comment string) Document
	SetLineComment(comment string) Document
	SetFootComment(comment string) Document
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

func (d *document) GetRootNode() (Node, error) {
	return NewNode(d.node.Content[0])
}

func (d *document) Bytes(indent int) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(indent)

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

/////////////////////////////////////////////////////////////////////
// comment API passthrough

func (d *document) HeadComment() string {
	return d.node.HeadComment
}

func (d *document) LineComment() string {
	return d.node.LineComment
}

func (d *document) FootComment() string {
	return d.node.FootComment
}

func (d *document) SetHeadComment(comment string) Document {
	d.node.HeadComment = comment
	return d
}

func (d *document) SetLineComment(comment string) Document {
	d.node.LineComment = comment
	return d
}

func (d *document) SetFootComment(comment string) Document {
	d.node.FootComment = comment
	return d
}

/////////////////////////////////////////////////////////////////////
// traversal - reading

func (d *document) Get(steps ...Step) (Node, bool) {
	n, err := d.GetRootNode()
	if err != nil {
		return nil, false
	}

	return n.Get(steps...)
}

func (d *document) GetKey(steps ...Step) (KeyNode, bool) {
	n, err := d.GetRootNode()
	if err != nil {
		return nil, false
	}

	return n.GetKey(steps...)
}

func (d *document) MustGet(steps ...Step) Node {
	n, err := d.GetRootNode()
	if err != nil {
		return nil
	}

	return n.MustGet(steps...)
}

/////////////////////////////////////////////////////////////////////
// traversal - writing

func (d *document) Set(value interface{}) error {
	n, err := d.GetRootNode()
	if err != nil {
		return err
	}

	return n.Set(value)
}

func (d *document) SetKey(key Step, value interface{}) (Node, error) {
	n, err := d.GetRootNode()
	if err != nil {
		return nil, err
	}

	return n.SetKey(key, value)
}

func (d *document) SetAt(path Path, value interface{}) (Node, error) {
	n, err := d.GetRootNode()
	if err != nil {
		return nil, err
	}

	return n.SetAt(path, value)
}

func (d *document) Replace(value interface{}) error {
	n, err := d.GetRootNode()
	if err != nil {
		return err
	}

	return n.Replace(value)
}

func (d *document) ReplaceKey(key Step, value interface{}) (Node, error) {
	n, err := d.GetRootNode()
	if err != nil {
		return nil, err
	}

	return n.ReplaceKey(key, value)
}

func (d *document) ReplaceAt(path Path, value interface{}) (Node, error) {
	n, err := d.GetRootNode()
	if err != nil {
		return nil, err
	}

	return n.ReplaceAt(path, value)
}

/////////////////////////////////////////////////////////////////////
// traversal - deleting

func (d *document) DeleteKey(steps ...Step) error {
	n, err := d.GetRootNode()
	if err != nil {
		return err
	}

	return n.DeleteKey(steps...)
}
