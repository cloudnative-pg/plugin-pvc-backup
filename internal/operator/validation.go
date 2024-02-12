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
	"fmt"

	"github.com/cloudnative-pg/cnpg-i/pkg/operator"

	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/metadata"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/pluginhelper"
)

const (
	imagePullPolicyParameter = "imagePullPolicy"
	imageNameParameter       = "image"
	pvcNameParameter         = "pvc"
	secretNameParameter      = "secretName"
	secretKeyParameter       = "secretKey"
)

// ValidateClusterCreate validates a cluster that is being created
func (Implementation) ValidateClusterCreate(
	_ context.Context,
	request *operator.OperatorValidateClusterCreateRequest,
) (*operator.OperatorValidateClusterCreateResult, error) {
	result := &operator.OperatorValidateClusterCreateResult{}

	helper, err := pluginhelper.NewFromJSONCluster(metadata.Data.Name, request.Definition)
	if err != nil {
		return nil, err
	}

	result.ValidationErrors = append(result.ValidationErrors, validateParameters(helper)...)

	return result, nil
}

// ValidateClusterChange validates a cluster that is being changed
func (Implementation) ValidateClusterChange(
	_ context.Context,
	request *operator.OperatorValidateClusterChangeRequest,
) (*operator.OperatorValidateClusterChangeResult, error) {
	result := &operator.OperatorValidateClusterChangeResult{}

	oldClusterHelper, err := pluginhelper.NewFromJSONCluster(metadata.Data.Name, request.OldCluster)
	if err != nil {
		return nil, fmt.Errorf("while parsing old cluster: %w", err)
	}

	newClusterHelper, err := pluginhelper.NewFromJSONCluster(metadata.Data.Name, request.NewCluster)
	if err != nil {
		return nil, fmt.Errorf("while parsing new cluster: %w", err)
	}

	result.ValidationErrors = append(result.ValidationErrors, validateParameters(newClusterHelper)...)

	if newClusterHelper.Parameters[pvcNameParameter] != oldClusterHelper.Parameters[pvcNameParameter] {
		result.ValidationErrors = append(
			result.ValidationErrors,
			newClusterHelper.ValidationErrorForParameter(pvcNameParameter, "cannot be changed"))
	}

	return result, nil
}

func validateParameters(helper *pluginhelper.Data) []*operator.ValidationError {
	result := make([]*operator.ValidationError, 0)

	if len(helper.Parameters[pvcNameParameter]) == 0 {
		result = append(
			result,
			helper.ValidationErrorForParameter(pvcNameParameter, "cannot be empty"))
	}

	if len(helper.Parameters[imageNameParameter]) == 0 {
		result = append(
			result,
			helper.ValidationErrorForParameter(imageNameParameter, "cannot be empty"))
	}

	if len(helper.Parameters[secretNameParameter]) == 0 {
		result = append(
			result,
			helper.ValidationErrorForParameter(secretNameParameter, "cannot be empty"))
	}

	if len(helper.Parameters[secretKeyParameter]) == 0 {
		result = append(
			result,
			helper.ValidationErrorForParameter(secretKeyParameter, "cannot be empty"))
	}

	return result
}
