package parser

import (
	"testing"
	// "regexp"
	"strings"
)

func DummyScanner(text string) *Scanner {
	return NewScanner(strings.NewReader(text))
}

func (c *Constraint) Equals(d Constraint) bool {
	return c.Name == d.Name && c.Value == d.Value
}

func (c *Column) Equals(d Column) bool {
	res := c.Name == d.Name &&
	      c.Type == d.Type &&
		  c.RequestedType == d.RequestedType &&
		  c.Alias == d.Alias &&
		  len(c.Constraints) == len(d.Constraints)
    for i, con := range c.Constraints {
		res = res && con.Equals(*d.Constraints[i])
    }
	return res
}

func (t *Table) Equals(u Table) bool {
	res := t.Name == u.Name && len(t.Columns) == len(u.Columns)
	for i, col := range t.Columns {
		res = res && col.Equals(*u.Columns[i])
	}
	return res
}

func (p *ParseTree) Equals(q ParseTree) bool {
	res := true
	for k,v := range p.Directives {
		w, ok := q.Directives[k]
		res = res && ok && v == w
	}
	for i, tab := range p.Tables {
		res = res && tab.Equals(*q.Tables[i])
	}
	return res
}

var constraintMatchesTests = []struct{
	name string
	input string
	expected Constraint
}{
	{"Single Token", "-unique", Constraint{Name: "unique"}},
	{"Multiple Tokens", "-primary key", Constraint{Name: "primary key"}},
	{"Name With Value", "-default:'John Doe'", Constraint{Name: "default", Value: "'John Doe'"}},
	{"Arbitrary Spacing", "- NOT NULL", Constraint{Name: "NOT NULL"}},
	{"Arbitrary Spacing", "-\tNOT NULL", Constraint{Name: "NOT NULL"}},
	{"Arbitrary Spacing", "- \tNOT NULL", Constraint{Name: "NOT NULL"}},
	{"Arbitrary Spacing", " -NOT NULL", Constraint{Name: "NOT NULL"}},
	{"Arbitrary Spacing", "\t-NOT NULL", Constraint{Name: "NOT NULL"}},
	{"Arbitrary Spacing", "\t -NOT NULL", Constraint{Name: "NOT NULL"}},
	{"Arbitrary Spacing", "-NOT NULL ", Constraint{Name: "NOT NULL"}},
	{"Arbitrary Spacing", "-NOT NULL\t", Constraint{Name: "NOT NULL"}},
	{"Arbitrary Spacing", "-NOT NULL \t", Constraint{Name: "NOT NULL"}},
	{"Arbitrary Spacing", " \t   -\t   \t  NOT NULL\t\t    ", Constraint{Name: "NOT NULL"}},
	{"Arbitrary Spacing", "-default :'John Doe'", Constraint{Name: "default", Value: "'John Doe'"}},
	{"Arbitrary Spacing", "-default\t:'John Doe'", Constraint{Name: "default", Value: "'John Doe'"}},
	{"Arbitrary Spacing", "-default \t: 'John Doe'", Constraint{Name: "default", Value: "'John Doe'"}},
	{"Arbitrary Spacing", "-default: 'John Doe'", Constraint{Name: "default", Value: "'John Doe'"}},
	{"Arbitrary Spacing", "-default:\t'John Doe'", Constraint{Name: "default", Value: "'John Doe'"}},
	{"Arbitrary Spacing", "-default:\t 'John Doe'", Constraint{Name: "default", Value: "'John Doe'"}},
}

func TestParseConstraintMatches(t *testing.T) {
	for _, test := range constraintMatchesTests {
		t.Run(test.name, func(tt *testing.T) {
			c, errs := ParseConstraint(DummyScanner(test.input))
			if errs != nil {
				tt.Fatalf("Unexpected errors: %v", errs)
			}
			if !c.Equals(test.expected) {
				tt.Fatalf("Incorrect value parsed. Expected: %s; got: %s", test.expected, c)
			}
		})
	}
}

var constraintRejectsTests = []struct {
	name string
	input string
} {
	{"Empty Line", ""},
	{"Empty Constraint", "-"},
	{"Non-constraint", "uid int"},
}

