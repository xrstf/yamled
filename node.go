// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package yamled

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type Node interface {
	yaml.Marshaler
	fmt.Stringer

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

	ToString() string
	ToInt() int
	ToSlice() []interface{}
	ToMap() map[string]interface{}
	To(val interface{}) error

	HeadComment() string
	LineComment() string
	FootComment() string

	SetHeadComment(comment string) Node
	SetLineComment(comment string) Node
	SetFootComment(comment string) Node
}

type node struct {
	node *yaml.Node
}

func NewNode(n *yaml.Node) (Node, error) {
	if n == nil {
		return nil, errors.New("node cannot be nil")
	}

	if n.Kind == yaml.DocumentNode {
		return nil, errors.New("node cannot be a DocumentNode")
	}

	return &node{
		node: n,
	}, nil
}

func NewNodeFromReader(r io.Reader) (Node, error) {
	var node yaml.Node

	if err := yaml.NewDecoder(r).Decode(&node); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	return NewNode(&node)
}

func (n *node) Bytes(indent int) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(indent)

	if err := encoder.Encode(n.node); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (n *node) Encode(encoder *yaml.Encoder) error {
	return encoder.Encode(n.node)
}

func (n *node) MarshalYAML() (interface{}, error) {
	return n.node, nil
}

func (n *node) String() string {
	return n.ToString()
}

/////////////////////////////////////////////////////////////////////
// comment API passthrough

func (n *node) HeadComment() string {
	return n.node.HeadComment
}

func (n *node) LineComment() string {
	return n.node.LineComment
}

func (n *node) FootComment() string {
	return n.node.FootComment
}

func (n *node) SetHeadComment(comment string) Node {
	n.node.HeadComment = comment
	return n
}

func (n *node) SetLineComment(comment string) Node {
	n.node.LineComment = comment
	return n
}

func (n *node) SetFootComment(comment string) Node {
	n.node.FootComment = comment
	return n
}

/////////////////////////////////////////////////////////////////////
// traversal - reading

func (n *node) Get(steps ...Step) (Node, bool) {
	node, found, _ := n.get(steps...)

	return node, found
}

func (n *node) GetKey(steps ...Step) (KeyNode, bool) {
	if len(steps) == 0 {
		return nil, false
	}

	var curNode *yaml.Node
	curNode = n.node

	if len(steps) > 1 {
		// traverse down up until the last step
		childNode, found, _ := n.get(steps[:len(steps)-1]...)
		if !found {
			return nil, false
		}

		asserted, ok := childNode.(*node)
		if !ok {
			panic("This should never happen.")
		}

		curNode = asserted.node
	}

	if curNode.Kind != yaml.MappingNode {
		return nil, false
	}

	step := steps[len(steps)-1]

	// mappings are represented as [keyNode, valueNode, keyNode, valueNode, ...]
	// in this node's content
	for i := 0; i < len(curNode.Content); i += 2 {
		kNode := curNode.Content[i]

		// safety check
		if kNode.Kind != yaml.ScalarNode {
			continue
		}

		// we found the key!
		if kNode.Value == step {
			return &keyNode{
				node: kNode,
			}, true
		}
	}

	// key not found
	return nil, false
}

func (n *node) MustGet(steps ...Step) Node {
	child, found, _ := n.get(steps...)
	if !found {
		return &node{nullNode()}
	}

	return child
}

// get is providing more information about why a step
// was not found than the regular .Get() function.
func (n *node) get(steps ...Step) (Node, bool, bool) {
	if n == nil {
		return nil, false, false
	}

	if len(steps) == 0 {
		return nil, false, false
	}

	// convenience feature, allow to chain steps in order to
	// avoid repeated .Get("key").Get("subkey") calls
	// with error checks in between
	if len(steps) > 1 {
		head := steps[0]
		tail := steps[1:]

		child, found, incompatibleKind := n.get(head)
		if !found {
			return child, found, incompatibleKind
		}

		childAsserted, ok := child.(*node)
		if !ok {
			panic("This should never happen.")
		}

		return childAsserted.get(tail...)
	}

	switch step := steps[0].(type) {
	// string means descending into an object
	case string:
		if n.node.Kind != yaml.MappingNode {
			return nil, false, true
		}

		// mappings are represented as [keyNode, valueNode, keyNode, valueNode, ...]
		// in this node's content
		for i := 0; i < len(n.node.Content); i += 2 {
			keyNode := n.node.Content[i]

			// safety check
			if keyNode.Kind != yaml.ScalarNode {
				continue
			}

			// we found the key! next content item will be the value
			if keyNode.Value == step {
				// safety check
				if i+1 >= len(n.node.Content) {
					return nil, false, false
				}

				node, err := NewNode(n.node.Content[i+1])
				if err != nil {
					return nil, false, false
				}

				// success!
				return node, true, false
			}
		}

		// key not found
		return nil, false, false

	// int means descending into an array
	case int:
		if n.node.Kind != yaml.SequenceNode {
			return nil, false, true
		}

		if step >= len(n.node.Content) {
			return nil, false, false
		}

		node, err := NewNode(n.node.Content[step])
		if err != nil {
			return nil, false, false
		}

		// success!
		return node, true, false
	}

	// cannot handle this type of step
	return nil, false, false
}

