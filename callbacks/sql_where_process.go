package callbacks

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

// executeFilter executes filtering using lancet
func executeFilter(result []unstructured.Unstructured, conditions []kom.Condition) []unstructured.Unstructured {

	// Final result, execute filtering one by one according to the condition list

	// 2. Group by and/or
	groupedConditions := groupByOperator(conditions)

	// 3. Currently changed to support simple and/or logic without considering parentheses.
	// TODO Need to handle complex logic cases with parentheses
	// keys and or
	keys := make([]string, 0, len(groupedConditions))
	for k := range groupedConditions {
		keys = append(keys, k)
	}

	// 3. Process map by visiting two groups of and/or
	for _, key := range keys {
		// This key == and or
		group := slice.Filter(conditions, func(index int, item kom.Condition) bool {
			return item.AndOr == key
		})
		// Filter by group, generally one group has the same and/or conditions
		result = evaluateCondition(result, group)
	}
	return result
}

func groupByDepth(conditions []kom.Condition) map[int][]kom.Condition {
	groups := make(map[int][]kom.Condition)
	for _, p := range conditions {
		depth := p.Depth
		groups[depth] = append(groups[depth], p)
	}
	return groups
}
func groupByOperator(conditions []kom.Condition) map[string][]kom.Condition {
	groups := make(map[string][]kom.Condition)
	for _, p := range conditions {
		operator := p.AndOr
		groups[operator] = append(groups[operator], p)
	}
	return groups
}
func evaluateCondition(result []unstructured.Unstructured, group []kom.Condition) []unstructured.Unstructured {

	// 一般一组 and or 都相同
	if group[0].AndOr == "OR" {
		return matchAny(result, group)
	} else {
		return matchAll(result, group)
	}
}

// matchAll checks if all conditions are met (AND logic)
func matchAll(result []unstructured.Unstructured, conditions []kom.Condition) []unstructured.Unstructured {
	return slice.Filter(result, func(index int, item unstructured.Unstructured) bool {
		// Iterate through all conditions, return true only if all conditions are met
		for _, c := range conditions {
			condition := matchCondition(item, c)
			klog.V(8).Infof("matchAny %s/%s  %s  %s  %s = %v", item.GetNamespace(), item.GetName(), c.Field, c.Operator, c.Value, condition)
			if !condition {
				return false // Return false immediately if any condition is not met
			}
		}
		return true // Return true if all conditions are met
	})
}

// matchAny checks if any condition is met (OR logic)
func matchAny(result []unstructured.Unstructured, conditions []kom.Condition) []unstructured.Unstructured {
	return slice.Filter(result, func(index int, item unstructured.Unstructured) bool {
		// Iterate through all conditions, return true if any condition is met
		for _, c := range conditions {
			condition := matchCondition(item, c)
			klog.V(8).Infof("matchAny %s/%s  %s  %s  %s = %v", item.GetNamespace(), item.GetName(), c.Field, c.Operator, c.Value, condition)
			if condition {
				// Return if any condition is met.
				// Don't return if not met, let it execute the next condition
				return true
			}
		}
		return false
	})
}

