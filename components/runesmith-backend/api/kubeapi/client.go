package kubeapi

import (
	"context"
	"fmt"
	"strconv"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	cs        kubernetes.Interface
	Namespace string
}

func NewInCluster(namespace string) (*Client, error) {
	if namespace == "" { // to test locally
		return nil, nil
	}
	cfg, err := rest.InClusterConfig() // todo KUBERNETES_SERVICE_HOST and Port must be set on local
	if err != nil {
		return nil, err
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{cs: cs, Namespace: namespace}, nil
}

func (c *Client) CreateFireEnchantmentJob( // temporary method for now until CRD implemented
	ctx context.Context,
	artifactID int,
	enchanterImage string,
	enchantmentCost int,
) (*batchv1.Job, error) {
	if c.Namespace == "" {
		return nil, fmt.Errorf("namespace must be set")
	}

	labels := map[string]string{
		"energy":        "fire",
		"workload-type": "enchantment",
		"artifact-id":   strconv.Itoa(artifactID),
	}

	ttl := int32(5)
	backoff := int32(0)
	termGrace := int64(10)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "enchantment-fire-",
			Labels:       labels,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			BackoffLimit:            &backoff,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy:                 corev1.RestartPolicyNever,
					TerminationGracePeriodSeconds: &termGrace,
					Tolerations: []corev1.Toleration{
						{
							Key:      "energy",
							Operator: corev1.TolerationOpEqual,
							Value:    "fire",
							Effect:   corev1.TaintEffectNoSchedule,
						},
					},
					NodeSelector: map[string]string{"energy": "fire"},
					Containers: []corev1.Container{
						{
							Name:            "runesmith-enchanter",
							Image:           enchanterImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{Name: "http", ContainerPort: 8080},
							},
							Env: []corev1.EnvVar{
								{
									Name: "POD_UID",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"},
									},
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
									},
								},
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"},
									},
								},
								{Name: "ARTIFACT_ID", Value: strconv.Itoa(artifactID)},
								{Name: "ENCHANTMENT_COST", Value: strconv.Itoa(enchantmentCost)},
								{Name: "SELF_REPORT", Value: "true"},
								{Name: "HTTP_PORT", Value: "8080"},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceName("manawell.io/fire"): resource.MustParse("2"),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.FromInt32(8080),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       2,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/readyz",
										Port: intstr.FromInt32(8080),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       1,
							},
						},
					},
				},
			},
		},
	}

	return c.cs.BatchV1().Jobs(c.Namespace).Create(ctx, job, metav1.CreateOptions{})
}
