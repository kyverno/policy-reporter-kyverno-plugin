package api

type Policy struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	UID       string `json:"uid,omitempty"`
}

type VerifyImage struct {
	Policy       *Policy `json:"policy"`
	Rule         string  `json:"rule"`
	Repository   string  `json:"repository"`
	Image        string  `json:"image"`
	Key          string  `json:"key"`
	Attestations string  `json:"attestations,omitempty"`
}
