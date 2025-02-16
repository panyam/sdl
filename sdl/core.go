package sdl

import "log"

type Attribs map[string]any

type Choice struct {
	Weight Fraction

	// Attributes at this node
	Attribs map[string]any

	Parent   *Choice
	Children []*Choice
}

type ChoiceMatcher func(ch *Choice) bool

// Perform a deep clone of this choice
func (c *Choice) Clone() *Choice {
	out := &Choice{
		Weight:  c.Weight,
		Attribs: c.Attribs,
	}

	var children []*Choice
	for _, ch := range c.Children {
		chclone := ch.Clone()
		chclone.Parent = out
		children = append(children, chclone)
	}
	out.Children = children
	return out
}

// Get the attribute of a choice.  The attribute of a choice also includes those of its parents
func (c *Choice) Attrib(key string) any {
	if val, ok := c.Attribs[key]; ok && val != nil {
		return val
	}
	if c.Parent != nil {
		return c.Parent.Attrib(key)
	}
	return nil
}

// Sum of direct weights of all children
func (c *Choice) TotalWeight() (out Fraction) {
	for _, child := range c.Children {
		out = out.Plus(child.Weight).Factorized()
	}
	return
}

// Ratio of this child within the parent
func (c *Choice) RelWeight() Fraction {
	curr := c.Weight
	if c.Parent != nil {
		curr = curr.DivBy(c.Parent.TotalWeight()).Factorized()
	}
	return curr
}

// Ratio of this child's weight across *all* choices in the tree
func (c *Choice) FinalWeight() Fraction {
	// Root is 100%
	if c.Parent == nil {
		return Frac(1, 1)
	}
	return c.RelWeight().Times(c.Parent.FinalWeight()).Factorized()
}

// For a given parent adds a new choice.  By using weights intead of probabilities it is easier to "adjust" things as we go
func (c *Choice) Add(weight any, attribs Attribs) *Choice {
	fracWeight, ok := weight.(Fraction)
	if !ok {
		// TODO - caller must check or return error
		log.Fatalf("Invalid weight: %v.  Must be a int or a Fraction", weight)
	}
	// To ensure we dont have double counting, the only constraint we impose is that parent and child should not have the same attributes
	// Since this is an outcome tree the parent outcome must be a union of all the child outcomes
	for pkey := range attribs {
		if c.Attrib(pkey) != nil {
			log.Fatalf("Key %s exists in both parent and child (weight: %s)", pkey, fracWeight.String())
		}
	}

	// Then add a new choice
	child := &Choice{
		Parent:  c,
		Weight:  fracWeight,
		Attribs: attribs,
	}
	c.Children = append(c.Children, child)
	return c
}

// Add a list of choices to a particular parent with the respective weights
func (c *Choice) AddChoices(weight any, attribs Attribs, rest ...any) (choices []*Choice) {
	choices = append(choices, c.Add(weight, attribs))
	if len(rest)%2 != 0 {
		// TODO - caller must check or return error
		panic("rest must have even number of items as Weight,Attribs pairs")
	}
	for i := 0; i < len(rest); i += 2 {
		attribs := rest[i+1]
		attribMap, ok := attribs.(map[string]any)
		if !ok {
			log.Fatalf("%d th attribs should be a map[string]any map: %v", i+1, attribs)
		}
		choices = append(choices, c.Add(rest[i], attribMap))
	}
	return choices
}

func (c *Choice) If(matcher ChoiceMatcher, body *Choice, otherwise *Choice) *Choice {
	return nil
}

func (c *Choice) Then(rest ...*Choice) *Choice {
	return nil
}

// This is a "switch" statement, eg:
// given current Outcome,
//
//	case 1 => do T1
//	case 2 => do T2
//	...
//	default: do TDef
/*
func (o *Outcome) Switch(caseFunc func(ch *Choice) bool, body *Outcome, args ...any) (partitions *Outcome) {
	var defaultOutcome *Outcome
	L := len(args)
	if L%2 != 0 {
		if _, ok := args[L-1].(*Outcome); !ok {
			panic("If args.Len is not even then last item in rest array should be the default Outcome")
		}
		defaultOutcome = args[L-1].(*Outcome)
	}

	// Way we would do this is split our current choice set into
	// those that match the cond and the rest
	splitNodes := func(matcher Matcher, choices []*Choice) (matched []*Choice, remaining []*Choice) {
		for _, ch := range choices {
			if ch.isLeaf {
				if matcher(ch) {
					matched = append(matched, ch)
				} else {
					remaining = append(remaining, ch)
				}
			}
		}
		return
	}

	var out Outcome
	var matched []*Choice
	rest := o.allNodes

	// How do we do each one of these?
	for i := -2; i < len(args); i += 2 {
		if i >= 0 {
			caseFunc = args[i].(Matcher)
			body = args[i+1].(*Outcome)
		}
		matched, rest = splitNodes(caseFunc, rest)
	}

	if defaultOutcome != nil && len(rest) > 0 {
		// The final case for our outcome list
	}
	return &out
}
*/
