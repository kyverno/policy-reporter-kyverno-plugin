package kubernetes

import (
	"log"
	"strings"
	"time"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type eventClient struct {
	client                k8s.Interface
	policyStore           *kyverno.PolicyStore
	restartWatchOnFailure time.Duration
	startUp               time.Time
	eventNamespace        string
}

func (e *eventClient) StartWatching(ctx context.Context) <-chan kyverno.PolicyViolation {
	violationChan := make(chan kyverno.PolicyViolation)

	go func() {
		for {
			e.watchEvents(ctx, violationChan)
			time.Sleep(e.restartWatchOnFailure)
		}
	}()

	return violationChan
}

func (e *eventClient) watchEvents(ctx context.Context, violationChan chan<- kyverno.PolicyViolation) {
	factory := informers.NewFilteredSharedInformerFactory(e.client, 0, e.eventNamespace, func(lo *v1.ListOptions) {
		lo.FieldSelector = fields.Set{
			"source": "kyverno-admission",
			"reason": "PolicyViolation",
			"type":   "Warning",
		}.AsSelector().String()
	})

	informer := factory.Core().V1().Events().Informer()

	informer.SetWatchErrorHandler(func(_ *cache.Reflector, _ error) {
		log.Println("[WARNING] Event watch failure - restarting")
	})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if event, ok := obj.(*corev1.Event); ok {
				if !strings.Contains(event.Message, "(blocked)") || e.startUp.After(event.CreationTimestamp.Time) {
					return
				}

				policy, ok := e.policyStore.Get(string(event.InvolvedObject.UID))
				if !ok {
					log.Printf("[ERROR] policy not found %s\n", event.InvolvedObject.Name)
					return
				}

				violationChan <- ConvertEvent(event, policy, false)
			}
		},
		UpdateFunc: func(old interface{}, obj interface{}) {
			if event, ok := obj.(*corev1.Event); ok {
				if !strings.Contains(event.Message, "(blocked)") || e.startUp.After(event.LastTimestamp.Time) {
					return
				}

				policy, ok := e.policyStore.Get(string(event.InvolvedObject.UID))
				if !ok {
					log.Printf("[ERROR] policy not found %s\n", event.InvolvedObject.Name)
					return
				}

				violationChan <- ConvertEvent(event, policy, true)
			}
		},
	})

	informer.Run(ctx.Done())
}

func ConvertEvent(event *corev1.Event, policy *kyverno.Policy, updated bool) kyverno.PolicyViolation {
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

	return kyverno.PolicyViolation{
		Resource: kyverno.ViolationResource{
			Kind:      strings.TrimSpace(parts[0]),
			Namespace: namespace,
			Name:      name,
		},
		Policy: kyverno.ViolationPolicy{
			Name:     policy.Name,
			Rule:     ruleName,
			Message:  message,
			Category: policy.Category,
			Severity: policy.Severity,
		},
		Timestamp: event.LastTimestamp.Time,
		Updated:   updated,
		Event: kyverno.ViolationEvent{
			Name: event.Name,
			UID:  string(event.UID),
		},
	}
}

type EventClient interface {
	StartWatching(ctx context.Context) <-chan kyverno.PolicyViolation
}

func NewEventClient(client k8s.Interface, policyStore *kyverno.PolicyStore, restartWatchOnFailure time.Duration, eventNamespace string) EventClient {
	return &eventClient{
		client:                client,
		policyStore:           policyStore,
		restartWatchOnFailure: restartWatchOnFailure,
		startUp:               time.Now(),
		eventNamespace:        eventNamespace,
	}
}