// matchCondition checks if a single condition matches
// If the field to be processed is a value rather than a list, compare directly.
// If the field to be processed corresponds to a k8s list, such as status.addresses[type=InternalIP].address
// status.addresses is an array, status.addresses[type=InternalIP].address then the extracted address is an array, but filtered by type
// status.addresses is an array, status.addresses.address then the extracted address is also an array, not filtered by type
// status:
//
//	addresses:
//	    - address: 172.18.0.2
//	      type: InternalIP
//	    - address: kind-control-plane
//	      type: Hostname
//
// For positive operators (like =, like, in, between), return true if any matching value is found.
// For negative operators (like !=, not in, not between), return true only if no values match.
func matchCondition(resource unstructured.Unstructured, condition kom.Condition) bool {
	klog.V(6).Infof("matchCondition  %s %s %s", condition.Field, condition.Operator, condition.Value)

	// Get field value
	fieldValues, found, err := getNestedFieldAsString(resource.Object, condition.Field)
	if err != nil || !found {
		klog.V(6).Infof("not found %s,%v", condition.Field, err)
		return false
	}

	// If the obtained value is a single value, not a list, process directly
	if len(fieldValues) == 1 {
		fieldValue := fieldValues[0]
		switch condition.Operator {
		case "=":
			if compareValue(fieldValue, condition.Value) {
				return true
			}
		case "!=":
			if compareValue(fieldValue, condition.Value) {
				return false // For negative condition, return false if a match is found
			}
		case "like":
			if compareLike(fieldValue, condition.Value) {
				return true
			}
		case "in":
			if compareIn(fieldValue, condition.Value) {
				return true
			}
		case "not in":
			if compareIn(fieldValue, condition.Value) {
				return false
			}
		case ">":
			if compareGreater(fieldValue, condition.Value) {
				return true
			}
		case "<":
			if compareLess(fieldValue, condition.Value) {
				return true
			}
		case ">=":
			if compareGreaterOrEqual(fieldValue, condition.Value) {
				return true
			}
		case "<=":
			if compareLessOrEqual(fieldValue, condition.Value) {
				return true
			}
		case "between":
			if compareBetween(fieldValue, condition.Value) {
				return true
			}
		case "not between":
			if compareBetween(fieldValue, condition.Value) {
				return false
			}
		default:
			return false
		}
	}

	// If the obtained value is a list, it's a list property in yaml, so we need to think comprehensively

	// Determine if it's a positive or negative condition
	isNegativeCondition := condition.Operator == "!=" || condition.Operator == "not in" || condition.Operator == "not between"

	// Process each field value
	for _, fieldValue := range fieldValues {
		// For negative conditions (!=, not in, not between), need to ensure all values don't meet the condition
		switch condition.Operator {
		case "=":
			if !isNegativeCondition && compareValue(fieldValue, condition.Value) {
				return true // For positive condition, return true if any value matches
			}
		case "!=":
			if isNegativeCondition && compareValue(fieldValue, condition.Value) {
				return false // For negative condition, return false if any value matches
			}
		case "like":
			if !isNegativeCondition && compareLike(fieldValue, condition.Value) {
				return true
			}
		case "in":
			if !isNegativeCondition && compareIn(fieldValue, condition.Value) {
				return true
			}
		case "not in":
			if isNegativeCondition && compareIn(fieldValue, condition.Value) {
				return false
			}
		case ">":
			if !isNegativeCondition && compareGreater(fieldValue, condition.Value) {
				return true
			}
		case "<":
			if !isNegativeCondition && compareLess(fieldValue, condition.Value) {
				return true
			}
		case ">=":
			if !isNegativeCondition && compareGreaterOrEqual(fieldValue, condition.Value) {
				return true
			}
		case "<=":
			if !isNegativeCondition && compareLessOrEqual(fieldValue, condition.Value) {
				return true
			}
		case "between":
			if !isNegativeCondition && compareBetween(fieldValue, condition.Value) {
				return true
			}
		case "not between":
			if isNegativeCondition && compareBetween(fieldValue, condition.Value) {
				return false
			}
		default:
			return false
		}
	}

	// For negative conditions: only consider condition not met if all values don't meet the condition
	if isNegativeCondition {
		return true // Return true if all values don't meet the condition
	}

	// Default return false, indicating no value meets the condition
	return false
}

