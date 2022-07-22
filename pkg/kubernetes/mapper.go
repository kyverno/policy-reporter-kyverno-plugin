package kubernetes

import (
	"strings"
	"time"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"gopkg.in/yaml.v2"
)

// Mapper converts maps into report structs
type Mapper interface {
	// MapPolicy maps a map into a Policy
	MapPolicy(reportMap map[string]interface{}) kyverno.Policy
}

type mapper struct{}

func (m *mapper) MapPolicy(policy map[string]interface{}) kyverno.Policy {
	metadata := policy["metadata"].(map[string]interface{})

	r := kyverno.Policy{
		Kind:      policy["kind"].(string),
		Rules:     make([]*kyverno.Rule, 0),
		Name:      toString(metadata["name"]),
		UID:       toString(metadata["uid"]),
		Namespace: toString(metadata["namespace"]),
	}

	if an, ok := metadata["annotations"]; ok {
		if annotations, ok := an.(map[string]interface{}); ok {
			if category, ok := annotations["policies.kyverno.io/category"]; ok {
				r.Category = category.(string)
			}

			if severity, ok := annotations["policies.kyverno.io/severity"]; ok {
				r.Severity = severity.(string)
			}

			if description, ok := annotations["policies.kyverno.io/description"]; ok {
				r.Description = description.(string)
			}

			if autogen, ok := annotations["pod-policies.kyverno.io/autogen-controllers"]; ok {
				r.AutogenControllers = strings.Split(autogen.(string), ",")
			}

			delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
		}
	}

	if creationTimestamp, err := m.mapCreationTime(metadata); err == nil {
		r.CreationTimestamp = creationTimestamp
	}

	spec := policy["spec"].(map[string]interface{})

	r.Background = toBool(spec["background"])
	r.ValidationFailureAction = toString(spec["validationFailureAction"])

	if rules, ok := spec["rules"].([]interface{}); ok {
		for _, ruleMap := range rules {
			r.Rules = append(r.Rules, m.mapRule(ruleMap.(map[string]interface{})))
		}
	}

	delete(metadata, "managedFields")
	delete(policy, "status")

	content, err := yaml.Marshal(policy)
	if err == nil {
		r.Content = string(content)
	}

	return r
}

func (m *mapper) mapRule(rule map[string]interface{}) *kyverno.Rule {
	r := &kyverno.Rule{
		Name: toString(rule["name"]),
	}

	if verifyImages, ok := rule["verifyImages"]; ok {
		r.Type = "validation"
		r.VerifyImages = make([]*kyverno.VerifyImage, 0)

		data := verifyImages.([]interface{})

		r.VerifyImages = make([]*kyverno.VerifyImage, 0, len(data))

		for _, d := range data {
			if verify, ok := d.(map[string]interface{}); ok {
				item := &kyverno.VerifyImage{
					Repository: toString(verify["repository"]),
					Image:      toString(verify["image"]),
					Key:        strings.TrimSpace(toString(verify["key"])),
				}

				if attestations, ok := verify["attestations"].([]interface{}); ok && len(attestations) > 0 {
					value, err := yaml.Marshal(map[string]interface{}{"attestations": attestations})
					if err == nil {
						item.Attestations = string(value)
					}
				}
				r.VerifyImages = append(r.VerifyImages, item)
			}
		}

		return r
	}
	if validate, ok := rule["validate"]; ok {
		r.Type = "validation"
		r.ValidateMessage = toString(validate.(map[string]interface{})["message"])

		return r
	}
	if generate, ok := rule["generate"]; ok {
		if len(generate.(map[string]interface{})) > 0 {
			r.Type = "generation"
			return r
		}
	}
	if mutate, ok := rule["mutate"]; ok {
		if len(mutate.(map[string]interface{})) > 0 {
			r.Type = "mutation"
			return r
		}
	}

	return r
}

func (m *mapper) mapCreationTime(metadata map[string]interface{}) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z", toString(metadata["creationTimestamp"]))
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
