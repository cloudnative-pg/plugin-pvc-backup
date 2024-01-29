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
)

// ValidateClusterCreate validates a cluster that is being created
func (Implementation) ValidateClusterCreate(
	_ context.Context,
	request *operator.OperatorValidateClusterCreateRequest,
) (*operator.OperatorValidateClusterCreateResult, error) {
	result := &operator.OperatorValidateClusterCreateResult{}

	helper, err := pluginhelper.NewFromCluster(metadata.Data.Name, request.Definition)
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

	oldClusterHelper, err := pluginhelper.NewFromCluster(metadata.Data.Name, request.OldCluster)
	if err != nil {
		return nil, fmt.Errorf("while parsing old cluster: %w", err)
	}

	newClusterHelper, err := pluginhelper.NewFromCluster(metadata.Data.Name, request.NewCluster)
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

	return result
}
