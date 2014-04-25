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
	populateCompleteViewContext(viewContext)

	sectionList := parseTemplateFromFile("../api-response-examples/v1/doctor/visit/review_v2_template.json", t)
	_, err := sectionList.Render(viewContext)
	if err != nil {
		t.Fatalf("Error rendering layout:%s", err)
	}
}

func TestRenderingLayoutForDoctorVisitReview_ContentLabels(t *testing.T) {
	viewContext := common.ViewContext(map[string]interface{}{})
	populateCompleteViewContext(viewContext)

	// change one of the content labels list content to populate CheckedUncheckedData items
	viewContext.Set("q_skin_description:question_summary", "testing5")
	viewContext.Set("q_skin_description:answers", []CheckedUncheckedData{
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
			IsChecked: true,
		},
	})

	sectionList := parseTemplateFromFile("../api-response-examples/v1/doctor/visit/review_v2_template.json", t)
	_, err := sectionList.Render(viewContext)
	if err != nil {
		t.Fatalf("Error rendering layout:%s", err)
	}

	// now change it to titlesubtitlesubtiems type with just the title set
	viewContext.Set("q_skin_description:question_summary", "testing3")
	viewContext.Set("q_skin_description:answers", []TitleSubtitleSubItemsData{
		TitleSubtitleSubItemsData{
			Title: "testing3",
		},
		TitleSubtitleSubItemsData{
			Title: "testing3",
		},
	})
	_, err = sectionList.Render(viewContext)
	if err != nil {
		t.Fatalf("Error rendering layout:%s", err)
	}
}

func TestRenderingLayoutForDoctorVisitReview_EmptyStateViews(t *testing.T) {
	viewContext := common.ViewContext(map[string]interface{}{})
	populateCompleteViewContext(viewContext)

	// delete certain entries and specify the empty state instead
	viewContext.Delete("patient_visit_alerts")
	viewContext.Set("patient_visit_alerts:empty_state_text", "No alerts specified")

	sectionList := parseTemplateFromFile("../api-response-examples/v1/doctor/visit/review_v2_template.json", t)
	_, err := sectionList.Render(viewContext)
	if err != nil {
		t.Fatalf("Error rendering layout:%s", err)
	}

	// do the same for the empty_title_subtitle_labels
	viewContext.Delete("q_changes_acne_worse:answers")
	viewContext.Set("q_changes_acne_worse:empty_state_text", "Patient chose not to answer")
	_, err = sectionList.Render(viewContext)
	if err != nil {
		t.Fatalf("Error rendering layout:%s", err)
	}

}

func populateCompleteViewContext(viewContext common.ViewContext) {
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

	viewContext.Set("q_other_conditions_acne:question_summary", "testing5")
	viewContext.Set("q_other_conditions_acne:answers", []CheckedUncheckedData{
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
	viewContext.Set("q_reason_visit:answers", "testing")

	viewContext.Set("q_onset_acne:question_summary", "testing5")
	viewContext.Set("q_onset_acne:answers", "testing")

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
	viewContext.Set("q_acne_worse:answers", "testing")

	viewContext.Set("q_changes_acne_worse:question_summary", "testing5")
	viewContext.Set("q_changes_acne_worse:answers", "testing")

	viewContext.Set("q_skin_description:question_summary", "testing5")
	viewContext.Set("q_skin_description:answers", []string{"testing", "testing1", "testing2"})

	viewContext.Set("q_acne_prev_treatment_types:question_summary", "testing5")
	viewContext.Set("q_acne_prev_treatment_types:answers", []string{"testing", "testing1", "testing2"})

	viewContext.Set("q_anything_else_acne:question_summary", "testing5")
	viewContext.Set("q_anything_else_acne:answers", "testing")
}
