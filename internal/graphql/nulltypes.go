package graphql

import (
	"strconv"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"gopkg.in/guregu/null.v4"
)

// SerializeNullString serializes `NullString` to a string
func SerializeNullString(value interface{}) interface{} {
	switch value := value.(type) {
	case null.String:
		return value.String
	case *null.String:
		v := *value
		return v.String
	default:
		return nil
	}
}

// SerializeNullInt serializes `NullInt` to a int
func SerializeNullInt(value interface{}) interface{} {
	switch value := value.(type) {
	case null.Int:
		return value.Int64
	case *null.Int:
		v := *value
		return v.Int64
	default:
		return nil
	}
}

// ParseNullString parses GraphQL variables from `string`
func ParseNullString(value interface{}) interface{} {
	switch value := value.(type) {
	case string:
		return null.StringFrom(value)
	case *string:
		return null.StringFromPtr(value)
	default:
		return nil
	}
}

// ParseNullInt parses GraphQL variables from `int`
func ParseNullInt(value interface{}) interface{} {
	switch value := value.(type) {
	case int64:
		return null.IntFrom(value)
	case *int64:
		return null.IntFromPtr(value)
	default:
		return nil
	}
}

// ParseLiteralNullString parses GraphQL AST value to `NullString`.
func ParseLiteralNullString(valueAST ast.Value) interface{} {
	switch valueAST := valueAST.(type) {
	case *ast.StringValue:
		return null.StringFrom(valueAST.Value)
	default:
		return nil
	}
}

// ParseLiteralNullInt parses GraphQL AST value to `NullInt`.
func ParseLiteralNullInt(valueAST ast.Value) interface{} {
	switch valueAST := valueAST.(type) {
	case *ast.IntValue:
		n, err := strconv.ParseInt(valueAST.Value, 10, 64)
		if err != nil {
			return nil
		}
		return null.IntFrom(n)
	default:
		return nil
	}
}

// NullableString graphql *Scalar type based of NullString
var NullableString = graphql.NewScalar(graphql.ScalarConfig{
	Name:         "NullableString",
	Description:  "The `NullableString` type repesents a nullable SQL string.",
	Serialize:    SerializeNullString,
	ParseValue:   ParseNullString,
	ParseLiteral: ParseLiteralNullString,
})

// NullableString graphql *Scalar type based of NullString
var NullableInt = graphql.NewScalar(graphql.ScalarConfig{
	Name:         "NullableInt",
	Description:  "The `NullableInt` type repesents a nullable SQL int.",
	Serialize:    SerializeNullInt,
	ParseValue:   ParseNullInt,
	ParseLiteral: ParseLiteralNullInt,
})
