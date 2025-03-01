// Copyright (c) 2018 Chef Software Inc. and/or applicable contributors
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

	habv1beta1 "github.com/biome-sh/biome-operator/pkg/apis/biome/v1beta1"

	"github.com/go-kit/kit/log/level"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

const persistentVolumeName = "persistent"

func (hc *BiomeController) newStatefulSet(h *habv1beta1.Biome) (*appsv1.StatefulSet, error) {
	hs := h.Spec.V1beta2

	// This value needs to be passed as a *int32, so we convert it, assign it to a
	// variable and afterwards pass a pointer to it.
	count := int32(hs.Count)

	// Set the service arguments we send to Biome.
	var bioArgs []string
	if hs.Service.Group != nil {
		// When a service is started without explicitly naming the group,
		// it's assigned to the default group.
		habArgs = append(habArgs,
			"--group", *hs.Service.Group)
	}

	if hs.Service.Channel != nil {
		// When a service is started without explicitly naming the channel,
		// it's assigned to the stable channel.
		habArgs = append(habArgs,
			"--channel", *hs.Service.Channel)
	}

	// As we want to label our pods with the
	// topology type we set standalone as the default one.
	// We do not need to pass this to biome, as if no topology
	// is set, biome by default sets standalone topology.
	topology := biov1beta1.TopologyStandalone

	if hs.Service.Topology == biov1beta1.TopologyLeader {
		topology = biov1beta1.TopologyLeader
	}

	path := fmt.Sprintf("%s/%s", configMapDir, peerFilename)

	habArgs = append(habArgs,
		"--topology", topology.String(),
		"--peer-watch-file", path,
	)

	// Runtime binding.
	// One Service connects to another forming a producer/consumer relationship.
	for _, bind := range hs.Service.Bind {
		// Pass --bind flag.
		bindArg := fmt.Sprintf("%s:%s.%s", bind.Name, bind.Service, bind.Group)
		habArgs = append(habArgs,
			"--bind", bindArg)
	}

	base := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: h.Name,
			Labels: map[string]string{
				habv1beta1.BiomeLabel:     "true",
				habv1beta1.BiomeNameLabel: h.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					APIVersion: biov1beta1.SchemeGroupVersion.String(),
					Kind:       biov1beta1.BiomeKind,
					Name:       h.Name,
					UID:        h.UID,
				},
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					habv1beta1.BiomeNameLabel: h.Name,
				},
			},
			Replicas:            &count,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						habv1beta1.BiomeLabel:         "true",
						habv1beta1.BiomeNameLabel:     h.Name,
						habv1beta1.TopologyLabel:        topology.String(),
						habv1beta1.BiomeTopologyLabel: topology.String(),
					},
				},
				Spec: apiv1.PodSpec{
					InitContainers: []apiv1.Container{},
					Containers: []apiv1.Container{
						{
							Name:  "biome-service",
							Image: hs.Image,
							Args:  bioArgs,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "config",
									MountPath: configMapDir,
									ReadOnly:  true,
								},
							},
							Env: hs.Env,
						},
					},
					// Define the volume for the ConfigMap.
					Volumes: []apiv1.Volume{
						{
							Name: "config",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: configMapName,
									},
									Items: []apiv1.KeyToPath{
										{
											Key:  peerFile,
											Path: peerFilename,
										},
									},
								},
							},
						},
					},
				},
			},
			// We delete pods manually in the controller when StatefulSet
			// objects are updated. Setting UpdateStrategy to OnDelete
			// prevents us messing with the StatefulSet controller when
			// StatefulSet is updated.
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.OnDeleteStatefulSetStrategyType,
			},
		},
	}

	spec := &base.Spec
	tSpec := &spec.Template.Spec

	// If we have a secret name present we should mount that secret.
	if hs.Service.ConfigSecretName != nil {
		// Let's make sure our secret is there before mounting it.
		secret, err := hc.config.KubernetesClientset.CoreV1().Secrets(h.Namespace).Get(*hs.Service.ConfigSecretName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		secretVolume := &apiv1.Volume{
			Name: userConfigFilename,
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName: secret.Name,
					Items: []apiv1.KeyToPath{
						{
							Key:  userTOMLFile,
							Path: userTOMLFile,
						},
					},
				},
			},
		}

		secretVolumeMount := &apiv1.VolumeMount{
			Name: userConfigFilename,
			// The Biome supervisor creates a directory for each service under /hab/svc/<servicename>.
			// We need to place the user.toml file in there in order for it to be detected.
			MountPath: fmt.Sprintf("/bio/user/%s/config", hs.Service.Name),
			ReadOnly:  false,
		}

		tSpec.Containers[0].VolumeMounts = append(tSpec.Containers[0].VolumeMounts, *secretVolumeMount)
		tSpec.Volumes = append(tSpec.Volumes, *secretVolume)
	}

	// If we have a file volume name present we should mount that secret.
	if hs.Service.FilesSecretName != nil {
		// Let's make sure our secret is there before mounting it.
		files, err := hc.config.KubernetesClientset.CoreV1().Secrets(h.Namespace).Get(*hs.Service.FilesSecretName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// In order to mount the volume such that bio can change the permissions, we need to
		//   #1. Create A Secret Volume based on the supplied secret
		//   #2. Create an EmptyDir Volume to hold /hab/svc/NAME/files
		//   #3. Add an initContainer to copy the files from the Secret Volume to the EmptyDir Volume
		//   #4. Mount only the EmptyDir Volume into the biome service container

		// #1
		filesSecretVolume := &apiv1.Volume{
			Name: "files-secrets",
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName: files.Name,
				},
			},
		}
		tSpec.Volumes = append(tSpec.Volumes, *filesSecretVolume)

		filesSecretVolumeMount := &apiv1.VolumeMount{
			Name:      "files-secrets",
			MountPath: "/mnt/files",
		}

		// #2
		filesVolume := &apiv1.Volume{
			Name: filesDirectoryName,
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		}
		tSpec.Volumes = append(tSpec.Volumes, *filesVolume)

		filesVolumeMount := &apiv1.VolumeMount{
			Name: filesDirectoryName,
			// The Biome supervisor creates a directory for each service under /hab/svc/<servicename>.
			// We need to place the files directory there.
			MountPath: fmt.Sprintf("/hab/svc/%s/files", hs.Service.Name),
			ReadOnly:  false,
		}

		// #3
		command := fmt.Sprintf("cp /mnt/files/* /hab/svc/%s/files", hs.Service.Name)
		initContainer := &apiv1.Container{
			Name:         "copy-files",
			Image:        "busybox",
			Command:      []string{"sh", "-c", command},
			VolumeMounts: []apiv1.VolumeMount{},
		}
		initContainer.VolumeMounts = append(initContainer.VolumeMounts, *filesSecretVolumeMount)
		initContainer.VolumeMounts = append(initContainer.VolumeMounts, *filesVolumeMount)

		// #4
		tSpec.InitContainers = append(tSpec.InitContainers, *initContainer)
		tSpec.Containers[0].VolumeMounts = append(tSpec.Containers[0].VolumeMounts, *filesVolumeMount)

	}

	// Mount Persistent Volume, if requested.
	if ps := hs.PersistentStorage; ps != nil {
		vm := &apiv1.VolumeMount{
			Name:      persistentVolumeName,
			MountPath: ps.MountPath,
		}

		tSpec.Containers[0].VolumeMounts = append(tSpec.Containers[0].VolumeMounts, *vm)

		q, err := resource.ParseQuantity(ps.Size)
		if err != nil {
			return nil, fmt.Errorf("Could not parse PersistentStorage.Size: %v", err)
		}

		spec.VolumeClaimTemplates = []apiv1.PersistentVolumeClaim{
			apiv1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      persistentVolumeName,
					Namespace: h.Namespace,
					Labels: map[string]string{
						habv1beta1.BiomeLabel:     "true",
						habv1beta1.BiomeNameLabel: h.Name,
					},
				},
				Spec: apiv1.PersistentVolumeClaimSpec{
					AccessModes: []apiv1.PersistentVolumeAccessMode{
						apiv1.ReadWriteOnce,
					},
					StorageClassName: &ps.StorageClassName,
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceStorage: q,
						},
					},
				},
			},
		}
	}

	// Handle ring key, if one is specified.
	if ringSecretName := hs.Service.RingSecretName; ringSecretName != nil {
		ringSecretName := *ringSecretName
		s, err := hc.config.KubernetesClientset.CoreV1().Secrets(apiv1.NamespaceDefault).Get(ringSecretName, metav1.GetOptions{})
		if err != nil {
			level.Error(hc.logger).Log("msg", "Could not find Secret containing ring key")
			return nil, err
		}

		// The filename under which the ring key is saved.
		ringKeyFile := fmt.Sprintf("%s.%s", ringSecretName, ringKeyFileExt)

		// Extract the bare ring name, by removing the revision.
		// Validation has already been performed by this point.
		ringName := ringRegexp.FindStringSubmatch(ringSecretName)[1]

		v := &apiv1.Volume{
			Name: ringSecretName,
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName: s.Name,
					Items: []apiv1.KeyToPath{
						{
							Key:  ringSecretKey,
							Path: ringKeyFile,
						},
					},
				},
			},
		}

		vm := &apiv1.VolumeMount{
			Name:      ringSecretName,
			MountPath: "/hab/cache/keys",
			// This directory cannot be made read-only, as the supervisor writes to
			// it during its operation.
			ReadOnly: false,
		}

		// Mount ring key file.
		tSpec.Volumes = append(tSpec.Volumes, *v)
		tSpec.Containers[0].VolumeMounts = append(tSpec.Containers[0].VolumeMounts, *vm)

		// Add --ring argument to supervisor invocation.
		tSpec.Containers[0].Args = append(tSpec.Containers[0].Args, "--ring", ringName)
	}

	return base, nil
}

