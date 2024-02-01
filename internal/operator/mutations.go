/*
Copyright The CloudNativePG Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package operator

import (
	"context"

	"github.com/cloudnative-pg/cnpg-i/pkg/operator"
	corev1 "k8s.io/api/core/v1"

	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/metadata"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/pluginhelper"
)

// MutateCluster is called to mutate a cluster with the defaulting webhook.
// This function is defaulting the "imagePullPolicy" plugin parameter
func (Implementation) MutateCluster(
	_ context.Context,
	request *operator.OperatorMutateClusterRequest,
) (*operator.OperatorMutateClusterResult, error) {
	helper, err := pluginhelper.NewFromCluster(metadata.Data.Name, request.Definition)
	if err != nil {
		return nil, err
	}

	mutatedCluster := helper.GetCluster().DeepCopy()
	for i := range mutatedCluster.Spec.Plugins {
		if mutatedCluster.Spec.Plugins[i].Name != metadata.Data.Name {
			continue
		}

		if mutatedCluster.Spec.Plugins[i].Parameters == nil {
			mutatedCluster.Spec.Plugins[i].Parameters = make(map[string]string)
		}

		if _, ok := mutatedCluster.Spec.Plugins[i].Parameters[imagePullPolicyParameter]; !ok {
			mutatedCluster.Spec.Plugins[i].Parameters[imagePullPolicyParameter] = string(corev1.PullAlways)
		}
	}

	patch, err := helper.CreateClusterJSONPatch(*mutatedCluster)
	if err != nil {
		return nil, err
	}

	return &operator.OperatorMutateClusterResult{
		JsonPatch: patch,
	}, nil
}

// MutatePod is called to mutate a Pod before it will be created
func (Implementation) MutatePod(
	_ context.Context,
	request *operator.OperatorMutatePodRequest,
) (*operator.OperatorMutatePodResult, error) {
	helper, err := pluginhelper.NewFromClusterAndPod(metadata.Data.Name, request.ClusterDefinition, request.PodDefinition)
	if err != nil {
		return nil, err
	}

	mutatedPod := helper.GetPod().DeepCopy()
	helper.InjectPluginVolume(mutatedPod)

	// Inject sidecar
	if len(mutatedPod.Spec.Containers) > 0 {
		mutatedPod.Spec.Containers = append(
			mutatedPod.Spec.Containers,
			getSidecarContainer(helper.Parameters))
	}

	// Inject backup volume
	if len(mutatedPod.Spec.Volumes) > 0 {
		mutatedPod.Spec.Volumes = append(
			mutatedPod.Spec.Volumes,
			getBackupVolume(helper.Parameters))
	}

	patch, err := helper.CreatePodJSONPatch(*mutatedPod)
	if err != nil {
		return nil, err
	}

	return &operator.OperatorMutatePodResult{
		JsonPatch: patch,
	}, nil
}
