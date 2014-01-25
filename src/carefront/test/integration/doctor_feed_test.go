package integration

import (
	"carefront/api"
	"carefront/apiservice"
	"carefront/settings"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDoctorFeed(t *testing.T) {
	if err := CheckIfRunningLocally(t); err == CannotRunTestLocally {
		return
	}
	testData := SetupIntegrationTest(t)
	defer TearDownIntegrationTest(t, testData)

	// get the current primary doctor
	doctorId := getDoctorIdOfCurrentPrimaryDoctor(testData, t)

	doctor, err := testData.DataApi.GetDoctorFromId(doctorId)
	if err != nil {
		t.Fatal("Unable to get doctor from doctor id " + err.Error())
	}

	doctorDisplayFeedTabs := getDoctorQueue(testData, doctor.AccountId, t)
	// there should be no sections, but just two empty tabs
	if doctorDisplayFeedTabs.Tabs == nil {
		t.Fatal("Expected there to be 2 sections instead got none")
	}

	if len(doctorDisplayFeedTabs.Tabs) != 2 {
		t.Fatalf("Expected there to be 2 sections instead got %d", len(doctorDisplayFeedTabs.Tabs))
	}

	for _, tab := range doctorDisplayFeedTabs.Tabs {
		if tab.Sections != nil && len(tab.Sections) != 0 {
			t.Fatalf("Expected there to be no sectioins containing items in the doctor's feed but instead got %d sections with items", len(tab.Sections))
		}
	}

	patientSignedupResponse := SignupRandomTestPatient(t, testData.DataApi, testData.AuthApi)
	// get patient to start a visit
	patientVisitResponse := CreatePatientVisitForPatient(patientSignedupResponse.Patient.PatientId, testData, t)

	// lets go ahead and insert several items into the doctor queue for this doctor
	doctorQueueItem := &api.DoctorQueueItem{}
	doctorQueueItem.DoctorId = doctor.DoctorId
	doctorQueueItem.ItemId = patientVisitResponse.PatientVisitId
	doctorQueueItem.Status = api.QUEUE_ITEM_STATUS_ONGOING
	insertIntoDoctorQueue(testData, doctorQueueItem, t)

	doctorQueueItem = &api.DoctorQueueItem{}
	doctorQueueItem.DoctorId = doctor.DoctorId
	doctorQueueItem.ItemId = patientVisitResponse.PatientVisitId
	doctorQueueItem.Status = api.QUEUE_ITEM_STATUS_PENDING
	insertIntoDoctorQueue(testData, doctorQueueItem, t)

	doctorQueueItem = &api.DoctorQueueItem{}
	doctorQueueItem.DoctorId = doctor.DoctorId
	doctorQueueItem.ItemId = patientVisitResponse.PatientVisitId
	doctorQueueItem.Status = api.QUEUE_ITEM_STATUS_PENDING
	insertIntoDoctorQueue(testData, doctorQueueItem, t)

	// lets go ahead and make a call to get the doctor feed
	doctorDisplayFeedTabs = getDoctorQueue(testData, doctor.AccountId, t)

	// ensure that there are two tabs as required
	if len(doctorDisplayFeedTabs.Tabs) != 2 {
		t.Fatalf("Expected two tabs but got %d", len(doctorDisplayFeedTabs.Tabs))
	}

	// ensure that all the items are in the pending tab
	for _, tab := range doctorDisplayFeedTabs.Tabs {
		switch tab.Title {
		case "Pending":
			if len(tab.Sections) != 2 {
				t.Fatal("Expect there to be 3 sections, one for upcoming visit and another for the rest of the visits")
			}

			// ensure that the first item has the button text set to Continue to indicate an ongoing itgem
			if tab.Sections[0].Items[0].Button.ButtonText != "Continue" {
				t.Fatal("Expected the first item in the list to be the ongoing item. ")
			}

			// ensure that all items in the pending section have the display type set as needed
			if tab.Sections[0].Items[0].DisplayTypes == nil || len(tab.Sections[0].Items[0].DisplayTypes) == 0 {
				t.Fatal("Expected there to exist a list of display types for the item but there arent any")
			} else if tab.Sections[0].Items[0].DisplayTypes[0] != api.DISPLAY_TYPE_TITLE_SUBTITLE_BUTTON {
				t.Fatalf("Expected the display type to be %s for this item in the queue but instead it was %s.", api.DISPLAY_TYPE_TITLE_SUBTITLE_BUTTON, tab.Sections[0].Items[0].DisplayTypes[0])
			}

			for _, item := range tab.Sections[1].Items {
				if item.DisplayTypes == nil || len(item.DisplayTypes) == 0 {
					t.Fatal("Expected there to exist a list of display types for the item but there arent any")
				} else if item.DisplayTypes[0] != api.DISPLAY_TYPE_TITLE_SUBTITLE_NONACTIONABLE {
					t.Fatalf("Expected the display type to be %s for this item in the queue but instead it was %s.", api.DISPLAY_TYPE_TITLE_SUBTITLE_BUTTON, item.DisplayTypes[0])
				}
			}

		case "Completed":
			if tab.Sections != nil && len(tab.Sections) != 0 {
				t.Fatal("Expected there to be no completed sections")
			}
		}
	}

	// test the clustering of completed tasks to ensure it is working as expected
	queueItems := make([]*api.DoctorQueueItem, 0)
	for i := 0; i < 10; i++ {
		queueItem := &api.DoctorQueueItem{}
		queueItem.DoctorId = doctor.DoctorId
		queueItem.ItemId = patientVisitResponse.PatientVisitId
		queueItem.Status = api.QUEUE_ITEM_STATUS_COMPLETED
		queueItem.EnqueueDate = time.Date(2013, 1, i, 0, 0, 0, 0, time.UTC)
		queueItems = append(queueItems, queueItem)
		insertIntoDoctorQueueWithEnqueuedDate(testData, queueItem, t)
	}

	queueItem := &api.DoctorQueueItem{}
	queueItem.DoctorId = doctor.DoctorId
	queueItem.ItemId = patientVisitResponse.PatientVisitId
	queueItem.Status = api.QUEUE_ITEM_STATUS_PHOTOS_REJECTED
	queueItem.EnqueueDate = time.Date(2013, 1, 10, 0, 0, 0, 0, time.UTC)
	queueItems = append(queueItems, queueItem)
	insertIntoDoctorQueueWithEnqueuedDate(testData, queueItem, t)

	doctorDisplayFeedTabs = getDoctorQueue(testData, doctor.AccountId, t)

	// now there should be items in the pending and completed tabs

	// ensure that there are two tabs as required
	if len(doctorDisplayFeedTabs.Tabs) != 2 {
		t.Fatalf("Expected two tabs but got %d", len(doctorDisplayFeedTabs.Tabs))
	}

	// ensure that all the items are in the pending tab
	for _, tab := range doctorDisplayFeedTabs.Tabs {
		switch tab.Title {
		case "Pending":
			if len(tab.Sections) != 2 {
				t.Fatal("Expect there to be 2 sections, one for upcoming visit and another for the rest of the visits")
			}
		case "Completed":
			if len(tab.Sections) != 11 {
				t.Fatalf("Expected there to be 10 completed sections. Instead there were %d", len(tab.Sections))
			}

			// in each of the sections there should be 1 item
			for i, section := range tab.Sections {
				if section.Items == nil {
					t.Fatal("Expected there to be 1 completed item in the section instead there were none")
				}

				if len(section.Items) != 1 {
					t.Fatalf("Expected there to be 1 completed item in the section, instead there were %d", len(section.Items))
				}

				// ensure that all items in the pending section have the display type set as needed
				if section.Items[0].DisplayTypes == nil || len(section.Items[0].DisplayTypes) == 0 {
					t.Fatal("Expected there to exist a list of display types for the item but there arent any")
				} else if i != 0 && section.Items[0].DisplayTypes[0] != api.DISPLAY_TYPE_TITLE_SUBTITLE_ACTIONABLE {
					t.Fatalf("Expected the display type to be %s for this item in the queue.", api.DISPLAY_TYPE_TITLE_SUBTITLE_ACTIONABLE)
				} else if i == 0 && section.Items[0].DisplayTypes[0] != api.DISPLAY_TYPE_TITLE_SUBTITLE_NONACTIONABLE {
					t.Fatalf("Expected the display type to be %s for this item in the queue.", api.DISPLAY_TYPE_TITLE_SUBTITLE_NONACTIONABLE)
				}
			}
		}
	}

	// lets go ahead and remove all items from the doctor queue
	_, err = testData.DB.Exec(`delete from doctor_queue`)
	if err != nil {
		t.Fatal("Unable to delete items from doctor queue")
	}

	// now, lets insert items to test the time left
	startingTime := time.Now().Add(-12 * time.Hour)
	differencesText := make([]string, 0)
	for i := 0; i < 5; i++ {
		queueItem := &api.DoctorQueueItem{}
		queueItem.DoctorId = doctor.DoctorId
		queueItem.ItemId = patientVisitResponse.PatientVisitId
		queueItem.Status = api.QUEUE_ITEM_STATUS_ONGOING
		queueItem.EnqueueDate = startingTime.Add(time.Hour)
		queueItems = append(queueItems, queueItem)
		insertIntoDoctorQueueWithEnqueuedDate(testData, queueItem, t)

		difference := queueItem.EnqueueDate.Add(settings.SLA_TO_SERVICE_CUSTOMER).Sub(time.Now())
		minutesLeft := int64(difference.Minutes()) - (60 * int64(difference.Hours()))
		differenceString := fmt.Sprintf("%dh %dm left", int64(difference.Hours()), int64(minutesLeft))
		differencesText = append(differencesText, differenceString)
	}

	doctorDisplayFeedTabs = getDoctorQueue(testData, doctor.AccountId, t)
	// lets go through the pending items and ensure that the time matches up
	for _, tab := range doctorDisplayFeedTabs.Tabs {
		if tab.Title == "Pending" {
			var i int64
			for _, section := range tab.Sections {
				for _, item := range section.Items {
					if differencesText[i] != item.Subtitle {
						t.Fatalf("Expected the subtitle to be '%s' but was '%s'", differencesText, item.Subtitle)
					}
					i += 1
				}
			}
		}
	}

}

func getDoctorQueue(testData TestData, doctorAccountId int64, t *testing.T) *apiservice.DisplayFeedTabs {
	doctorQueueHandler := &apiservice.DoctorQueueHandler{DataApi: testData.DataApi}
	ts := httptest.NewServer(doctorQueueHandler)
	defer ts.Close()

	resp, err := authGet(ts.URL, doctorAccountId)
	if err != nil {
		t.Fatal("Unable to get doctor feed for doctor: " + err.Error())
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Unable to read body of response: " + err.Error())
	}

	CheckSuccessfulStatusCode(resp, "Unable to make successful call to get doctor feed "+string(respBody), t)

	doctorDisplayFeedTabs := &apiservice.DisplayFeedTabs{}
	err = json.Unmarshal(respBody, doctorDisplayFeedTabs)
	if err != nil {
		t.Fatal("Unable to unmarshal response body into tabs " + err.Error())
	}

	return doctorDisplayFeedTabs
}

func insertIntoDoctorQueue(testData TestData, doctorQueueItem *api.DoctorQueueItem, t *testing.T) {
	_, err := testData.DB.Exec(fmt.Sprintf(`insert into doctor_queue (doctor_id, event_type, item_id, status) 
												values (?, 'PATIENT_VISIT', ?, '%s')`, doctorQueueItem.Status), doctorQueueItem.DoctorId, doctorQueueItem.ItemId)
	if err != nil {
		t.Fatal("Unable to insert item into doctor queue: " + err.Error())
	}
}

func insertIntoDoctorQueueWithEnqueuedDate(testData TestData, doctorQueueItem *api.DoctorQueueItem, t *testing.T) {
	_, err := testData.DB.Exec(fmt.Sprintf(`insert into doctor_queue (doctor_id, event_type, item_id, status, enqueue_date) 
												values (?, 'PATIENT_VISIT', ?, '%s', ?)`, doctorQueueItem.Status), doctorQueueItem.DoctorId, doctorQueueItem.ItemId, doctorQueueItem.EnqueueDate)
	if err != nil {
		t.Fatal("Unable to insert item into doctor queue: " + err.Error())
	}
}
