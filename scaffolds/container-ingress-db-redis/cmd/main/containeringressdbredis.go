package main

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ContainerIngressDBRedisAPIVersion = "templates.stolos.cloud/v1"
	KindContainerIngressDBRedis       = "ContainerIngressDBRedis"
)

// ContainerIngressDBRedis wires together a backend deployment, ingress, PostgreSQL, and cache layer.
type ContainerIngressDBRedis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ContainerIngressDBRedisSpec `json:"spec"`
}

// ContainerIngressDBRedisSpec defines backend, ingress, database, and cache knobs.
type ContainerIngressDBRedisSpec struct {
	Image         string       `json:"image"`
	Replicas      int32        `json:"replicas,omitempty" Default:"2"`
	ContainerPort int32        `json:"containerPort,omitempty" Default:"8080"`
	Host          string       `json:"host"`
	Path          string       `json:"path,omitempty" Default:"/"`
	TLSSecretName string       `json:"tlsSecretName,omitempty"`
	Database      DatabaseSpec `json:"database"`
	Cache         CacheSpec    `json:"cache"`
}

// DatabaseSpec matches the CNPG inputs reused across scaffolds.
type DatabaseSpec struct {
	ClusterName     string `json:"clusterName"`
	DatabaseName    string `json:"databaseName"`
	Instances       int32  `json:"instances,omitempty" Default:"1"`
	StorageSize     string `json:"storageSize,omitempty" Default:"\"10Gi\""`
	PostgresVersion string `json:"postgresVersion,omitempty" Default:"\"16\""`
}

// CacheSpec configures Redis / Valkey deployment options.
type CacheSpec struct {
	Flavor string `json:"flavor,omitempty" Default:"\"redis\""`
	Port   int32  `json:"port,omitempty" Default:"6379"`
}

func (c ContainerIngressDBRedis) MarshalJSON() ([]byte, error) {
	c.APIVersion = ContainerIngressDBRedisAPIVersion
	c.Kind = KindContainerIngressDBRedis
	type Alias ContainerIngressDBRedis
	return json.Marshal(Alias(c))
}

func (c *ContainerIngressDBRedis) UnmarshalJSON(data []byte) error {
	type Alias ContainerIngressDBRedis
	if err := json.Unmarshal(data, (*Alias)(c)); err != nil {
		return err
	}
	if c.APIVersion != "" && c.APIVersion != ContainerIngressDBRedisAPIVersion {
		return fmt.Errorf("unexpected api version: expected %s but got %s", ContainerIngressDBRedisAPIVersion, c.APIVersion)
	}
	if c.Kind != "" && c.Kind != KindContainerIngressDBRedis {
		return fmt.Errorf("unexpected kind: expected %s but got %s", KindContainerIngressDBRedis, c.Kind)
	}
	return nil
}
