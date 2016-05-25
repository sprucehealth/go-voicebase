package saml

import (
	"strings"
	"testing"

	"github.com/sprucehealth/backend/libs/ptr"
	"github.com/sprucehealth/backend/test"
)

// TestQuestionMultiTriage makes sure that different answers can
// point to different triage steps.
func TestQuestionMultiTriage(t *testing.T) {
	r := strings.NewReader(`
		[triage "systemic"]
			[screen]
				[type “screen_type_warning_popup”]
				[content header title “We're going to have to end your visit here."]
				[body text “Your symptoms and medical history suggest that you may need more immediate medical attention than we can currently provide. A local emergency department is an appropriate option, as is your primary care provider.”]
				[bottom button title “Next Steps”]
			[end screen]
			[screen]
				[type “screen_type_triage”]
				[title “Next Steps”]
				[triage abandon]
				[content header title “You should seek in-person medical evaluation today. A local emergency department is an appropriate option, as is your primary care provider (if you can be seen immediately today).”]
				[body text “If you have health insurance, you should contact your insurance company to find out which providers are covered under your plan. Locate your insurance card and call the listed Member Services number. A representative will help you locate your nearest in-network emergency department. If you are too ill to accomplish this, call 911 and go to the nearest emergency department.\n\nIf you do not have health insurance, go to the most convenient emergency department.”]
				[bottom button title “I Understand”]
			[end screen]
		[end triage]

		[triage "bugs"]
			[screen]
				[type “screen_type_warning_popup”]
				[content header title “We're going to have to end your visit here."]
				[body text “Your symptoms and medical history suggest that you may need more immediate medical attention than we can currently provide. A local emergency department is an appropriate option, as is your primary care provider.”]
				[bottom button title “Next Steps”]
			[end screen]
			[screen]
				[type “screen_type_triage”]
				[title “Next Steps”]
				[triage abandon]
				[content header title “You should seek in-person medical evaluation today. A local urgent care is an appropriate option, as is your primary care provider.”]
				[body text “If you have health insurance, you should contact your insurance company to find out which providers are covered under your plan. Locate your insurance card and call the listed Member Services number. A representative will help you locate your nearest in-network urgent care center or other appropriate provider. If you are too ill to accomplish this, call 911 and go to the nearest emergency department.\n\nIf you do not have health insurance, go to the most convenient urgent care or emergency department. ”]
				[bottom button title “I Understand”]
			[end screen]
		[end triage]

		[patient section "Section"]

		[MD section "Dr Dr Dr Dr"]

		Main) Question?
			[summary "Summary Text"]
			Answer1 → triage:systemic
			Answer2 → triage:bugs
			Answer3
	`)
	intake, err := Parse(r)
	if err != nil {
		t.Fatal(err)
	}
	screens := intake.Sections[0].Subsections[0].Screens
	if len(screens) != 5 {
		t.Fatalf("Expected 5 screens, got %d. Should be 1 for the question and 2 per triage screen.", len(screens))
	}
	s1 := screens[1:3]
	s2 := screens[3:5]
	// The order of the screens is non-deterministic which is fine but makes testing annoying. So order them.
	if s1[0].Condition.String() > s2[0].Condition.String() {
		s1, s2 = s2, s1
	}
	if e := "screen_type_warning_popup"; s1[0].Type != e {
		t.Fatalf("Expected screen type '%s' got '%s'", e, s1[0].Type)
	} else if e := "(summary_text any [summary_text_answer1])"; s1[0].Condition.String() != e {
		t.Fatalf("Expected condition '%s' got '%s'", e, s1[0].Condition.String())
	} else if e := "(summary_text any [summary_text_answer1])"; s1[1].Condition.String() != e {
		t.Fatalf("Expected condition '%s' got '%s'", e, s1[1].Condition.String())
	} else if e := "screen_type_triage"; s1[1].Type != e {
		t.Fatalf("Expected screen type '%s' got '%s'", e, s1[0].Type)
	}
	if e := "screen_type_warning_popup"; s2[0].Type != e {
		t.Fatalf("Expected screen type '%s' got '%s'", e, s2[0].Type)
	} else if e := "(summary_text any [summary_text_answer2])"; s2[0].Condition.String() != e {
		t.Fatalf("Expected condition '%s' got '%s'", e, s2[0].Condition.String())
	} else if e := "(summary_text any [summary_text_answer2])"; s2[1].Condition.String() != e {
		t.Fatalf("Expected condition '%s' got '%s'", e, s2[1].Condition.String())
	} else if e := "screen_type_triage"; s2[1].Type != e {
		t.Fatalf("Expected screen type '%s' got '%s'", e, s2[0].Type)
	}
}

