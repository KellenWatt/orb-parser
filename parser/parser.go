package parser

import (
	// "fmt"
	"strings"
	"regexp"
	"io"
	"strconv"
)

type ParseError struct {
	msg string
	line int
}

func NewError(msg string, s *Scanner) *ParseError {
	return &ParseError{msg, s.Line()}
}

func (p *ParseError) Error() string {
	return strconv.Itoa(p.line) + ":" + p.msg
}

type ParseTree struct {
	Directives map[string]string
	Tables []*Table
}

func NewParseTree() *ParseTree {
	return &ParseTree{make(map[string]string), nil}
}

func (p *ParseTree) String() string {
	var out string
	for d, v := range p.Directives {
		out += "#" + d + "=" + v + "\n"
	}
	for _, t := range p.Tables {
		out += "\n" + t.String()
	}
	return out
}

type Table struct {
	Name string
	Columns []*Column
}

func (t *Table) String() string {
	s := "[" + t.Name + "]"
	for _, c := range t.Columns {
		s += "\n" + c.String()
	}
	return s + "\n"
}

type Column struct {
	Name string
	Type string
	RequestedType string
	Alias string
	Constraints []*Constraint
}

func (c *Column) String() string {
	s := c.Name + " " + c.Type
	if c.RequestedType != "" {
		s += " using " + c.RequestedType
	}
	if c.Alias != "" {
		s += "\n-alias: " + c.Alias
	}
	for _, con := range c.Constraints {
		s += "\n" + con.String()
	}
	return s
}

type Constraint struct {
	Name string
	Value string
}

func (c *Constraint) String() string {
	out := "-" + c.Name
	if c.Value != "" {
		out += ": " + c.Value
	}
	return out
}


func Parse(stream io.Reader) (*ParseTree, []error) {
	tree := NewParseTree()
	input := NewScanner(stream)
	if input.EOF() {
		return tree, nil
	}

	var errors []error

	for input.Scan() {
		if input.Text() == "" {
			continue
		} else if strings.HasPrefix(input.Text(), "#") {
			input.Backtrack()
			if kind, value, valid := ParseDirective(input); valid {
				tree.Directives[kind] = value
			}
			// else skip it because it's a comment
			// possibly account for incorrect directives in the future
		} else if strings.HasPrefix(input.Text(), "[") {
			input.Backtrack()
			table, errs := ParseTable(input)
			if errs != nil {
				errors = append(errors, errs...)
			} else {
				tree.Tables = append(tree.Tables, table)
			}
		} else {
			errors = append(errors, NewError("Invalid token outside table definition", input))
		}
	}

	return tree, errors
}
var directiveFormat = regexp.MustCompile(`^#\s*([-A-Za-z_]+)\s*=\s*([A-Za-z0-9_]+)\s*$`)

func ParseDirective(input *Scanner) (string, string, bool) {
	input.Scan()
	matches := directiveFormat.FindStringSubmatch(input.Text())
	if matches != nil {
		return matches[1], matches[2], true
	}
	// Isn't a compiler directive
	return "", "", false
}

var tableName = regexp.MustCompile(`^\[([-A-Za-z0-9_]+)\]\s*$`)

func ParseTable(input *Scanner) (*Table, []error) {
	input.Scan()
	matches := tableName.FindStringSubmatch(input.Text())
	if matches == nil {
		return nil, []error{NewError("Invalid table name", input)}
	}

	table := &Table{Name: matches[1]}
	var errors []error
	for input.Scan() {
		if input.Text() == "" {
			return table, errors
		}
		input.Backtrack()
		col, errs := ParseColumn(input)
		if errs != nil {
			errors = append(errors, errs...)
		} else if col != nil {
			table.Columns = append(table.Columns, col)
		} else {
			input.Backtrack()
			break
		}
	}
	return table, errors
}

// DB types can have spaces, but not internal types
var columnFormat = regexp.MustCompile(`^([^-\s]+?)\s+([^\s]+?)(?:\s+using\s+([-A-Za-z0-9_]+.*?))?\s*$`)
var subIndicator = regexp.MustCompile(`^\s*-`)

func ParseColumn(input *Scanner) (*Column, []error) {
	input.Scan()
	column := &Column{}
	var errors []error
	matches := columnFormat.FindStringSubmatch(input.Text())
	if matches != nil {
		column.Name = matches[1]
		column.Type = matches[2]
		column.RequestedType = matches[3]
		for input.Scan() {
			if strings.Contains(input.Text(), "alias") {
				input.Backtrack()
				alias, errs := ParseAlias(input)
				if errs != nil {
					errors = append(errors, errs...)
				}
				column.Alias = alias
			} else if subIndicator.MatchString(input.Text()) {
				input.Backtrack()
				constraint, errs := ParseConstraint(input)
				if errs != nil {
					errors = append(errors, errs...)
				} else if constraint != nil {
					column.Constraints = append(column.Constraints, constraint)
				}
			} else {
				input.Backtrack()
				break
			}
		}
	} else {
		errors = append(errors, NewError("Invalid column definition", input))
	}
	return column, errors
}

var aliasFormat = regexp.MustCompile(`^\s*-\s*alias\s*:\s*([^\s]+?)\s*$`)

func ParseAlias(input *Scanner) (string, []error) {
	input.Scan()
	matches := aliasFormat.FindStringSubmatch(input.Text())
	if matches != nil {
		return matches[1], nil
	} else {
		return "", []error{NewError("Ill-formed alias", input)}
	}
}

// Can legally match aliases, but that's a bad idea
var constraintFormat = regexp.MustCompile(`\s*-\s*(.+?)(?:\s*:\s*(.+?))?\s*$`)

func ParseConstraint(input *Scanner) (*Constraint, []error) {
	input.Scan()
	c := &Constraint{}
	matches := constraintFormat.FindStringSubmatch(input.Text())
	if matches != nil {
		c.Name = matches[1]
		c.Value = matches[2]
		return c, nil
	}
	return nil, []error{NewError("Ill-formed constraint", input)}
}
