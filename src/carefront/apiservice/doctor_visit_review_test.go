package apiservice

import (
	"carefront/common"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/SpruceHealth/mapstructure"
)

func TestParsingTemplateForDoctorVisitReview(t *testing.T) {

	parseTemplateFromFile("../api-response-examples/v1/doctor/visit/review_v2_template.json", t)
}

func TestParsingLayoutForDoctorVisitReview(t *testing.T) {
	parseTemplateFromFile("../api-response-examples/v1/doctor/visit/review_v2.json", t)
}

func parseTemplateFromFile(fileLocation string, t *testing.T) DVisitReviewSectionListView {
	fileContents, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		t.Fatalf("error parsing file: %s", err)
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(fileContents, &jsonData)
	if err != nil {
		t.Fatalf("error unmarshalling file contents into json: %s", err)
	}

	sectionList := &DVisitReviewSectionListView{}
	decoderConfig := &mapstructure.DecoderConfig{
		Result:  sectionList,
		TagName: "json",
	}
	if err := decoderConfig.SetRegistry(dVisitReviewViewTypeRegistry.Map()); err != nil {
		t.Fatalf("Error setting registry for decoder config: %s", err)
	}

	d, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		t.Fatalf("error creating new decoder: %s", err)
	}

	err = d.Decode(jsonData)
	if err != nil {
		t.Fatalf("error decoding template into native go structures: %s", err)
	}

	return *sectionList
}

func TestRenderingLayoutForDoctorVisitReview(t *testing.T) {
	viewContext := common.ViewContext(map[string]interface{}{})
	viewContext.Set("patient_visit_photos", []PhotoData{
		PhotoData{
			Title:          "Left Photo",
			PlaceholderUrl: "testing",
		},
		PhotoData{
			Title:          "Right Photo",
			PlaceholderUrl: "testing",
		},
	})

	viewContext.Set("patient_visit_alerts", []string{
		"testing1",
		"testing2",
		"testing3",
	})

	viewContext.Set("q_allergic_medication_entry:question_summary", "testing")
	viewContext.Set("q_allergic_medication_entry:answers", []string{
		"testing1",
		"testing2",
		"testing3",
	})

	viewContext.Set("q_current_medications_entry:question_summary", "testing3")
	viewContext.Set("q_current_medications_entry:answers", []TitleSubtitleSubItemsData{
		TitleSubtitleSubItemsData{
			Title:    "testing3",
			Subtitle: "testing3",
			SubItems: []string{
				"testing3",
				"testing3",
				"testing3",
			},
		},
		TitleSubtitleSubItemsData{
			Title:    "testing3",
			Subtitle: "testing3",
			SubItems: []string{
				"testing3",
				"testing3",
				"testing3",
			},
		},
	})

	viewContext.Set("q_list_prev_skin_condition_diagnosis:question_summary", "testing4")
	viewContext.Set("q_list_prev_skin_condition_diagnosis:answers", []string{
		"testing1",
		"testing2",
		"testing3",
	})

	viewContext.Set("q_other_skin_condition_entry:question_summary", "testing5")
	viewContext.Set("q_other_skin_condition_entry:answers", []CheckedUncheckedData{
		CheckedUncheckedData{
			Value:     "val1",
			IsChecked: true,
		},
		CheckedUncheckedData{
			Value:     "val2",
			IsChecked: false,
		},
		CheckedUncheckedData{
			Value:     "val3",
			IsChecked: false,
		},
		CheckedUncheckedData{
			Value:     "val4",
			IsChecked: false,
		},
	})

	viewContext.Set("q_reason_visit:question_summary", "testing5")
	viewContext.Set("q_reason_visit:answer", "testing")

	viewContext.Set("q_onset_acne:question_summary", "testing5")
	viewContext.Set("q_onset_acne:answer", "testing")

	viewContext.Set("q_acne_location:question_summary", "testing5")
	viewContext.Set("q_acne_location:answers", []CheckedUncheckedData{
		CheckedUncheckedData{
			Value:     "val1",
			IsChecked: true,
		},
		CheckedUncheckedData{
			Value:     "val2",
			IsChecked: false,
		},
		CheckedUncheckedData{
			Value:     "val3",
			IsChecked: true,
		},
		CheckedUncheckedData{
			Value:     "val4",
			IsChecked: false,
		},
	})

	viewContext.Set("q_acne_symptoms:question_summary", "testing5")
	viewContext.Set("q_acne_symptoms:answers", []CheckedUncheckedData{
		CheckedUncheckedData{
			Value:     "val1",
			IsChecked: true,
		},
		CheckedUncheckedData{
			Value:     "val2",
			IsChecked: false,
		},
		CheckedUncheckedData{
			Value:     "val3",
			IsChecked: true,
		},
		CheckedUncheckedData{
			Value:     "val4",
			IsChecked: true,
		},
	})

	viewContext.Set("q_acne_prev_treatment_list:question_summary", "testing3")
	viewContext.Set("q_acne_prev_treatment_list:answers", []TitleSubtitleSubItemsData{
		TitleSubtitleSubItemsData{
			Title:    "testing3",
			Subtitle: "testing3",
			SubItems: []string{
				"testing3",
				"testing3",
				"testing3",
			},
		},
		TitleSubtitleSubItemsData{
			Title:    "testing3",
			Subtitle: "testing3",
			SubItems: []string{
				"testing3",
				"testing3",
				"testing3",
			},
		},
	})

	viewContext.Set("q_acne_prev_otc_treatment_list:question_summary", "testing3")
	viewContext.Set("q_acne_prev_otc_treatment_list:answers", []TitleSubtitleSubItemsData{
		TitleSubtitleSubItemsData{
			Title:    "testing3",
			Subtitle: "testing3",
			SubItems: []string{
				"testing3",
				"testing3",
				"testing3",
			},
		},
		TitleSubtitleSubItemsData{
			Title:    "testing3",
			Subtitle: "testing3",
			SubItems: []string{
				"testing3",
				"testing3",
				"testing3",
			},
		},
	})

	viewContext.Set("q_acne_worse:question_summary", "testing5")
	viewContext.Set("q_acne_worse:answer", "testing")

	viewContext.Set("q_changes_acne_worse:question_summary", "testing5")
	viewContext.Set("q_changes_acne_worse:answer", "testing")

	viewContext.Set("q_skin_description:question_summary", "testing5")
	viewContext.Set("q_skin_description:answers", []string{"testing", "testing1", "testing2"})

	viewContext.Set("q_acne_prev_treatment_types:question_summary", "testing5")
	viewContext.Set("q_acne_prev_treatment_types:answers", []string{"testing", "testing1", "testing2"})

	viewContext.Set("q_anything_else_acne:question_summary", "testing5")
	viewContext.Set("q_anything_else_acne:answer", "testing")

	sectionList := parseTemplateFromFile("../api-response-examples/v1/doctor/visit/review_v2_template.json", t)
	_, err := sectionList.Render(viewContext)
	if err != nil {
		t.Fatalf("Error rendering layout:%s", err)
	}
}
