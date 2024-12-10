// Copyright (c) 2024 ScyllaDB.
package controllerhelpers

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/utils/ptr"
)

type objectForAnnotationsPatch struct {
	objectMetaForAnnotationsPatch `json:"metadata"`
}
type objectMetaForAnnotationsPatch struct {
	ResourceVersion string             `json:"resourceVersion"`
	Annotations     map[string]*string `json:"annotations"`
}

func PrepareSetAnnotationPatch(obj metav1.Object, annotationKey string, annotationValue *string) ([]byte, error) {
	newAnnotations := make(map[string]*string, len(obj.GetAnnotations())+1)
	for k, v := range obj.GetAnnotations() {
		newAnnotations[k] = ptr.To(v)
	}
	newAnnotations[annotationKey] = annotationValue
	patch, err := json.Marshal(objectForAnnotationsPatch{
		objectMetaForAnnotationsPatch: objectMetaForAnnotationsPatch{
			ResourceVersion: obj.GetResourceVersion(),
			Annotations:     newAnnotations,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("can't marshal object for set annotation patch: %w", err)
	}
	return patch, nil
}

func HasAnnotation(obj metav1.Object, annotationKey string) bool {
	_, ok := obj.GetAnnotations()[annotationKey]
	return ok
}

func HasMatchingAnnotation(obj metav1.Object, annotationKey string, annotationValue string) bool {
	val, ok := obj.GetAnnotations()[annotationKey]
	return ok && val == annotationValue
}
