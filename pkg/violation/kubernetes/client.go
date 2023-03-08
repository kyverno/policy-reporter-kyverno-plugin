package kubernetes

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/fasthash/fnv1a"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/violation"
)

type eventClient struct {
	publisher      *violation.Publisher
	factory        informers.SharedInformerFactory
	policyStore    *kyverno.PolicyStore
	eventNamespace string
}

func (e *eventClient) Run(stopper chan struct{}) error {
	startUp := time.Now()
	informer := e.factory.Core().V1().Events().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if event, ok := obj.(*corev1.Event); ok {
				if !strings.Contains(event.Message, "(blocked)") || startUp.After(event.CreationTimestamp.Time) {
					return
				}

				policy, ok := e.policyStore.Get(generateID(event.InvolvedObject))
				if !ok {
					log.Printf("[ERROR] policy not found %s\n", event.InvolvedObject.Name)
					return
				}

				e.publisher.Publish(ConvertEvent(event, policy, false))
			}
		},
		UpdateFunc: func(old interface{}, obj interface{}) {
			if event, ok := obj.(*corev1.Event); ok {
				if !strings.Contains(event.Message, "(blocked)") || startUp.After(event.LastTimestamp.Time) {
					return
				}

				policy, ok := e.policyStore.Get(generateID(event.InvolvedObject))
				if !ok {
					log.Printf("[ERROR] policy not found %s\n", event.InvolvedObject.Name)
					return
				}

				e.publisher.Publish(ConvertEvent(event, policy, true))
			}
		},
	})

	e.factory.Start(stopper)

	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		return fmt.Errorf("failed to sync events")
	}

	return nil
}

func ConvertEvent(event *corev1.Event, policy kyverno.Policy, updated bool) violation.PolicyViolation {
	parts := strings.Split(event.Message, " ")
	resourceParts := strings.Split(parts[1][0:len(parts[1])-1], "/")

	var namespace, name string

	if len(resourceParts) == 2 {
		namespace = strings.TrimSpace(resourceParts[0])
		name = strings.TrimSpace(resourceParts[1])
	} else {
		name = strings.TrimSpace(resourceParts[0])
	}

	ruleName := strings.TrimSpace(parts[2][1 : len(parts[2])-1])

	message := event.Message
	for _, rule := range policy.Rules {
		if rule.Name == ruleName {
			message = policy.Rules[0].ValidateMessage
		}
	}

	return violation.PolicyViolation{
		Resource: violation.Resource{
			Kind:      strings.TrimSpace(parts[0]),
			Namespace: namespace,
			Name:      name,
		},
		Policy: violation.Policy{
			Name:     policy.Name,
			Rule:     ruleName,
			Message:  message,
			Category: policy.Category,
			Severity: policy.Severity,
		},
		Timestamp: event.LastTimestamp.Time,
		Updated:   updated,
		Event: violation.Event{
			Name: event.Name,
			UID:  string(event.UID),
		},
	}
}

func NewClient(client k8s.Interface, publisher *violation.Publisher, policyStore *kyverno.PolicyStore, eventNamespace string) violation.EventClient {
	factory := informers.NewFilteredSharedInformerFactory(client, 0, eventNamespace, func(lo *v1.ListOptions) {
		lo.FieldSelector = fields.Set{
			"source": "kyverno-admission",
			"reason": "PolicyViolation",
			"type":   "Warning",
		}.AsSelector().String()
	})

	return &eventClient{
		publisher:   publisher,
		factory:     factory,
		policyStore: policyStore,
	}
}

func generateID(object corev1.ObjectReference) string {
	h1 := fnv1a.Init64
	h1 = fnv1a.AddString64(h1, object.Name)
	h1 = fnv1a.AddString64(h1, object.Namespace)

	return strconv.FormatUint(h1, 10)
}
