package app_url

import (
	"github.com/sprucehealth/backend/libs/golog"
	"net/url"
	"strings"
)

var (
	PatientVisitQueueIcon       = &SpruceAsset{name: "patient_visit_queue_icon"}
	IconBlueTreatmentPlan       = &SpruceAsset{name: "icon_blue_treatment_plan"}
	IconReply                   = &SpruceAsset{name: "icon_reply"}
	IconHomeVisitNormal         = &SpruceAsset{name: "icon_home_visit_normal"}
	IconHomeTreatmentPlanNormal = &SpruceAsset{name: "icon_home_treatmentplan_normal"}
	IconHomeConversationNormal  = &SpruceAsset{name: "icon_home_conversation_normal"}
	TmpSignature                = &SpruceAsset{name: "tmp_signature"}
	IconRX                      = &SpruceAsset{name: "icon_plan_rx"}
	IconOTC                     = &SpruceAsset{name: "icon_plan_otc"}
	IconMessage                 = &SpruceAsset{name: "icon_message"}
)

type SpruceAsset struct {
	name string
}

func (s SpruceAsset) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(spruceImageUrl)+len(s.name)+2)
	b = append(b, '"')
	b = append(b, []byte(spruceImageUrl)...)
	b = append(b, []byte(s.name)...)
	b = append(b, '"')

	return b, nil
}

func (s SpruceAsset) String() string {
	return spruceImageUrl + s.name
}

func (s *SpruceAsset) UnmarshalJSON(data []byte) error {
	if len(data) < 3 {
		return nil
	}
	incomingUrl := string(data[1 : len(data)-1])
	spruceUrlComponents, err := url.Parse(incomingUrl)
	if err != nil {
		golog.Errorf("Unable to parse url for spruce asset %s", err)
		return err
	}
	pathComponents := strings.Split(spruceUrlComponents.Path, "/")
	if len(pathComponents) < 3 {
		golog.Errorf("Unable to break path %#v into its components when attempting to unmarshal %s", pathComponents, incomingUrl)
		return nil
	}
	s.name = pathComponents[2]
	return nil
}
