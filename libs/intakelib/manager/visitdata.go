package manager

import (
	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"github.com/sprucehealth/backend/libs/errors"
	"github.com/sprucehealth/backend/libs/intakelib/protobuf/intake"
)

type platform string

const (
	android platform = "android"
	ios     platform = "ios"
)

type visitData struct {
	patientVisitID string
	isSubmitted    bool
	layoutData     dataMap
	userFields     *userFields
	platform       platform
}

// unmarshal parses the incoming data into the visitData by using dataType
// to determine how to parse the incoming data.
func (v *visitData) unmarshal(dataType string, data []byte) error {
	var vd intake.VisitData
	if err := proto.Unmarshal(data, &vd); err != nil {
		return errors.Trace(err)
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(vd.Layout, &jsonMap); err != nil {
		return errors.Trace(err)
	}

	v.patientVisitID = *vd.PatientVisitId
	v.layoutData = jsonMap
	v.isSubmitted = *vd.IsSubmitted
	v.userFields = &userFields{
		fields: make(map[string]interface{}),
	}
	for _, pair := range vd.Pairs {
		if err := v.userFields.set(*pair.Key, *pair.Value); err != nil {
			return err
		}
	}

	// if there are any preferences included in the layout data then lets go ahead and add those
	// to the user fields
	preferences, ok := jsonMap["preferences"]
	if ok {
		preferenceMap := preferences.(map[string]interface{})
		for preferenceKey, preferenceValue := range preferenceMap {
			v.userFields.fields["preference."+preferenceKey] = preferenceValue
		}
	}

	switch *vd.Platform {
	case intake.VisitData_ANDROID:
		v.platform = android
	case intake.VisitData_IOS:
		v.platform = ios
	}

	return nil
}
