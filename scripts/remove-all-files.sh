#!/usr/bin/env bash
##
## Copyright The CloudNativePG Contributors
##
## Licensed under the Apache License, Version 2.0 (the "License");
## you may not use this file except in compliance with the License.
## You may obtain a copy of the License at
##
##     http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software
## distributed under the License is distributed on an "AS IS" BASIS,
## WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
## See the License for the specific language governing permissions and
## limitations under the License.
##

set -eu

if [ -z "${1-}" ]; then 
    echo "Use: $0 [pvc-name]"
    exit 1
fi

kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: cleanup-files
spec:
  template:
    metadata:
      creationTimestamp: null
    spec:
      volumes:
      - name: backups
        persistentVolumeClaim:
          claimName: $1
      containers:
      - image: alpine
        name: cleanup-files
        command:
        - sh
        - -c
        - "rm -rf /backups/*"
        volumeMounts:
        - name: backups
          mountPath: /backups
        resources: {}
      restartPolicy: Never
EOF

kubectl wait --for=condition=complete job/cleanup-files
kubectl delete job cleanup-files
