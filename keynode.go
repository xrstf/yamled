// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package yamled

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// KeyNode is a specialized interface that makes setting comments
// for mapping keys more useful, as setting a comment on the
// value node itself might be undesirable.
type KeyNode interface {
	fmt.Stringer

	HeadComment() string
	LineComment() string
	FootComment() string

	SetHeadComment(comment string) KeyNode
	SetLineComment(comment string) KeyNode
	SetFootComment(comment string) KeyNode
}

type keyNode struct {
	node *yaml.Node
}

func (n *keyNode) String() string {
	return n.node.Value
}

func (n *keyNode) HeadComment() string {
	return n.node.HeadComment
}

func (n *keyNode) LineComment() string {
	return n.node.LineComment
}

func (n *keyNode) FootComment() string {
	return n.node.FootComment
}

func (n *keyNode) SetHeadComment(comment string) KeyNode {
	n.node.HeadComment = comment
	return n
}

func (n *keyNode) SetLineComment(comment string) KeyNode {
	n.node.LineComment = comment
	return n
}

func (n *keyNode) SetFootComment(comment string) KeyNode {
	n.node.FootComment = comment
	return n
}