func TestPhotoQuestion_Tips(t *testing.T) {
	r := strings.NewReader(`
		[patient section "Section"]

		[MD section "Dr Dr Dr Dr"]

		Main) Face [photo]
		[photo slot "Face Front"]
			[tip "Center your face in the dotted lines."]
			[photo missing error message "A photo of the front of your face is required to continue."]
			[initial camera direction "front"]
		[photo slot "Side"]
			[tip (inline) "Turn your face to the side."]
			[tip subtext (inline) "Just move your face, not your phone."]
			[tip style (inline) "point_left"]
		[photo slot "Other Side"]
			[tip (inline) "Now turn to the other side."]
			[tip subtext (inline) "Just move your face, not your phone."]
			[tip style (inline) "point_right"]
	`)

	intake, err := Parse(r)
	if err != nil {
		t.Fatal(err)
	}

	screens := intake.Sections[0].Subsections[0].Screens
	if len(screens) != 1 {
		t.Fatalf("Expected 1 screens, got %d. Should be 1 for the question and 2 per triage screen.", len(screens))
	}

	if len(screens[0].Questions) != 1 {
		t.Fatalf("Expected 1 question got %d", len(screens[0].Questions))
	} else if screens[0].Questions[0].Details.Type != "q_type_photo_section" {
		t.Fatalf("Expected q_type_photo_section got %s", screens[0].Questions[0].Details.Type)
	} else if len(screens[0].Questions[0].Details.PhotoSlots) != 3 {
		t.Fatalf("Expected 3 slots got %d", len(screens[0].Questions[0].Details.PhotoSlots))
	}

	// expect inline tip to be specified
	slots := screens[0].Questions[0].Details.MediaSlots
	if slots[0].ClientData.Tip != "Center your face in the dotted lines." {
		t.Fatal("Expected tip to exist but didnt")
	}

	if slots[1].ClientData.Tips["inline"] == nil {
		t.Fatal("Expected inline tip to exist but didnt")
	} else if slots[1].ClientData.Tips["inline"].Tip != "Turn your face to the side." {
		t.Fatal("Expected tip text for inline to exist but didnt")
	} else if slots[1].ClientData.Tips["inline"].TipSubtext != "Just move your face, not your phone." {
		t.Fatal("Expected tip subtext for inline to exist but didnt")
	} else if slots[1].ClientData.Tips["inline"].TipStyle != "point_left" {
		t.Fatal("Expected tip style for inline to exist but didnt")
	}

}

func TestMediaQuestion(t *testing.T) {
	r := strings.NewReader(`
		[patient section "Section"]

		[MD section "Dr Dr Dr Dr"]

		Main) Face [media]
		[video slot "Face Front"]
			[tip "Center your face in the dotted lines."]
			[media missing error message "A photo of the front of your face is required to continue."]
			[initial camera direction "front"]
		[video slot "Side"]
			[tip (inline) "Turn your face to the side."]
			[tip subtext (inline) "Just move your face, not your phone."]
			[tip style (inline) "point_left"]
		[photo slot "Other Side"]
			[tip (inline) "Now turn to the other side."]
			[tip subtext (inline) "Just move your face, not your phone."]
			[tip style (inline) "point_right"]
	`)

	intake, err := Parse(r)
	if err != nil {
		t.Fatal(err)
	}

	test.Equals(t, &Intake{
		Sections: []*Section{
			{
				Title: "Section",
				Subsections: []*Subsection{
					{
						Title: "Dr Dr Dr Dr",
						Screens: []*Screen{
							{
								Questions: []*Question{
									{
										Details: &QuestionDetails{
											Tag:      "face",
											Text:     "Face",
											Required: ptr.Bool(true),
											Type:     QuestionTypeMediaSection,
											MediaSlots: []*MediaSlot{
												{
													Name: "Face Front",
													Type: "video",
													ClientData: &MediaSlotClientData{
														InitialCameraDirection:   "front",
														MediaMissingErrorMessage: "A photo of the front of your face is required to continue.",
														MediaTip: MediaTip{
															Tip: "Center your face in the dotted lines.",
														},
														Tips: make(map[string]*MediaTip),
													},
												},
												{
													Name: "Side",
													Type: "video",
													ClientData: &MediaSlotClientData{
														Tips: map[string]*MediaTip{
															"inline": &MediaTip{
																Tip:        "Turn your face to the side.",
																TipSubtext: "Just move your face, not your phone.",
																TipStyle:   "point_left",
															},
														},
													},
												},
												{
													Name: "Other Side",
													Type: "photo",
													ClientData: &MediaSlotClientData{
														Tips: map[string]*MediaTip{
															"inline": &MediaTip{
																Tip:        "Now turn to the other side.",
																TipSubtext: "Just move your face, not your phone.",
																TipStyle:   "point_right",
															},
														},
													},
												},
											},
											PhotoSlots: []*MediaSlot{
												{
													Name: "Other Side",
													ClientData: &MediaSlotClientData{
														Tips: map[string]*MediaTip{
															"inline": &MediaTip{
																Tip:        "Now turn to the other side.",
																TipSubtext: "Just move your face, not your phone.",
																TipStyle:   "point_right",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, intake)
}
