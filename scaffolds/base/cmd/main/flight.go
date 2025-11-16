package main

import (
	_ "embed"
	"encoding/json"
	"io"
	"os"

	stolos_yoke "github.com/stolos-cloud/stolos/yoke-base/pkg/stolos-yoke"
	corev1 "k8s.io/api/core/v1"
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

	stolos_yoke.Run[Base](airway, run)
}

func run() ([]byte, error) {
	var base Base // Yoke will pass your Custom Resource instance here via stdin.
	if err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&base); err != nil && err != io.EOF {
		return nil, err
	}

	// Validation step (Customize)
	if err := validateSpec(&base); err != nil && err != io.EOF {
		return nil, err
	}

	// Create the k8s resources for your application.
	return json.Marshal([]flight.Resource{
		createResources(base),
	})
}

func validateSpec(base *Base) error {
	// TODO : Validate the spec and set sane defaults

	if base != nil {
		return nil
	}

	return nil
}

// TODO : Implement functions which return standard k8s resources to create.
func createResources(base Base) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		Data: map[string]string{
			"HelloWorld": base.Spec.SomeProperty,
		},
	}
}
