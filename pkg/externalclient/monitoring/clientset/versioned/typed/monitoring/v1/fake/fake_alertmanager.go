// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1 "github.com/scylladb/scylla-operator/pkg/externalapi/monitoring/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeAlertmanagers implements AlertmanagerInterface
type FakeAlertmanagers struct {
	Fake *FakeMonitoringV1
	ns   string
}

var alertmanagersResource = v1.SchemeGroupVersion.WithResource("alertmanagers")

var alertmanagersKind = v1.SchemeGroupVersion.WithKind("Alertmanager")

// Get takes name of the alertmanager, and returns the corresponding alertmanager object, and an error if there is any.
func (c *FakeAlertmanagers) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.Alertmanager, err error) {
	emptyResult := &v1.Alertmanager{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(alertmanagersResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.Alertmanager), err
}

// List takes label and field selectors, and returns the list of Alertmanagers that match those selectors.
func (c *FakeAlertmanagers) List(ctx context.Context, opts metav1.ListOptions) (result *v1.AlertmanagerList, err error) {
	emptyResult := &v1.AlertmanagerList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(alertmanagersResource, alertmanagersKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.AlertmanagerList{ListMeta: obj.(*v1.AlertmanagerList).ListMeta}
	for _, item := range obj.(*v1.AlertmanagerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested alertmanagers.
func (c *FakeAlertmanagers) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(alertmanagersResource, c.ns, opts))

}

// Create takes the representation of a alertmanager and creates it.  Returns the server's representation of the alertmanager, and an error, if there is any.
func (c *FakeAlertmanagers) Create(ctx context.Context, alertmanager *v1.Alertmanager, opts metav1.CreateOptions) (result *v1.Alertmanager, err error) {
	emptyResult := &v1.Alertmanager{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(alertmanagersResource, c.ns, alertmanager, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.Alertmanager), err
}

// Update takes the representation of a alertmanager and updates it. Returns the server's representation of the alertmanager, and an error, if there is any.
func (c *FakeAlertmanagers) Update(ctx context.Context, alertmanager *v1.Alertmanager, opts metav1.UpdateOptions) (result *v1.Alertmanager, err error) {
	emptyResult := &v1.Alertmanager{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(alertmanagersResource, c.ns, alertmanager, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.Alertmanager), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeAlertmanagers) UpdateStatus(ctx context.Context, alertmanager *v1.Alertmanager, opts metav1.UpdateOptions) (result *v1.Alertmanager, err error) {
	emptyResult := &v1.Alertmanager{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(alertmanagersResource, "status", c.ns, alertmanager, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.Alertmanager), err
}

// Delete takes name of the alertmanager and deletes it. Returns an error if one occurs.
func (c *FakeAlertmanagers) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(alertmanagersResource, c.ns, name, opts), &v1.Alertmanager{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAlertmanagers) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(alertmanagersResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.AlertmanagerList{})
	return err
}

// Patch applies the patch and returns the patched alertmanager.
func (c *FakeAlertmanagers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Alertmanager, err error) {
	emptyResult := &v1.Alertmanager{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(alertmanagersResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.Alertmanager), err
}

// GetScale takes name of the alertmanager, and returns the corresponding scale object, and an error if there is any.
func (c *FakeAlertmanagers) GetScale(ctx context.Context, alertmanagerName string, options metav1.GetOptions) (result *autoscalingv1.Scale, err error) {
	emptyResult := &autoscalingv1.Scale{}
	obj, err := c.Fake.
		Invokes(testing.NewGetSubresourceActionWithOptions(alertmanagersResource, c.ns, "scale", alertmanagerName, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*autoscalingv1.Scale), err
}

// UpdateScale takes the representation of a scale and updates it. Returns the server's representation of the scale, and an error, if there is any.
func (c *FakeAlertmanagers) UpdateScale(ctx context.Context, alertmanagerName string, scale *autoscalingv1.Scale, opts metav1.UpdateOptions) (result *autoscalingv1.Scale, err error) {
	emptyResult := &autoscalingv1.Scale{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(alertmanagersResource, "scale", c.ns, scale, opts), &autoscalingv1.Scale{})

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*autoscalingv1.Scale), err
}
