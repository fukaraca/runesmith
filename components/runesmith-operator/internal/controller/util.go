package controller

import (
	"fmt"
	"time"

	enchv1 "github.com/fukaraca/runesmith/components/runesmith-operator/api/v1"
	"github.com/fukaraca/runesmith/shared"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ptrStatus struct {
	namespacedName             client.ObjectKey
	phase                      *shared.EnchantmentPhase
	progress                   *string
	expiresAt, completionTime  *metav1.Time
	active, failed, successful *int
}

// markCompletion is helper to keep state uniform, it is planned to use only one reconcile and just before the reconcile
func markCompletion(enchantment *enchv1.Enchantment, ptr *ptrStatus) {
	if enchantment.Status.CompletionTime != nil {
		return
	}
	now := time.Now()
	if enchantment.Spec.Retention.TTLSecondsAfterFinished != nil {
		ttl := metav1.NewTime(now.Add(time.Duration(*enchantment.Spec.Retention.TTLSecondsAfterFinished) * time.Second))
		ptr.expiresAt = &ttl
	}
	nowT := metav1.NewTime(now)
	ptr.completionTime = &nowT
}

func determineNodeSelector(req *enchv1.EnchantmentSpecArtifactRequirement) map[string]string {
	nodeSelector := make(map[string]string)
	nodeSelector[lblKeyEnergy] = req.EnergyType.String()
	return nodeSelector
}

func determineTolerations(req *enchv1.EnchantmentSpecArtifactRequirement) []corev1.Toleration {
	return []corev1.Toleration{{
		Key:      lblKeyEnergy,
		Operator: corev1.TolerationOpEqual,
		Value:    req.EnergyType.String(),
		Effect:   corev1.TaintEffectNoSchedule,
	}}
}

func generateJobName(enchantment *enchv1.Enchantment, energyType shared.Elemental) string {
	return fmt.Sprintf("ejob-%d-%s-", enchantment.Spec.OrderID, energyType)
}

func isJobEnchanting(jobs *batchv1.JobList) bool {
	for i := range jobs.Items {
		if jobs.Items[i].Spec.Suspend != nil && *jobs.Items[i].Spec.Suspend {
			return false
		}
	}
	return true
}
