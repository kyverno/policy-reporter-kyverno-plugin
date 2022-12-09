package reporting

import "strings"

type Summary struct {
	Error   int
	Pass    int
	Fail    int
	Warning int
}

type Resource struct {
	Kind       string
	APIVersion string
	Name       string
	Status     string
}

type Rule struct {
	Summary   *Summary
	Resources []*Resource
}

type Group struct {
	Name    string
	Policy  *Policy
	Summary *Summary
	Rules   map[string]*Rule
}

type Policy struct {
	Title       string
	Category    string
	Description string
	Severity    string
}

type Validation struct {
	Name   string
	Policy *Policy
	Groups map[string]*Group
}

type Filter struct {
	Namespaces   []string
	Policies     []string
	ClusterScope bool
}

func (f Filter) IncludesPolicy(policy string) bool {
	if len(f.Policies) == 0 {
		return true
	}

	return Contains(policy, f.Policies)
}

func (f Filter) IncludesNamespace(namespace string) bool {
	if len(f.Namespaces) == 0 || namespace == "" {
		return true
	}

	return Contains(namespace, f.Namespaces)
}

func Contains(source string, sources []string) bool {
	for _, s := range sources {
		if strings.EqualFold(s, source) {
			return true
		}
	}

	return false
}
