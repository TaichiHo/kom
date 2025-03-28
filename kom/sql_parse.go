package kom

import (
	"fmt"
	"reflect"

	"github.com/weibaohui/kom/utils"
	"github.com/xwb1989/sqlparser"
	"k8s.io/klog/v2"
)

// Parse WHERE expression
func parseWhereExpr(conditions []Condition, depth int, andor string, expr sqlparser.Expr) []Condition {
	klog.V(6).Infof("expr type [%v],string %s, type [%s]", reflect.TypeOf(expr), sqlparser.String(expr), andor)
	d := depth + 1 // Increment depth
	switch node := expr.(type) {
	case *sqlparser.ComparisonExpr:
		// Handle comparison expressions (e.g., age > 80)
		cond := Condition{
			Depth:    depth,
			AndOr:    andor,
			Field:    utils.TrimQuotes(sqlparser.String(node.Left)),
			Operator: node.Operator,
			Value:    utils.TrimQuotes(sqlparser.String(node.Right)),
		}
		conditions = append(conditions, cond)
	case *sqlparser.ParenExpr:
		// Handle parentheses expressions
		// Expression inside parentheses is an independent sub-expression, increase depth
		conditions = parseWhereExpr(conditions, d+1, "AND", node.Expr)

	case *sqlparser.AndExpr:
		// Recursively parse AND expressions
		// Pass "AND" to both sides
		conditions = parseWhereExpr(conditions, d, "AND", node.Left)
		conditions = parseWhereExpr(conditions, d, "AND", node.Right)
	case *sqlparser.RangeCond:
		// Parse "between 1 and 3" expressions
		cond := Condition{
			Depth:    depth,
			AndOr:    andor,
			Field:    utils.TrimQuotes(sqlparser.String(node.Left)),                                                                        // Left field
			Operator: node.Operator,                                                                                                        // Operator (BETWEEN)
			Value:    fmt.Sprintf("%s and %s", utils.TrimQuotes(sqlparser.String(node.From)), utils.TrimQuotes(sqlparser.String(node.To))), // Range value
		}
		conditions = append(conditions, cond)
	case *sqlparser.OrExpr:
		// Recursively parse OR expressions
		// Pass "OR" to both sides
		conditions = parseWhereExpr(conditions, d, "OR", node.Left)
		conditions = parseWhereExpr(conditions, d, "OR", node.Right)

	default:
		// Other expressions
		fmt.Printf("Unhandled expression at depth %d: %s\n", depth, sqlparser.String(expr))
	}
	return conditions
}
