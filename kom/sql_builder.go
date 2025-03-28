package kom

import (
	"fmt"
	"log"
	"strings"

	"github.com/weibaohui/kom/utils"
	"github.com/xwb1989/sqlparser"
	"k8s.io/klog/v2"
)

// Sql TODO Insert Update Delete
// Currently supports Select
// Parses SQL into function calls, implementing support for native SQL statements
//
// Example:
// select * from pod where pod.name='?', 'abc'
func (k *Kubectl) Sql(sql string, values ...interface{}) *Kubectl {
	tx := k.getInstance()
	tx.AllNamespace()

	sql = formatSql(sql, values)

	// Add backticks to convert metadata.name to `metadata.name`
	// Many k8s fields are similar to JSON fields and need to be wrapped in backticks
	// to avoid being treated as db.table format
	// sql = NewSqlParse(sql).AddBackticks()

	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		klog.Errorf("Error parsing SQL:%s,%v", sql, err)
		tx.Error = err
		return tx
	}

	var conditions []Condition // Store parsed conditions

	// Assert as *sqlparser.Select type
	selectStmt, ok := stmt.(*sqlparser.Select)
	if !ok {
		log.Fatalf("Not a SELECT statement")
	}
	// Get From clause from Select statement as Resource
	from := sqlparser.String(selectStmt.From)
	gvk := k.Tools().FindGVKByTableNameInApiResources(from)
	if gvk == nil {
		tx.Error = fmt.Errorf("resource %s not found both in api-resource and crd", from)
		klog.V(6).Infof("resource %s not found both in api-resource and crd", from)
		names := k.Tools().ListAvailableTableNames()
		klog.V(6).Infof("Available resource: %s", names)
		return tx
	}

	// Set GVK
	tx.GVK(gvk.Group, gvk.Version, gvk.Kind)

	// Get LIMIT clause information
	limit := selectStmt.Limit
	if limit != nil {
		// Get Rowcount and Offset from LIMIT
		rowCount := sqlparser.String(limit.Rowcount)
		offset := sqlparser.String(limit.Offset)

		tx.Limit(utils.ToInt(rowCount))
		tx.Offset(utils.ToInt(offset))
	}
	// Parse Where clause to get execution conditions
	conditions = parseWhereExpr(conditions, 0, "AND", selectStmt.Where.Expr)

	// Detect value types in conditions
	for i, cond := range conditions {
		conditions[i].ValueType, conditions[i].Value = utils.DetectType(cond.Value)
	}
	tx.Statement.Filter.Conditions = conditions

	// Set order fields
	orderBy := selectStmt.OrderBy
	if orderBy != nil {
		tx.Statement.Filter.Order = sqlparser.String(orderBy)
	}

	tx.Statement.Filter.Parsed = true
	return tx
}

func (k *Kubectl) From(tableName string) *Kubectl {
	tx := k.getInstance()
	gvk := k.Tools().FindGVKByTableNameInApiResources(tableName)
	if gvk == nil {
		tx.Error = fmt.Errorf("resource %s not found both in api-resource and crd", tableName)
		klog.V(6).Infof("resource %s not found both in api-resource and crd", tableName)
		names := k.Tools().ListAvailableTableNames()
		klog.V(6).Infof("Available resource: %s", names)
		return tx
	}
	tx.Statement.Filter.From = tableName
	// Set GVK
	tx.GVK(gvk.Group, gvk.Version, gvk.Kind)
	return tx
}
func (k *Kubectl) Where(condition string, values ...interface{}) *Kubectl {
	tx := k.getInstance()
	originalSql := tx.Statement.Filter.Sql
	sql := formatSql(condition, values)

	trimSql := strings.ReplaceAll(sql, " ", "")
	if trimSql == "(())" || trimSql == "()" || trimSql == "" {
		// No content
		return tx
	}
	if originalSql != "" {
		sql = originalSql + " and ( " + sql + " ) "
	} else {
		sql = fmt.Sprintf(" select * from fake where ( %s )", sql)
	}

	// Add backticks to convert metadata.name to `metadata.name`
	// Many k8s fields are similar to JSON fields and need to be wrapped in backticks
	// to avoid being treated as db.table format
	// sql = NewSqlParse(sql).AddBackticks()

	tx.Statement.Filter.Sql = sql

	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		klog.Errorf("Error parsing SQL:%s,%v", sql, err)
		tx.Error = err
		return tx
	}

	var conditions []Condition // Store parsed conditions

	// Assert as *sqlparser.Select type
	selectStmt, ok := stmt.(*sqlparser.Select)
	if !ok {
		klog.Errorf("not select parsing SQL:%s,%v", sql, err)
		tx.Error = err
		return tx
	}

	// Parse Where clause to get execution conditions
	conditions = parseWhereExpr(conditions, 0, "AND", selectStmt.Where.Expr)

	// Detect value types in conditions
	for i, cond := range conditions {
		conditions[i].ValueType, conditions[i].Value = utils.DetectType(cond.Value)
	}

	tx.Statement.Filter.Conditions = conditions

	tx.Statement.Filter.Parsed = true

	return tx
}

// formatSql formats SQL with placeholders
// Example:
// select * from pod where pod.name='?', 'abc'
func formatSql(condition string, values []interface{}) string {
	// Replace placeholders (?) in condition with values
	for _, value := range values {
		// Safely format values, e.g., add single quotes for strings
		switch v := value.(type) {
		case string:
			condition = strings.Replace(condition, "?", fmt.Sprintf("'%s'", v), 1)
		default:
			condition = strings.Replace(condition, "?", fmt.Sprintf("%v", v), 1)
		}
	}
	return condition
}

// Order sets the order clause
// Examples:
// Order(" id desc")
// Order(" date asc")
func (k *Kubectl) Order(order string) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Filter.Order = order
	return tx
}
func (k *Kubectl) Limit(limit int) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Filter.Limit = limit
	return tx
}
func (k *Kubectl) Offset(offset int) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Filter.Offset = offset
	return tx
}

// Skip is an alias for Offset
func (k *Kubectl) Skip(skip int) *Kubectl {
	return k.Offset(skip)
}
