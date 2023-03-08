package kubernetes

import (
	"strings"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	apiV1 "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/api/kyverno/v1"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

// Mapper converts maps into report structs
type Mapper interface {
	// MapPolicy maps a map into a Policy
	MapPolicy(apiV1.PolicyInterface, *unstructured.Unstructured) kyverno.Policy
}

type mapper struct{}

func (m *mapper) MapPolicy(policy apiV1.PolicyInterface, content *unstructured.Unstructured) kyverno.Policy {
	r := kyverno.Policy{
		Kind:       policy.GetKind(),
		APIVersion: policy.GetAPIVersion(),
		Rules:      make([]*kyverno.Rule, 0),
		Name:       policy.GetName(),
		UID:        string(policy.GetUID()),
		Namespace:  policy.GetNamespace(),
	}

	annotations := policy.GetAnnotations()

	if category, ok := annotations[apiV1.AnnotationPolicyCategory]; ok {
		r.Category = category
	}

	if severity, ok := annotations[apiV1.AnnotationPolicySeverity]; ok {
		r.Severity = severity
	}

	if description, ok := annotations["policies.kyverno.io/description"]; ok {
		r.Description = description
	}

	if autogen, ok := annotations[apiV1.PodControllersAnnotation]; ok {
		r.AutogenControllers = strings.Split(autogen, ",")
	}

	r.CreationTimestamp = policy.GetCreationTimestamp().Time

	if policy.GetSpec() != nil {
		r.Background = policy.GetSpec().Background
		r.ValidationFailureAction = string(policy.GetSpec().ValidationFailureAction)

		for _, rule := range policy.GetSpec().Rules {
			r.Rules = append(r.Rules, m.mapRule(rule))
		}
	}

	r.Content = mapContent(content)

	return r
}

func (m *mapper) mapRule(rule apiV1.Rule) *kyverno.Rule {
	r := &kyverno.Rule{
		Name: rule.Name,
	}

	if len(rule.VerifyImages) > 0 {
		r.Type = "validation"
		r.VerifyImages = make([]*kyverno.VerifyImage, 0)

		r.VerifyImages = make([]*kyverno.VerifyImage, 0, len(rule.VerifyImages))

		for _, verify := range rule.VerifyImages {
			item := &kyverno.VerifyImage{
				Repository: verify.Repository,
				Image:      verify.Image,
				Key:        strings.TrimSpace(verify.Key),
			}

			if item.Image == "" && len(verify.ImageReferences) > 0 {
				item.Image = strings.Join(verify.ImageReferences, ", ")
			}

			r.VerifyImages = append(r.VerifyImages, item)
		}

		return r
	}
	if rule.HasValidate() {
		r.Type = "validation"
		r.ValidateMessage = rule.Validation.Message

		return r
	}
	if rule.HasGenerate() {
		r.Type = "generation"
		return r
	}
	if rule.HasMutate() {
		r.Type = "mutation"
		return r
	}

	return r
}

func mapContent(policy *unstructured.Unstructured) string {
	if policy == nil {
		return ""
	}

	metadata := policy.Object["metadata"].(map[string]interface{})

	delete(metadata, "managedFields")
	delete(metadata, "creationTimestamp")
	delete(metadata, "generation")
	delete(metadata, "resourceVersion")
	delete(metadata, "uid")

	delete(policy.Object, "status")

	if annotations, ok := metadata["annotations"]; ok {
		delete(annotations.(map[string]interface{}), "kubectl.kubernetes.io/last-applied-configuration")
	}

	content, err := yaml.Marshal(policy.Object)
	if err != nil {
		return ""
	}

	return string(content)
}

// NewMapper creates an new Mapper instance
func NewMapper() Mapper {
	return &mapper{}
}

func toString(value any) string {
	if v, ok := value.(string); ok {
		return v
	}

	return ""
}

func toBool(value any) bool {
	if v, ok := value.(bool); ok {
		return v
	}

	return false
}
