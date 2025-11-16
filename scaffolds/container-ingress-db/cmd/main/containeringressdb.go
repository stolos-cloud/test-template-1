package main

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ContainerIngressDBAPIVersion = "templates.stolos.cloud/v1"
	KindContainerIngressDB       = "ContainerIngressDB"
)

// ContainerIngressDB models a workload exposed via ingress with a managed PostgreSQL cluster.
type ContainerIngressDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ContainerIngressDBSpec `json:"spec"`
}

// ContainerIngressDBSpec configures the backend workload, ingress, and database cluster.
type ContainerIngressDBSpec struct {
	Image         string       `json:"image"`
	Replicas      int32        `json:"replicas,omitempty" Default:"2"`
	ContainerPort int32        `json:"containerPort,omitempty" Default:"8080"`
	Host          string       `json:"host"`
	Path          string       `json:"path,omitempty" Default:"/"`
	TLSSecretName string       `json:"tlsSecretName,omitempty"`
	Database      DatabaseSpec `json:"database"`
}

// DatabaseSpec holds CNPG configuration options.
type DatabaseSpec struct {
	ClusterName     string `json:"clusterName"`
	DatabaseName    string `json:"databaseName"`
	Instances       int32  `json:"instances,omitempty" Default:"1"`
	StorageSize     string `json:"storageSize,omitempty" Default:"\"10Gi\""`
	PostgresVersion string `json:"postgresVersion,omitempty" Default:"\"16\""`
}

func (c ContainerIngressDB) MarshalJSON() ([]byte, error) {
	c.APIVersion = ContainerIngressDBAPIVersion
	c.Kind = KindContainerIngressDB
	type Alias ContainerIngressDB
	return json.Marshal(Alias(c))
}

func (c *ContainerIngressDB) UnmarshalJSON(data []byte) error {
	type Alias ContainerIngressDB
	if err := json.Unmarshal(data, (*Alias)(c)); err != nil {
		return err
	}
	if c.APIVersion != "" && c.APIVersion != ContainerIngressDBAPIVersion {
		return fmt.Errorf("unexpected api version: expected %s but got %s", ContainerIngressDBAPIVersion, c.APIVersion)
	}
	if c.Kind != "" && c.Kind != KindContainerIngressDB {
		return fmt.Errorf("unexpected kind: expected %s but got %s", KindContainerIngressDB, c.Kind)
	}
	return nil
}
