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

package clusterwide

import (
	"fmt"
	"strings"
	"testing"
	"time"

	habv1beta1 "github.com/biome-sh/biome-operator/pkg/apis/biome/v1beta1"
	utils "github.com/biome-sh/biome-operator/test/e2e/v1beta1/framework"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	serviceStartupWaitTime = 1 * time.Minute
	secretUpdateTimeout    = 2 * time.Minute
	secretUpdateQueryTime  = 10 * time.Second

	configMapName = "peer-watch-file"
)

// TestBind tests that the operator correctly created two Biome Services and bound them together.
func TestBind(t *testing.T) {
	// Get Biome object from Biome go example.
	web, err := utils.ConvertBiome("resources/bind-config/webapp.yml")
	if err != nil {
		t.Fatal(err)
	}

	if err := framework.CreateBiome(web); err != nil {
		t.Fatal(err)
	}

	// Get Biome object from Biome db example.
	db, err := utils.ConvertBiome("resources/bind-config/db.yml")
	if err != nil {
		t.Fatal(err)
	}

	if err := framework.CreateBiome(db); err != nil {
		t.Fatal(err)
	}

	// Get Service object from example file.
	svc, err := utils.ConvertService("resources/bind-config/service.yml")
	if err != nil {
		t.Fatal(err)
	}

	// Create Service.
	_, err = framework.KubeClient.CoreV1().Services(TestNSClusterwide).Create(svc)
	if err != nil {
		t.Fatal(err)
	}
	// Delete Service so it doesn't interfere with other tests.
	defer (func(name string) {
		if err := framework.DeleteService(name); err != nil {
			t.Fatal(err)
		}
	})(svc.Name)

	// Get Secret object from example file.
	sec, err := utils.ConvertSecret("resources/bind-config/secret.yml")
	if err != nil {
		t.Fatal(err)
	}

	// Create Secret.
	sec, err = framework.KubeClient.CoreV1().Secrets(TestNSClusterwide).Create(sec)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for resources to be ready.
	if err := framework.WaitForResources(habv1beta1.BiomeNameLabel, web.ObjectMeta.Name, 1); err != nil {
		t.Fatal(err)
	}
	if err := framework.WaitForResources(habv1beta1.BiomeNameLabel, db.ObjectMeta.Name, 1); err != nil {
		t.Fatal(err)
	}

	// Wait until endpoints are ready.
	if err := framework.WaitForEndpoints(svc.ObjectMeta.Name); err != nil {
		t.Fatal(err)
	}

	time.Sleep(serviceStartupWaitTime)

	loadBalancerIP, err := framework.GetLoadBalancerIP(svc.ObjectMeta.Name)
	if err != nil {
		t.Fatal(err)
	}
	// Get response from Biome Service.
	url := fmt.Sprintf("http://%s:5555/", loadBalancerIP)

	body, err := utils.QueryService(url)
	if err != nil {
		t.Fatal(err)
	}

	// This msg is set in the config of the biome/bindgo-bio Go Biome Service.
	expectedMsg := "hello from port: 4444"
	actualMsg := body
	// actualMsg can contain whitespace and newlines or different formatting,
	// the only thing we need to check is it contains the expectedMsg.
	if !strings.Contains(actualMsg, expectedMsg) {
		t.Fatalf("Biome Service msg does not match one in default.toml. Expected: \"%s\", got: \"%s\"", expectedMsg, actualMsg)
	}

	// Test `user.toml` updates.

	// Update secret.
	newPort := "port = 6333"

	sec.Data["user.toml"] = []byte(newPort)
	if _, err = framework.KubeClient.CoreV1().Secrets(TestNSClusterwide).Update(sec); err != nil {
		t.Fatalf("Could not update Secret: \"%s\"", err)
	}

	// Wait for SecretVolume to be updated.
	ticker := time.NewTicker(secretUpdateQueryTime)
	defer ticker.Stop()
	timer := time.NewTimer(secretUpdateTimeout)
	defer timer.Stop()

	// Update the message set in the config of the biome/bindgo-bio Go Biome Service.
	expectedMsg = fmt.Sprintf("hello from port: %v", 6333)
	for {
		// Check that the port differs after the update.
		actualMsg, err := utils.QueryService(url)
		if err != nil {
			t.Fatal(err)
		}

		// actualMsg can contain whitespace and newlines or different formatting,
		// the only thing we need to check is it contains the expectedMsg.
		if strings.Contains(actualMsg, expectedMsg) {
			break
		}

		fail := func() {
			t.Fatalf("Configuration update did not go through. Expected: \"%s\", got: \"%s\"", expectedMsg, actualMsg)
		}

		select {
		case <-timer.C:
			fail()
		case <-ticker.C:
			// This is to avoid infinite loops when go
			// decides to always pick the ticker channel,
			// even when timer channel is ready too.
			select {
			case <-timer.C:
				fail()
			default:
			}
		}
	}
}

