package kom

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

type annotate struct {
	kubectl *Kubectl
}

func (a *annotate) Annotate(s string) error {
	annotateStr := ""
	if strings.HasSuffix(s, "-") {
		// Case when deleting a label
		annotateStr = fmt.Sprintf(`{"%s":null}`, strings.TrimSuffix(s, "-"))
	} else {
		if !strings.Contains(s, "=") {
			return fmt.Errorf("invalid annotate format (must k=v)")
		}
		parts := strings.Split(s, "=")
		if len(parts) != 2 {
			return fmt.Errorf("invalid annotate format (must k=v)")
		}
		// Build map
		annotateStr = fmt.Sprintf(`{"%s":"%s"}`, strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}

	var item interface{}
	patchData := fmt.Sprintf(`{"metadata":{"annotations":%s}}`, annotateStr)
	err := a.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return err
}
