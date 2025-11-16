package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

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

	stolos_yoke.Run[FullStack](manifest.Spec, run)
}

func run() ([]byte, error) {
	var resource FullStack
	if err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&resource); err != nil && err != io.EOF {
		return nil, err
	}

	if err := validateSpec(&resource); err != nil && err != io.EOF {
		return nil, err
	}

	resources := []flight.Resource{
		createBackendDeployment(resource),
		createBackendService(resource),
		createBackendIngress(resource),
		createDatabaseCluster(resource),
		createCacheDeployment(resource),
		createCacheService(resource),
		createFrontendConfigMap(resource),
		createFrontendDeployment(resource),
		createFrontendService(resource),
		createFrontendIngress(resource),
	}

	return json.Marshal(resources)
}

func validateSpec(resource *FullStack) error {
	if resource.Spec.Backend.Image == "" {
		return fmt.Errorf("spec.backend.image is required")
	}
	if resource.Spec.Backend.Host == "" {
		return fmt.Errorf("spec.backend.host is required")
	}
	if resource.Spec.Frontend.Host == "" {
		return fmt.Errorf("spec.frontend.host is required")
	}
	if resource.Spec.Database.ClusterName == "" {
		return fmt.Errorf("spec.database.clusterName is required")
	}
	if resource.Spec.Database.DatabaseName == "" {
		return fmt.Errorf("spec.database.databaseName is required")
	}
	if resource.Spec.Cache.Flavor == "" {
		resource.Spec.Cache.Flavor = "redis"
	}
	if resource.Spec.Cache.Port == 0 {
		resource.Spec.Cache.Port = 6379
	}
	if resource.Spec.Backend.Replicas <= 0 {
		resource.Spec.Backend.Replicas = 2
	}
	if resource.Spec.Backend.ContainerPort == 0 {
		resource.Spec.Backend.ContainerPort = 8080
	}
	if resource.Spec.Backend.Path == "" {
		resource.Spec.Backend.Path = "/api"
	}
	if resource.Spec.Frontend.Replicas <= 0 {
		resource.Spec.Frontend.Replicas = 1
	}
	if resource.Spec.Frontend.Path == "" {
		resource.Spec.Frontend.Path = "/"
	}
	if resource.Spec.Frontend.Image == "" {
		resource.Spec.Frontend.Image = "nginx:stable-alpine"
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
	if resource.Spec.Frontend.StaticContent == "" {
		resource.Spec.Frontend.StaticContent = fmt.Sprintf(`<!doctype html>
<html>
  <head>
    <title>%s</title>
  </head>
  <body>
    <h1>%s</h1>
    <p>Your backend API is available at https://%s%s</p>
  </body>
</html>`, resource.Name, resource.Name, resource.Spec.Backend.Host, resource.Spec.Backend.Path)
	}
	return nil
}

func createBackendDeployment(resource FullStack) *appsv1.Deployment {
	replicas := resource.Spec.Backend.Replicas
	labels := map[string]string{"app": resource.Name}
	dbHost := fmt.Sprintf("%s-rw", resource.Spec.Database.ClusterName)
	cacheService := fmt.Sprintf("%s-cache", resource.Name)

	return &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.Identifier(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{Name: resource.Name, Namespace: resource.Namespace, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  resource.Name,
							Image: resource.Spec.Backend.Image,
							Env: []corev1.EnvVar{
								{Name: "DATABASE_HOST", Value: dbHost},
								{Name: "DATABASE_NAME", Value: resource.Spec.Database.DatabaseName},
								{Name: "DATABASE_PORT", Value: "5432"},
								{Name: "CACHE_HOST", Value: cacheService},
								{Name: "CACHE_PORT", Value: fmt.Sprintf("%d", resource.Spec.Cache.Port)},
							},
							Ports: []corev1.ContainerPort{{ContainerPort: resource.Spec.Backend.ContainerPort}},
						},
					},
				},
			},
		},
	}
}

func createBackendService(resource FullStack) *corev1.Service {
	labels := map[string]string{"app": resource.Name}
	return &corev1.Service{
		TypeMeta:   metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.Identifier(), Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{Name: resource.Name, Namespace: resource.Namespace, Labels: labels},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports:    []corev1.ServicePort{{Port: 80, TargetPort: intstr.FromInt(int(resource.Spec.Backend.ContainerPort))}},
		},
	}
}