// TestBiomeDelete tests Biome deletion.
func TestBiomeDelete(t *testing.T) {
	// Get Biome object from Biome go example.
	biome, err := utils.ConvertBiome("resources/standalone/biome.yml")
	if err != nil {
		t.Fatal(err)
	}

	if err := framework.CreateBiome(biome); err != nil {
		t.Fatal(err)
	}

	// Wait for resources to be ready.
	if err := framework.WaitForResources(habv1beta1.BiomeNameLabel, biome.ObjectMeta.Name, 1); err != nil {
		t.Fatal(err)
	}

	// Delete Biome.
	if err := framework.DeleteBiome(biome.ObjectMeta.Name, TestNSClusterwide); err != nil {
		t.Fatal(err)
	}

	// Wait for resources to be deleted.
	if err := framework.WaitForResources(habv1beta1.BiomeNameLabel, biome.ObjectMeta.Name, 0); err != nil {
		t.Fatal(err)
	}

	// Check if all the resources the operator creates are deleted.
	// We do not care about secrets being deleted, as the user needs to delete those manually.
	d, err := framework.KubeClient.AppsV1beta1().Deployments(TestNSClusterwide).Get(biome.ObjectMeta.Name, metav1.GetOptions{})
	if err == nil && d != nil {
		t.Fatal("Deployment was not deleted.")
	}

	// The CM with the peer IP should still be alive, despite the Biome being deleted as it was created outside of the scope of a Biome.
	_, err = framework.KubeClient.CoreV1().ConfigMaps(TestNSClusterwide).Get(configMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPersistentStorage(t *testing.T) {
	ephemeral, err := utils.ConvertBiome("resources/standalone/biome.yml")
	if err != nil {
		t.Fatal(err)
	}

	persisted, err := utils.ConvertBiome("resources/persistent/biome.yml")
	if err != nil {
		t.Fatal(err)
	}

	if err := framework.CreateBiome(ephemeral); err != nil {
		t.Fatal(err)
	}

	if err := framework.CreateBiome(persisted); err != nil {
		t.Fatal(err)
	}

	// Delete all PVCs at the end of the test.
	// For dynamically provisioned PVs (as is the case on minikube), this will
	// also delete the PVs.
	defer (func(name string) {
		ls := labels.SelectorFromSet(labels.Set(map[string]string{
			habv1beta1.BiomeNameLabel: name,
		}))

		lo := metav1.ListOptions{
			LabelSelector: ls.String(),
		}

		err := framework.KubeClient.CoreV1().PersistentVolumeClaims(TestNSClusterwide).DeleteCollection(&metav1.DeleteOptions{}, lo)
		if err != nil {
			t.Fatal(err)
		}
	})(persisted.Name)

	// Delete the ephemeral resource created
	defer (func(name string) {
		if err := framework.DeleteBiome(name, TestNSClusterwide); err != nil {
			t.Fatal(err)
		}
	})(ephemeral.Name)

	if err := framework.WaitForResources(habv1beta1.BiomeNameLabel, persisted.Name, 1); err != nil {
		t.Fatal(err)
	}

	// Test that persistence is only enabled if requested
	ephemeralSTS, err := framework.KubeClient.AppsV1beta2().StatefulSets(TestNSClusterwide).Get(ephemeral.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(ephemeralSTS.Spec.VolumeClaimTemplates) != 0 {
		t.Fatal("PersistentVolumeClaims created for ephemeral StatefulSet")
	}

	persistedSTS, err := framework.KubeClient.AppsV1beta2().StatefulSets(TestNSClusterwide).Get(persisted.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(persistedSTS.Spec.VolumeClaimTemplates) == 0 {
		t.Fatal("No PersistentVolumeClaims created for persistent StatefulSet")
	}
}
