/*
Copyright 2021.

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
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	databaserv1alpha1 "github.com/slamdev/databaser/api/v1alpha1"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=databaser.slamdev.github.com,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databaser.slamdev.github.com,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databaser.slamdev.github.com,resources=databases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Database object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("database", req.NamespacedName)

	db := &databaserv1alpha1.Database{}
	if err := r.Client.Get(ctx, req.NamespacedName, db); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if err := controllerutil.SetControllerReference(db, db, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	instance := &databaserv1alpha1.DatabaseInstance{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: "", Name: db.Spec.DatabaseInstanceRef.Name}, db); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, r.updateErrorStatus(ctx, db, "no corresponding database instance found")
		}
		return ctrl.Result{}, err
	}
	if instance.Status.Phase != "connected" {
		return ctrl.Result{}, r.updateErrorStatus(ctx, db, "corresponding database is not initialized")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databaserv1alpha1.Database{}).
		Complete(r)
}

func (r *DatabaseReconciler) updateErrorStatus(ctx context.Context, db *databaserv1alpha1.Database, msg string) error {
	db.Status.Phase = "failed"
	db.Status.LastError = msg
	return r.Client.Status().Update(ctx, db)
}

func (r *DatabaseReconciler) updateConnectedStatus(ctx context.Context, db *databaserv1alpha1.Database) error {
	db.Status.Phase = "connected"
	db.Status.LastError = ""
	return r.Client.Status().Update(ctx, db)
}