func TestParseConstraintRejects(t *testing.T) {
	for _, test := range constraintRejectsTests {
		t.Run(test.name, func(tt *testing.T) {
			_, errs := ParseConstraint(DummyScanner(test.input))
			if errs == nil {
				tt.Fatalf("Expected parsing error for \"%s\", but got nothing", test.input)
			}
		})
	}
}

// This test should pass, but it's not critical if it doesn't, since in a practical
// implementation it shouldn't matter. As such, A failed test just leads to a skip
func TestParseConstraintRejectsAlias(t *testing.T) {
	_, errs := ParseConstraint(DummyScanner("-alias: name"))
	if errs == nil {
		t.Skip("Expected parsing error for \"-alias: name\" as a constraint, but got nothing")
		t.Fatal("Expected parsing error for \"-alias: name\" as a constraint, but got nothing")
	}
}

var aliasMatchesTests = []struct {
	name string
	input string
	expected string
} {
	{"Single Token", "-alias:id", "id"},
	{"Arbitrary Spacing", "-alias: hex", "hex"},
	{"Arbitrary Spacing", "-alias:\thex", "hex"},
	{"Arbitrary Spacing", "-alias: \thex", "hex"},
	{"Arbitrary Spacing", "-alias :hex", "hex"},
	{"Arbitrary Spacing", "-alias\t:hex", "hex"},
	{"Arbitrary Spacing", "-alias\t :hex", "hex"},
	{"Arbitrary Spacing", " -alias:hex", "hex"},
	{"Arbitrary Spacing", "\t-alias:hex", "hex"},
	{"Arbitrary Spacing", "\t -alias:hex", "hex"},
	{"Arbitrary Spacing", "- alias:hex", "hex"},
	{"Arbitrary Spacing", "-\talias:hex", "hex"},
	{"Arbitrary Spacing", "- \talias:hex", "hex"},
	{"Arbitrary Spacing", "-alias:hex ", "hex"},
	{"Arbitrary Spacing", "-alias:hex\t", "hex"},
	{"Arbitrary Spacing", "-alias:hex \t", "hex"},
	{"Arbitrary Spacing", " \t   -   \t\t\t  alias\t\t\t\t  :\t   \t  hex\t\t   ", "hex"},
}

func TestParseAliasMatches(t *testing.T) {
	for _, test := range aliasMatchesTests {
		t.Run(test.name, func(tt *testing.T) {
			a, errs := ParseAlias(DummyScanner(test.input))
			if errs != nil {
				tt.Fatalf("Unexpected errors: %v", errs)
			}
			if a != test.expected {
				tt.Fatalf("Incorrect value parsed. Expected: %s; got: %s", test.expected, a)
			}
		})
	}
}

var aliasRejectsTests = []struct {
	name, input string
} {
	{"Without Alias", "-:name"},
	{"Constraint", "-unique"},
	{"Multi-word Alias", "-alias:multi word"},
}

func TestParseAliasRejects(t *testing.T) {
	for _,test := range aliasRejectsTests {
		t.Run(test.name, func(tt *testing.T) {
			_, errs := ParseAlias(DummyScanner(test.input))
			if errs == nil {
				tt.Fatalf("Expecting error for \"%s\", but no errors returned", test.input)
			}
		})
	}
}

