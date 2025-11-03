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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/types"
	netcrackercomv2 "github.com/Netcracker/qubership-redis/redis-operator/api/v2"
	"github.com/Netcracker/qubership-redis/redis-operator/api/v2/impl"
)

// DbaasRedisAdapterReconciler reconciles a DbaasRedisAdapter object
type DbaasRedisAdapterReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Reconciler reconcile.Reconciler
}

//+kubebuilder:rbac:groups=netcracker.com,resources=dbaasredisadapters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=netcracker.com,resources=dbaasredisadapters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=netcracker.com,resources=dbaasredisadapters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DbaasRedisAdapter object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *DbaasRedisAdapterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	return r.Reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DbaasRedisAdapterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Reconciler = newReconciler(mgr)
	return ctrl.NewControllerManagedBy(mgr).
		For(&netcrackercomv2.DbaasRedisAdapter{}).
		Complete(r)
}

func newReconciler(mgr ctrl.Manager) reconcile.Reconciler {
	return &core.ReconcileCommonService{
		Client:           mgr.GetClient(),
		KubeConfig:       mgr.GetConfig(),
		Scheme:           mgr.GetScheme(),
		Executor:         core.DefaultExecutor(),
		Builder:          &impl.RedisServiceBuilder{},
		PredeployBuilder: &impl.RedisPreDeployBuilder{},
		Reconciler:       NewCommonReconciler(),
	}
}

var _ reconcile.Reconciler = &core.ReconcileCommonService{}

type RedisReconciler struct {
	Instance *netcrackercomv2.DbaasRedisAdapter
}

func (s *RedisReconciler) GetConsulRegistration() *types.ConsulRegistration {
	return nil
}

func (s *RedisReconciler) GetConsulServiceRegistrations() map[string]*types.AgentServiceRegistration {
	return nil
}

// TODO maybe add not only CommonReconciler but spmething else to distinguish DR and non DR services
func NewCommonReconciler() core.CommonReconciler {
	return &RedisReconciler{}
}

// TODO Deployment credential change procedure
func (s *RedisReconciler) GetAdminSecretName() string {
	return ""
}

func (s *RedisReconciler) UpdatePassWithFullReconcile() bool {
	return false
}

type noopExecutable struct{}

func (n noopExecutable) Validate(ctx core.ExecutionContext) error {
	return nil
}

func (n noopExecutable) Condition(ctx core.ExecutionContext) (bool, error) {
	return true, nil
}

func (n noopExecutable) Execute(ctx core.ExecutionContext) error {
	return nil
}

func (s *RedisReconciler) UpdatePassword() core.Executable {
	return noopExecutable{}
}

func (s *RedisReconciler) SetServiceInstance(client client.Client, request reconcile.Request) {
	redisServiceList := &netcrackercomv2.DbaasRedisAdapterList{}
	err := core.ListRuntimeObjectsByNamespace(redisServiceList, client, request.Namespace)
	if err != nil {
		if errors.IsNotFound(err) {

		}
	}
	msCount := len(redisServiceList.Items)
	if msCount != 1 {
	}
	s.Instance = &redisServiceList.Items[0]
}

func (s *RedisReconciler) GetStatus() *types.ServiceStatusCondition {
	if len(s.Instance.Status.Conditions) > 0 {
		return &s.Instance.Status.Conditions[0]
	}
	return nil
}

func (s *RedisReconciler) UpdateStatus(condition types.ServiceStatusCondition) {
	s.Instance.Status.Conditions = []types.ServiceStatusCondition{condition}
}

func (s *RedisReconciler) GetSpec() interface{} {
	return s.Instance.Spec
}

func (s *RedisReconciler) GetInstance() client.Object {
	return s.Instance
}

func (s *RedisReconciler) GetDeploymentVersion() string {
	return s.Instance.Spec.DeploymentVersion
}

func (s *RedisReconciler) GetVaultRegistration() *types.VaultRegistration {
	return &s.Instance.Spec.VaultRegistration
}

func (s *RedisReconciler) UpdateDRStatus(status types.DisasterRecoveryStatus) {

}

func (s *RedisReconciler) GetMessage() string {
	if len(s.Instance.Status.Conditions) > 0 {
		return s.Instance.Status.Conditions[0].Message
	}

	return ""
}

func (s *RedisReconciler) GetConfigMapName() string {
	return "last-applied-configuration-info"
}
