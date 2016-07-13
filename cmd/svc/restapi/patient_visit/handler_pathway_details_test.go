package patient_visit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/test/config"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/cfg"
	"github.com/sprucehealth/backend/libs/ptr"
	"github.com/sprucehealth/backend/libs/test"
)

type pathwayDetailsHandlerDataAPI struct {
	api.DataAPI
	pathways       map[string]*common.Pathway
	pathwayCases   []*common.PatientCase
	pathwayDoctors map[string][]*common.Doctor
	careTeams      map[int64]*common.PatientCareTeam
	itemCost       *common.ItemCost
	visits         []*common.PatientVisit
}

// pathwayDetailsRes is a simplified response version of the pathway details handler response.
// It's not possible to use the existing response struct because it uses interfaces.
type pathwayDetailsRes struct {
	Pathways []struct {
		PathwayTag string `json:"pathway_id"`
		Screen     struct {
			Type           string        `json:"type"`
			Title          string        `json:"title"`
			ContentText    string        `json:"content_text,omitempty"`
			ContentSubtext string        `json:"content_subtext,omitempty"`
			Views          []interface{} `json:"views"`
		} `json:"screen"`
		FAQ *struct {
			Views []interface{} `json:"views"`
		} `json:"faq"`
		AgeRestrictions []*pathwayAgeRestriction `json:"age_restrictions,omitempty"`
	} `json:"pathway_details_screens"`
}

func (api *pathwayDetailsHandlerDataAPI) GetPatientIDFromAccountID(accountID int64) (common.PatientID, error) {
	return common.NewPatientID(1), nil
}

func (api *pathwayDetailsHandlerDataAPI) GetCasesForPatient(patientID common.PatientID, states []string) ([]*common.PatientCase, error) {
	return api.pathwayCases, nil
}

func (api *pathwayDetailsHandlerDataAPI) PathwaysForTags(tags []string, opts api.PathwayOption) (map[string]*common.Pathway, error) {
	ps := make(map[string]*common.Pathway, len(tags))
	for _, tag := range tags {
		if p := api.pathways[tag]; p != nil {
			ps[tag] = p
		}
	}
	return ps, nil
}

func (api *pathwayDetailsHandlerDataAPI) CaseCareTeams(caseIDs []int64) (map[int64]*common.PatientCareTeam, error) {
	return api.careTeams, nil
}

func (api *pathwayDetailsHandlerDataAPI) DoctorsForPathway(pathwayTag string, limit int) ([]*common.Doctor, error) {
	return api.pathwayDoctors[pathwayTag], nil
}

func (api *pathwayDetailsHandlerDataAPI) GetActiveItemCost(skuType string) (*common.ItemCost, error) {
	return api.itemCost, nil
}
func (api *pathwayDetailsHandlerDataAPI) SKUForPathway(pathwayTag string, category common.SKUCategoryType) (*common.SKU, error) {
	return &common.SKU{}, nil
}
func (api *pathwayDetailsHandlerDataAPI) AvailableDoctorIDs(n int) ([]int64, error) {
	return []int64{1, 2, 3}, nil
}
func (api *pathwayDetailsHandlerDataAPI) VisitsSubmittedForPatientSince(patientID common.PatientID, since time.Time) ([]*common.PatientVisit, error) {
	return api.visits, nil
}