func createBackendIngress(resource FullStack) *networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		TypeMeta:   metav1.TypeMeta{APIVersion: networkingv1.SchemeGroupVersion.Identifier(), Kind: "Ingress"},
		ObjectMeta: metav1.ObjectMeta{Name: resource.Name, Namespace: resource.Namespace},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: resource.Spec.Backend.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{Path: resource.Spec.Backend.Path, PathType: &pathType, Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{Name: resource.Name, Port: networkingv1.ServiceBackendPort{Number: 80}}}},
							},
						},
					},
				},
			},
		},
	}
	if resource.Spec.Backend.TLSSecretName != "" {
		ingress.Spec.TLS = []networkingv1.IngressTLS{{Hosts: []string{resource.Spec.Backend.Host}, SecretName: resource.Spec.Backend.TLSSecretName}}
	}
	return ingress
}

func createDatabaseCluster(resource FullStack) *unstructured.Unstructured {
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

func createCacheDeployment(resource FullStack) *appsv1.Deployment {
	name := fmt.Sprintf("%s-cache", resource.Name)
	labels := map[string]string{"app": name}
	replicas := int32(1)
	image := resolveCacheImage(resource.Spec.Cache.Flavor)

	return &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.Identifier(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: resource.Namespace, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "cache", Image: image, Ports: []corev1.ContainerPort{{ContainerPort: resource.Spec.Cache.Port}}},
					},
				},
			},
		},
	}
}

func createCacheService(resource FullStack) *corev1.Service {
	name := fmt.Sprintf("%s-cache", resource.Name)
	labels := map[string]string{"app": name}
	return &corev1.Service{
		TypeMeta:   metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.Identifier(), Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: resource.Namespace, Labels: labels},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports:    []corev1.ServicePort{{Port: resource.Spec.Cache.Port, TargetPort: intstr.FromInt(int(resource.Spec.Cache.Port))}},
		},
	}
}

func createFrontendConfigMap(resource FullStack) *corev1.ConfigMap {
	name := fmt.Sprintf("%s-frontend", resource.Name)
	return &corev1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.Identifier(), Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: resource.Namespace},
		Data:       map[string]string{"index.html": resource.Spec.Frontend.StaticContent},
	}
}

func createFrontendDeployment(resource FullStack) *appsv1.Deployment {
	name := fmt.Sprintf("%s-frontend", resource.Name)
	labels := map[string]string{"app": name}
	replicas := resource.Spec.Frontend.Replicas

	return &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.Identifier(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: resource.Namespace, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{Name: "site", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: name}}}},
					},
					Containers: []corev1.Container{
						{
							Name:         "frontend",
							Image:        resource.Spec.Frontend.Image,
							Ports:        []corev1.ContainerPort{{ContainerPort: 80}},
							VolumeMounts: []corev1.VolumeMount{{Name: "site", MountPath: "/usr/share/nginx/html", ReadOnly: true}},
						},
					},
				},
			},
		},
	}
}

func createFrontendService(resource FullStack) *corev1.Service {
	name := fmt.Sprintf("%s-frontend", resource.Name)
	labels := map[string]string{"app": name}
	return &corev1.Service{
		TypeMeta:   metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.Identifier(), Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: resource.Namespace, Labels: labels},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports:    []corev1.ServicePort{{Port: 80, TargetPort: intstr.FromInt(80)}},
		},
	}
}

func createFrontendIngress(resource FullStack) *networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	name := fmt.Sprintf("%s-frontend", resource.Name)
	ingress := &networkingv1.Ingress{
		TypeMeta:   metav1.TypeMeta{APIVersion: networkingv1.SchemeGroupVersion.Identifier(), Kind: "Ingress"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: resource.Namespace},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: resource.Spec.Frontend.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{Path: resource.Spec.Frontend.Path, PathType: &pathType, Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{Name: name, Port: networkingv1.ServiceBackendPort{Number: 80}}}},
							},
						},
					},
				},
			},
		},
	}
	if resource.Spec.Frontend.TLSSecretName != "" {
		ingress.Spec.TLS = []networkingv1.IngressTLS{{Hosts: []string{resource.Spec.Frontend.Host}, SecretName: resource.Spec.Frontend.TLSSecretName}}
	}
	return ingress
}

func resolveCacheImage(flavor string) string {
	switch strings.ToLower(flavor) {
	case "valkey":
		return "docker.io/valkey/valkey:1.7"
	default:
		return "docker.io/redis:7.2"
	}
}