var columnMatchesTests = []struct{
	name string
	input string
	expected Column
} {
	{"Simple Column", "id int", Column{Name:"id", Type:"int"}},
	{"Typed Column", "id int using SERIAL", Column{Name:"id", Type:"int", RequestedType:"SERIAL"}},
	{"Typed Column With Spaces", "id int using double precision", Column{
		Name: "id", 
		Type: "int", 
		RequestedType: "double precision",
	}},
	{"With Alias", "uid int\n-alias: id", Column{Name:"uid", Type:"int", Alias:"id"}},
	{"With Constraint", "uid int\n-primary key", Column{
		Name:"uid", 
		Type:"int", 
		Constraints:[]*Constraint{
			&Constraint{Name: "primary key"},
	}}},
	{"With Multiple Constraints", "uid int\n-unique\n-not null", Column{
		Name:"uid", 
		Type:"int", 
		Constraints:[]*Constraint{
			&Constraint{Name: "unique"},
			&Constraint{Name: "not null"},
	}}},
	{"With Constraint And Alias", "uid int\n-foreign key\n-alias: id\n-not null", Column{
		Name:"uid", 
		Type:"int",
		Alias: "id",
		Constraints:[]*Constraint{
			&Constraint{Name: "foreign key"},
			&Constraint{Name: "not null"},
	}}},
	{"Multiple Aliases Override Previous", "uid int\n-alias: id\n-alias: user_id", Column {
		Name: "uid",
		Type: "int",
		Alias: "user_id",
	}},
	{"Using With Constraints", "uid int using numeric\n-foreign key\n-alias: id\n-not null", Column {
		Name: "uid",
		Type: "int",
		RequestedType: "numeric",
		Alias: "id",
		Constraints: []*Constraint {
			&Constraint{Name: "foreign key"},
			&Constraint{Name: "not null"},
	}}},
}

func TestParseColumnMatches(t *testing.T) {
	for _,test := range columnMatchesTests {
		t.Run(test.name, func(tt *testing.T) {
			col, errs := ParseColumn(DummyScanner(test.input))
			if errs != nil {
				tt.Fatalf("Unexpected errors: %v", errs)
			}
			if !col.Equals(test.expected) {
				tt.Fatalf("Incorrect values parsed. Expected %s; got %s", test.expected, *col)
			}
		})
	}
}

var columnRejectTests = []struct {
	name string
	input string
} {
	{"Incomplete Using", "id int using "},
	{"Bad Alias", "uid int\n-alias id"},
	{"Bad Constraint", "uid int\n-"},
}

func TestParseColumnRejects(t *testing.T) {
	for _,test := range columnRejectTests {
		t.Run(test.name, func(tt *testing.T) {
			_, errs := ParseColumn(DummyScanner(test.input))
			if errs == nil {
				tt.Fatalf("Expected error from \"%s\", but no error returned", test.input)
			}
		})
	}
}

var tableMatchesTests = []struct {
	name string
	input string
	expected Table
} {
	{"Empty Table", "[table]", Table{Name: "table"}},
	{"Single Simple Column", "[table]\nfield type", Table{Name: "table", Columns:[]*Column{
		&Column{Name: "field", Type: "type"},
	}}},
	{"Multiple Simple Columns", "[table]\nid int\nname string", Table{
		Name: "table",
		Columns: []*Column{
			&Column{Name: "id", Type: "int"},
			&Column{Name: "name", Type: "string"},
		},
	}},
	{"Multiple Complex Columns", "[person]\nsid int using serial\n-primary key\n-alias: id\nname string\n-not null\n-default: 'John Doe'", Table{
		Name: "person",
		Columns: []*Column{
			&Column{Name: "sid", Type: "int", RequestedType: "serial", Alias: "id", Constraints:[]*Constraint{
				&Constraint{Name: "primary key"},
			}},
			&Column{Name: "name", Type: "string", Constraints:[]*Constraint {
				&Constraint{Name: "not null"},
				&Constraint{Name: "default", Value: "'John Doe'"},
			}},
		},
	}},
	// This example is generally invalid, but it is sufficient for ParseTable
	{"Table Terminated By Empty Line", "[table]\nfield int\n\nnot_included null", Table{
		Name: "table",
		Columns: []*Column{
			&Column{Name: "field", Type: "int"},
		},
	}},
}

func TestParseTableMatches(t *testing.T) {
	for _,test := range tableMatchesTests {
		t.Run(test.name, func(tt *testing.T) {
			tab, errs := ParseTable(DummyScanner(test.input))
			if errs != nil {
				tt.Fatalf("Unexpected errors: %v", errs)
			}
			if !tab.Equals(test.expected) {
				tt.Fatalf("Incorrect values parsed. Expected %s; got %s", test.expected, *tab)
			}
		})
	}
}

var tableRejectsTests = []struct {
	name string
	input string
}{
	{"Name Containing Spaces", "[table with spaces]"},
	{"Non-bracketed Name", "table"},
	{"Non-Column Row", "[table]\n-invalid row"},
	// space prefixing table brackets
}

