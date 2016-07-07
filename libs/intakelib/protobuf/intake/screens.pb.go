// Code generated by protoc-gen-gogo.
// source: screens.proto
// DO NOT EDIT!

package intake

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type ScreenData_Type int32

const (
	ScreenData_QUESTION       ScreenData_Type = 0
	ScreenData_TRIAGE         ScreenData_Type = 1
	ScreenData_IMAGE_POPUP    ScreenData_Type = 2
	ScreenData_MEDIA          ScreenData_Type = 3
	ScreenData_PHARMACY       ScreenData_Type = 4
	ScreenData_GENERIC_POPUP  ScreenData_Type = 5
	ScreenData_VISIT_OVERVIEW ScreenData_Type = 6
	ScreenData_REVIEW         ScreenData_Type = 7
)

var ScreenData_Type_name = map[int32]string{
	0: "QUESTION",
	1: "TRIAGE",
	2: "IMAGE_POPUP",
	3: "MEDIA",
	4: "PHARMACY",
	5: "GENERIC_POPUP",
	6: "VISIT_OVERVIEW",
	7: "REVIEW",
}
var ScreenData_Type_value = map[string]int32{
	"QUESTION":       0,
	"TRIAGE":         1,
	"IMAGE_POPUP":    2,
	"MEDIA":          3,
	"PHARMACY":       4,
	"GENERIC_POPUP":  5,
	"VISIT_OVERVIEW": 6,
	"REVIEW":         7,
}

func (x ScreenData_Type) Enum() *ScreenData_Type {
	p := new(ScreenData_Type)
	*p = x
	return p
}
func (x ScreenData_Type) String() string {
	return proto.EnumName(ScreenData_Type_name, int32(x))
}
func (x *ScreenData_Type) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(ScreenData_Type_value, data, "ScreenData_Type")
	if err != nil {
		return err
	}
	*x = ScreenData_Type(value)
	return nil
}

type VisitOverviewScreen_Section_FilledState int32

const (
	VisitOverviewScreen_Section_FILLED_STATE_UNDEFINED VisitOverviewScreen_Section_FilledState = 0
	VisitOverviewScreen_Section_FILLED                 VisitOverviewScreen_Section_FilledState = 1
	VisitOverviewScreen_Section_UNFILLED               VisitOverviewScreen_Section_FilledState = 2
)

var VisitOverviewScreen_Section_FilledState_name = map[int32]string{
	0: "FILLED_STATE_UNDEFINED",
	1: "FILLED",
	2: "UNFILLED",
}
var VisitOverviewScreen_Section_FilledState_value = map[string]int32{
	"FILLED_STATE_UNDEFINED": 0,
	"FILLED":                 1,
	"UNFILLED":               2,
}

func (x VisitOverviewScreen_Section_FilledState) Enum() *VisitOverviewScreen_Section_FilledState {
	p := new(VisitOverviewScreen_Section_FilledState)
	*p = x
	return p
}
func (x VisitOverviewScreen_Section_FilledState) String() string {
	return proto.EnumName(VisitOverviewScreen_Section_FilledState_name, int32(x))
}
func (x *VisitOverviewScreen_Section_FilledState) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(VisitOverviewScreen_Section_FilledState_value, data, "VisitOverviewScreen_Section_FilledState")
	if err != nil {
		return err
	}
	*x = VisitOverviewScreen_Section_FilledState(value)
	return nil
}

type VisitOverviewScreen_Section_EnabledState int32

const (
	VisitOverviewScreen_Section_ENABLED_STATE_UNDEFINED VisitOverviewScreen_Section_EnabledState = 0
	VisitOverviewScreen_Section_ENABLED                 VisitOverviewScreen_Section_EnabledState = 1
	VisitOverviewScreen_Section_DISABLED                VisitOverviewScreen_Section_EnabledState = 2
)

var VisitOverviewScreen_Section_EnabledState_name = map[int32]string{
	0: "ENABLED_STATE_UNDEFINED",
	1: "ENABLED",
	2: "DISABLED",
}
var VisitOverviewScreen_Section_EnabledState_value = map[string]int32{
	"ENABLED_STATE_UNDEFINED": 0,
	"ENABLED":                 1,
	"DISABLED":                2,
}

