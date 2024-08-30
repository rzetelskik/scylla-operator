// Copyright (c) 2024 ScyllaDB.

package naming

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ManualRef(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func ObjRef(obj metav1.Object) string {
	namespace := obj.GetNamespace()
	if len(namespace) == 0 {
		return obj.GetName()
	}

	return ManualRef(obj.GetNamespace(), obj.GetName())
}