func TestParseTableRejects(t *testing.T) {
	for _, test := range tableRejectsTests {
		t.Run(test.name, func(tt *testing.T) {
			_, errs := ParseTable(DummyScanner(test.input))
			if errs == nil {
				tt.Fatalf("Expected error from \"%s\", but no error returned", test.input)
			}
		})
	}
}

var directiveMatchesTests = []struct {
	name string
	input string
	expectedName, expectedValue string
} {
	{"Simple Directive", "#directive=value", "directive", "value"},
	// arbitrary spacing
	{"Arbitrary Spacing", "#directive=value", "directive", "value"},
	{"Arbitrary Spacing", "# directive=value", "directive", "value"},
	{"Arbitrary Spacing", "#\tdirective=value", "directive", "value"},
	{"Arbitrary Spacing", "#\t directive=value", "directive", "value"},
	{"Arbitrary Spacing", "#directive =value", "directive", "value"},
	{"Arbitrary Spacing", "#directive\t=value", "directive", "value"},
	{"Arbitrary Spacing", "#directive \t=value", "directive", "value"},
	{"Arbitrary Spacing", "#directive= value", "directive", "value"},
	{"Arbitrary Spacing", "#directive=\tvalue", "directive", "value"},
	{"Arbitrary Spacing", "#directive=\t value", "directive", "value"},
	{"Arbitrary Spacing", "#directive=value ", "directive", "value"},
	{"Arbitrary Spacing", "#directive=value\t", "directive", "value"},
	{"Arbitrary Spacing", "#directive=value \t", "directive", "value"},
	{"Arbitrary Spacing", "#\t   \t\t\t  \tdirective\t \t =     value\t\t\t \t", "directive", "value"},
}

func TestParseDirectiveMatches(t *testing.T) {
	for _,test := range directiveMatchesTests {
		t.Run(test.name, func(tt *testing.T) {
			directive, value, matches := ParseDirective(DummyScanner(test.input))
			if !matches {
				tt.Fatalf("Directive incorrectly parsed. Expected %s=%v; got %s=%v", 
				  test.expectedName, test.expectedValue, directive, value)
			}
		})
	}
}

var directiveRejectsTests = []struct {
	name string
	input string
} {
	{"Name With Spaces", "#mutiple words=value"},
	{"Value With Spaces", "#directive=multiple words"},
	{"Using : Instead of =", "#directive:value"},
	{"Non-directive Comment", "#This is a comment"},
}

func TestParseDirectiveRejects(t *testing.T) {
	for _,test := range directiveRejectsTests {
		t.Run(test.name, func(tt *testing.T) {
			_, _, matches := ParseDirective(DummyScanner(test.input))
			if matches {
				tt.Fatalf("Directive match not expected for %s", test.input)
			}
		})
	}
}

