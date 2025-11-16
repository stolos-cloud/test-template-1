package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"

	stolos_yoke "github.com/stolos-cloud/stolos/yoke-base/pkg/stolos-yoke"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/yokecd/yoke/pkg/flight"
)

//go:embed "AirwayInputs.yml"
var airwayInputsYml []byte

type airwayInputsManifest struct {
	Spec stolos_yoke.AirwayInputs `json:"spec"`
}

func main() {
	jsonBytes, err := yaml.ToJSON(airwayInputsYml)
	if err != nil {
		panic(err)
	}

	var manifest airwayInputsManifest
	if err := json.Unmarshal(jsonBytes, &manifest); err != nil {
		panic(err)
	}

	stolos_yoke.Run[ContainerIngressDB](manifest.Spec, run)
}

func run() ([]byte, error) {
	var resource ContainerIngressDB
	if err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&resource); err != nil && err != io.EOF {
		return nil, err
	}

	if err := validateSpec(&resource); err != nil && err != io.EOF {
		return nil, err
	}

	resources := []flight.Resource{
		createDeployment(resource),
		createService(resource),
		createIngress(resource),
		createCNPGCluster(resource),
	}

	return json.Marshal(resources)
}

func validateSpec(resource *ContainerIngressDB) error {
	if resource.Spec.Image == "" {
		return fmt.Errorf("spec.image is required")
	}
	if resource.Spec.Host == "" {
		return fmt.Errorf("spec.host is required")
	}
	if resource.Spec.Database.ClusterName == "" {
		return fmt.Errorf("spec.database.clusterName is required")
	}
	if resource.Spec.Database.DatabaseName == "" {
		return fmt.Errorf("spec.database.databaseName is required")
	}
	if resource.Spec.Replicas <= 0 {
		resource.Spec.Replicas = 2
	}
	if resource.Spec.ContainerPort == 0 {
		resource.Spec.ContainerPort = 8080
	}
	if resource.Spec.Path == "" {
		resource.Spec.Path = "/"
	}
	if resource.Spec.Database.Instances <= 0 {
		resource.Spec.Database.Instances = 1
	}
	if resource.Spec.Database.StorageSize == "" {
		resource.Spec.Database.StorageSize = "10Gi"
	}
	if resource.Spec.Database.PostgresVersion == "" {
		resource.Spec.Database.PostgresVersion = "16"
	}
	return nil
}

func createDeployment(resource ContainerIngressDB) *appsv1.Deployment {
	replicas := resource.Spec.Replicas
	labels := map[string]string{"app": resource.Name}
	dbHost := fmt.Sprintf("%s-rw", resource.Spec.Database.ClusterName)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.Identifier(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.Name,
			Namespace: resource.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  resource.Name,
							Image: resource.Spec.Image,
							Env: []corev1.EnvVar{
								{Name: "DATABASE_HOST", Value: dbHost},
								{Name: "DATABASE_NAME", Value: resource.Spec.Database.DatabaseName},
								{Name: "DATABASE_PORT", Value: "5432"},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: resource.Spec.ContainerPort},
							},
						},
					},
				},
			},
		},
	}
}

func createService(resource ContainerIngressDB) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.Identifier(), Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.Name,
			Namespace: resource.Namespace,
			Labels:    map[string]string{"app": resource.Name},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": resource.Name},
			Ports: []corev1.ServicePort{
				{Port: 80, TargetPort: intstr.FromInt(int(resource.Spec.ContainerPort))},
			},
		},
	}
}

func createIngress(resource ContainerIngressDB) *networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{APIVersion: networkingv1.SchemeGroupVersion.Identifier(), Kind: "Ingress"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.Name,
			Namespace: resource.Namespace,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: resource.Spec.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     resource.Spec.Path,
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: resource.Name,
											Port: networkingv1.ServiceBackendPort{Number: 80},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if resource.Spec.TLSSecretName != "" {
		ingress.Spec.TLS = []networkingv1.IngressTLS{
			{SecretName: resource.Spec.TLSSecretName, Hosts: []string{resource.Spec.Host}},
		}
	}

	return ingress
}

func createCNPGCluster(resource ContainerIngressDB) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "postgresql.cnpg.io/v1",
			"kind":       "Cluster",
			"metadata": map[string]interface{}{
				"name":      resource.Spec.Database.ClusterName,
				"namespace": resource.Namespace,
			},
			"spec": map[string]interface{}{
				"instances": resource.Spec.Database.Instances,
				"imageName": fmt.Sprintf("ghcr.io/cloudnative-pg/postgresql:%s", resource.Spec.Database.PostgresVersion),
				"storage": map[string]interface{}{
					"size": resource.Spec.Database.StorageSize,
				},
				"bootstrap": map[string]interface{}{
					"initdb": map[string]interface{}{
						"database": resource.Spec.Database.DatabaseName,
					},
				},
			},
		},
	}
}
