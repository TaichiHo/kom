package kom

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

type label struct {
	kubectl *Kubectl
}

func (l *label) Label(s string) error {
	labelStr := ""
	if strings.HasSuffix(s, "-") {
		// Case when deleting a label
		labelStr = fmt.Sprintf(`{"%s":null}`, strings.TrimSuffix(s, "-"))
	} else {
		if !strings.Contains(s, "=") {
			return fmt.Errorf("invalid label format (must k=v)")
		}
		parts := strings.Split(s, "=")
		if len(parts) != 2 {
			return fmt.Errorf("invalid label format (must k=v)")
		}
		// Build map
		labelStr = fmt.Sprintf(`{"%s":"%s"}`, strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}

	var item interface{}
	patchData := fmt.Sprintf(`{"metadata":{"labels":%s}}`, labelStr)
	err := l.kubectl.Patch(&item, types.StrategicMergePatchType, patchData).Error
	return err
}
