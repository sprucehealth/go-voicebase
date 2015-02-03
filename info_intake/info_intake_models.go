package info_intake

import "github.com/sprucehealth/backend/common"

const (
	FORMATTED_FIELD_DOCTOR_LAST_NAME = "doctor_last_name"
	FORMATTED_TITLE_FIELD            = "title"
	QUESTION_TYPE_AUTOCOMPLETE       = "q_type_autocomplete"
	QUESTION_TYPE_COMPOUND           = "q_type_compound"
	QUESTION_TYPE_FREE_TEXT          = "q_type_free_text"
	QUESTION_TYPE_MULTIPLE_CHOICE    = "q_type_multiple_choice"
	QUESTION_TYPE_MULTIPLE_PHOTO     = "q_type_multiple_photo"
	QUESTION_TYPE_PHOTO              = "q_type_photo"
	QUESTION_TYPE_SEGMENTED_CONTROL  = "q_type_segmented_control"
	QUESTION_TYPE_SINGLE_ENTRY       = "q_type_single_entry"
	QUESTION_TYPE_SINGLE_PHOTO       = "q_type_single_photo"
	QUESTION_TYPE_SINGLE_SELECT      = "q_type_single_select"
	QUESTION_TYPE_PHOTO_SECTION      = "q_type_photo_section"
)

type Condition struct {
	OperationTag         string       `json:"op,omitempty"`
	IsServerCondition    bool         `json:"server_condition,omitempty"`
	GenderField          string       `json:"gender,omitempty"`
	QuestionTag          string       `json:"question,omitempty"`
	QuestionID           int64        `json:"question_id,string,omitempty"`
	PotentialAnswersId   []string     `json:"potential_answers_id,omitempty"`
	PotentialAnswersTags []string     `json:"potential_answers,omitempty"`
	FieldTag             string       `json:"field,omitempty"`
	ValueTag             string       `json:"value,omitempty"`
	Operands             []*Condition `json:"operands,omitempty"`
}

type TipSection struct {
	TipsSectionTag   string   `json:"tips_section_tag"`
	TipsSectionTitle string   `json:"tips_section_title,omitempty"`
	TipsSubtext      string   `json:"tips_subtext, omitempty"`
	PhotoTipsTags    []string `json:"photo_tips,omitempty"`
	TipsTags         []string `json:"tips"`
	Tips             []string `json:"tips_text"`
}

type PotentialAnswer struct {
	AnswerID      int64                  `json:"potential_answer_id,string,omitempty"`
	Answer        string                 `json:"potential_answer,omitempty"`
	AnswerSummary string                 `json:"potential_answer_summary,omitempty"`
	AnswerType    string                 `json:"answer_type,omitempty"`
	Ordering      int64                  `json:"ordering,string"`
	ToAlert       bool                   `json:"to_alert"`
	AnswerTag     string                 `json:"answer_tag"`
	ClientData    map[string]interface{} `json:"client_data,omitempty"`
}

type Question struct {
	QuestionTag                  string                 `json:"question"`
	QuestionID                   int64                  `json:"question_id,string,omitempty"`
	QuestionTitle                string                 `json:"question_title,omitempty"`
	QuestionTitleHasTokens       bool                   `json:"question_title_has_tokens"`
	QuestionType                 string                 `json:"question_type,omitempty"`
	FormattedFieldTags           []string               `json:"formatted_field_tags,omitempty"`
	QuestionSubText              string                 `json:"question_subtext,omitempty"`
	QuestionSummary              string                 `json:"question_summary,omitempty"`
	AdditionalFields             map[string]interface{} `json:"additional_fields,omitempty"`
	DisplayStyles                []string               `json:"display_styles,omitempty"`
	ParentQuestionId             int64                  `json:"parent_question_id,string,omitempty"`
	PotentialAnswers             []*PotentialAnswer     `json:"potential_answers,omitempty"`
	Answers                      []common.Answer        `json:"answers,omitempty"`
	ConditionBlock               *Condition             `json:"condition,omitempty"`
	Tips                         *TipSection            `json:"tips,omitempty"`
	ToPrefill                    bool                   `json:"to_prefill"`
	PrefilledWithPreviousAnswers bool                   `json:"prefilled_with_previous_answers"`
	Required                     bool                   `json:"required"`
	ToAlert                      bool                   `json:"to_alert"`
	AlertFormattedText           string                 `json:"alert_text"`
	PhotoSlots                   []*PhotoSlot           `json:"photo_slots,omitempty"`
	SubQuestionsConfig           *SubQuestionsConfig    `json:"subquestions_config,omitempty"`
}