// compareValue 比较值是否相等，不区分大小写
func compareValue(fieldValue string, value interface{}) bool {
	klog.V(8).Infof("compareValue (=) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))

	switch v := value.(type) {
	case string:
		return strings.ToLower(fieldValue) == strings.ToLower(v)
	case float64, int, int64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat == v
	case time.Time:
		// 需要将 fieldValue 转换为 time 类型进行比较
		fieldTime, err := time.Parse(time.RFC3339, fieldValue)
		if err != nil {
			return false
		}
		return fieldTime.Equal(v)
	default:
		return false
	}
}

// compareLike 判断字符串是否匹配,不区分大小写
func compareLike(fieldValue string, value interface{}) bool {
	klog.V(6).Infof("compareLike (like) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
	// 处理为小写
	fieldValue = strings.ToLower(fieldValue)

	targetValue := fmt.Sprintf("%v", value)

	// 提取值
	val := strings.TrimPrefix(targetValue, "%")
	val = strings.TrimSuffix(val, "%")
	// 处理为小写
	val = strings.ToLower(val)

	// 判断是否包含%
	if strings.HasSuffix(targetValue, "%") && strings.HasPrefix(targetValue, "%") {
		// 以%开头，以%结尾，表示包含即可
		return strings.Contains(fieldValue, val)

	} else if strings.HasSuffix(targetValue, "%") {
		// abc%， 只以%结尾，表示开头必须是abc
		return strings.HasPrefix(fieldValue, val)
	} else if strings.HasPrefix(targetValue, "%") {
		// %abc 只以%开头，表示结尾必须是abc
		return strings.HasSuffix(fieldValue, val)
	} else {
		// 不包含%，表示必须相等
		return fieldValue == val
	}
}

// compareGreater 比较数值是否大于
func compareGreater(fieldValue string, value interface{}) bool {
	klog.V(6).Infof("compareGreater(>) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
	switch v := value.(type) {
	case float64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat > v
	case int, int64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat > float64(v.(int))
	case string:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		valueFloat, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		return fieldValFloat > valueFloat
	case time.Time:
		fieldValTime, err := utils.ParseTime(fieldValue)
		if err != nil {
			return false
		}

		return fieldValTime.After(v)
	default:
		klog.V(6).Infof("%s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
		return false
	}
}

// compareLess 比较数值是否小于
func compareLess(fieldValue string, value interface{}) bool {
	klog.V(6).Infof("compareLess(<) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))

	switch v := value.(type) {
	case float64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat < v
	case int, int64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat < float64(v.(int))
	case string:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		valueFloat, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		return fieldValFloat < valueFloat
	case time.Time:
		fieldValTime, err := utils.ParseTime(fieldValue)
		if err != nil {
			return false
		}
		return fieldValTime.Before(v)
	default:
		klog.V(6).Infof("%s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
		return false
	}
}

// compareGreaterOrEqual 比较数值是否大于或等于
func compareGreaterOrEqual(fieldValue string, value interface{}) bool {
	klog.V(6).Infof("compareGreaterOrEqual(>=) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
	switch v := value.(type) {
	case float64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat >= v
	case int, int64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat >= float64(v.(int))
	case string:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		valueFloat, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		return fieldValFloat >= valueFloat
	case time.Time:
		fieldValTime, err := utils.ParseTime(fieldValue)
		if err != nil {
			return false
		}
		return fieldValTime.After(v) || fieldValTime.Equal(v)
	default:
		klog.V(6).Infof("%s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
		return false
	}

}

// compareLessOrEqual 比较数值是否小于或等于
func compareLessOrEqual(fieldValue string, value interface{}) bool {
	klog.V(6).Infof("compareLessOrEqual(<=) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))

	switch v := value.(type) {
	case float64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat <= v
	case int, int64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat <= float64(v.(int))
	case string:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		valueFloat, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		return fieldValFloat <= valueFloat
	case time.Time:
		fieldValTime, err := utils.ParseTime(fieldValue)
		if err != nil {
			return false
		}
		return fieldValTime.Before(v) || fieldValTime.Equal(v)
	default:
		klog.V(6).Infof("%s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
		return false
	}
}

// compareIn 判断值是否在列表中
func compareIn(fieldValue string, value interface{}) bool {

	klog.V(6).Infof("compareIn(in []) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))

	// value 类型字符串 = (1,2,3,4) ，这个格式决定了只能是string类型
	// 如何判断fieldValue 是否在1,2,3,4范围内?
	if str, ok := value.(string); ok {
		// 去掉首尾的括号
		str = strings.TrimPrefix(str, "(")
		str = strings.TrimSuffix(str, ")")
		// 以逗号分割
		values := strings.Split(str, ",")
		for _, v := range values {
			v = strings.Trim(v, " ")
			v = utils.TrimQuotes(v)
			// 时间、字符串、数字
			// 只有相等，才能返回，因为in操作符，是or的关系。一个不行，需要判断下一个。

			// 先按数字比较
			fieldValueNum, err1 := strconv.ParseFloat(fieldValue, 64)
			toNum, err2 := strconv.ParseFloat(v, 64)
			if err1 == nil && err2 == nil {
				if fieldValueNum == toNum {
					return true
				}
			}

			// 时间不能简单判断，而要判断是否日期、小时、分钟，是否in。
			// 是否包含时间部分，如果包含，就是精确匹配。如果不不含，就是判断日期
			fieldValueTime, err1 := utils.ParseTime(fieldValue)
			toTime, err2 := utils.ParseTime(v)
			if err1 == nil && err2 == nil {

				// 判断目标时间字符串是否包含时间部分（即时分秒）
				if hasTimeComponent(v) {
					// 逐级比较时间分量（小时、分钟、秒）
					if fieldValueTime.Hour() == toTime.Hour() &&
						fieldValueTime.Minute() == toTime.Minute() &&
						fieldValueTime.Second() == toTime.Second() {
						return true
					}
				}
				// 比较日期部分（年、月、日）
				if isSameDate(fieldValueTime, toTime) {
					return true
				}
			}

			if fieldValue == v {
				return true
			}

		}
	}
	return false
}

// 判断是否包含时间部分
func hasTimeComponent(value string) bool {
	return strings.Contains(value, ":") // 如果字符串中包含冒号，说明有时间部分
}

// 判断两个时间的日期部分是否相同
func isSameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// compareBetween 判断值是否在范围内
func compareBetween(fieldValue string, value interface{}) bool {
	klog.V(6).Infof("compareBetween (between x and y) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))

	// value格式 举例: 1 and 5 这个格式决定了只能是string类型
	// 正则表达式匹配 "from AND to"
	var from, to string
	re := regexp.MustCompile(`(?i)(.+?)\s+AND\s+(.+)`)
	matches := re.FindStringSubmatch(fmt.Sprintf("%v", value))

	// 如果匹配成功，提取出 from 和 to
	if len(matches) == 3 {
		from = matches[1]
		to = matches[2]
	}

	// 判断 from to 是否为时间类型、数字类、还是字符串
	// 数字类型，要做 fieldValue 要转换为对应的类型，并进行>=from <=to的判断
	// 1. 尝试作为数字比较
	if fieldValNum, err := strconv.ParseFloat(fieldValue, 64); err == nil {
		fromNum, err1 := strconv.ParseFloat(from, 64)
		toNum, err2 := strconv.ParseFloat(to, 64)
		if err1 == nil && err2 == nil {
			klog.V(6).Infof("compareBetween(between x and y) as number %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
			return fieldValNum >= fromNum && fieldValNum <= toNum
		} else {
			klog.V(6).Infof("compareBetween(between x and y) as number  error %v %v", err1, err2)
		}

	}

	// 2. 尝试作为时间比较
	if fieldValTime, err := utils.ParseTime(fieldValue); err == nil {
		fromTime, err1 := utils.ParseTime(from)
		toTime, err2 := utils.ParseTime(to)
		if err1 == nil && err2 == nil {
			klog.V(6).Infof("compareBetween(between x and y) as date %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
			return (fieldValTime.Equal(fromTime) || fieldValTime.After(fromTime)) &&
				(fieldValTime.Equal(toTime) || fieldValTime.Before(toTime))
		} else {
			klog.V(6).Infof("compareBetween(between x and y) as date  error %v %v", err1, err2)
		}
	}

	// 3. 作为字符串比较
	return fieldValue >= from && fieldValue <= to
}

// getNestedFieldAsString 获取嵌套字段值，支持数组筛选并处理数组返回值
func getNestedFieldAsString(obj interface{}, path string) ([]string, bool, error) {
	fields, arrayCondition, err := parsePathWithCondition(path)
	if err != nil {
		return nil, false, err
	}
	return getFieldValues(obj, fields, arrayCondition)
}

// parsePathWithCondition 解析路径，支持数组条件筛选
func parsePathWithCondition(path string) ([]string, map[string]string, error) {
	// 用 . 分割路径
	parts := strings.Split(path, ".")
	var arrayCondition map[string]string

	// 检查是否包含条件（如 [type=InternalIP]）
	if len(parts) > 1 && strings.Contains(parts[1], "[") {
		// 提取字段和条件
		conditionPart := parts[1]
		startIndex := strings.Index(conditionPart, "[")
		endIndex := strings.Index(conditionPart, "]")

		// 提取字段和条件
		field := conditionPart[:startIndex]
		condition := conditionPart[startIndex+1 : endIndex]
		conditionParts := strings.Split(condition, "=")

		if len(conditionParts) == 2 {
			arrayCondition = map[string]string{conditionParts[0]: conditionParts[1]}
			// 重新组装路径，去掉条件部分
			parts[1] = field
		}
	}

	return parts, arrayCondition, nil
}

// getFieldValues 递归获取字段值，支持数组筛选并返回多个值
func getFieldValues(obj interface{}, fields []string, arrayCondition map[string]string) ([]string, bool, error) {
	if len(fields) == 0 {
		if obj != nil {
			return []string{fmt.Sprintf("%v", obj)}, true, nil
		}
		return nil, false, nil
	}

	currentField := fields[0]
	remainingFields := fields[1:]

	switch v := obj.(type) {
	case map[string]interface{}:
		// 从 map 中获取值
		if val, exists := v[currentField]; exists {
			return getFieldValues(val, remainingFields, arrayCondition)
		}
		return nil, false, nil
	case []map[string]interface{}:
		// 遍历数组，筛选符合条件的项
		var results []string
		for _, item := range v {
			if matchCondition2(item, arrayCondition) {
				// 条件匹配，递归获取剩余字段
				if val, found, err := getFieldValues(item, remainingFields, nil); found || err != nil {
					results = append(results, val...)
				}
			}
		}
		return results, len(results) > 0, nil
	case []interface{}:
		// 如果是 interface{} 数组，逐项判断
		var results []string
		for _, item := range v {
			if val, found, err := getFieldValues(item, fields, arrayCondition); found || err != nil {
				results = append(results, val...)
			}
		}
		return results, len(results) > 0, nil
	case string:
		if len(fields) == 0 {
			return []string{v}, true, nil
		}
		return nil, false, nil
	case bool, int, int64, float64:
		if len(fields) == 0 {
			return []string{fmt.Sprintf("%v", v)}, true, nil
		}
		return nil, false, nil
	default:
		return nil, false, nil
	}
}

// matchCondition 检查数组中的元素是否符合条件
func matchCondition2(value interface{}, condition map[string]string) bool {
	for key, val := range condition {
		if valueMap, ok := value.(map[string]interface{}); ok {
			if mapVal, exists := valueMap[key]; exists && fmt.Sprintf("%v", mapVal) == val {
				continue
			}
			return false
		}
	}
	return true
}