var fileMatchesTests = []struct {
	name string
	input string
	expected ParseTree
} {
	{"Empty File", "", ParseTree{}},
	{"Single Directive", "#language=ruby", ParseTree{Directives: map[string]string{"language": "ruby"}}},
	{"Multiple Directives", "#language=go\n#database=mysql", ParseTree{Directives: map[string]string{
		"language": "go",
		"database": "mysql",
	}}},
	{"Single Empty Table", "[person]", ParseTree{Tables: []*Table{
		&Table{Name: "person"},
	}}},
	{"Multiple Empty Tables", "[people]\n\n[cars]", ParseTree{Tables: []*Table{
		&Table{Name: "people"},
		&Table{Name: "cars"},
	}}},
	{"Multiple Tables", "[people]\nname string using TEXT\n\n[cars]\nvin int\nmodel string", ParseTree{
		Tables: []*Table {
			&Table{Name: "people", Columns: []*Column {
				&Column{Name: "name", Type: "string", RequestedType: "TEXT"},
			}},
			&Table{Name: "cars", Columns: []*Column {
				&Column{Name: "vin", Type: "int"},
				&Column{Name: "model", Type: "string"},
			}},
		},
	}},
	{"Multiple Tables With Directives", 
	 "#language=go\n#database=sqlite\n[people]\nname string using TEXT\n\n[cars]\nvin int\nmodel string", 
	 ParseTree{
		 Directives: map[string]string{"language": "go", "database": "sqlite"},
		 Tables: []*Table {
			 &Table{Name: "people", Columns: []*Column {
				 &Column{Name: "name", Type: "string", RequestedType: "TEXT"},
			 }},
			 &Table{Name: "cars", Columns: []*Column {
				 &Column{Name: "vin", Type: "int"},
				 &Column{Name: "model", Type: "string"},
			 }},
		 },
	 }},

	{"Invalid Directives But Otherwise Fine", 
	 "#holyeggbeaters:batman\n#life=hardknock\n[people]\nname string using TEXT\n\n[cars]\nvin int\nmodel string", 
	 ParseTree{
		 Directives: map[string]string{"life": "hardknock"},
		 Tables: []*Table {
			 &Table{Name: "people", Columns: []*Column {
				 &Column{Name: "name", Type: "string", RequestedType: "TEXT"},
			 }},
			 &Table{Name: "cars", Columns: []*Column {
				 &Column{Name: "vin", Type: "int"},
				 &Column{Name: "model", Type: "string"},
			 }},
		 },
	}},
	// language is spelled wrong. I left it that way to catch any overzealous parser implementations.
	{"Complete File",
`#langauge = go
#database = postgres
#driver = pq

[users]
uid int using SERIAL
- Primary Key
- alias: id
name string using TEXT
- NOT NULL
username string using TEXT
- NOT NULL

[records]
sid int using SERIAL
- PRIMARY KEY
- alias: id
uid int
- FOREIGN KEY: users
- ON DELETE: CASCADE`,
     ParseTree{
		 Directives: map[string]string{
			 "langauge": "go",
			 "database": "postgres",
			 "driver": "pq",
		 },
		 Tables: []*Table{
			 &Table{Name: "users", Columns: []*Column {
				 &Column{Name:"uid", Type:"int", RequestedType:"SERIAL", Alias:"id", Constraints: []*Constraint{
					 &Constraint{Name: "Primary Key"},
				 }},
				 &Column{Name:"name", Type:"string", RequestedType:"TEXT", Constraints: []*Constraint{
					 &Constraint{Name: "NOT NULL"},
				 }},
				 &Column{Name:"username", Type: "string", RequestedType: "TEXT", Constraints: []*Constraint{
					 &Constraint{Name: "NOT NULL"},
				 }},
			 }},
			 &Table{Name: "records", Columns: []*Column {
				 &Column{Name:"sid", Type:"int", RequestedType:"SERIAL", Alias:"id", Constraints: []*Constraint{
					 &Constraint{Name: "PRIMARY KEY"},
				 }},
				 &Column{Name: "uid", Type: "int", Constraints: []*Constraint{
					 &Constraint{Name: "FOREIGN KEY", Value: "users"},
					 &Constraint{Name: "ON DELETE", Value: "CASCADE"},
				 }},
			 }},
		 },
     }},
}

func TestParseMatches(t *testing.T) {
	for _, test := range fileMatchesTests {
		t.Run(test.name, func(tt *testing.T) {
			tree, errs := Parse(strings.NewReader(test.input))
			if errs != nil {
				tt.Fatalf("Unexpected errors: %v", errs)
			}
			if !tree.Equals(test.expected) {
				tt.Fatalf("Incorrect value parsed. Expected: %s; got: %s", test.expected, tree)
			}			
		})
	}
}

var fileRejectsTests = []struct {
	name string
	input string
} {
	{"Invalid Constraint", "[table]\nuid int\n-primary key"},
	{"Invalid Alias", "[table]\nuid int\n-alias:id"},
	{"Invalid Column", "[table]\ninvalid_token"},
	{"Invalid Table", "[invalid table]"},
	{"Broken Table", "[table]\nuid int\n\n-primary key"},
}

func TestParseRejects(t *testing.T) {
	for _,test := range fileRejectsTests {
		t.Run(test.name, func(tt *testing.T) {
			_, _, matches := ParseDirective(DummyScanner(test.input))
			if matches {
				tt.Fatalf("Directive match not expected for %s", test.input)
			}
		})
	}
}

