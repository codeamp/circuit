package codeamp

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/codeamp/circuit/plugins/codeamp/constants"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/transistor"
)

func (x *CodeAmp) ComplainIfNotInStaging(r *model.Release, p *model.Project) error {
	fmt.Println("ComplainIfNotInStaging")

	// get staging and production environments
	stagingEnv := model.Environment{}
	if err := x.DB.Where("key = ?", constants.StagingEnvironment).Find(&stagingEnv).Error; err != nil {
		return err
	}

	prodEnv := model.Environment{}
	if err := x.DB.Where("key = ?", constants.StagingEnvironment).Find(&prodEnv).Error; err != nil {
		return err
	}

	// check if input release's environment matches prodEnv. if it doesn't, then there's nothing to complain about
	// as we only care about releases going to prod environment that they have a corresponding
	// successful, staging environment release
	inputReleaseEnv := model.Environment{}
	if err := x.DB.Where("id = ?", r.EnvironmentID).Find(&inputReleaseEnv).Error; err != nil {
		return err
	}

	if inputReleaseEnv.Key != prodEnv.Key {
		return fmt.Errorf("Desired prod env %s does not match input release env %s", prodEnv.Key, inputReleaseEnv.Key)
	}

	f := model.Feature{}
	if err := x.DB.Where("id = ?", r.HeadFeatureID).Find(&f).Error; err != nil {
		return err
	}

	// get a list of all features that were created after the input release's feature
	features := []model.Feature{}
	if err := x.DB.Where("created_at >= ?", f.Model.CreatedAt).Find(&features).Error; err != nil {
		return err
	}

	// prepend input release's head feature into features, so it's inclusive of the input release's
	// head feature when finding a corresponding staging release
	features = append([]model.Feature{f}, features...)

	releaseFoundInStaging := false
	for _, feature := range features {
		// check if successful release exists in staging environment
		stagingRelease := model.Release{}
		if err := x.DB.Where("state = ? and head_feature_id = ? and environment_id = ?", transistor.GetState("complete"), feature.Model.ID, stagingEnv.Model.ID).Find(&stagingRelease).Error; err != nil {
			// if it's any error besides IsRecordNotFound, terminate the function execution here
			if !gorm.IsRecordNotFoundError(err) {
				return err
			}
		} else {
			// a successful release in staging is found
			releaseFoundInStaging = true
			break
		}
	}

	if !releaseFoundInStaging {
		// send notification
		fmt.Println("send notification")
	}

	return nil
}
