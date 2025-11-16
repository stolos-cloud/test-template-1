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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	airway := manifest.Spec

	stolos_yoke.Run[ContainerDeployment](airway, run)
}

func run() ([]byte, error) {
	var deployment ContainerDeployment // Yoke will pass your Custom Resource instance here via stdin.
	if err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&deployment); err != nil && err != io.EOF {
		return nil, err
	}

	// Validation and defaulting
	if err := validateSpec(&deployment); err != nil && err != io.EOF {
		return nil, err
	}

	// Create the k8s resources for your application.
	return json.Marshal([]flight.Resource{
		createDeployment(deployment),
	})
}

func validateSpec(deployment *ContainerDeployment) error {
	if deployment.Spec.Image == "" {
		return fmt.Errorf("spec.image is required")
	}
	if deployment.Spec.Replicas < 0 {
		return fmt.Errorf("spec.replicas cannot be negative")
	}

	// Defaulting
	if deployment.Spec.Replicas == 0 {
		deployment.Spec.Replicas = 1
	}
	if deployment.Spec.Port == 0 {
		deployment.Spec.Port = 80
	}

	return nil
}

func createDeployment(resource ContainerDeployment) *appsv1.Deployment {
	labels := map[string]string{"app": resource.Name}
	replicas := resource.Spec.Replicas

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.Identifier(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.Name,
			Namespace: resource.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  resource.Name,
							Image: resource.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: resource.Spec.Port,
								},
							},
						},
					},
				},
			},
		},
	}
}
