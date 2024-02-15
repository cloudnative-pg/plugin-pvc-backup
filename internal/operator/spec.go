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
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const pgPath = "/var/lib/postgresql"

func getSidecarContainer(pgPod *corev1.Pod, parameters map[string]string) corev1.Container {
	result := corev1.Container{
		Name: "plugin-pvc-backup",
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "scratch-data",
				MountPath: "/controller",
			},
			{
				Name:      "plugins",
				MountPath: "/plugins",
			},
			{
				Name:      "backups",
				MountPath: "/backup",
			},
		},
		Image:           parameters["image"],
		ImagePullPolicy: corev1.PullPolicy(parameters[imagePullPolicyParameter]),
		Env: []corev1.EnvVar{
			{
				Name: "KOPIA_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: parameters["secretName"],
						},
						Key: parameters["secretKey"],
					},
				},
			},
		},
	}

	volumeMounts := pgPod.Spec.Containers[0].VolumeMounts
	for i := range volumeMounts {
		if strings.HasPrefix(volumeMounts[i].MountPath, pgPath) {
			result.VolumeMounts = append(result.VolumeMounts, volumeMounts[i])
		}
	}

	return result
}

func getBackupVolume(parameters map[string]string) corev1.Volume {
	return corev1.Volume{
		Name: "backups",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: parameters[pvcNameParameter],
			},
		},
	}
}