/////////////////////////////////////////////////////////////////////
// traversal - writing

func (n *node) Set(value interface{}) error {
	return n.set(value, true)
}

func (n *node) Replace(value interface{}) error {
	return n.set(value, false)
}

func (n *node) set(value interface{}, forbidKindChange bool) error {
	parsed, err := createNode(value)
	if err != nil {
		return err
	}

	return n.setNode(parsed, forbidKindChange)
}

func (n *node) setNode(newNode *yaml.Node, forbidKindChange bool) error {
	if forbidKindChange && !compatibleKinds(newNode, n.node) {
		return errors.New("cannot set a new node kind without replacing the node")
	}

	deepCopyNode(n.node, *newNode)

	return nil
}

func (n *node) SetKey(key Step, value interface{}) (Node, error) {
	return n.setKey(key, value, true)
}

func (n *node) ReplaceKey(key Step, value interface{}) (Node, error) {
	return n.setKey(key, value, false)
}

func (n *node) setKey(key Step, value interface{}, forbidKindChange bool) (Node, error) {
	newNode, err := createNode(value)
	if err != nil {
		return nil, err
	}

	if err := n.setKeyNode(key, newNode, forbidKindChange); err != nil {
		return nil, err
	}

	return NewNode(newNode)
}

func (n *node) setKeyNode(key Step, newNode *yaml.Node, forbidKindChange bool) error {
	switch n.node.Kind {
	case yaml.MappingNode:
		step, ok := key.(string)
		if !ok {
			return errors.New("invalid key type, must be string")
		}

		// try to find the key
		for i := 0; i < len(n.node.Content); i += 2 {
			keyNode := n.node.Content[i]

			// safety check
			if keyNode.Kind != yaml.ScalarNode {
				continue
			}

			// we found the key! next content item will be the value
			if keyNode.Value == step {
				// safety check
				if i+1 >= len(n.node.Content) {
					return errors.New("found key node, but current object has no value node")
				}

				if forbidKindChange && !compatibleKinds(n.node.Content[i+1], newNode) {
					return errors.New("cannot change the node's kind")
				}

				// success!
				n.node.Content[i+1] = newNode
				return nil
			}
		}

		// key was not yet found, let's insert one automagically
		n.node.Content = append(n.node.Content,
			stringNode(step),
			newNode,
		)

		// success!
		return nil

	case yaml.SequenceNode:
		step, ok := key.(int)
		if !ok {
			return errors.New("invalid key type, must be int")
		}

		if step < 0 {
			return errors.New("step must be >= 0")
		}

		// insert enough empty nodes to fill up the content
		for step >= len(n.node.Content) {
			n.node.Content = append(n.node.Content, nullNode())
		}

		if forbidKindChange && !compatibleKinds(n.node.Content[step], newNode) {
			return errors.New("cannot change the node's kind")
		}

		n.node.Content[step] = newNode

		// success!
		return nil

	default:
		return errors.New("node is neither sequence nor mapping node, cannot set a child value")
	}
}

func (n *node) SetAt(path Path, value interface{}) (Node, error) {
	return n.setAt(path, value, true)
}

func (n *node) ReplaceAt(path Path, value interface{}) (Node, error) {
	return n.setAt(path, value, false)
}

func (n *node) setAt(path Path, value interface{}, forbidKindChange bool) (Node, error) {
	if len(path) == 0 {
		return nil, errors.New("path cannot be empty")
	}

	if err := path.Validate(); err != nil {
		return nil, err
	}

	// stop recursing
	if len(path) == 1 {
		return n.setKey(path[0], value, forbidKindChange)
	}

	head, tail := path.Consume()

	childNode, keyFound, incompatibleKind := n.get(head)
	if incompatibleKind {
		if forbidKindChange {
			return nil, errors.New("current node cannot be traversed into and changing the kind is disabled")
		}

		// replace current node with a compatible, empty one
		newEmptyNode, err := createFittingEmptyNode(head)
		if err != nil {
			return nil, err
		}

		deepCopyNode(n.node, *newEmptyNode)

		// the key cannot possibly exist now
		keyFound = false
	}

	if !keyFound {
		// insert an empty node that fits whatever we want to recurse into *next*
		newEmptyNode, err := createFittingEmptyNode(tail.Start())
		if err != nil {
			return nil, err
		}

		if err := n.setKeyNode(head, newEmptyNode, false); err != nil {
			return nil, err
		}

		childNode = &node{newEmptyNode}
	}

	childAsserted, ok := childNode.(*node)
	if !ok {
		panic("This should never happen.")
	}

	return childAsserted.setAt(tail, value, forbidKindChange)
}

