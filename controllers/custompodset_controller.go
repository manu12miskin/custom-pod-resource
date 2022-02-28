/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	custompodsetv1 "github.com/manu12miskin/custom-pod-resource/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomPodSetReconciler reconciles a CustomPodSet object
type CustomPodSetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=demo.my.domain,resources=custompodsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=demo.my.domain,resources=custompodsets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=demo.my.domain,resources=custompodsets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CustomPodSet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *CustomPodSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	_ = context.Background()
	reqLogger := r.Log.WithValues("CustomPodset", req.NamespacedName)

	// Fetch the PodSet instance
	instance := &custompodsetv1.CustomPodSet{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	// List all pods owned by this PodSet instance
	podSet := instance
	podList := &corev1.PodList{}
	lbs := map[string]string{
		"app":     podSet.Name,
		"version": "v0.1",
	}
	labelSelector := labels.SelectorFromSet(lbs)
	listOps := &client.ListOptions{Namespace: podSet.Namespace, LabelSelector: labelSelector}
	if err = r.List(context.TODO(), podList, listOps); err != nil {
		return reconcile.Result{}, err
	}

	// Count the pods that are pending or running as available
	var available []corev1.Pod
	for _, pod := range podList.Items {
		if pod.ObjectMeta.DeletionTimestamp != nil {
			continue
		}
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
			available = append(available, pod)
		}
	}
	numAvailable := int32(len(available))
	availableNames := []string{}
	for _, pod := range available {
		availableNames = append(availableNames, pod.ObjectMeta.Name)
	}

	// Update the status if necessary
	status := custompodsetv1.CustomPodSetStatus{
		PodNames:          availableNames,
		AvailableReplicas: numAvailable,
	}
	if !reflect.DeepEqual(podSet.Status, status) {
		podSet.Status = status
		err = r.Status().Update(context.TODO(), podSet)
		if err != nil {
			reqLogger.Error(err, "Failed to update PodSet status")
			return reconcile.Result{}, err
		}
	}

	if numAvailable > podSet.Spec.Replicas {
		reqLogger.Info("Scaling down pods", "Currently available", numAvailable, "Required replicas", podSet.Spec.Replicas)
		diff := numAvailable - podSet.Spec.Replicas
		dpods := available[:diff]
		for _, dpod := range dpods {
			err = r.Delete(context.TODO(), &dpod)
			if err != nil {
				reqLogger.Error(err, "Failed to delete pod", "pod.name", dpod.Name)
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{Requeue: true}, nil
	}

	if numAvailable < podSet.Spec.Replicas {
		reqLogger.Info("Scaling up pods", "Currently available", numAvailable, "Required replicas", podSet.Spec.Replicas)
		// Define a new Pod object
		pod := newPodForCR(podSet)
		// Set PodSet instance as the owner and controller
		if err := controllerutil.SetControllerReference(podSet, pod, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		err = r.Create(context.TODO(), pod)
		if err != nil {
			reqLogger.Error(err, "Failed to create pod", "pod.name", pod.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *custompodsetv1.PodSet) *corev1.Pod {
	labels := map[string]string{
		"app":     cr.Name,
		"version": "v0.1",
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cr.Name + "-pod",
			Namespace:    cr.Namespace,
			Labels:       labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "customPodContainer",
					Image:   cr.Spec.Image,
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomPodSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&custompodsetv1.CustomPodSet{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
