package main

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	FullStackAPIVersion = "templates.stolos.cloud/v1"
	KindFullStack       = "FullStack"
)

// FullStack wires backend, database, cache, and frontend resources.
type FullStack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              FullStackSpec `json:"spec"`
}

// FullStackSpec enumerates nested config sections.
type FullStackSpec struct {
	Backend  BackendSpec  `json:"backend"`
	Frontend FrontendSpec `json:"frontend"`
	Database DatabaseSpec `json:"database"`
	Cache    CacheSpec    `json:"cache"`
}

// BackendSpec configures the API deployment and ingress.
type BackendSpec struct {
	Image         string `json:"image"`
	Replicas      int32  `json:"replicas,omitempty" Default:"2"`
	ContainerPort int32  `json:"containerPort,omitempty" Default:"8080"`
	Host          string `json:"host"`
	Path          string `json:"path,omitempty" Default:"\"/api\""`
	TLSSecretName string `json:"tlsSecretName,omitempty"`
}

// FrontendSpec configures the nginx deployment + ingress.
type FrontendSpec struct {
	Host          string `json:"host"`
	Path          string `json:"path,omitempty" Default:"\"/\""`
	TLSSecretName string `json:"tlsSecretName,omitempty"`
	Image         string `json:"image,omitempty" Default:"\"nginx:stable-alpine\""`
	Replicas      int32  `json:"replicas,omitempty" Default:"1"`
	StaticContent string `json:"staticContent,omitempty"`
}

// DatabaseSpec describes the CNPG cluster inputs.
type DatabaseSpec struct {
	ClusterName     string `json:"clusterName"`
	DatabaseName    string `json:"databaseName"`
	Instances       int32  `json:"instances,omitempty" Default:"1"`
	StorageSize     string `json:"storageSize,omitempty" Default:"\"10Gi\""`
	PostgresVersion string `json:"postgresVersion,omitempty" Default:"\"16\""`
}

// CacheSpec configures Redis / Valkey.
type CacheSpec struct {
	Flavor string `json:"flavor,omitempty" Default:"\"redis\""`
	Port   int32  `json:"port,omitempty" Default:"6379"`
}

func (f FullStack) MarshalJSON() ([]byte, error) {
	f.APIVersion = FullStackAPIVersion
	f.Kind = KindFullStack
	type Alias FullStack
	return json.Marshal(Alias(f))
}

func (f *FullStack) UnmarshalJSON(data []byte) error {
	type Alias FullStack
	if err := json.Unmarshal(data, (*Alias)(f)); err != nil {
		return err
	}
	if f.APIVersion != "" && f.APIVersion != FullStackAPIVersion {
		return fmt.Errorf("unexpected api version: expected %s but got %s", FullStackAPIVersion, f.APIVersion)
	}
	if f.Kind != "" && f.Kind != KindFullStack {
		return fmt.Errorf("unexpected kind: expected %s but got %s", KindFullStack, f.Kind)
	}
	return nil
}
