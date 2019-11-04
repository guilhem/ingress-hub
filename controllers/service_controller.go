/*
Copyright 2019 Guilhem Lettron <guilhem@barpilot.io>.

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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	boardingbridgev1alpha1 "github.com/guilhem/boardingbridge/api/v1alpha1"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	serviceFinalizer = "service.boardingbridge.barpilot.io"
)

// +kubebuilder:rbac:groups=boardingbridge.barpilot.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=boardingbridge.barpilot.io,resources=services/status,verbs=get;update;patch

func (r *ServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("namespace", req.NamespacedName)

	var instance boardingbridgev1alpha1.Service
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		log.Error(err, "unable to fetch KInamespace")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, ignoreNotFound(err)
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(instance.ObjectMeta.Finalizers, serviceFinalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, serviceFinalizer)
			if err := r.Update(ctx, &instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(instance.ObjectMeta.Finalizers, serviceFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			var coreService *corev1.Service
			if err := r.Get(ctx, types.NamespacedName{Name: instance.Name}, coreService); err != nil {
				if !apierrs.IsNotFound(err) {
					return ctrl.Result{}, err
				}
				// Service not found
			} else {
				if err := r.Delete(ctx, coreService); ignoreNotFound(err) != nil {
					log.Error(err, "unable to delete Service", "Service", coreService)
					return ctrl.Result{}, err
				}
			}

			// remove our finalizer from the list and update it.
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, serviceFinalizer)
			if err := r.Update(ctx, &instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	constructService := func(instance *boardingbridgev1alpha1.Service) (*corev1.Service, error) {
		// // Generate an UUID
		// uuid, err := uuid.NewRandom()
		// if err != nil {
		// 	return nil, err
		// }

		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.Name,
				Namespace: instance.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{
					{
						Name:     "test",
						Protocol: corev1.ProtocolTCP,
						Port:     8080,
					},
				},
			},
		}

		if err := ctrl.SetControllerReference(instance, service, r.Scheme); err != nil {
			return nil, err
		}

		return service, nil
	}
	// +kubebuilder:docs-gen:collapse=constructService

	service, err := constructService(&instance)
	if err != nil {
		log.Error(err, "unable to construct service")
		return ctrl.Result{}, nil
	}

	if err := r.Create(ctx, service); err != nil {
		log.Error(err, "unable to create Namespace", "namespace", service)
		return ctrl.Result{}, err
	}

	service.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{
		{
			IP: "1.2.3.4",
		},
	}

	if err := r.Status().Update(ctx, service); err != nil {
		log.Error(err, "unable to update service status")
		return ctrl.Result{}, err
	}

	// instance.Status. = namespace.Name
	// if err := r.Status().Update(ctx, &instance); err != nil {
	// 	log.Error(err, "unable to update KInamespace status")
	// 	return ctrl
	// }

	// your logic here

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&boardingbridgev1alpha1.Service{}).
		Complete(r)
}

/*
We generally want to ignore (not requeue) NotFound errors, since we'll get a
reconciliation request once the object exists, and requeuing in the meantime
won't help.
*/
func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
