/*
Copyright 2025.

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

package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/fukaraca/runesmith/shared"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	enchv1 "github.com/fukaraca/runesmith/components/runesmith-operator/api/v1"
)

const (
	lblKeyEnergy   = "energy"
	lblKeyWorkload = "workload-type"
	jobOwnerIndex  = "enchantmentIndex"
	localKueue     = "runesmith-queue"
)

// EnchantmentReconciler reconciles a Enchantment object
type EnchantmentReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Image    string
}

// +kubebuilder:rbac:groups=enchantment.runesmith.io,resources=enchantments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=enchantment.runesmith.io,resources=enchantments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=enchantment.runesmith.io,resources=enchantments/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Enchantment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *EnchantmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	ptr := &ptrStatus{namespacedName: req.NamespacedName}

	ench := &enchv1.Enchantment{}
	err := r.Get(ctx, req.NamespacedName, ench)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("enchantment resource not found. ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get Enchantment")
		return ctrl.Result{}, err
	}
	if !ench.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}
	phase := ench.Status.Phase
	if phase == "" {
		phase = shared.ScheduledAS
	}

	switch phase {
	case shared.ScheduledAS:
		// check jobs doesn't exist
		var jobs batchv1.JobList
		if err = r.List(ctx, &jobs,
			client.InNamespace(ench.Namespace),
			client.MatchingFields{jobOwnerIndex: string(ench.UID)},
		); err != nil {
			return ctrl.Result{}, err
		}
		if len(jobs.Items) > 0 {
			if !isJobEnchanting(&jobs) {
				return ctrl.Result{}, nil
			}
			ptr.phase = shared.EnchantingAS.Ptr()
			if err = r.reconcileStatus(ctx, ptr); err != nil {
				logger.Error(err, "failed to update Enchantment status", "from", shared.ScheduledAS, "to", shared.EnchantingAS)
				return ctrl.Result{RequeueAfter: time.Second}, nil
			}
			return ctrl.Result{}, nil
		}

		// create jobs
		return r.createJobs(ctx, ench, ptr)
	case shared.EnchantingAS, shared.RequeuedAS:
		// list jobs
		var jobs batchv1.JobList
		if err = r.List(ctx, &jobs,
			client.InNamespace(ench.Namespace),
			client.MatchingFields{jobOwnerIndex: string(ench.UID)},
		); err != nil {
			return ctrl.Result{}, err
		}

		if len(jobs.Items) == 0 {
			logger.Error(err, "unexpected items.len")
			ptr.phase = shared.FailedAS.Ptr()
			markCompletion(ench, ptr)
			if statusErr := r.reconcileStatus(ctx, ptr); statusErr != nil {
				logger.Error(statusErr, "failed to update Enchantment status", "from", ench.Status.Phase, "to", shared.FailedAS)
				return ctrl.Result{RequeueAfter: time.Second}, nil
			}
			return ctrl.Result{}, fmt.Errorf("unexpected items.len")
		}

		var completedCount, failedCount, activeCount, suspendedCount int

		for _, job := range jobs.Items {
			if job.Status.Failed > 0 {
				r.Recorder.Eventf(ench, corev1.EventTypeWarning, "JobFailed", "Job %s failed", job.Name)
				failedCount++
				continue
			}

			if job.Status.Succeeded > 0 && job.Status.Active == 0 {
				r.Recorder.Eventf(ench, corev1.EventTypeNormal, "JobSucceeded", "Job %s succeeded", job.Name)
				completedCount++
				continue
			}

			if job.Spec.Suspend != nil && *job.Spec.Suspend {
				r.Recorder.Eventf(ench, corev1.EventTypeNormal, "JobSuspended", "Job %s suspended", job.Name)
				suspendedCount++
				continue
			}

			if job.Status.Active > 0 {
				activeCount++
			}
		}

		progress := fmt.Sprintf("%d/%d", completedCount, len(ench.Spec.Artifact.Requirements))
		ptr.successful = &completedCount
		ptr.failed = &failedCount
		ptr.active = &activeCount
		ptr.progress = &progress
		var state shared.EnchantmentPhase

		switch {
		case failedCount > 0:
			// any job failed means enchantment failed, maybe we can work on detailed reconicle
			state = shared.FailedAS
			logger.Info("enchantment failed", "name", ench.Name, "failed jobs", failedCount)
			markCompletion(ench, ptr)
		case completedCount == len(jobs.Items):
			// all jobs completed successfully
			state = shared.CompletedAS
			logger.Info("enchantment completed", "name", ench.Name)
			markCompletion(ench, ptr)
		case suspendedCount > 0:
			// any job suspended/requeued means enchantment is requeued
			state = shared.RequeuedAS
			logger.Info("enchantment requeued", "name", ench.Name)
		default:
			if phase == shared.RequeuedAS {
				r.Recorder.Eventf(ench, corev1.EventTypeNormal, "JobResumed", "A pendingJob resumed but we don't know which one")
			}
			state = shared.EnchantingAS
		}

		ptr.phase = state.Ptr()
		if err = r.reconcileStatus(ctx, ptr); err != nil {
			logger.Error(err, "failed to update Enchantment status", "from", ench.Status.Phase, "to", state)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
		if activeCount > 0 {
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}

		return ctrl.Result{}, nil
	case shared.FailedAS, shared.CompletedAS:
		if ench.Status.ExpiresAt == nil {
			markCompletion(ench, ptr)
			if err = r.reconcileStatus(ctx, ptr); err != nil {
				logger.Error(err, "failed to update Enchantment ttl")
				return ctrl.Result{RequeueAfter: time.Second}, nil
			}
			return ctrl.Result{RequeueAfter: time.Until(ench.Status.ExpiresAt.Time)}, nil
		}
		if time.Now().After(ench.Status.ExpiresAt.Time) {
			policy := metav1.DeletePropagationForeground // or Background
			if err = r.Delete(ctx, ench, &client.DeleteOptions{PropagationPolicy: &policy}); err != nil {
				logger.Error(err, "failed to delete dependent objects", "from", ench.Status.Phase)
				return ctrl.Result{}, err
			}
			logger.Info("enchantment deleted", "name", ench.Name, "last state", ench.Status.Phase)
			return ctrl.Result{}, nil
		}
		logger.Info("enchantment completed", "name", ench.Name, "last state", ench.Status.Phase)
		return ctrl.Result{RequeueAfter: time.Until(ench.Status.ExpiresAt.Time)}, nil
	}

	// Job exists, update enchantment status based on job status
	return ctrl.Result{}, nil
}

// createJob creates a new Job for the Enchantment
func (r *EnchantmentReconciler) createJobs(ctx context.Context, enchantment *enchv1.Enchantment, ptr *ptrStatus) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	for i, ess := range enchantment.Spec.Artifact.Requirements {
		jobNameStub := generateJobName(enchantment, ess.EnergyType)
		nodeSelector := determineNodeSelector(&ess) // redundant
		tolerations := determineTolerations(&ess)
		suspend := true // TODO kueue expects on suspend
		backOff := int32(0)

		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: jobNameStub,
				Namespace:    enchantment.Namespace,
				Labels: map[string]string{
					lblKeyEnergy:                           ess.EnergyType.String(),
					lblKeyWorkload:                         "enchantment",
					"artifact-order-id":                    strconv.Itoa(enchantment.Spec.OrderID),
					"kueue.x-k8s.io/queue-name":            localKueue,
					"kueue.x-k8s.io/priority-class":        enchantment.Spec.Artifact.Tier.Lower(),
					"kueue.x-k8s.io/max-exec-time-seconds": "360",
				},
			},
			Spec: batchv1.JobSpec{
				Suspend:      &suspend,
				BackoffLimit: &backOff,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							lblKeyEnergy:        ess.EnergyType.String(),
							lblKeyWorkload:      "enchantment",
							"artifact-order-id": strconv.Itoa(enchantment.Spec.OrderID),
						},
					},
					Spec: corev1.PodSpec{
						RestartPolicy: corev1.RestartPolicyNever,
						NodeSelector:  nodeSelector,
						Tolerations:   tolerations,
						Containers: []corev1.Container{
							{
								Name:            "runesmith-enchanter",
								Image:           r.Image,
								ImagePullPolicy: corev1.PullIfNotPresent,
								Ports: []corev1.ContainerPort{
									{Name: "http", ContainerPort: 8080}, // TODO Parameterize
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
									{Name: "ARTIFACT_ID", Value: strconv.Itoa(enchantment.Spec.Artifact.ID)},
									{Name: "ENCHANTMENT_COST", Value: strconv.Itoa(enchantment.Spec.Cost)},
									{Name: "SELF_REPORT", Value: strconv.FormatBool(*enchantment.Spec.SelfReport)},
									{Name: "HTTP_PORT", Value: "8080"},
								},
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceName(ess.ResourceName): resource.MustParse(strconv.Itoa(ess.Limit)),
									},
								},
								LivenessProbe: &corev1.Probe{
									ProbeHandler: corev1.ProbeHandler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/healthz",
											Port: intstr.FromInt32(8080),
										},
									},
									InitialDelaySeconds: 2,
									PeriodSeconds:       2,
								},
								ReadinessProbe: &corev1.Probe{
									ProbeHandler: corev1.ProbeHandler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/readyz",
											Port: intstr.FromInt32(8080),
										},
									},
									InitialDelaySeconds: 2, // TODO Parameterize
									PeriodSeconds:       1,
								},
							},
						},
					},
				},
			},
		}

		if err := controllerutil.SetControllerReference(enchantment, job, r.Scheme); err != nil {
			logger.Error(err, "Failed to set owner reference on Job")
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, job); err != nil {
			r.Recorder.Eventf(enchantment, corev1.EventTypeWarning, "JobCreateFailed", "Error: %v", err)
			logger.Error(err, "Failed to create Job")

			progress := fmt.Sprintf("%d/%d", i+1, len(enchantment.Spec.Artifact.Requirements))
			ptr.phase = shared.FailedAS.Ptr()
			ptr.progress = &progress
			markCompletion(enchantment, ptr)
			if statusErr := r.reconcileStatus(ctx, ptr); statusErr != nil {
				logger.Error(statusErr, "Failed to update Enchantment status")
			}
			return ctrl.Result{}, err
		}
		r.Recorder.Eventf(enchantment, corev1.EventTypeNormal, "JobsCreated", "created %d jobs", len(enchantment.Spec.Artifact.Requirements))
		logger.Info("Successfully created Job", "job", job.Name)
	}

	progress := fmt.Sprintf("%d/%d", 0, len(enchantment.Spec.Artifact.Requirements))
	ptr.progress = &progress
	ptr.phase = shared.ScheduledAS.Ptr()
	if err := r.reconcileStatus(ctx, ptr); err != nil {
		logger.Error(err, "Failed to update Enchantment status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// reconcileStatus patches sub resource Status. status.Phase is required
func (r *EnchantmentReconciler) reconcileStatus(ctx context.Context, p *ptrStatus) error {
	if p.phase == nil {
		return fmt.Errorf("reconcileStatus failed: missing phase")
	}
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var ench enchv1.Enchantment
		if err := r.Client.Get(ctx, p.namespacedName, &ench); err != nil {
			return err
		}
		original := ench.DeepCopy()

		ench.Status.Phase = *p.phase
		if p.expiresAt != nil {
			ench.Status.ExpiresAt = p.expiresAt
		}
		if p.active != nil {
			ench.Status.ActiveJobs = *p.active
		}
		if p.failed != nil {
			ench.Status.FailedJobs = *p.failed
		}
		if p.progress != nil {
			ench.Status.Progress = *p.progress
		}
		if p.completionTime != nil {
			ench.Status.CompletionTime = p.completionTime
		}
		if p.successful != nil {
			ench.Status.SucceededJobs = *p.successful
		}

		return r.Client.Status().Patch(ctx, &ench, client.MergeFrom(original))
	}); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnchantmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(),
		&batchv1.Job{}, jobOwnerIndex,
		func(obj client.Object) []string {
			j := obj.(*batchv1.Job)
			if owner := metav1.GetControllerOf(j); owner != nil &&
				owner.APIVersion == enchv1.GroupVersion.String() &&
				owner.Kind == "Enchantment" {
				return []string{string(owner.UID)}
			}
			return nil
		}); err != nil {
		return err
	}
	r.Recorder = mgr.GetEventRecorderFor("runesmith-operator")

	return ctrl.NewControllerManagedBy(mgr).
		For(&enchv1.Enchantment{}).
		Owns(&batchv1.Job{}).
		Named("runesmith-operator").
		Complete(r)
}