/////////////////////////////////////////////////////////////////////
// traversal - deleting

func (n *node) DeleteKey(steps ...Step) error {
	if len(steps) == 0 {
		return errors.New("path cannot be empty")
	}

	if len(steps) > 1 {
		// re-use the existing recursion helper in Get()
		// to get to the node right before the last step.
		headPath := steps[:len(steps)-1]

		node, exists := n.Get(headPath...)
		if !exists {
			return nil
		}

		// and then recurse to actually remove the  key
		return node.DeleteKey(steps[len(steps)-1])
	}

	switch step := steps[0].(type) {
	// string means we remove a key from an object (mapping)
	case string:
		if n.node.Kind != yaml.MappingNode {
			return nil
		}

		// mappings are represented as [keyNode, valueNode, keyNode, valueNode, ...]
		// in this node's content
		keyIndex := -1

		for i := 0; i < len(n.node.Content); i += 2 {
			keyNode := n.node.Content[i]

			// safety check
			if keyNode.Kind != yaml.ScalarNode {
				continue
			}

			// we found the key! next content item will be the value
			if keyNode.Value == step {
				// safety check
				if i+1 >= len(n.node.Content) {
					return nil
				}

				keyIndex = i
				break
			}
		}

		// key not found
		if keyIndex == -1 {
			return nil
		}

		// remove the key node and the value node
		n.node.Content = append(n.node.Content[:keyIndex], n.node.Content[keyIndex+2:]...)

		// success
		return nil

	// int means removing an item from an array
	// (this shrinks the array and does not leave gaps)
	case int:
		if n.node.Kind != yaml.SequenceNode {
			return nil
		}

		if step >= len(n.node.Content) {
			return nil
		}

		// remove the array item
		n.node.Content = append(n.node.Content[:step], n.node.Content[step+1:]...)

		// success!
		return nil

	default:
		return fmt.Errorf("cannot handle %T steps", step)
	}
}

/////////////////////////////////////////////////////////////////////
// conversions

func (n *node) ToString() string {
	if n.node.Kind != yaml.ScalarNode {
		return ""
	}

	var s string
	if err := n.To(&s); err != nil {
		return ""
	}

	return s
}

func (n *node) ToInt() int {
	if n.node.Kind != yaml.ScalarNode {
		return 0
	}

	var i int
	if err := n.To(&i); err != nil {
		return 0
	}

	return i
}

func (n *node) ToSlice() []interface{} {
	if n.node.Kind != yaml.SequenceNode {
		return nil
	}

	var values []interface{}
	if err := n.To(&values); err != nil {
		return nil
	}

	return values
}

func (n *node) ToMap() map[string]interface{} {
	if n.node.Kind != yaml.MappingNode {
		return nil
	}

	var values map[string]interface{}
	if err := n.To(&values); err != nil {
		return nil
	}

	return values
}

func (n *node) To(val interface{}) error {
	return n.node.Decode(val)
}

/////////////////////////////////////////////////////////////////////
// helpers

func compatibleKinds(a, b *yaml.Node) bool {
	return a.Kind == b.Kind || isNullNode(a) || isNullNode(b)
}

func createFittingEmptyNode(s Step) (*yaml.Node, error) {
	if s == nil {
		return nullNode(), nil
	}

	switch s.(type) {
	case string:
		return mappingNode(), nil

	case int:
		return sequenceNode(), nil

	default:
		return nil, fmt.Errorf("cannot handle %T steps when traversing paths", s)
	}
}

func createNode(value interface{}) (*yaml.Node, error) {
	var buf bytes.Buffer
	if err := yaml.NewEncoder(&buf).Encode(value); err != nil {
		return nil, err
	}

	var node yaml.Node
	if err := yaml.NewDecoder(&buf).Decode(&node); err != nil {
		return nil, err
	}

	// safety check
	if node.Kind != yaml.DocumentNode {
		return nil, fmt.Errorf("expected a Document node from decoding a YAML string, but got %v", KindName(node.Kind))
	}

	return node.Content[0], nil
}

func deepCopyNode(dst *yaml.Node, src yaml.Node) {
	dst.Kind = src.Kind
	dst.Style = src.Style
	dst.Tag = src.Tag
	dst.Value = src.Value
	dst.Anchor = src.Anchor
	dst.Alias = src.Alias
	dst.Content = src.Content
	dst.HeadComment = src.HeadComment
	dst.LineComment = src.LineComment
	dst.FootComment = src.FootComment
	dst.Line = src.Line
	dst.Column = src.Column
}