func (hc *BiomeController) cacheStatefulSets() {
	hc.stsInformer = hc.config.KubeInformerFactory.Apps().V1().StatefulSets().Informer()

	hc.stsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    hc.handleStsAdd,
		UpdateFunc: hc.handleStsUpdate,
		DeleteFunc: hc.handleStsDelete,
	})

	hc.stsInformerSynced = hc.stsInformer.HasSynced
}

func (hc *BiomeController) handleStsAdd(obj interface{}) {
	sts, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		level.Error(hc.logger).Log("msg", "Failed to type assert StatefulSet", "obj", obj)
		return
	}

	h, err := hc.getBiomeFromLabeledResource(sts)
	if err != nil {
		level.Error(hc.logger).Log("msg", "Could not find Biome for StatefulSet", "name", sts.Name)
		return
	}

	hc.enqueue(h)
}

func (hc *BiomeController) handleStsUpdate(oldObj, newObj interface{}) {
	oldSTS, ok := oldObj.(*appsv1.StatefulSet)
	if !ok {
		level.Error(hc.logger).Log("msg", "Failed to type assert StatefulSet", "obj", oldObj)
		return
	}

	newSTS, ok := newObj.(*appsv1.StatefulSet)
	if !ok {
		level.Error(hc.logger).Log("msg", "Failed to type assert StatefulSet", "obj", newObj)
		return
	}

	if oldSTS.ResourceVersion == newSTS.ResourceVersion {
		level.Debug(hc.logger).Log("msg", "Update ignored as it didn't change ResourceVersion", "statefulset", newSTS.Name)
		return
	}

	h, err := hc.getBiomeFromLabeledResource(newSTS)
	if err != nil {
		level.Error(hc.logger).Log("msg", "Could not find Biome for StatefulSet", "name", newSTS.Name)
		return
	}

	hc.enqueue(h)
}

func (hc *BiomeController) handleStsDelete(obj interface{}) {
	sts, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		level.Error(hc.logger).Log("msg", "Failed to type assert StatefulSet", "obj", obj)
		return
	}

	h, err := hc.getBiomeFromLabeledResource(sts)
	if err != nil {
		// Could not find Biome, it must have already been removed.
		level.Debug(hc.logger).Log("msg", "Could not find Biome for StatefulSet", "name", sts.Name)
		return
	}

	hc.enqueue(h)
}
