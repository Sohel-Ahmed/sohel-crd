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
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	demov1 "my.domain/demo-operator/api/v1"
)

// SohelReconciler reconciles a Sohel object
type SohelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=demo.my.domain,resources=sohels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=demo.my.domain,resources=sohels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=demo.my.domain,resources=sohels/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Sohel object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *SohelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Enter Reconcile", "req", req)

	sohel := &demov1.Sohel{}
	r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, sohel)

	l.Info("Enter Reconcile", "spec", sohel.Spec, "status", sohel.Status)

	if sohel.Spec.Name != sohel.Status.Name {
		sohel.Status.Name = sohel.Spec.Name
		r.Status().Update(ctx, sohel)
	}

	r.ReconcilePVC(ctx, sohel, l)

	return ctrl.Result{}, nil
}

func (r *SohelReconciler) ReconcilePVC(ctx context.Context, sohel *demov1.Sohel, l logr.Logger) error {
	pvc := &v1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{Name: sohel.Name, Namespace: sohel.Namespace}, pvc)
	if err == nil {
		l.Info("PVC found")
		return nil
	}

	if !errors.IsNotFound(err) {
		return err
	}

	l.Info("PVC Not Found")

	storageClass := "standard"
	storageReqeust, err := resource.ParseQuantity(fmt.Sprintf("%dGi", sohel.Spec.Size))

	pvc_new := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: sohel.Namespace,
			Name:      sohel.Name,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
			AccessModes:      []v1.PersistentVolumeAccessMode{"ReadWriteOnce"},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{"storage": storageReqeust},
			},
		},
	}

	return r.Create(ctx, pvc_new)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SohelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demov1.Sohel{}).
		Complete(r)
}
