package k8s

import (
	"fmt"
	"github.com/openfaas/faas/gateway/requests"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"path/filepath"
	"time"
)

const (
	ProbePath         = "com.openfaas.health.http.path"
	ProbeInitialDelay = "com.openfaas.health.http.initialDelay"
)

type FunctionProbes struct {
	Liveness  *corev1.Probe
	Readiness *corev1.Probe
}

func (f *Factory) MakeProbes(r requests.CreateFunctionRequest) (*FunctionProbes, error) {
	var handler corev1.Handler
	httpPath := "/_/health"
	initialDelaySeconds := int32(f.Config.LivenessProbe.InitialDelaySeconds)

	if r.Annotations != nil {
		annotations := *r.Annotations
		if path, ok := annotations[ProbePath]; ok {
			httpPath = path
		}
		if delay, ok := annotations[ProbeInitialDelay]; ok {
			d, err := time.ParseDuration(delay)
			if err != nil {
				return nil, fmt.Errorf("invalid %s duration format: %v", ProbeInitialDelay, err)
			}
			initialDelaySeconds = int32(d.Seconds())
		}
	}

	if f.Config.HTTPProbe {
		handler = corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: httpPath,
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: int32(f.Config.Port),
				},
			},
		}
	} else {
		path := filepath.Join("/tmp/", ".lock")
		handler = corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"cat", path},
			},
		}
	}

	probes := FunctionProbes{}
	probes.Readiness = &corev1.Probe{
		Handler:             handler,
		InitialDelaySeconds: initialDelaySeconds,
		TimeoutSeconds:      int32(f.Config.ReadinessProbe.TimeoutSeconds),
		PeriodSeconds:       int32(f.Config.ReadinessProbe.PeriodSeconds),
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}

	probes.Liveness = &corev1.Probe{
		Handler:             handler,
		InitialDelaySeconds: initialDelaySeconds,
		TimeoutSeconds:      int32(f.Config.LivenessProbe.TimeoutSeconds),
		PeriodSeconds:       int32(f.Config.LivenessProbe.PeriodSeconds),
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}

	return &probes, nil
}