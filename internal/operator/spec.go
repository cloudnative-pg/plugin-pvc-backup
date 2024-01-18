package operator

import corev1 "k8s.io/api/core/v1"

func getSidecarContainer(parameters map[string]string) corev1.Container {
	return corev1.Container{
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
			{
				Name:      "pgdata",
				MountPath: "/var/lib/postgresql/data",
			},
		},
		Image:           parameters["image"],
		ImagePullPolicy: corev1.PullPolicy(parameters[imagePullPolicyParameter]),
	}
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
