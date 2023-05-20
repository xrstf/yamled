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

	Get(path ...Step) (Node, error)
	Set(path Path, value interface{}) error

	// debugging only
	GetNode() *yaml.Node
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

func (n *node) GetNode() *yaml.Node {
	return n.node
}

func (n *node) MarshalYAML() (interface{}, error) {
	return n.node, nil
}

/////////////////////////////////////////////////////////////////////
// traversal - reading

func (n *node) Get(path ...Step) (Node, error) {
	flattened, err := flattenPath(path)
	if err != nil {
		return nil, err
	}

	if err := flattened.Validate(); err != nil {
		return nil, err
	}

	result, getErr := nodeGet(n.node, flattened)
	if getErr != nil {
		return nil, getErr
	}

	return NewNode(result)
}

var (
	ErrStepNotFound     = errors.New("step cannot be found")
	ErrNodeNotAMapping  = errors.New("node is not a mapping node")
	ErrNodeNotASequence = errors.New("node is not a sequence node")
)

type TraversalError struct {
	Path Path
	Err  error
}

func (e *TraversalError) Error() string {
	return fmt.Sprintf("failed to traverse path at %v: %v", e.Path.String(), e.Err)
}

func IsNotFound(err error) bool {
	e := &TraversalError{}
	if errors.As(err, &e) {
		return errors.Is(e.Err, ErrStepNotFound) || errors.Is(e.Err, ErrNodeNotAMapping) || errors.Is(e.Err, ErrNodeNotASequence)
	}
	return false
}

func nodeGet(n *yaml.Node, path Path) (*yaml.Node, error) {
	if len(path) == 0 {
		return n, nil
	}

	// descend further
	currentStep, remainingPath := path.Consume()

	makeErr := func(err error) *TraversalError {
		terr := &TraversalError{}
		if errors.As(err, &terr) {
			terr.Path = terr.Path.Prepend(currentStep)
			return terr
		}

		return &TraversalError{
			Path: Path{currentStep},
			Err:  err,
		}
	}

	switch step := currentStep.(type) {
	// string means descending into an object
	case string:
		if n.Kind != yaml.MappingNode {
			return nil, makeErr(fmt.Errorf("%w: found a %v", ErrNodeNotAMapping, KindName(n.Kind)))
		}

		for i := 0; i < len(n.Content); i += 2 {
			keyNode := n.Content[i]

			// safety check
			if keyNode.Kind != yaml.ScalarNode {
				continue
			}

			// we found the key! next content item will be the value
			if keyNode.Value == step {
				// safety check
				if i+1 >= len(n.Content) {
					return nil, makeErr(errors.New("found key node, but current object has no value node"))
				}

				result, err := nodeGet(n.Content[i+1], remainingPath)
				if err != nil {
					return nil, makeErr(err)
				}

				return result, nil
			}
		}

		return nil, makeErr(ErrStepNotFound)

	// int means descending into an array
	case int:
		if n.Kind != yaml.SequenceNode {
			return nil, makeErr(fmt.Errorf("%w: found a %v", ErrNodeNotASequence, KindName(n.Kind)))
		}

		if step >= len(n.Content) {
			return nil, makeErr(ErrStepNotFound)
		}

		result, err := nodeGet(n.Content[step], remainingPath)
		if err != nil {
			return nil, makeErr(err)
		}

		return result, nil
	}

	return nil, makeErr(fmt.Errorf("cannot handle %T steps when traversing paths", currentStep))
}

/////////////////////////////////////////////////////////////////////
// traversal - writing

func (n *node) Set(path Path, value interface{}) error {
	if len(path) == 0 {
		return errors.New("path cannot be empty")
	}

	if err := path.Validate(); err != nil {
		return err
	}

	return nodeSet(n.node, path, value)
}

func nodeSet(node *yaml.Node, path Path, value interface{}) error {
	currentStep, remainingPath := path.Consume()

	makeErr := func(err error) *TraversalError {
		terr := &TraversalError{}
		if errors.As(err, &terr) {
			terr.Path = terr.Path.Prepend(currentStep)
			return terr
		}

		return &TraversalError{
			Path: Path{currentStep},
			Err:  err,
		}
	}

	// if we're on the leaf level, we will insert the desired value;
	// otherwise we inject a synthetic empty value baed on the step type
	var (
		newNode *yaml.Node
		err     error
	)

	if len(remainingPath) == 0 {
		newNode, err = createNode(value)
	} else {
		newNode, err = createFittingEmptyNode(remainingPath.Start())
	}

	if err != nil {
		return makeErr(err)
	}

	//////////////////////////////////////////////////
	// step 1: prepare the current node to accomodate the new value
	//         this does not yet insert the actual value, but instead
	//         a synthetic empty value based on the _next_ step

	switch step := currentStep.(type) {
	// string means descending into an object
	case string:
		// safety check
		if node.Kind != yaml.MappingNode {
			return makeErr(fmt.Errorf("%w: found a %v", ErrNodeNotAMapping, KindName(node.Kind)))
		}

		// try to find the key in the current node
		success := false

		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]

			// safety check
			if keyNode.Kind != yaml.ScalarNode {
				continue
			}

			// we found the key! next content item will be the value
			if keyNode.Value == step {
				// safety check
				if i+1 >= len(node.Content) {
					return makeErr(errors.New("found key node, but current object has no value node"))
				}

				node.Content[i+1] = newNode
				success = true
				break
			}
		}

		// key was not yet found, let's insert one automagically
		if !success {
			node.Content = append(node.Content,
				stringNode(step),
				newNode,
			)
		}

	// int means descending into an array
	case int:
		if node.Kind != yaml.SequenceNode {
			return makeErr(fmt.Errorf("%w: found a %v", ErrNodeNotASequence, KindName(node.Kind)))
		}

		// insert enough empty nodes to fill up the content
		for step >= len(node.Content) {
			node.Content = append(node.Content, nullNode())
		}

		node.Content[step] = newNode

	default:
		makeErr(fmt.Errorf("cannot handle %T steps when traversing paths", currentStep))
	}

	//////////////////////////////////////////////////
	// step 2: descend further, if needed

	if len(remainingPath) > 0 {
		if err := nodeSet(newNode, remainingPath, value); err != nil {
			return makeErr(err)
		}
	}

	return nil
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