type SubQuestionsConfig struct {
	Screens   []*Screen   `json:"screens,omitempty"`
	Questions []*Question `json:"questions,omitempty"`
}

type PhotoSlot struct {
	ID         int64                  `json:"id,string"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Required   bool                   `json:"required"`
	ClientData map[string]interface{} `json:"client_data"`
}

type Screen struct {
	HeaderTitle          string `json:"header_title,omitempty"`
	HeaderTitleHasTokens *bool  `json:"header_title_has_tokens,omitempty"`
	Subtitle             string `json:"header_subtitle,omitempty"`
	// HeaderSummary is used only in the case of photo screens to specifiy a summarized version
	// of the title to be used on the doctor review side.
	HeaderSummary  string      `json:"header_summary,omitempty"`
	Questions      []*Question `json:"questions,omitempty"`
	ScreenType     string      `json:"screen_type,omitempty"`
	ConditionBlock *Condition  `json:"condition,omitempty"`

	// Body represents the screen's body contents used in certain situations to communicate
	// information to the user rather than the screen containing questions to be asked from the user.
	Body *Body `json:"body,omitempty"`

	// BottomButtonTitle is used to indicate the title of the bottom button on the screen
	// for the client. If not present, the assumption is that the client has a default defined
	// for the button title.
	BottomButtonTitle string `json:"bottom_button_title,omitempty"`

	// ContentHeaderTitle is used to indicate the title of the content on the screen. This is a better named
	// version of the HeaderTitle included above
	ContentHeaderTitle string `json:"content_header_title,omitempty"`

	// ScreenHeaderTitle is used to indicate the title of the screen (presumably shown in the nav bar but
	// dependent on the client)
	ScreenHeaderTitle string `json:"screen_title,omitempty"`
}

type Section struct {
	SectionTag        string      `json:"section"`
	SectionId         string      `json:"section_id,omitempty"`
	SectionTitle      string      `json:"section_title,omitempty"`
	Questions         []*Question `json:"questions,omitempty"`
	Screens           []*Screen   `json:"screens,omitempty"`
	SectionTransition *Transition `json:"transition,omitempty"`
}

type Transition struct {
	Title    string    `json:"title"`
	Message  string    `json:"message,omitempty"`
	ImageURL string    `json:"image_url,omitempty"`
	Buttons  []*Button `json:"buttons,omitempty"`
}

type Button struct {
	Text   string `json:"button_text,omitempty"`
	TapUrl string `json:"tap_url,omitempty"`
	Style  string `json:"style,omitempty"`
}

type Body struct {
	Text   string  `json:"text,omitempty"`
	Button *Button `json:"button,omitempy"`
}

type VisitOverviewHeader struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	IconURL  string `json:"icon_url,omitempty"`
}

type VisitMessage struct {
	Title       string `json:"title"`
	Placeholder string `json:"placeholder"`
}

type CheckoutText struct {
	Header string `json:"header_text"`
	Footer string `json:"footer_text"`
}

type SubmissionConfirmationText struct {
	Title  string `json:"title"`
	Top    string `json:"top_text"`
	Bottom string `json:"bottom_text"`
	Button string `json:"button_title"`
}

type TransitionItem struct {
	Message string    `json:"message"`
	Buttons []*Button `json:"buttons"`
}
