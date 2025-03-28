package callbacks

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/stream"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

func List(k *kom.Kubectl) error {

	stmt := k.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	ctx := stmt.Context
	conditions := stmt.Filter.Conditions
	namespaceList := stmt.NamespaceList

	opts := stmt.ListOptions
	listOptions := metav1.ListOptions{}
	if len(opts) > 0 {
		listOptions = opts[0]
	}

	// Use reflection to get the value of dest
	destValue := reflect.ValueOf(stmt.Dest)

	// Ensure dest is a pointer to a slice
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Slice {
		// Handle error: dest is not a pointer to a slice
		return fmt.Errorf("Please pass in an array type")
	}
	// Get the element type of the slice
	elemType := destValue.Elem().Type().Elem()

	cacheKey := fmt.Sprintf("%s/%s/%s/%s", ns, gvr.Group, gvr.Resource, gvr.Version)
	list, err := utils.GetOrSetCache(stmt.ClusterCache(), cacheKey, stmt.CacheTTL, func() (list *unstructured.UnstructuredList, err error) {
		// TODO Change list retrieval to use Option to solve large data volume retrieval issues
		if namespaced {
			if stmt.AllNamespace || len(namespaceList) > 1 {
				// All namespaces or multiple namespaces provided
				// client-go doesn't support cross-namespace queries, so get all and filter later
				ns = metav1.NamespaceAll
				list, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).List(ctx, listOptions)
			} else {
				// Not all namespaces and no multiple namespaces provided
				if ns == "" {
					ns = metav1.NamespaceDefault
				}
				list, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).List(ctx, listOptions)
			}
		} else {
			// Cluster-level query, no namespace needed
			list, err = stmt.Kubectl.DynamicClient().Resource(gvr).List(ctx, listOptions)
		}
		return
	})
	if err != nil {
		return err
	}
	if list == nil {
		// Return directly if empty
		return fmt.Errorf("list is nil")
	}
	if list.Items == nil {
		// Return directly if empty
		return fmt.Errorf("list Items is nil")
	}

	// Filter results, execute where conditions
	result := executeFilter(list.Items, conditions)
	if stmt.TotalCount != nil {
		*stmt.TotalCount = int64(len(result))
	}

	if stmt.Filter.Order != "" {
		// Execute OrderBy on results
		klog.V(6).Infof("order by = %s", stmt.Filter.Order)
		executeOrderBy(result, stmt.Filter.Order)
	} else {
		// Default sort by creation time in descending order
		utils.SortByCreationTime(result)
	}

	// Clear previous values first
	destValue.Elem().Set(reflect.MakeSlice(destValue.Elem().Type(), 0, 0))
	streamTmp := stream.FromSlice(result)
	// Check if there's a filter, use filter first to form a final list.Items
	if stmt.Filter.Offset > 0 {
		streamTmp = streamTmp.Skip(stmt.Filter.Offset)
	}
	if stmt.Filter.Limit > 0 {
		streamTmp = streamTmp.Limit(stmt.Filter.Limit)
	}

	for _, item := range streamTmp.ToSlice() {

		obj := item.DeepCopy()
		if stmt.RemoveManagedFields {
			utils.RemoveManagedFields(obj)
		}
		// Create new pointer to element type
		newElemPtr := reflect.New(elemType)
		// Convert unstructured to original target type
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, newElemPtr.Interface())
		// Add pointer value to slice
		destValue.Elem().Set(reflect.Append(destValue.Elem(), newElemPtr.Elem()))

	}
	stmt.RowsAffected = int64(len(list.Items))

	if err != nil {
		return err
	}
	return nil
}

func executeOrderBy(result []unstructured.Unstructured, order string) {
	// order by `metadata.name` asc, `metadata.host` asc
	// TODO Currently only implemented single field sorting, haven't figured out multi-field sorting yet
	order = strings.TrimPrefix(strings.TrimSpace(order), "order by")
	order = strings.TrimSpace(order)
	orders := strings.Split(order, ",")
	for _, ord := range orders {
		var field string
		var desc bool
		// Determine sort direction
		if strings.Contains(ord, "desc") {
			desc = true
			field = strings.ReplaceAll(ord, "desc", "")
		} else {
			field = strings.ReplaceAll(ord, "asc", "")
		}
		field = strings.TrimSpace(field)
		field = strings.TrimSpace(utils.TrimQuotes(field))
		klog.V(6).Infof("Sorting by field: %s, Desc: %v", field, desc)

		slice.SortBy(result, func(a, b unstructured.Unstructured) bool {
			// Get field values
			aFieldValues, found, err := getNestedFieldAsString(a.Object, field)
			if err != nil || !found {
				return false
			}
			bFieldValues, found, err := getNestedFieldAsString(b.Object, field)
			if err != nil || !found {
				return false
			}

			// order by must convert array to single value
			if len(aFieldValues) > 1 || len(bFieldValues) > 1 {
				return false
			}
			if len(aFieldValues) == 0 || len(bFieldValues) == 0 {
				return false
			}
			aFieldValue := aFieldValues[0]
			bFieldValue := bFieldValues[0]

			t, va := utils.DetectType(aFieldValue)
			_, vb := utils.DetectType(bFieldValue)

			switch t {
			case utils.TypeString:
				if desc {
					return va.(string) > vb.(string)
				}
				return va.(string) < vb.(string)
			case utils.TypeNumber:
				if desc {
					return va.(float64) > vb.(float64)
				}
				return va.(float64) < vb.(float64)
			case utils.TypeTime:
				tva, err := utils.ParseTime(fmt.Sprintf("%s", va))
				if err != nil {
					return false
				}
				tvb, err := utils.ParseTime(fmt.Sprintf("%s", vb))
				if err != nil {
					return false
				}
				if desc {
					return tva.After(tvb)
				}
				return tva.Before(tvb)
			default:
				return false
			}
		})
	}
}
