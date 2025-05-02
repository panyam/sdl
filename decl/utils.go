package decl

import (
	"fmt"
	"strconv"
)

// ParseLiteralValue converts a LiteralExpr value string to a basic Go type.
func ParseLiteralValue(lit *LiteralExpr) (any, error) {
	switch lit.Kind {
	case "STRING":
		return lit.Value, nil
	case "INT":
		return strconv.ParseInt(lit.Value, 10, 64)
	case "FLOAT":
		return strconv.ParseFloat(lit.Value, 64)
	case "BOOL":
		return strconv.ParseBool(lit.Value)
	// TODO: case "DURATION":
	default:
		return nil, fmt.Errorf("cannot parse literal kind %s yet", lit.Kind)
	}
}
