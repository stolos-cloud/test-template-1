package main

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ContainerIngressAPIVersion = "templates.stolos.cloud/v1"
	KindContainerIngress       = "ContainerIngress"
)

// ContainerIngress defines a container workload exposed via an Ingress.
type ContainerIngress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ContainerIngressSpec `json:"spec"`
}

// ContainerIngressSpec configures the Deployment, Service, and Ingress resources.
type ContainerIngressSpec struct {
	Image         string `json:"image"`
	Replicas      int32  `json:"replicas,omitempty" Default:"1"`
	ContainerPort int32  `json:"containerPort,omitempty" Default:"8080"`
	Host          string `json:"host"`
	Path          string `json:"path,omitempty" Default:"\"/\""`
	TLSSecretName string `json:"tlsSecretName,omitempty"`
}

func (c ContainerIngress) MarshalJSON() ([]byte, error) {
	c.APIVersion = ContainerIngressAPIVersion
	c.Kind = KindContainerIngress
	type Alias ContainerIngress
	return json.Marshal(Alias(c))
}

func (c *ContainerIngress) UnmarshalJSON(data []byte) error {
	type Alias ContainerIngress
	if err := json.Unmarshal(data, (*Alias)(c)); err != nil {
		return err
	}
	if c.APIVersion != "" && c.APIVersion != ContainerIngressAPIVersion {
		return fmt.Errorf("unexpected api version: expected %s but got %s", ContainerIngressAPIVersion, c.APIVersion)
	}
	if c.Kind != "" && c.Kind != KindContainerIngress {
		return fmt.Errorf("unexpected kind: expected %s but got %s", KindContainerIngress, c.Kind)
	}
	return nil
}