func TestPathwayDetailsHandler(t *testing.T) {
	dataAPI := &pathwayDetailsHandlerDataAPI{
		pathways: map[string]*common.Pathway{
			"acne": {
				ID:   1,
				Tag:  "acne",
				Name: "Acne",
				Details: &common.PathwayDetails{
					WhatIsIncluded: []string{"Cheese"},
					WhoWillTreatMe: "George Carlin",
					RightForMe:     "Probably",
					DidYouKnow:     []string{"BEEEEES"},
					FAQ: []common.FAQ{
						{Question: "Why?", Answer: "Because"},
					},
					AgeRestrictions: []*common.PathwayAgeRestriction{
						{
							MaxAgeOfRange: ptr.Int(12),
							VisitAllowed:  false,
							Alert: &common.PathwayAlert{
								Message: "Sorry!",
							},
						},
						{
							MaxAgeOfRange: ptr.Int(17),
							VisitAllowed:  true,
						},
						{
							MaxAgeOfRange: ptr.Int(70),
							VisitAllowed:  true,
						},
						{
							MaxAgeOfRange: nil,
							VisitAllowed:  false,
							Alert: &common.PathwayAlert{
								Message: "Not Sorry!",
							},
						},
					},
				},
			},
			"arachnophobia": {
				ID:   2,
				Tag:  "arachnophobia",
				Name: "Arachnophobia",
			},
			"hypochondria": {
				ID:   3,
				Tag:  "hypochondria",
				Name: "Hypochondria",
			},
			"eczema": {
				ID:   4,
				Tag:  "eczema",
				Name: "Eczema",
			},
		},
		itemCost: &common.ItemCost{
			LineItems: []*common.LineItem{
				{
					Cost: common.Cost{
						Currency: "USD",
						Amount:   4000,
					},
				},
			},
		},
		pathwayCases: []*common.PatientCase{
			{
				ID:         encoding.DeprecatedNewObjectID(111),
				Name:       "Acne",
				PathwayTag: "acne",
				Status:     common.PCStatusActive,
			},
			{
				ID:         encoding.DeprecatedNewObjectID(222),
				Name:       "Arachnophobia",
				PathwayTag: "arachnophobia",
				Status:     common.PCStatusOpen,
			},
			{
				ID:         encoding.DeprecatedNewObjectID(333),
				Name:       "Eczema",
				PathwayTag: "eczema",
				Status:     common.PCStatusActive,
			},
		},
		pathwayDoctors: map[string][]*common.Doctor{
			"acne": {
				{},
			},
		},
		careTeams: map[int64]*common.PatientCareTeam{
			111: {
				Assignments: []*common.CareProviderAssignment{
					{
						ProviderRole:     api.RoleDoctor,
						ShortDisplayName: "Dr. Jones",
					},
				},
			},
		},
	}

	cfgStore, err := cfg.NewLocalStore([]*cfg.ValueDef{config.GlobalFirstVisitFreeDisabled})
	test.OK(t, err)
	h := NewPathwayDetailsHandler(dataAPI, "api.spruce.local", cfgStore)

	// Unauthenticated
	r, err := http.NewRequest("GET", "/?pathway_id=acne,arachnophobia,hypochondria,eczema", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 got %d", w.Code)
	}
	res := &pathwayDetailsRes{}
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if len(res.Pathways) != 4 {
		t.Fatalf("Expected 4 pathways, got %d", len(res.Pathways))
	}
	for _, p := range res.Pathways {
		switch p.PathwayTag {
		default:
			t.Fatalf("Unepxected pathway tag %s", p.PathwayTag)
		case "acne":
			if p.Screen.Type != "merchandising" {
				t.Fatal("Expected acne pathway screen type to be merchandising")
			} else if p.FAQ == nil || len(p.FAQ.Views) == 0 {
				t.Fatalf("Expected acne patchway to have an FAQ: %+v", p.FAQ)
			} else if len(p.Screen.Views) != 4 {
				t.Fatalf("Expected 4 views within screen but got %d", len(p.Screen.Views))
			}
			test.Equals(t, 4, len(p.AgeRestrictions))
			test.Equals(t, 12, *p.AgeRestrictions[0].MaxAgeOfRange)
			test.Equals(t, false, p.AgeRestrictions[0].VisitAllowed)
			test.Equals(t, 17, *p.AgeRestrictions[1].MaxAgeOfRange)
			test.Equals(t, true, p.AgeRestrictions[1].VisitAllowed)
		case "arachnophobia":
			if p.Screen.Type != "generic_message" {
				t.Fatal("Expected arachnophobia pathway screen type to be generic_message")
			}
		case "hypochondria":
			if p.Screen.Type != "generic_message" {
				t.Fatal("Expected hypochondria pathway screen type to be generic_message")
			}
		case "eczema":
			if p.Screen.Type != "generic_message" {
				t.Fatalf("Expected eczema screen type to be generic_message")
			}
		}

	}

	// Unauthenticated with launch promo
	cfgStore, err = cfg.NewLocalStore([]*cfg.ValueDef{config.GlobalFirstVisitFreeEnabled})
	test.OK(t, err)
	h = NewPathwayDetailsHandler(dataAPI, "api.spruce.local", cfgStore)
	r, err = http.NewRequest("GET", "/?pathway_id=acne", nil)
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 got %d", w.Code)
	}
	res = &pathwayDetailsRes{}
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if len(res.Pathways) != 1 {
		t.Fatalf("Expected 1 pathways, got %d", len(res.Pathways))
	}
	for _, p := range res.Pathways {
		switch p.PathwayTag {
		default:
			t.Fatalf("Unepxected pathway tag %s", p.PathwayTag)
		case "acne":
			if p.Screen.Type != "merchandising" {
				t.Fatal("Expected acne pathway screen type to be merchandising")
			} else if p.FAQ == nil || len(p.FAQ.Views) == 0 {
				t.Fatalf("Expected acne patchway to have an FAQ: %+v", p.FAQ)
			} else if len(p.Screen.Views) != 5 {
				t.Fatalf("Expected 5 views within the screen but got %d", len(p.Screen.Views))
			}

			card := p.Screen.Views[0].(map[string]interface{})
			if card["title"] != "Limited time offer" {
				t.Fatalf("Expected limited time offer card but got %s", card["title"])
			}
		}
	}

	// Authenticated
	r, err = http.NewRequest("GET", "/?pathway_id=acne,arachnophobia,hypochondria,eczema", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := apiservice.CtxWithAccount(context.Background(), &common.Account{ID: 1, Role: api.RolePatient})
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r.WithContext(ctx))
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 got %d", w.Code)
	}
	res = &pathwayDetailsRes{}
	if err := json.NewDecoder(w.Body).Decode(res); err != nil {
		t.Fatal(err)
	}
	if len(res.Pathways) != 4 {
		t.Fatalf("Expected 4 pathways, got %d", len(res.Pathways))
	}
	for _, p := range res.Pathways {
		switch p.PathwayTag {
		default:
			t.Fatalf("Unepxected pathway tag %s", p.PathwayTag)
		case "acne":
			if p.Screen.Type != "generic_message" {
				t.Fatal("Expected pathway 1 screen type to be generic_message")
			}
			if !strings.Contains(p.Screen.ContentText, "an existing") {
				t.Fatalf("Expected an existing active case message, got '%s'", p.Screen.ContentText)
			}
		case "arachnophobia":
			if p.Screen.Type != "generic_message" {
				t.Fatal("Expected arachnophobia pathway screen type to be generic_message")
			}
			if !strings.Contains(p.Screen.ContentText, "visit in progress") {
				t.Fatalf("Expected an open visit message, got '%s'", p.Screen.ContentText)
			}
		case "hypochondria":
			if p.Screen.Type != "generic_message" {
				t.Fatal("Expected pathway 2 screen type to be generic_message")
			}
		case "eczema":
			if p.Screen.Type != "generic_message" {
				t.Fatal("Expected pathway screen type to be generic_message")
			}
			if !strings.Contains(p.Screen.ContentText, "pending review") {
				t.Fatalf("Expected a pending review message, got '%s'", p.Screen.ContentText)
			}
		}
	}

	// Authenticated with launch promo (no visits since launch)
	dataAPI.pathways["hypochondria"].Details = dataAPI.pathways["acne"].Details
	r, err = http.NewRequest("GET", "/?pathway_id=hypochondria", nil)
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r.WithContext(ctx))
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 got %d", w.Code)
	}
	res = &pathwayDetailsRes{}
	if err := json.NewDecoder(w.Body).Decode(res); err != nil {
		t.Fatal(err)
	}
	if len(res.Pathways) != 1 {
		t.Fatalf("Expected 1 pathways, got %d", len(res.Pathways))
	}
	for _, p := range res.Pathways {
		switch p.PathwayTag {
		default:
			t.Fatalf("Unepxected pathway tag %s", p.PathwayTag)
		case "hypochondria":
			if p.Screen.Type != "merchandising" {
				t.Fatalf("Expected hypochondria pathway screen type to be merchandising but was %s", p.Screen.Type)
			} else if len(p.Screen.Views) != 5 {
				t.Fatalf("Expected 5 views within the screen but got %d", len(p.Screen.Views))
			}

			card := p.Screen.Views[0].(map[string]interface{})
			if card["title"] != "Limited time offer" {
				t.Fatalf("Expected limited time offer card but got %s", card["title"])
			}
		}
	}

	// Authenticated with launch promo (visits since launch)
	dataAPI.visits = []*common.PatientVisit{{}}
	r, err = http.NewRequest("GET", "/?pathway_id=hypochondria", nil)
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r.WithContext(ctx))
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 got %d", w.Code)
	}
	res = &pathwayDetailsRes{}
	if err := json.NewDecoder(w.Body).Decode(res); err != nil {
		t.Fatal(err)
	}
	if len(res.Pathways) != 1 {
		t.Fatalf("Expected 1 pathways, got %d", len(res.Pathways))
	}
	for _, p := range res.Pathways {
		switch p.PathwayTag {
		default:
			t.Fatalf("Unepxected pathway tag %s", p.PathwayTag)
		case "hypochondria":
			if p.Screen.Type != "merchandising" {
				t.Fatalf("Expected hypochondria pathway screen type to be merchandising but was %s", p.Screen.Type)
			} else if len(p.Screen.Views) != 4 {
				t.Fatalf("Expected 4 views within the screen but got %d", len(p.Screen.Views))
			}
		}
	}
}
