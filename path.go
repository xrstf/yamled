// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package yamled

import (
	"fmt"
	"strings"
)

type Step interface{}

type Path []Step

func (p Path) Append(s ...Step) Path {
	return append(p, s...)
}

func (p Path) Prepend(s ...Step) Path {
	return append(s, p)
}

// Parent returns the path except for the last element.
func (p Path) Parent() Path {
	if len(p) < 1 {
		return nil
	}

	return p[0 : len(p)-1]
}

func (p Path) Start() Step {
	if len(p) == 0 {
		return nil
	}

	return p[0]
}

func (p Path) End() Step {
	if len(p) == 0 {
		return nil
	}

	return p[len(p)-1]
}

func (p Path) Consume() (Step, Path) {
	if len(p) == 0 {
		return nil, nil
	}

	return p[0], p[1:]
}

func (p Path) String() string {
	parts := []string{}

	for _, p := range p {
		if s, ok := p.(string); ok {
			parts = append(parts, s)
			continue
		}

		if i, ok := p.(int); ok {
			parts = append(parts, fmt.Sprintf("[%d]", i))
			continue
		}

		parts = append(parts, fmt.Sprintf("%v", p))
	}

	return strings.Join(parts, ".")
}

func (p Path) Validate() error {
	errors := []string{}

	for _, s := range p {
		switch step := s.(type) {
		case string:
			// NOP
		case int:
			if step < 0 {
				errors = append(errors, fmt.Sprintf("%d is invalid, steps must be >= 0", step))
			}
		default:
			errors = append(errors, fmt.Sprintf("cannot handle %T steps", step))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("invalid path: %s", strings.Join(errors, "; "))
	}

	return nil
}