func (x VisitOverviewScreen_Section_EnabledState) Enum() *VisitOverviewScreen_Section_EnabledState {
	p := new(VisitOverviewScreen_Section_EnabledState)
	*p = x
	return p
}
func (x VisitOverviewScreen_Section_EnabledState) String() string {
	return proto.EnumName(VisitOverviewScreen_Section_EnabledState_name, int32(x))
}
func (x *VisitOverviewScreen_Section_EnabledState) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(VisitOverviewScreen_Section_EnabledState_value, data, "VisitOverviewScreen_Section_EnabledState")
	if err != nil {
		return err
	}
	*x = VisitOverviewScreen_Section_EnabledState(value)
	return nil
}

// ScreenIDData is a typed representation of the screenID
// so as to determine the type of the screen and its ID.
type ScreenIDData struct {
	Type             *ScreenData_Type `protobuf:"varint,1,req,name=type,enum=intake.ScreenData_Type" json:"type,omitempty"`
	Id               *string          `protobuf:"bytes,2,req,name=id" json:"id,omitempty"`
	XXX_unrecognized []byte           `json:"-"`
}

func (m *ScreenIDData) Reset()         { *m = ScreenIDData{} }
func (m *ScreenIDData) String() string { return proto.CompactTextString(m) }
func (*ScreenIDData) ProtoMessage()    {}

func (m *ScreenIDData) GetType() ScreenData_Type {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return ScreenData_QUESTION
}

func (m *ScreenIDData) GetId() string {
	if m != nil && m.Id != nil {
		return *m.Id
	}
	return ""
}

