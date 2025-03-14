package web

import (
	"regexp"
	"strings"
)

type patternType int8

const (
	_PATTERN_STATIC    patternType = iota // /home
	_PATTERN_REGEXP                       // /:id([0-9]+)
	_PATTERN_PATH_EXT                     // /*.*
	_PATTERN_HOLDER                       // /:user
	_PATTERN_MATCH_ALL                    // /*
)

// Leaf represents a leaf route information.
type Leaf struct {
	parent *Tree

	typ        patternType
	pattern    string
	rawPattern string // Contains wildcard instead of regexp
	wildcards  []string
	reg        *regexp.Regexp
	optional   bool

	handle Handle
}

// Tree represents a router tree in Macaron.
type Tree struct {
	parent *Tree

	typ        patternType
	pattern    string
	rawPattern string
	wildcards  []string
	reg        *regexp.Regexp

	subtrees []*Tree
	leaves   []*Leaf
}

var wildcardPattern = regexp.MustCompile(`:[a-zA-Z0-9]+`)

func NewLeaf(parent *Tree, pattern string, handle Handle) *Leaf {
	typ, rawPattern, wildcards, reg := checkPattern(pattern)
	optional := false
	if len(pattern) > 0 && pattern[0] == '?' {
		optional = true
	}
	return &Leaf{parent, typ, pattern, rawPattern, wildcards, reg, optional, handle}
}

func NewSubtree(parent *Tree, pattern string) *Tree {
	typ, rawPattern, wildcards, reg := checkPattern(pattern)
	return &Tree{parent, typ, pattern, rawPattern, wildcards, reg, make([]*Tree, 0, 5), make([]*Leaf, 0, 5)}
}

func NewTree() *Tree {
	return NewSubtree(nil, "")
}

func (t *Tree) Add(pattern string, handle Handle) *Leaf {
	pattern = strings.TrimSuffix(pattern, "/")
	return t.addNextSegment(pattern, handle)
}

func (t *Tree) addNextSegment(pattern string, handle Handle) *Leaf {
	pattern = strings.TrimPrefix(pattern, "/")

	i := strings.Index(pattern, "/")
	if i == -1 {
		return t.addLeaf(pattern, handle)
	}
	return t.addSubtree(pattern[:i], pattern[i+1:], handle)
}

func (t *Tree) addLeaf(pattern string, handle Handle) *Leaf {
	for i := 0; i < len(t.leaves); i++ {
		if t.leaves[i].pattern == pattern {
			return t.leaves[i]
		}
	}

	leaf := NewLeaf(t, pattern, handle)

	// Add exact same leaf to grandparent/parent level without optional.
	if leaf.optional {
		parent := leaf.parent
		if parent.parent != nil {
			parent.parent.addLeaf(parent.pattern, handle)
		} else {
			parent.addLeaf("", handle) // Root tree can add as empty pattern.
		}
	}

	i := 0
	for ; i < len(t.leaves); i++ {
		if leaf.typ < t.leaves[i].typ {
			break
		}
	}

	if i == len(t.leaves) {
		t.leaves = append(t.leaves, leaf)
	} else {
		t.leaves = append(t.leaves[:i], append([]*Leaf{leaf}, t.leaves[i:]...)...)
	}
	return leaf
}

func (t *Tree) addSubtree(segment, pattern string, handle Handle) *Leaf {
	for i := 0; i < len(t.subtrees); i++ {
		if t.subtrees[i].pattern == segment {
			return t.subtrees[i].addNextSegment(pattern, handle)
		}
	}

	subtree := NewSubtree(t, segment)
	i := 0
	for ; i < len(t.subtrees); i++ {
		if subtree.typ < t.subtrees[i].typ {
			break
		}
	}

	if i == len(t.subtrees) {
		t.subtrees = append(t.subtrees, subtree)
	} else {
		t.subtrees = append(t.subtrees[:i], append([]*Tree{subtree}, t.subtrees[i:]...)...)
	}
	return subtree.addNextSegment(pattern, handle)
}

func checkPattern(pattern string) (typ patternType, rawPattern string, wildcards []string, reg *regexp.Regexp) {
	pattern = strings.TrimLeft(pattern, "?")
	rawPattern = getRawPattern(pattern)

	if pattern == "*" {
		typ = _PATTERN_MATCH_ALL
	} else if pattern == "*.*" {
		typ = _PATTERN_PATH_EXT
	} else if strings.Contains(pattern, ":") {
		typ = _PATTERN_REGEXP
		pattern, wildcards = getWildcards(pattern)
		if pattern == "(.+)" {
			typ = _PATTERN_HOLDER
		} else {
			reg = regexp.MustCompile(pattern)
		}
	}
	return typ, rawPattern, wildcards, reg
}

func getWildcards(pattern string) (string, []string) {
	wildcards := make([]string, 0, 2)

	// Keep getting next wildcard until nothing is left.
	var wildcard string
	for {
		wildcard, pattern = getNextWildcard(pattern)
		if len(wildcard) > 0 {
			wildcards = append(wildcards, wildcard)
		} else {
			break
		}
	}

	return pattern, wildcards
}

// getNextWildcard tries to find next wildcard and update pattern with corresponding regexp.
func getNextWildcard(pattern string) (wildcard, _ string) {
	pos := wildcardPattern.FindStringIndex(pattern)
	if pos == nil {
		return "", pattern
	}
	wildcard = pattern[pos[0]:pos[1]]

	// Reach last character or no regexp is given.
	if len(pattern) == pos[1] {
		return wildcard, strings.Replace(pattern, wildcard, `(.+)`, 1)
	} else if pattern[pos[1]] != '(' {
		switch {
		case isSpecialRegexp(pattern, ":int", pos):
			pattern = strings.Replace(pattern, ":int", "([0-9]+)", 1)
		case isSpecialRegexp(pattern, ":string", pos):
			pattern = strings.Replace(pattern, ":string", "([\\w]+)", 1)
		default:
			return wildcard, strings.Replace(pattern, wildcard, `(.+)`, 1)
		}
	}

	// Cut out placeholder directly.
	return wildcard, pattern[:pos[0]] + pattern[pos[1]:]
}

// getRawPattern removes all regexp but keeps wildcards for building URL path.
func getRawPattern(rawPattern string) string {
	rawPattern = strings.ReplaceAll(rawPattern, ":int", "")
	rawPattern = strings.ReplaceAll(rawPattern, ":string", "")

	for {
		startIdx := strings.Index(rawPattern, "(")
		if startIdx == -1 {
			break
		}

		closeIdx := strings.Index(rawPattern, ")")
		if closeIdx > -1 {
			rawPattern = rawPattern[:startIdx] + rawPattern[closeIdx+1:]
		}
	}
	return rawPattern
}

func isSpecialRegexp(pattern, regStr string, pos []int) bool {
	return len(pattern) >= pos[1]+len(regStr) && pattern[pos[1]:pos[1]+len(regStr)] == regStr
}
