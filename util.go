// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package yamled

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func KindName(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "Document"
	case yaml.SequenceNode:
		return "Sequence"
	case yaml.MappingNode:
		return "Mapping"
	case yaml.ScalarNode:
		return "Scalar"
	case yaml.AliasNode:
		return "Alias"
	default:
		return fmt.Sprintf("?Kind(%d)?", k)
	}
}

func StyleName(s yaml.Style) string {
	switch s {
	case yaml.TaggedStyle:
		return "Tagged"
	case yaml.DoubleQuotedStyle:
		return "DoubleQuoted"
	case yaml.SingleQuotedStyle:
		return "SingleQuoted"
	case yaml.LiteralStyle:
		return "Literal"
	case yaml.FoldedStyle:
		return "Folded"
	case yaml.FlowStyle:
		return "Flow"
	default:
		return "NoStyle" // fmt.Sprintf("?style(%d)?", s)
	}
}

func nullNode() *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!null",
		Value: "null",
	}
}

func mappingNode() *yaml.Node {
	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Tag:     "!!map",
		Content: []*yaml.Node{},
	}
}

func sequenceNode() *yaml.Node {
	return &yaml.Node{
		Kind:    yaml.SequenceNode,
		Tag:     "!!seq",
		Content: []*yaml.Node{},
	}
}

func stringNode(value string) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: value,
	}
}