// ScreenData is a typed representation of the data contained
// in serialized form, where the type is used to lookup the model to use
// to deserialize the information into.
type ScreenData struct {
	Type *ScreenData_Type `protobuf:"varint,1,req,name=type,enum=intake.ScreenData_Type" json:"type,omitempty"`
	Data []byte           `protobuf:"bytes,2,opt,name=data" json:"data,omitempty"`
	// progress represents the progress the user has made
	// on a scale of 0 - 1.0.
	Progress         *float32 `protobuf:"fixed32,3,opt,name=progress" json:"progress,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *ScreenData) Reset()         { *m = ScreenData{} }
func (m *ScreenData) String() string { return proto.CompactTextString(m) }
func (*ScreenData) ProtoMessage()    {}

func (m *ScreenData) GetType() ScreenData_Type {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return ScreenData_QUESTION
}

func (m *ScreenData) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *ScreenData) GetProgress() float32 {
	if m != nil && m.Progress != nil {
		return *m.Progress
	}
	return 0
}

// CommonScreenInfo represents a the common screen info across
// all screens that is communicated to the client.
type CommonScreenInfo struct {
	// id represents the unique identifier for this particular screen.
	Id *string `protobuf:"bytes,1,req,name=id" json:"id,omitempty"`
	// title represents the screen title to display.
	Title *string `protobuf:"bytes,2,opt,name=title" json:"title,omitempty"`
	// is_triage_screen indicates whether the client is expected to triage
	// the user out of the visit on this screen.
	IsTriageScreen *bool `protobuf:"varint,3,opt,name=is_triage_screen" json:"is_triage_screen,omitempty"`
	// triage_pathway_id indicates the pathway the user is to be triaged to.
	// Note that this field is expected to be set alongside the is_triage_screen being true.
	TriagePathwayId *string `protobuf:"bytes,4,opt,name=triage_pathway_id" json:"triage_pathway_id,omitempty"`
	// triage_parameters represents data that is intended to be passed back to the server
	// on the triage visit call. It is data represented in json form.
	TriageParametersJson []byte `protobuf:"bytes,5,opt,name=triage_parameters_json" json:"triage_parameters_json,omitempty"`
	XXX_unrecognized     []byte `json:"-"`
}

func (m *CommonScreenInfo) Reset()         { *m = CommonScreenInfo{} }
func (m *CommonScreenInfo) String() string { return proto.CompactTextString(m) }
func (*CommonScreenInfo) ProtoMessage()    {}

func (m *CommonScreenInfo) GetId() string {
	if m != nil && m.Id != nil {
		return *m.Id
	}
	return ""
}

func (m *CommonScreenInfo) GetTitle() string {
	if m != nil && m.Title != nil {
		return *m.Title
	}
	return ""
}

func (m *CommonScreenInfo) GetIsTriageScreen() bool {
	if m != nil && m.IsTriageScreen != nil {
		return *m.IsTriageScreen
	}
	return false
}

func (m *CommonScreenInfo) GetTriagePathwayId() string {
	if m != nil && m.TriagePathwayId != nil {
		return *m.TriagePathwayId
	}
	return ""
}

func (m *CommonScreenInfo) GetTriageParametersJson() []byte {
	if m != nil {
		return m.TriageParametersJson
	}
	return nil
}

// QuestionScreen represents a screen which contains a list of questions.
type QuestionScreen struct {
	Questions             []*QuestionData   `protobuf:"bytes,1,rep,name=questions" json:"questions,omitempty"`
	ScreenInfo            *CommonScreenInfo `protobuf:"bytes,2,req,name=screen_info" json:"screen_info,omitempty"`
	ContentHeaderTitle    *string           `protobuf:"bytes,3,opt,name=content_header_title" json:"content_header_title,omitempty"`
	ContentHeaderSubtitle *string           `protobuf:"bytes,4,opt,name=content_header_subtitle" json:"content_header_subtitle,omitempty"`
	// info_popup represents the information to show in a popup if present.
	ContentHeaderInfoPopup *InfoPopup `protobuf:"bytes,5,opt,name=content_header_info_popup" json:"content_header_info_popup,omitempty"`
	XXX_unrecognized       []byte     `json:"-"`
}

func (m *QuestionScreen) Reset()         { *m = QuestionScreen{} }
func (m *QuestionScreen) String() string { return proto.CompactTextString(m) }
func (*QuestionScreen) ProtoMessage()    {}

func (m *QuestionScreen) GetQuestions() []*QuestionData {
	if m != nil {
		return m.Questions
	}
	return nil
}

func (m *QuestionScreen) GetScreenInfo() *CommonScreenInfo {
	if m != nil {
		return m.ScreenInfo
	}
	return nil
}

func (m *QuestionScreen) GetContentHeaderTitle() string {
	if m != nil && m.ContentHeaderTitle != nil {
		return *m.ContentHeaderTitle
	}
	return ""
}

func (m *QuestionScreen) GetContentHeaderSubtitle() string {
	if m != nil && m.ContentHeaderSubtitle != nil {
		return *m.ContentHeaderSubtitle
	}
	return ""
}

func (m *QuestionScreen) GetContentHeaderInfoPopup() *InfoPopup {
	if m != nil {
		return m.ContentHeaderInfoPopup
	}
	return nil
}

// PhotoScreen represents a screen containing photo section questions.
type MediaScreen struct {
	ScreenInfo            *CommonScreenInfo       `protobuf:"bytes,1,req,name=screen_info" json:"screen_info,omitempty"`
	MediaQuestions        []*MediaSectionQuestion `protobuf:"bytes,2,rep,name=media_questions" json:"media_questions,omitempty"`
	ContentHeaderTitle    *string                 `protobuf:"bytes,3,opt,name=content_header_title" json:"content_header_title,omitempty"`
	ContentHeaderSubtitle *string                 `protobuf:"bytes,4,opt,name=content_header_subtitle" json:"content_header_subtitle,omitempty"`
	// info_popup represents the information to show in a popup if present.
	ContentHeaderInfoPopup *InfoPopup `protobuf:"bytes,5,opt,name=content_header_info_popup" json:"content_header_info_popup,omitempty"`
	// bottom_container represents the content to be displayed below the photo
	// section questions on the screen.
	BottomContainer  *MediaScreen_ImageTextBox `protobuf:"bytes,6,opt,name=bottom_container" json:"bottom_container,omitempty"`
	XXX_unrecognized []byte                    `json:"-"`
}

func (m *MediaScreen) Reset()         { *m = MediaScreen{} }
func (m *MediaScreen) String() string { return proto.CompactTextString(m) }
func (*MediaScreen) ProtoMessage()    {}

func (m *MediaScreen) GetScreenInfo() *CommonScreenInfo {
	if m != nil {
		return m.ScreenInfo
	}
	return nil
}

func (m *MediaScreen) GetMediaQuestions() []*MediaSectionQuestion {
	if m != nil {
		return m.MediaQuestions
	}
	return nil
}

func (m *MediaScreen) GetContentHeaderTitle() string {
	if m != nil && m.ContentHeaderTitle != nil {
		return *m.ContentHeaderTitle
	}
	return ""
}

func (m *MediaScreen) GetContentHeaderSubtitle() string {
	if m != nil && m.ContentHeaderSubtitle != nil {
		return *m.ContentHeaderSubtitle
	}
	return ""
}

func (m *MediaScreen) GetContentHeaderInfoPopup() *InfoPopup {
	if m != nil {
		return m.ContentHeaderInfoPopup
	}
	return nil
}

func (m *MediaScreen) GetBottomContainer() *MediaScreen_ImageTextBox {
	if m != nil {
		return m.BottomContainer
	}
	return nil
}

type MediaScreen_ImageTextBox struct {
	ImageLink        *string `protobuf:"bytes,1,opt,name=image_link" json:"image_link,omitempty"`
	Text             *string `protobuf:"bytes,2,opt,name=text" json:"text,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *MediaScreen_ImageTextBox) Reset()         { *m = MediaScreen_ImageTextBox{} }
func (m *MediaScreen_ImageTextBox) String() string { return proto.CompactTextString(m) }
func (*MediaScreen_ImageTextBox) ProtoMessage()    {}

func (m *MediaScreen_ImageTextBox) GetImageLink() string {
	if m != nil && m.ImageLink != nil {
		return *m.ImageLink
	}
	return ""
}

func (m *MediaScreen_ImageTextBox) GetText() string {
	if m != nil && m.Text != nil {
		return *m.Text
	}
	return ""
}

// PharmacyScreen represents a screen on which the patient is to enter their pharmacy information.
type PharmacyScreen struct {
	ScreenInfo       *CommonScreenInfo `protobuf:"bytes,1,req,name=screen_info" json:"screen_info,omitempty"`
	XXX_unrecognized []byte            `json:"-"`
}

func (m *PharmacyScreen) Reset()         { *m = PharmacyScreen{} }
func (m *PharmacyScreen) String() string { return proto.CompactTextString(m) }
func (*PharmacyScreen) ProtoMessage()    {}

func (m *PharmacyScreen) GetScreenInfo() *CommonScreenInfo {
	if m != nil {
		return m.ScreenInfo
	}
	return nil
}

// TriageScreen represnets a screen shown to the user
// before triaging them out of the visit.
type TriageScreen struct {
	ScreenInfo         *CommonScreenInfo `protobuf:"bytes,1,req,name=screen_info" json:"screen_info,omitempty"`
	ContentHeaderTitle *string           `protobuf:"bytes,2,opt,name=content_header_title" json:"content_header_title,omitempty"`
	BottomButtonTitle  *string           `protobuf:"bytes,3,opt,name=bottom_button_title" json:"bottom_button_title,omitempty"`
	Body               *Body             `protobuf:"bytes,4,opt,name=body" json:"body,omitempty"`
	XXX_unrecognized   []byte            `json:"-"`
}

func (m *TriageScreen) Reset()         { *m = TriageScreen{} }
func (m *TriageScreen) String() string { return proto.CompactTextString(m) }
func (*TriageScreen) ProtoMessage()    {}

func (m *TriageScreen) GetScreenInfo() *CommonScreenInfo {
	if m != nil {
		return m.ScreenInfo
	}
	return nil
}

func (m *TriageScreen) GetContentHeaderTitle() string {
	if m != nil && m.ContentHeaderTitle != nil {
		return *m.ContentHeaderTitle
	}
	return ""
}

func (m *TriageScreen) GetBottomButtonTitle() string {
	if m != nil && m.BottomButtonTitle != nil {
		return *m.BottomButtonTitle
	}
	return ""
}

func (m *TriageScreen) GetBody() *Body {
	if m != nil {
		return m.Body
	}
	return nil
}

// ImagePopupScreen represents a popup that contains an image along with text.
type ImagePopupScreen struct {
	ScreenInfo         *CommonScreenInfo `protobuf:"bytes,1,req,name=screen_info" json:"screen_info,omitempty"`
	Body               *Body             `protobuf:"bytes,2,opt,name=body" json:"body,omitempty"`
	ImageWidth         *float32          `protobuf:"fixed32,3,opt,name=image_width" json:"image_width,omitempty"`
	ImageHeight        *float32          `protobuf:"fixed32,4,opt,name=image_height" json:"image_height,omitempty"`
	ImageLink          *string           `protobuf:"bytes,5,opt,name=image_link" json:"image_link,omitempty"`
	BottomButtonTitle  *string           `protobuf:"bytes,6,opt,name=bottom_button_title" json:"bottom_button_title,omitempty"`
	ContentHeaderTitle *string           `protobuf:"bytes,7,opt,name=content_header_title" json:"content_header_title,omitempty"`
	XXX_unrecognized   []byte            `json:"-"`
}

func (m *ImagePopupScreen) Reset()         { *m = ImagePopupScreen{} }
func (m *ImagePopupScreen) String() string { return proto.CompactTextString(m) }
func (*ImagePopupScreen) ProtoMessage()    {}

func (m *ImagePopupScreen) GetScreenInfo() *CommonScreenInfo {
	if m != nil {
		return m.ScreenInfo
	}
	return nil
}

func (m *ImagePopupScreen) GetBody() *Body {
	if m != nil {
		return m.Body
	}
	return nil
}

func (m *ImagePopupScreen) GetImageWidth() float32 {
	if m != nil && m.ImageWidth != nil {
		return *m.ImageWidth
	}
	return 0
}

func (m *ImagePopupScreen) GetImageHeight() float32 {
	if m != nil && m.ImageHeight != nil {
		return *m.ImageHeight
	}
	return 0
}

func (m *ImagePopupScreen) GetImageLink() string {
	if m != nil && m.ImageLink != nil {
		return *m.ImageLink
	}
	return ""
}

func (m *ImagePopupScreen) GetBottomButtonTitle() string {
	if m != nil && m.BottomButtonTitle != nil {
		return *m.BottomButtonTitle
	}
	return ""
}

func (m *ImagePopupScreen) GetContentHeaderTitle() string {
	if m != nil && m.ContentHeaderTitle != nil {
		return *m.ContentHeaderTitle
	}
	return ""
}

// GenericPopupScreen represents a screen that is a container
// of generic views implemented in other context of the app.
type GenericPopupScreen struct {
	ScreenInfo *CommonScreenInfo `protobuf:"bytes,1,req,name=screen_info" json:"screen_info,omitempty"`
	// view_data_json represents an object containing
	// a "views" field which is an array of view types supported and understood
	// by the client. The data is represented in json form.
	ViewDataJson     []byte `protobuf:"bytes,2,req,name=view_data_json" json:"view_data_json,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *GenericPopupScreen) Reset()         { *m = GenericPopupScreen{} }
func (m *GenericPopupScreen) String() string { return proto.CompactTextString(m) }
func (*GenericPopupScreen) ProtoMessage()    {}

func (m *GenericPopupScreen) GetScreenInfo() *CommonScreenInfo {
	if m != nil {
		return m.ScreenInfo
	}
	return nil
}

func (m *GenericPopupScreen) GetViewDataJson() []byte {
	if m != nil {
		return m.ViewDataJson
	}
	return nil
}

// VisitOverviewScreen represents the content for the screen shown
// to capture progress between sections.
type VisitOverviewScreen struct {
	// id represents the unique identifier for the screen. In this case it will be the same
	// across various states of the overview screen.
	Id *string `protobuf:"bytes,1,req,name=id" json:"id,omitempty"`
	// header represents the content to show in the header section of the visit overview.
	Header *VisitOverviewScreen_Header `protobuf:"bytes,2,req,name=header" json:"header,omitempty"`
	// text represents the text to show below the header.
	Text *string `protobuf:"bytes,3,req,name=text" json:"text,omitempty"`
	// bottom_button represents the title and action to take when the bottom button on the visit overview
	// is tapped given that the action can be different pertaining to the state of the visit.
	BottomButton *Button `protobuf:"bytes,4,req,name=bottom_button" json:"bottom_button,omitempty"`
	// sections represent the variable number of sections to display on the visit overview.
	Sections         []*VisitOverviewScreen_Section `protobuf:"bytes,5,rep,name=sections" json:"sections,omitempty"`
	XXX_unrecognized []byte                         `json:"-"`
}

func (m *VisitOverviewScreen) Reset()         { *m = VisitOverviewScreen{} }
func (m *VisitOverviewScreen) String() string { return proto.CompactTextString(m) }
func (*VisitOverviewScreen) ProtoMessage()    {}

func (m *VisitOverviewScreen) GetId() string {
	if m != nil && m.Id != nil {
		return *m.Id
	}
	return ""
}

func (m *VisitOverviewScreen) GetHeader() *VisitOverviewScreen_Header {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *VisitOverviewScreen) GetText() string {
	if m != nil && m.Text != nil {
		return *m.Text
	}
	return ""
}

func (m *VisitOverviewScreen) GetBottomButton() *Button {
	if m != nil {
		return m.BottomButton
	}
	return nil
}

func (m *VisitOverviewScreen) GetSections() []*VisitOverviewScreen_Section {
	if m != nil {
		return m.Sections
	}
	return nil
}

// Header represents the content shown in the top section of the
// visit overview.
type VisitOverviewScreen_Header struct {
	Title            *string `protobuf:"bytes,1,req,name=title" json:"title,omitempty"`
	Subtitle         *string `protobuf:"bytes,2,opt,name=subtitle" json:"subtitle,omitempty"`
	ImageLink        *string `protobuf:"bytes,3,opt,name=image_link" json:"image_link,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *VisitOverviewScreen_Header) Reset()         { *m = VisitOverviewScreen_Header{} }
func (m *VisitOverviewScreen_Header) String() string { return proto.CompactTextString(m) }
func (*VisitOverviewScreen_Header) ProtoMessage()    {}

func (m *VisitOverviewScreen_Header) GetTitle() string {
	if m != nil && m.Title != nil {
		return *m.Title
	}
	return ""
}

func (m *VisitOverviewScreen_Header) GetSubtitle() string {
	if m != nil && m.Subtitle != nil {
		return *m.Subtitle
	}
	return ""
}

func (m *VisitOverviewScreen_Header) GetImageLink() string {
	if m != nil && m.ImageLink != nil {
		return *m.ImageLink
	}
	return ""
}

// Section represents the content and metadata about sections
// contained within the layout.
type VisitOverviewScreen_Section struct {
	// tap_link is used to navigate the user to the first screen within the section when the user taps on the section.
	TapLink *string `protobuf:"bytes,1,req,name=tap_link" json:"tap_link,omitempty"`
	// name of the section.
	Name *string `protobuf:"bytes,2,req,name=name" json:"name,omitempty"`
	// current_filled_state represents whether the section is filled or not at the time of showing the visit overview screen.
	CurrentFilledState *VisitOverviewScreen_Section_FilledState `protobuf:"varint,3,req,name=current_filled_state,enum=intake.VisitOverviewScreen_Section_FilledState" json:"current_filled_state,omitempty"`
	// prev_filled_state is used to animate the filled state from prev->current if prev is present.
	PrevFilledState *VisitOverviewScreen_Section_FilledState `protobuf:"varint,4,opt,name=prev_filled_state,enum=intake.VisitOverviewScreen_Section_FilledState,def=0" json:"prev_filled_state,omitempty"`
	// current_enabled_state represents whether the section is enabled at the time of showing the visit overview screen.
	CurrentEnabledState *VisitOverviewScreen_Section_EnabledState `protobuf:"varint,5,req,name=current_enabled_state,enum=intake.VisitOverviewScreen_Section_EnabledState" json:"current_enabled_state,omitempty"`
	// prev_enabled_state is used to animate the enabled state from prev->current if prev is present.
	PrevEnabledState *VisitOverviewScreen_Section_EnabledState `protobuf:"varint,6,opt,name=prev_enabled_state,enum=intake.VisitOverviewScreen_Section_EnabledState,def=0" json:"prev_enabled_state,omitempty"`
	XXX_unrecognized []byte                                    `json:"-"`
}

func (m *VisitOverviewScreen_Section) Reset()         { *m = VisitOverviewScreen_Section{} }
func (m *VisitOverviewScreen_Section) String() string { return proto.CompactTextString(m) }
func (*VisitOverviewScreen_Section) ProtoMessage()    {}

const Default_VisitOverviewScreen_Section_PrevFilledState VisitOverviewScreen_Section_FilledState = VisitOverviewScreen_Section_FILLED_STATE_UNDEFINED
const Default_VisitOverviewScreen_Section_PrevEnabledState VisitOverviewScreen_Section_EnabledState = VisitOverviewScreen_Section_ENABLED_STATE_UNDEFINED

func (m *VisitOverviewScreen_Section) GetTapLink() string {
	if m != nil && m.TapLink != nil {
		return *m.TapLink
	}
	return ""
}

func (m *VisitOverviewScreen_Section) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *VisitOverviewScreen_Section) GetCurrentFilledState() VisitOverviewScreen_Section_FilledState {
	if m != nil && m.CurrentFilledState != nil {
		return *m.CurrentFilledState
	}
	return VisitOverviewScreen_Section_FILLED_STATE_UNDEFINED
}

func (m *VisitOverviewScreen_Section) GetPrevFilledState() VisitOverviewScreen_Section_FilledState {
	if m != nil && m.PrevFilledState != nil {
		return *m.PrevFilledState
	}
	return Default_VisitOverviewScreen_Section_PrevFilledState
}

func (m *VisitOverviewScreen_Section) GetCurrentEnabledState() VisitOverviewScreen_Section_EnabledState {
	if m != nil && m.CurrentEnabledState != nil {
		return *m.CurrentEnabledState
	}
	return VisitOverviewScreen_Section_ENABLED_STATE_UNDEFINED
}

func (m *VisitOverviewScreen_Section) GetPrevEnabledState() VisitOverviewScreen_Section_EnabledState {
	if m != nil && m.PrevEnabledState != nil {
		return *m.PrevEnabledState
	}
	return Default_VisitOverviewScreen_Section_PrevEnabledState
}

func init() {
	proto.RegisterType((*ScreenIDData)(nil), "intake.ScreenIDData")
	proto.RegisterType((*ScreenData)(nil), "intake.ScreenData")
	proto.RegisterType((*CommonScreenInfo)(nil), "intake.CommonScreenInfo")
	proto.RegisterType((*QuestionScreen)(nil), "intake.QuestionScreen")
	proto.RegisterType((*MediaScreen)(nil), "intake.MediaScreen")
	proto.RegisterType((*MediaScreen_ImageTextBox)(nil), "intake.MediaScreen.ImageTextBox")
	proto.RegisterType((*PharmacyScreen)(nil), "intake.PharmacyScreen")
	proto.RegisterType((*TriageScreen)(nil), "intake.TriageScreen")
	proto.RegisterType((*ImagePopupScreen)(nil), "intake.ImagePopupScreen")
	proto.RegisterType((*GenericPopupScreen)(nil), "intake.GenericPopupScreen")
	proto.RegisterType((*VisitOverviewScreen)(nil), "intake.VisitOverviewScreen")
	proto.RegisterType((*VisitOverviewScreen_Header)(nil), "intake.VisitOverviewScreen.Header")
	proto.RegisterType((*VisitOverviewScreen_Section)(nil), "intake.VisitOverviewScreen.Section")
	proto.RegisterEnum("intake.ScreenData_Type", ScreenData_Type_name, ScreenData_Type_value)
	proto.RegisterEnum("intake.VisitOverviewScreen_Section_FilledState", VisitOverviewScreen_Section_FilledState_name, VisitOverviewScreen_Section_FilledState_value)
	proto.RegisterEnum("intake.VisitOverviewScreen_Section_EnabledState", VisitOverviewScreen_Section_EnabledState_name, VisitOverviewScreen_Section_EnabledState_value)
}
