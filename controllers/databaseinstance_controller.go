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
	"fmt"
	"github.com/slamdev/databaser/pkg"
	"github.com/slamdev/databaser/pkg/clickhouse"
	"github.com/slamdev/databaser/pkg/postgres"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	databaserv1alpha1 "github.com/slamdev/databaser/api/v1alpha1"
)

// DatabaseInstanceReconciler reconciles a DatabaseInstance object
type DatabaseInstanceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=databaser.slamdev.github.com,resources=databaseinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databaser.slamdev.github.com,resources=databaseinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databaser.slamdev.github.com,resources=databaseinstances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DatabaseInstance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *DatabaseInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("databaseinstance", req.NamespacedName)

	instance := &databaserv1alpha1.DatabaseInstance{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if err := controllerutil.SetControllerReference(instance, instance, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	if instance.Spec.Clikhouse != nil && instance.Spec.Postgres != nil {
		return ctrl.Result{}, r.updateErrorStatus(ctx, instance, "only one connection spec is allowed")
	} else if instance.Spec.Clikhouse != nil || instance.Spec.Postgres != nil {
		return ctrl.Result{}, r.updateErrorStatus(ctx, instance, "at least one connection spec should be defined")
	}

	if instance.Spec.Clikhouse != nil {
		if err := r.validateClickhouseConnection(ctx, *instance.Spec.Clikhouse); err != nil {
			return ctrl.Result{}, r.updateErrorStatus(ctx, instance, err.Error())
		}
	}

	if instance.Spec.Clikhouse != nil {
		if err := r.validatePostgresConnection(ctx, *instance.Spec.Postgres); err != nil {
			return ctrl.Result{}, r.updateErrorStatus(ctx, instance, err.Error())
		}
	}

	return ctrl.Result{RequeueAfter: time.Second * 60}, r.updateConnectedStatus(ctx, instance)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databaserv1alpha1.DatabaseInstance{}).
		Complete(r)
}

func (r *DatabaseInstanceReconciler) updateErrorStatus(ctx context.Context, instance *databaserv1alpha1.DatabaseInstance, msg string) error {
	instance.Status.Phase = "failed"
	instance.Status.LastError = msg
	return r.Client.Status().Update(ctx, instance)
}

func (r *DatabaseInstanceReconciler) updateConnectedStatus(ctx context.Context, instance *databaserv1alpha1.DatabaseInstance) error {
	instance.Status.Phase = "connected"
	instance.Status.LastError = ""
	return r.Client.Status().Update(ctx, instance)
}

func (r *DatabaseInstanceReconciler) validateClickhouseConnection(ctx context.Context, spec databaserv1alpha1.ClikhouseSpec) error {
	sqlParams, err := r.parseSqlParams(ctx, spec.SqlParams)
	if err != nil {
		return err
	}
	c, err := pkg.CreateClickhouseSqlConnection(ctx, clickhouse.Params{
		User:     sqlParams.Username,
		Password: sqlParams.Password,
		Host:     sqlParams.Host,
		Port:     sqlParams.Port,
	})
	if err != nil {
		return err
	}
	return c.Close()
}

func (r *DatabaseInstanceReconciler) validatePostgresConnection(ctx context.Context, spec databaserv1alpha1.PostgresSpec) error {
	sqlParams, err := r.parseSqlParams(ctx, spec.SqlParams)
	if err != nil {
		return err
	}
	if spec.AuthDBRef != nil {
		if spec.AuthDB, err = r.getParamValue(ctx, *spec.AuthDBRef, "authdb"); err != nil {
			return err
		}
	}
	c, err := pkg.CreatePostgresSqlConnection(ctx, postgres.Params{
		User:     sqlParams.Username,
		Password: sqlParams.Password,
		Host:     sqlParams.Host,
		Port:     sqlParams.Port,
		AuthDB:   spec.AuthDB,
	})
	if err != nil {
		return err
	}
	return c.Close()
}

func (r *DatabaseInstanceReconciler) parseSqlParams(ctx context.Context, params databaserv1alpha1.SqlParams) (databaserv1alpha1.SqlParams, error) {
	var err error
	if params.HostRef != nil {
		if params.Host, err = r.getParamValue(ctx, *params.HostRef, "hostname", "host"); err != nil {
			return databaserv1alpha1.SqlParams{}, err
		}
	}
	if params.PortRef != nil {
		var port string
		if port, err = r.getParamValue(ctx, *params.PortRef, "port"); err != nil {
			return databaserv1alpha1.SqlParams{}, err
		}
		if params.Port, err = strconv.Atoi(port); err != nil {
			return databaserv1alpha1.SqlParams{}, err
		}
	}
	if params.UsernameRef != nil {
		if params.Username, err = r.getParamValue(ctx, *params.UsernameRef, "user", "username"); err != nil {
			return databaserv1alpha1.SqlParams{}, err
		}
	}
	if params.PasswordRef != nil {
		if params.Password, err = r.getParamValue(ctx, *params.PasswordRef, "pass", "password"); err != nil {
			return databaserv1alpha1.SqlParams{}, err
		}
	}
	return params, nil
}

func (r *DatabaseInstanceReconciler) getParamValue(ctx context.Context, ref databaserv1alpha1.ParamRef, fallbacks ...string) (string, error) {
	var keys []string
	if ref.Key != "" {
		keys = []string{ref.Key}
	} else {
		keys = fallbacks
	}

	if ref.Kind == "ConfigMap" {
		instance := &v1.ConfigMap{}
		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: ref.Namespace, Name: ref.Name}, instance); err != nil {
			return "", err
		}
		for _, key := range keys {
			if val, ok := instance.Data[key]; ok {
				return val, nil
			}
		}
	} else if ref.Kind == "Secret" {
		instance := &v1.Secret{}
		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: ref.Namespace, Name: ref.Name}, instance); err != nil {
			return "", err
		}
		for _, key := range keys {
			if val, ok := instance.Data[key]; ok {
				return string(val), nil
			}
		}
	}
	return "", fmt.Errorf("don't know how to handle %s kind", ref.Kind)
}
