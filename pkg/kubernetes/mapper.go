package kubernetes

import (
	"errors"
	"strings"
	"time"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"gopkg.in/yaml.v2"
)

// Mapper converts maps into report structs
type Mapper interface {
	// MapPolicy maps a map into a Policy
	MapPolicy(reportMap map[string]interface{}) *kyverno.Policy
}

type mapper struct{}

func (m *mapper) MapPolicy(policy map[string]interface{}) *kyverno.Policy {
	r := &kyverno.Policy{
		Kind:  policy["kind"].(string),
		Rules: make([]*kyverno.Rule, 0),
	}

	metadata := policy["metadata"].(map[string]interface{})

	if name, ok := metadata["name"]; ok {
		r.Name = name.(string)
	}

	if uid, ok := metadata["uid"]; ok {
		r.UID = uid.(string)
	}

	if namespace, ok := metadata["namespace"]; ok {
		r.Namespace = namespace.(string)
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

	if creationTimestamp, err := m.mapCreationTime(policy); err == nil {
		r.CreationTimestamp = creationTimestamp
	}

	spec := policy["spec"].(map[string]interface{})

	if background, ok := spec["background"]; ok {
		r.Background = background.(bool)
	}
	if validation, ok := spec["validationFailureAction"]; ok {
		r.ValidationFailureAction = validation.(string)
	}
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
	r := &kyverno.Rule{}

	if name, ok := rule["name"]; ok {
		if n, ok := name.(string); ok {
			r.Name = n
		}
	}
	if verifyImages, ok := rule["verifyImages"]; ok {
		r.Type = "validation"
		r.VerifyImages = make([]*kyverno.VerifyImage, 0)

		data := verifyImages.([]interface{})

		r.VerifyImages = make([]*kyverno.VerifyImage, 0, len(data))

		for _, d := range data {
			if verify, ok := d.(map[string]interface{}); ok {
				item := &kyverno.VerifyImage{}
				if repo, ok := verify["repository"].(string); ok {
					item.Repository = repo
				}
				if image, ok := verify["image"].(string); ok {
					item.Image = image
				}
				if key, ok := verify["key"].(string); ok {
					item.Key = strings.TrimSpace(key)
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

		message := validate.(map[string]interface{})["message"]
		if m, ok := message.(string); ok {
			r.ValidateMessage = m
		}

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

func (m *mapper) mapCreationTime(result map[string]interface{}) (time.Time, error) {
	if metadata, ok := result["metadata"].(map[string]interface{}); ok {
		if created, ok2 := metadata["creationTimestamp"].(string); ok2 {
			return time.Parse("2006-01-02T15:04:05Z", created)
		}

		return time.Time{}, errors.New("Missing creationTimestamp in Metadata")
	}

	return time.Time{}, errors.New("Missing metadata")
}

// NewMapper creates an new Mapper instance
func NewMapper() Mapper {
	return &mapper{}
}
