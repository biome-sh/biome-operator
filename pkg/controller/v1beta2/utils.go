// Copyright (c) 2017 Chef Software Inc. and/or applicable contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1beta2

import (
	"fmt"
	"reflect"
	"time"

	"github.com/biome-sh/biome-operator/pkg/apis/biome"
	habv1beta1 "github.com/biome-sh/biome-operator/pkg/apis/biome/v1beta1"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	pollInterval   = 500 * time.Millisecond
	timeOut        = 10 * time.Second
	biomeCRDName = biov1beta1.BiomeResourcePlural + "." + biome.GroupName
)

type keyNotFoundError struct {
	key string
}

func (err keyNotFoundError) Error() string {
	return fmt.Sprintf("could not find Object with key %s in the cache", err.key)
}

func validateCustomObject(h biov1beta1.Biome) error {
	spec := h.Spec.V1beta2
	if spec == nil {
		return fmt.Errorf("missing `spec.v1beta2` field")
	}

	switch spec.Service.Topology {
	case biov1beta1.TopologyStandalone:
	case biov1beta1.TopologyLeader:
	default:
		return fmt.Errorf("unknown topology: %s", spec.Service.Topology)
	}

	if rsn := spec.Service.RingSecretName; rsn != nil {
		rsn := *rsn
		ringParts := ringRegexp.FindStringSubmatch(rsn)

		// The ringParts slice should have a second element for the capturing group
		// in the ringRegexp regular expression, containing the ring's name.
		if len(ringParts) < 2 {
			return fmt.Errorf("malformed ring secret name: %s", rsn)
		}
	}

	return nil
}

// listOptions adds filtering for Biome objects by adding a requirement
// for the Biome label.
func listOptions() func(*metav1.ListOptions) {
	ls := labels.SelectorFromSet(labels.Set(map[string]string{
		habv1beta1.BiomeLabel: "true",
	}))

	return func(options *metav1.ListOptions) {
		options.LabelSelector = ls.String()
	}
}

func CreateCRD(clientset apiextensionsclient.Interface) (*apiextensionsv1beta1.CustomResourceDefinition, error) {
	name := biov1beta1.Kind(habv1beta1.BiomeResourcePlural)

	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name.String(),
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   biov1beta1.SchemeGroupVersion.Group,
			Version: biov1beta1.SchemeGroupVersion.Version,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:     biov1beta1.BiomeResourcePlural,
				Kind:       reflect.TypeOf(habv1beta1.Biome{}).Name(),
				ShortNames: []string{habv1beta1.BiomeShortName},
			},
		},
	}

	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		return nil, err
	}

	// wait for CRD being established.
	err = wait.Poll(pollInterval, timeOut, func() (bool, error) {
		crd, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(biomeCRDName, metav1.GetOptions{})

		if err != nil {
			return false, err
		}

		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					return true, err
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					// TODO re-introduce logging?
					// fmt.Printf("Error: Name conflict: %v\n", cond.Reason)
				}
			}
		}

		return false, err
	})

	// delete CRD if there was an error.
	if err != nil {
		deleteErr := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(biomeCRDName, nil)
		if deleteErr != nil {
			return nil, errors.NewAggregate([]error{err, deleteErr})
		}

		return nil, err
	}

	return crd, nil
}

func checkCustomVersionMatch(v *string) error {
	if v == nil {
		return fmt.Errorf("no CustomVersion field provided")
	}
	if *v == "v1beta2" {
		return nil
	}

	return fmt.Errorf("wrong CustomVersion: %v", *v)
}
