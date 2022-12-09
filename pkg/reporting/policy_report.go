package reporting

import (
	"context"
	"strings"

	v1 "github.com/kyverno/kyverno/api/kyverno/v1"
	kyverno "github.com/kyverno/kyverno/api/kyverno/v1alpha2"
	"github.com/kyverno/kyverno/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/reporting/kubernetes"
	"golang.org/x/exp/slices"
)

const (
	titleAnnotation       = "policies.kyverno.io/title"
	categoryAnnotation    = "policies.kyverno.io/category"
	descriptionAnnotation = "policies.kyverno.io/description"
	severityAnnotation    = "policies.kyverno.io/severity"
)

type PolicyReportGenerator interface {
	PerPolicyData(ctx context.Context, filter Filter) ([]*Validation, error)
	PerNamespaceData(ctx context.Context, filter Filter) ([]*Validation, error)
}

type policyReportGenerator struct {
	policyClient *kubernetes.PolicyClient
	reportClient *kubernetes.ReportClient
}

func (g *policyReportGenerator) PerPolicyData(ctx context.Context, filter Filter) ([]*Validation, error) {
	cPolicies, err := g.policyClient.CusterPolicies(ctx)
	if err != nil {
		return nil, err
	}

	mapping := make(map[string]*Validation)
	for _, pol := range cPolicies {
		if !filter.IncludesPolicy(pol.Name) {
			continue
		}

		v := &Validation{
			Name:   pol.Name,
			Policy: mapClusterPolicy(pol),
			Groups: make(map[string]*Group),
		}

		mapping[pol.Name] = v
	}

	policies, err := g.policyClient.Policies(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, pol := range policies {
		if !filter.IncludesPolicy(pol.Name) || !filter.IncludesNamespace(pol.Namespace) {
			continue
		}

		v := &Validation{
			Name:   pol.Name,
			Policy: mapPolicy(pol),
			Groups: make(map[string]*Group),
		}

		mapping[pol.Name] = v
	}

	reports, err := g.reports(ctx)
	if err != nil {
		return nil, err
	}

	for _, polr := range reports {
		if !filter.IncludesNamespace(polr.GetNamespace()) {
			continue
		}

		if polr.GetNamespace() == "" && !filter.ClusterScope {
			continue
		}

		for _, result := range polr.GetResults() {
			val, ok := mapping[result.Policy]
			if !ok {
				continue
			}

			if result.Result == v1alpha2.StatusSkip {
				continue
			}

			_, ok = val.Groups[polr.GetNamespace()]
			if !ok {
				val.Groups[polr.GetNamespace()] = &Group{
					Rules:   make(map[string]*Rule),
					Summary: &Summary{},
				}
			}

			rule := result.Rule
			if strings.HasPrefix(rule, "autogen-") {
				rule = strings.TrimPrefix(rule, "autogen-")
			}

			_, ok = val.Groups[polr.GetNamespace()].Rules[rule]
			if !ok {
				val.Groups[polr.GetNamespace()].Rules[rule] = &Rule{
					Summary:   &Summary{},
					Resources: make([]*Resource, 0),
				}
			}

			if len(result.Resources) > 0 {
				val.Groups[polr.GetNamespace()].Rules[rule].Resources = append(val.Groups[polr.GetNamespace()].Rules[rule].Resources, mapResource(result))
			}

			increaseSummary(result.Result, val.Groups[polr.GetNamespace()].Rules[rule].Summary)
			increaseSummary(result.Result, val.Groups[polr.GetNamespace()].Summary)

		}
	}

	data := make([]*Validation, 0, len(mapping))
	for _, v := range mapping {
		if len(v.Groups) == 0 {
			continue
		}

		data = append(data, v)
	}

	slices.SortFunc(data, func(a, b *Validation) bool {
		if a.Policy.Category != b.Policy.Category {
			return a.Policy.Category < b.Policy.Category
		}

		return a.Name < b.Name
	})

	return data, nil
}

func (g *policyReportGenerator) PerNamespaceData(ctx context.Context, filter Filter) ([]*Validation, error) {
	mapping := make(map[string]*Validation)

	cPolicies, err := g.policyClient.CusterPolicies(ctx)
	if err != nil {
		return nil, err
	}

	groups := make(map[string]*Policy)
	for _, pol := range cPolicies {
		if !filter.IncludesPolicy(pol.Name) {
			continue
		}

		groups[pol.Name] = mapClusterPolicy(pol)
	}

	policies, err := g.policyClient.Policies(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, pol := range policies {
		if !filter.IncludesPolicy(pol.Name) || !filter.IncludesNamespace(pol.Namespace) {
			continue
		}

		groups[pol.Name] = mapPolicy(pol)
	}

	reports, err := g.reports(ctx)
	if err != nil {
		return nil, err
	}

	for _, polr := range reports {
		if !filter.IncludesNamespace(polr.GetNamespace()) {
			continue
		}

		if polr.GetNamespace() == "" && !filter.ClusterScope {
			continue
		}

		val, ok := mapping[polr.GetNamespace()]
		if !ok {
			val = &Validation{
				Name:   polr.GetNamespace(),
				Groups: make(map[string]*Group),
			}

			mapping[polr.GetNamespace()] = val
		}

		for _, result := range polr.GetResults() {
			if !filter.IncludesPolicy(result.Policy) {
				continue
			}

			if result.Result == v1alpha2.StatusSkip {
				continue
			}

			_, ok = val.Groups[result.Policy]
			if cache, found := groups[result.Policy]; !ok && found {
				val.Groups[result.Policy] = &Group{
					Name:    result.Policy,
					Policy:  cache,
					Rules:   make(map[string]*Rule),
					Summary: &Summary{},
				}
			} else if !ok {
				val.Groups[result.Policy] = &Group{
					Rules:   make(map[string]*Rule),
					Summary: &Summary{},
				}
			}

			rule := result.Rule
			if strings.HasPrefix(rule, "autogen-") {
				rule = strings.TrimPrefix(rule, "autogen-")
			}

			ruleObj, ok := val.Groups[result.Policy].Rules[rule]
			if !ok {
				ruleObj = &Rule{
					Summary:   &Summary{},
					Resources: make([]*Resource, 0),
				}

				val.Groups[result.Policy].Rules[rule] = ruleObj
			}

			if len(result.Resources) > 0 {
				ruleObj.Resources = append(ruleObj.Resources, mapResource(result))
			}

			increaseSummary(result.Result, ruleObj.Summary)
			increaseSummary(result.Result, val.Groups[result.Policy].Summary)

		}
	}

	data := make([]*Validation, 0, len(mapping))
	for _, v := range mapping {
		if len(v.Groups) == 0 {
			continue
		}

		data = append(data, v)
	}

	slices.SortFunc(data, func(a, b *Validation) bool {
		return a.Name < b.Name
	})

	return data, nil
}

func (g *policyReportGenerator) reports(ctx context.Context) ([]kyverno.ReportInterface, error) {
	reports := []kyverno.ReportInterface{}

	nsReports, err := g.reportClient.PolicyReports(ctx)
	if err != nil {
		return reports, err
	}

	for i := range nsReports {
		reports = append(reports, &nsReports[i])
	}

	cReports, err := g.reportClient.ClusterPolicyReports(ctx)
	if err != nil {
		return reports, err
	}

	for i := range cReports {
		reports = append(reports, &cReports[i])
	}

	return reports, nil
}

func NewPolicyReportGenerator(policyClient *kubernetes.PolicyClient, reportClient *kubernetes.ReportClient) PolicyReportGenerator {
	return &policyReportGenerator{
		policyClient,
		reportClient,
	}
}

func increaseSummary(result v1alpha2.PolicyResult, sum *Summary) {

	switch result {
	case v1alpha2.StatusPass:
		sum.Pass++
		break
	case v1alpha2.StatusWarn:
		sum.Warning++
		break
	case v1alpha2.StatusFail:
		sum.Fail++
		break
	case v1alpha2.StatusError:
		sum.Error++
		break
	}
}

func mapResource(result v1alpha2.PolicyReportResult) *Resource {
	return &Resource{
		Kind:       result.Resources[0].Kind,
		APIVersion: result.Resources[0].APIVersion,
		Name:       result.Resources[0].Name,
		Status:     string(result.Result),
	}
}

func mapPolicy(pol v1.Policy) *Policy {
	v := &Policy{
		Title:       pol.Annotations[titleAnnotation],
		Description: pol.Annotations[descriptionAnnotation],
		Category:    pol.Annotations[categoryAnnotation],
		Severity:    pol.Annotations[severityAnnotation],
	}

	if v.Title == "" {
		v.Title = pol.Name
	}

	if v.Category == "" {
		v.Category = "Custom"
	}

	return v
}

func mapClusterPolicy(pol v1.ClusterPolicy) *Policy {
	v := &Policy{
		Title:       pol.Annotations[titleAnnotation],
		Description: pol.Annotations[descriptionAnnotation],
		Category:    pol.Annotations[categoryAnnotation],
		Severity:    pol.Annotations[severityAnnotation],
	}

	if v.Title == "" {
		v.Title = pol.Name
	}

	if v.Category == "" {
		v.Category = "Custom"
	}

	return v
}
