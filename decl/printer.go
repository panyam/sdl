package decl

import (
	"fmt"
	"strings"
)

type CodePrinter interface {
	Indent(n int)
	Unindent(n int)
	Print(str string)
	Printf(fmt string, args ...any)
	Println(str string)
}

func WithIndent(n int, cp CodePrinter, block func(cp CodePrinter)) {
	cp.Indent(n)
	defer cp.Unindent(n)
	block(cp)
}

type codePrinter struct {
	indent      int
	line        int
	col         int
	builder     strings.Builder
	linebuilder strings.Builder
}

func (c *codePrinter) Indent(n int) {
	c.indent += n
}

func (c *codePrinter) Unindent(n int) {
	c.indent -= n
	if c.indent < 0 {
		c.indent = 0
	}
}

func (c *codePrinter) Print(str string) {
	lines := strings.Split(str, "\n")
	for idx, l := range lines {
		if c.col == 0 {
			// new line has started so add the indent string
			c.linebuilder.WriteString(c.IndentString())
			c.linebuilder.WriteString(l)
			c.col += len(l)
		}
		hasMore := idx < len(lines)
		if hasMore {
			c.line++
			c.col = 0
			c.builder.WriteString(c.linebuilder.String())
			c.builder.WriteRune('\n')
			c.linebuilder.Reset()
		}
	}
}

func (c *codePrinter) Println(str string) {
	c.Print(str + "\n")
}

func (c *codePrinter) Printf(format string, args ...any) {
	c.Print(fmt.Sprintf(format, args...))
}

func (c *codePrinter) IndentString() string {
	out := ""
	for range c.indent {
		out += "  "
	}
	return out
}

func NewCodePrinter() CodePrinter {
	return &codePrinter{}
}

func PPrint(node Node) {
	cp := &codePrinter{}
	node.PrettyPrint(cp)
	fmt.Println(cp.builder.String())
}
