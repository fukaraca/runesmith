package kubeapi

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fukaraca/runesmith/components/runesmith-backend/config"
	"github.com/fukaraca/runesmith/components/runesmith-backend/service/artifactory"
	enchantmentv1 "github.com/fukaraca/runesmith/components/runesmith-operator/api/v1"
	"github.com/fukaraca/runesmith/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	set        kubernetes.Interface
	cont       client.Client
	scheme     *runtime.Scheme
	Namespace  string
	restConfig *rest.Config
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
	sch := runtime.NewScheme()
	_ = enchantmentv1.AddToScheme(sch)
	cont, err := client.New(cfg, client.Options{Scheme: sch})
	if err != nil {
		return nil, err
	}
	return &Client{set: cs, Namespace: namespace, cont: cont, restConfig: cfg, scheme: sch}, nil
}

func (c *Client) CreateEnchantment( // temporary method for now until CRD implemented
	ctx context.Context,
	artifact *artifactory.Artifact,
	enchConfig config.Enchanter,
	item shared.MagicalItem,
) (*enchantmentv1.Enchantment, error) {
	if c.Namespace == "" {
		return nil, fmt.Errorf("namespace must be set")
	}

	labels := map[string]string{
		"artifact-id":      strconv.Itoa(artifact.ID),
		"artifact-item-id": strconv.Itoa(artifact.ItemID),
	}

	reqs := make([]enchantmentv1.EnchantmentSpecArtifactRequirement, 0)
	for e, i := range item.RequiredList() {
		reqs = append(reqs, enchantmentv1.EnchantmentSpecArtifactRequirement{
			EnergyType:   e,
			ResourceName: e.Resource(),
			Limit:        i,
		})
	}

	ttl := 5
	selfReport := true

	enchantment := &enchantmentv1.Enchantment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: c.generateName(artifact.ID),
			Namespace:    c.Namespace,
			Labels:       labels,
		},
		Spec: enchantmentv1.EnchantmentSpec{
			Retention: enchantmentv1.EnchantmentRetentionPolicy{
				TTLSecondsAfterFinished: &ttl},
			OrderID: artifact.ID,
			Artifact: enchantmentv1.EnchantmentSpecArtifact{
				ID:           item.ID,
				Name:         item.Name,
				Tier:         item.Tier,
				Requirements: reqs,
				Priority:     item.Priority,
			},
			Cost:       enchConfig.Cost,
			SelfReport: &selfReport,
		},
		Status: enchantmentv1.EnchantmentStatus{
			Phase: shared.ScheduledAS, // It doesn't matter anyway
		},
	}
	err := c.cont.Create(ctx, enchantment)
	if err != nil {
		return nil, err
	}

	return enchantment, nil
}

func (c *Client) generateName(orderID int) string {
	return fmt.Sprintf("ench-artifact-%d-", orderID)
}
