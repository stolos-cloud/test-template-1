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

	stolos_yoke.Run[ContainerIngress](manifest.Spec, run)
}

func run() ([]byte, error) {
	var resource ContainerIngress
	if err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&resource); err != nil && err != io.EOF {
		return nil, err
	}

	if err := validateSpec(&resource); err != nil && err != io.EOF {
		return nil, err
	}

	deployment := createDeployment(resource)
	service := createService(resource)
	ingress := createIngress(resource)

	return json.Marshal([]flight.Resource{deployment, service, ingress})
}

func validateSpec(resource *ContainerIngress) error {
	if resource.Spec.Image == "" {
		return fmt.Errorf("spec.image is required")
	}
	if resource.Spec.Host == "" {
		return fmt.Errorf("spec.host is required")
	}
	if resource.Spec.Replicas == 0 {
		resource.Spec.Replicas = 1
	}
	if resource.Spec.ContainerPort == 0 {
		resource.Spec.ContainerPort = 8080
	}
	if resource.Spec.Path == "" {
		resource.Spec.Path = "/"
	}
	return nil
}

func createDeployment(resource ContainerIngress) *appsv1.Deployment {
	replicas := resource.Spec.Replicas
	labels := map[string]string{"app": resource.Name}

	return &appsv1.Deployment{
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
							Ports: []corev1.ContainerPort{{ContainerPort: resource.Spec.ContainerPort}},
						},
					},
				},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.Name,
			Namespace: resource.Namespace,
			Labels:    labels,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.Identifier(),
			Kind:       "Deployment",
		},
	}
}

func createService(resource ContainerIngress) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.Identifier(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.Name,
			Namespace: resource.Namespace,
			Labels:    map[string]string{"app": resource.Name},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": resource.Name},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(int(resource.Spec.ContainerPort)),
				},
			},
		},
	}
}

func createIngress(resource ContainerIngress) *networkingv1.Ingress {
	ingress := &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkingv1.SchemeGroupVersion.Identifier(),
			Kind:       "Ingress",
		},
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
									PathType: func() *networkingv1.PathType { pt := networkingv1.PathTypePrefix; return &pt }(),
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
			{
				Hosts:      []string{resource.Spec.Host},
				SecretName: resource.Spec.TLSSecretName,
			},
		}
	}

	return ingress
}
