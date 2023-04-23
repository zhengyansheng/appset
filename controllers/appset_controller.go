/*
Copyright 2023 zhengyansheng.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appv1 "github.com/zhengyansheng/appset/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// AppSetReconciler reconciles a AppSet object
type AppSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.hh.org,resources=appsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.hh.org,resources=appsets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.hh.org,resources=appsets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppSet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AppSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	appSet := &appv1.AppSet{}
	if err := r.Get(ctx, req.NamespacedName, appSet); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		//return ctrl.Result{RequeueAfter: time.Second * 10}, err
		return ctrl.Result{}, err
	}

	foundDeployment := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, foundDeployment); err != nil {
		if apierrors.IsNotFound(err) {
			// Req object not found, Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			// Return and don't requeue

			// create deployment
			if err := r.CreateDeployment(ctx, appSet); err != nil {
				return ctrl.Result{}, err
			}

			// create service
			service := &corev1.Service{}
			if err := r.Get(ctx, req.NamespacedName, service); err != nil {
				if apierrors.IsNotFound(err) {
					if err := r.CreateService(ctx, appSet); err != nil {
						return ctrl.Result{}, err
					}
				}
			}

			// create ingress
			ingress := &networkv1.Ingress{}
			if err := r.Get(ctx, req.NamespacedName, ingress); err != nil {
				if apierrors.IsNotFound(err) {
					if err := r.CreateIngress(ctx, appSet); err != nil {
						return ctrl.Result{}, err
					}
				}
			}

			// update status
			if err = r.updateStatus(ctx, appSet); err != nil {
				return ctrl.Result{}, err
			}

			return reconcile.Result{}, nil
		} else {
			klog.Errorf("requesting app set operator err %v", err)
			// Error reading the object - requeue the request.
			return reconcile.Result{Requeue: true}, nil
		}
	}
	klog.Infof("app set, foo: %v", appSet.Spec.Name)

	// https://github.com/caoyingjunz/podset-operator/blob/master/controllers/podset_controller.go
	// https://github.com/zhengyansheng/learning/blob/master/kubernetes/operators/elasticweb-operator/controllers/elasticweb_controller.go

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.AppSet{}).
		Complete(r)
}
