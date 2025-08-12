package utils

import (
	"strings"

	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	v12 "github.com/Netcracker/qubership-redis/redis-operator/api/v2"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SecretTemplate(name string, values map[string]string, namespace string) *v1.Secret {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: make(map[string]string),
		},
	}
	data := map[string][]byte{}
	for key, value := range values {
		data[key] = []byte(value)
	}

	secret.Data = data

	return secret
}

func CreateRuntimeObjectContextWrapper(ctx core.ExecutionContext, object client.Object, meta metav1.ObjectMeta) error {
	scheme := ctx.Get(constants.ContextSchema).(*runtime.Scheme)
	spec := ctx.Get(constants.ContextSpec).(*v12.DbaasRedisAdapter)
	helper := ctx.Get(constants.KubernetesHelperImpl).(core.KubernetesHelper)
	specPointer := &(*spec)

	return helper.CreateRuntimeObject(scheme, specPointer, object, meta)
}

type SimpleCtxExecutable struct {
	core.DefaultExecutable
	StepName    string
	ExecuteFunc func(ctx core.ExecutionContext, cr *v12.DbaasRedisAdapter, log *zap.Logger) error
}

func (r *SimpleCtxExecutable) Execute(ctx core.ExecutionContext) error {
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	cr := ctx.Get(constants.ContextSpec).(*v12.DbaasRedisAdapter)

	log.Info(r.StepName + " step started")
	defer log.Info(r.StepName + " step finished")
	return r.ExecuteFunc(ctx, cr, log)
}

func GetHTTPPort(tlsEnabled bool) int32 {
	var port int32 = 8080
	if tlsEnabled {
		port = 8443
	}
	return port
}

func GetHTTPProtocol(tlsEnabled bool) string {
	if tlsEnabled {
		return "https"
	}
	return "http"
}

func IsTLSEnableForDBAAS(aggregatorRegistrationAddress string, tlsEnabled bool) bool {
	if !strings.Contains(aggregatorRegistrationAddress, "https") {
		return false
	}

	return tlsEnabled
}
