package twilio

import (
	"fmt"
	"html"
	"net/url"
	"testing"
	"time"

	"context"

	"github.com/sprucehealth/backend/cmd/svc/excomms/internal/dal"
	dalmock "github.com/sprucehealth/backend/cmd/svc/excomms/internal/dal/mock"
	"github.com/sprucehealth/backend/cmd/svc/excomms/internal/models"
	"github.com/sprucehealth/backend/cmd/svc/excomms/internal/rawmsg"
	excommsSettings "github.com/sprucehealth/backend/cmd/svc/excomms/settings"
	"github.com/sprucehealth/backend/libs/clock"
	"github.com/sprucehealth/backend/libs/conc"
	"github.com/sprucehealth/backend/libs/phone"
	"github.com/sprucehealth/backend/libs/ptr"
	"github.com/sprucehealth/backend/libs/sig"
	"github.com/sprucehealth/backend/libs/test"
	"github.com/sprucehealth/backend/libs/testhelpers/mock"
	"github.com/sprucehealth/backend/libs/urlutil"
	"github.com/sprucehealth/backend/svc/directory"
	directorymock "github.com/sprucehealth/backend/svc/directory/mock"
	"github.com/sprucehealth/backend/svc/excomms"
	"github.com/sprucehealth/backend/svc/settings"
	settingsmock "github.com/sprucehealth/backend/svc/settings/mock"
)

func TestIncoming_InvalidPhoneNumber(t *testing.T) {
	es := NewEventHandler(nil, nil, nil, &mockSNS_Twilio{}, clock.New(), nil, "https://test.com", "", "", "", nil, nil)
	params := &rawmsg.TwilioParams{
		From: "+97143430391",
		To:   "+17348465522",
	}

	twiml, err := processIncomingCall(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response><Say voice="alice">Sorry, your call cannot be completed as dialed.</Say></Response>`)

	if twiml != expected {
		t.Fatalf("\nExpected: %s\nGot: %s", expected, twiml)
	}

}

func TestIncoming_Organization(t *testing.T) {
	orgID := "12345"
	providerPersonalPhone := "+14152222222"
	patientPhone := "+14151111111"
	practicePhoneNumber := "+14150000000"
	callSID := "12345"

	md := &mockDirectoryService_Twilio{
		entitiesList: []*directory.Entity{
			{
				ID:   orgID,
				Type: directory.EntityType_ORGANIZATION,
				Contacts: []*directory.Contact{
					{
						ContactType: directory.ContactType_PHONE,
						Value:       practicePhoneNumber,
						Provisioned: true,
					},
				},
			},
		},
	}

	mdal := dalmock.New(t)
	defer mdal.Finish()

	mdal.Expect(mock.NewExpectation(mdal.CreateIncomingCall, &models.IncomingCall{
		OrganizationID: orgID,
		Source:         phone.Number(patientPhone),
		Destination:    phone.Number(practicePhoneNumber),
		CallSID:        callSID,
	}))

	msettings := settingsmock.New(t)
	defer msettings.Finish()

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key:    excommsSettings.ConfigKeySendCallsToVoicemail,
				Subkey: practicePhoneNumber,
			},
			{
				Key:    excommsSettings.ConfigKeyAfterHoursVociemailEnabled,
				Subkey: practicePhoneNumber,
			},
			{
				Key:    excommsSettings.ConfigKeyForwardingListTimeout,
				Subkey: practicePhoneNumber,
			},
			{
				Key:    excommsSettings.ConfigKeyForwardingList,
				Subkey: practicePhoneNumber,
			},
			{
				Key:    excommsSettings.ConfigKeyPauseBeforeCallConnect,
				Subkey: practicePhoneNumber,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Type: settings.ConfigType_BOOLEAN,
				Value: &settings.Value_Boolean{
					Boolean: &settings.BooleanValue{
						Value: false,
					},
				},
			},
			{
				Type: settings.ConfigType_BOOLEAN,
				Value: &settings.Value_Boolean{
					Boolean: &settings.BooleanValue{
						Value: false,
					},
				},
			},
			{
				Type: settings.ConfigType_INTEGER,
				Value: &settings.Value_Integer{
					Integer: &settings.IntegerValue{
						Value: 50,
					},
				},
			},
			{
				Key: &settings.ConfigKey{
					Key:    excommsSettings.ConfigKeyForwardingList,
					Subkey: practicePhoneNumber,
				},
				Type: settings.ConfigType_STRING_LIST,
				Value: &settings.Value_StringList{
					StringList: &settings.StringListValue{
						Values: []string{providerPersonalPhone},
					},
				},
			},
			{
				Type: settings.ConfigType_INTEGER,
				Value: &settings.Value_Integer{
					Integer: &settings.IntegerValue{
						Value: 0,
					},
				},
			},
		},
	}, nil))

	sig, err := sig.NewSigner([][]byte{[]byte("key")}, nil)
	test.OK(t, err)
	signer := urlutil.NewSigner("apiDomain", sig, clock.New())

	es := NewEventHandler(md, msettings, mdal, &mockSNS_Twilio{}, clock.New(), nil, "https://test.com", "", "", "", nil, signer)
	params := &rawmsg.TwilioParams{
		From:    patientPhone,
		To:      practicePhoneNumber,
		CallSID: callSID,
	}

	twiml, err := processIncomingCall(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response><Dial action="/twilio/call/process_incoming_call_status" timeout="50" callerId="%s"><Number url="/twilio/call/provider_call_connected">%s</Number></Dial></Response>`, practicePhoneNumber, providerPersonalPhone)

	if twiml != expected {
		t.Fatalf("\nExpected: %s\nGot: %s", expected, twiml)
	}
}

func TestIncoming_Organization_MultipleContacts(t *testing.T) {
	orgID := "12345"
	listedNumber1 := "+14152222222"
	listedNumber2 := "+14153333333"
	listedNumber3 := "+14154444444"
	patientPhone := "+14151111111"
	practicePhoneNumber := "+14150000000"
	callSID := "12345"

	md := &mockDirectoryService_Twilio{
		entitiesList: []*directory.Entity{
			{
				ID:   orgID,
				Type: directory.EntityType_ORGANIZATION,
				Contacts: []*directory.Contact{
					{
						ContactType: directory.ContactType_PHONE,
						Provisioned: true,
						Value:       practicePhoneNumber,
					},
				},
			},
		},
	}

	mdal := dalmock.New(t)
	defer mdal.Finish()

	mdal.Expect(mock.NewExpectation(mdal.CreateIncomingCall, &models.IncomingCall{
		OrganizationID: orgID,
		Source:         phone.Number(patientPhone),
		Destination:    phone.Number(practicePhoneNumber),
		CallSID:        callSID,
	}))

	msettings := settingsmock.New(t)
	defer msettings.Finish()

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key:    excommsSettings.ConfigKeySendCallsToVoicemail,
				Subkey: practicePhoneNumber,
			},
			{
				Key:    excommsSettings.ConfigKeyAfterHoursVociemailEnabled,
				Subkey: practicePhoneNumber,
			},
			{
				Key:    excommsSettings.ConfigKeyForwardingListTimeout,
				Subkey: practicePhoneNumber,
			},
			{
				Key:    excommsSettings.ConfigKeyForwardingList,
				Subkey: practicePhoneNumber,
			},
			{
				Key:    excommsSettings.ConfigKeyPauseBeforeCallConnect,
				Subkey: practicePhoneNumber,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Type: settings.ConfigType_BOOLEAN,
				Value: &settings.Value_Boolean{
					Boolean: &settings.BooleanValue{
						Value: false,
					},
				},
			},
			{
				Type: settings.ConfigType_BOOLEAN,
				Value: &settings.Value_Boolean{
					Boolean: &settings.BooleanValue{
						Value: false,
					},
				},
			},
			{
				Type: settings.ConfigType_INTEGER,
				Value: &settings.Value_Integer{
					Integer: &settings.IntegerValue{
						Value: 30,
					},
				},
			},
			{
				Key: &settings.ConfigKey{
					Key:    excommsSettings.ConfigKeyForwardingList,
					Subkey: practicePhoneNumber,
				},
				Type: settings.ConfigType_STRING_LIST,
				Value: &settings.Value_StringList{
					StringList: &settings.StringListValue{
						Values: []string{listedNumber1, listedNumber2, listedNumber3},
					},
				},
			},
			{
				Type: settings.ConfigType_INTEGER,
				Value: &settings.Value_Integer{
					Integer: &settings.IntegerValue{
						Value: 0,
					},
				},
			},
		},
	}, nil))

	sig, err := sig.NewSigner([][]byte{[]byte("key")}, nil)
	test.OK(t, err)
	signer := urlutil.NewSigner("apiDomain", sig, clock.New())

	es := NewEventHandler(md, msettings, mdal, &mockSNS_Twilio{}, clock.New(), nil, "https://test.com", "", "", "", nil, signer)
	params := &rawmsg.TwilioParams{
		From:    patientPhone,
		To:      practicePhoneNumber,
		CallSID: callSID,
	}

	twiml, err := processIncomingCall(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response><Dial action="/twilio/call/process_incoming_call_status" timeout="30" callerId="+14150000000"><Number url="/twilio/call/provider_call_connected">+14152222222</Number><Number url="/twilio/call/provider_call_connected">+14153333333</Number><Number url="/twilio/call/provider_call_connected">+14154444444</Number></Dial></Response>`)

	if twiml != expected {
		t.Fatalf("\nExpected: %s\nGot: %s", expected, twiml)
	}
}

func TestIncoming_Organization_MultipleContacts_SendCallsToVoicemail(t *testing.T) {
	orgID := "12345"
	patientPhone := "+14151111111"
	practicePhoneNumber := "+14150000000"
	callSID := "12345"

	md := &mockDirectoryService_Twilio{
		entitiesList: []*directory.Entity{
			{
				ID:   orgID,
				Type: directory.EntityType_ORGANIZATION,
				Info: &directory.EntityInfo{
					DisplayName: "Dewabi Corp",
				},
				Contacts: []*directory.Contact{
					{
						ContactType: directory.ContactType_PHONE,
						Provisioned: true,
						Value:       practicePhoneNumber,
					},
				},
			},
		},
	}

	mdal := dalmock.New(t)
	defer mdal.Finish()

	mdal.Expect(mock.NewExpectation(mdal.CreateIncomingCall, &models.IncomingCall{
		OrganizationID: orgID,
		Source:         phone.Number(patientPhone),
		Destination:    phone.Number(practicePhoneNumber),
		CallSID:        callSID,
	}))

	msettings := settingsmock.New(t)
	defer msettings.Finish()

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key:    excommsSettings.ConfigKeySendCallsToVoicemail,
				Subkey: practicePhoneNumber,
			},
			{
				Key:    excommsSettings.ConfigKeyAfterHoursVociemailEnabled,
				Subkey: practicePhoneNumber,
			},
			{
				Key:    excommsSettings.ConfigKeyForwardingListTimeout,
				Subkey: practicePhoneNumber,
			},
			{
				Key:    excommsSettings.ConfigKeyForwardingList,
				Subkey: practicePhoneNumber,
			},
			{
				Key:    excommsSettings.ConfigKeyPauseBeforeCallConnect,
				Subkey: practicePhoneNumber,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Type: settings.ConfigType_BOOLEAN,
				Value: &settings.Value_Boolean{
					Boolean: &settings.BooleanValue{
						Value: true,
					},
				},
			},
			{
				Type: settings.ConfigType_BOOLEAN,
				Value: &settings.Value_Boolean{
					Boolean: &settings.BooleanValue{
						Value: false,
					},
				},
			},
			{
				Type: settings.ConfigType_INTEGER,
				Value: &settings.Value_Integer{
					Integer: &settings.IntegerValue{
						Value: 30,
					},
				},
			},
			{
				Key: &settings.ConfigKey{
					Key:    excommsSettings.ConfigKeyForwardingList,
					Subkey: practicePhoneNumber,
				},
				Type: settings.ConfigType_STRING_LIST,
				Value: &settings.Value_StringList{
					StringList: &settings.StringListValue{
						Values: []string{},
					},
				},
			},
			{
				Type: settings.ConfigType_INTEGER,
				Value: &settings.Value_Integer{
					Integer: &settings.IntegerValue{
						Value: 0,
					},
				},
			},
		},
	}, nil))

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key:    excommsSettings.ConfigKeyVoicemailOption,
				Subkey: practicePhoneNumber,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Key: &settings.ConfigKey{
					Key: excommsSettings.ConfigKeyVoicemailOption,
				},
				Type: settings.ConfigType_SINGLE_SELECT,
				Value: &settings.Value_SingleSelect{
					SingleSelect: &settings.SingleSelectValue{
						Item: &settings.ItemValue{
							ID: excommsSettings.VoicemailOptionDefault,
						},
					},
				},
			},
		},
	}, nil))

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key: excommsSettings.ConfigKeyTranscribeVoicemail,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Key: &settings.ConfigKey{
					Key: excommsSettings.ConfigKeyTranscribeVoicemail,
				},
				Type: settings.ConfigType_BOOLEAN,
				Value: &settings.Value_Boolean{
					Boolean: &settings.BooleanValue{
						Value: true,
					},
				},
			},
		},
	}, nil))

	mdal.Expect(mock.NewExpectation(mdal.UpdateIncomingCall, callSID, &dal.IncomingCallUpdate{
		SentToVoicemail: ptr.Bool(true),
	}).WithReturns(int64(1), nil))

	sig, err := sig.NewSigner([][]byte{[]byte("key")}, nil)
	test.OK(t, err)
	signer := urlutil.NewSigner("apiDomain", sig, clock.New())

	es := NewEventHandler(md, msettings, mdal, &mockSNS_Twilio{}, clock.New(), nil, "https://test.com", "", "", "", nil, signer)
	params := &rawmsg.TwilioParams{
		From:    patientPhone,
		To:      practicePhoneNumber,
		CallSID: callSID,
	}

	twiml, err := processIncomingCall(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<Response><Say voice="alice">You have reached Dewabi Corp. Please leave a message after the tone. Speak slowly and clearly as your message will be transcribed.</Say><Record action="/twilio/call/no_op" timeout="60" maxLength="3600" transcribeCallback="/twilio/call/process_voicemail" playBeep="true"></Record></Response>`

	if expected != twiml {
		t.Fatalf("\nExpected: %s\nGot: %s", expected, twiml)
	}
}

func TestProviderCallConnected(t *testing.T) {
	patientPhone := "+12061111111"
	practicePhone := "+12062222222"
	providerPhone := "+12063333333"
	orgID := "o1"

	// the params are intended to simulate the dial leg of the call
	// where the call shows up as originating from the practice phone to
	// the number of the provider in the forwarding list
	params := &rawmsg.TwilioParams{
		From:          practicePhone,
		To:            providerPhone,
		ParentCallSID: "12345",
	}

	mdal := dalmock.New(t)
	defer mdal.Finish()

	mdal.Expect(mock.NewExpectation(mdal.LookupIncomingCall, params.ParentCallSID).WithReturns(&models.IncomingCall{
		OrganizationID: orgID,
		Source:         phone.Number(patientPhone),
		Destination:    phone.Number(practicePhone),
		CallSID:        params.ParentCallSID,
	}, nil))

	mdirectory := directorymock.New(t)
	defer mdirectory.Finish()

	mdirectory.Expect(mock.NewExpectation(mdirectory.LookupEntitiesByContact, &directory.LookupEntitiesByContactRequest{
		ContactValue: patientPhone,
		RequestedInformation: &directory.RequestedInformation{
			Depth: 0,
			EntityInformation: []directory.EntityInformation{
				directory.EntityInformation_MEMBERSHIPS,
			},
		},
		Statuses:  []directory.EntityStatus{directory.EntityStatus_ACTIVE},
		RootTypes: []directory.EntityType{directory.EntityType_EXTERNAL},
	}).WithReturns(&directory.LookupEntitiesByContactResponse{
		Entities: []*directory.Entity{
			{
				Type: directory.EntityType_EXTERNAL,
				Info: &directory.EntityInfo{
					FirstName:   "J",
					LastName:    "S",
					DisplayName: "JS",
				},
				Memberships: []*directory.Entity{
					{
						ID:   orgID,
						Type: directory.EntityType_ORGANIZATION,
					},
				},
			},
		},
	}, nil))

	sig, err := sig.NewSigner([][]byte{[]byte("key")}, nil)
	test.OK(t, err)
	signer := urlutil.NewSigner("apiDomain", sig, clock.New())

	es := NewEventHandler(mdirectory, nil, mdal, &mockSNS_Twilio{}, clock.New(), nil, "https://test.com", "", "", "", nil, signer)

	twiml, err := providerCallConnected(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<Response><Gather action="/twilio/call/provider_entered_digits" method="POST" timeout="10" numDigits="1"><Say voice="alice">You have an incoming call from JS</Say><Say voice="alice">Press 1 to answer.</Say></Gather><Hangup></Hangup></Response>`

	if twiml != expected {
		t.Fatalf("\nExpected: %s\nGot: %s", expected, twiml)
	}
}

func TestProviderCallConnected_NoName(t *testing.T) {
	patientPhone := "+12061111111"
	practicePhone := "+12062222222"
	providerPhone := "+12063333333"
	orgID := "o1"

	// the params are intended to simulate the dial leg of the call
	// where the call shows up as originating from the practice phone to
	// the number of the provider in the forwarding list
	params := &rawmsg.TwilioParams{
		From:          practicePhone,
		To:            providerPhone,
		ParentCallSID: "12345",
	}

	mdal := dalmock.New(t)
	defer mdal.Finish()

	mdal.Expect(mock.NewExpectation(mdal.LookupIncomingCall, params.ParentCallSID).WithReturns(&models.IncomingCall{
		OrganizationID: orgID,
		Source:         phone.Number(patientPhone),
		Destination:    phone.Number(practicePhone),
		CallSID:        params.ParentCallSID,
	}, nil))

	mdirectory := directorymock.New(t)
	defer mdirectory.Finish()

	mdirectory.Expect(mock.NewExpectation(mdirectory.LookupEntitiesByContact, &directory.LookupEntitiesByContactRequest{
		ContactValue: patientPhone,
		RequestedInformation: &directory.RequestedInformation{
			Depth: 0,
			EntityInformation: []directory.EntityInformation{
				directory.EntityInformation_MEMBERSHIPS,
			},
		},
		Statuses:  []directory.EntityStatus{directory.EntityStatus_ACTIVE},
		RootTypes: []directory.EntityType{directory.EntityType_EXTERNAL},
	}).WithReturns(&directory.LookupEntitiesByContactResponse{
		Entities: []*directory.Entity{
			{
				Type: directory.EntityType_EXTERNAL,
				Info: &directory.EntityInfo{},
				Memberships: []*directory.Entity{
					{
						ID:   orgID,
						Type: directory.EntityType_ORGANIZATION,
					},
				},
			},
		},
	}, nil))

	sig, err := sig.NewSigner([][]byte{[]byte("key")}, nil)
	test.OK(t, err)
	signer := urlutil.NewSigner("apiDomain", sig, clock.New())

	es := NewEventHandler(mdirectory, nil, mdal, &mockSNS_Twilio{}, clock.New(), nil, "https://test.com", "", "", "", nil, signer)

	twiml, err := providerCallConnected(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<Response><Gather action="/twilio/call/provider_entered_digits" method="POST" timeout="10" numDigits="1"><Say voice="alice">You have an incoming call from 206 111 1111</Say><Say voice="alice">Press 1 to answer.</Say></Gather><Hangup></Hangup></Response>`

	if twiml != expected {
		t.Fatalf("\nExpected: %s\nGot: %s", expected, twiml)
	}
}

func TestProviderEnteredDigits_Entered1(t *testing.T) {
	params := &rawmsg.TwilioParams{
		From:          "+14151111111",
		To:            "+14152222222",
		Digits:        "1",
		CallSID:       "callSID",
		ParentCallSID: "parentCallSID",
	}
	mdal := dalmock.New(t)
	defer mdal.Finish()

	mdal.Expect(mock.NewExpectation(mdal.UpdateIncomingCall, params.ParentCallSID, &dal.IncomingCallUpdate{
		Answered: ptr.Bool(true),
	}).WithReturns(int64(1), nil))

	es := NewEventHandler(nil, nil, mdal, &mockSNS_Twilio{}, clock.New(), nil, "https://test.com", "", "", "", nil, nil)
	twiml, err := providerEnteredDigits(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatal(err)
	}

	expected := `<?xml version="1.0" encoding="UTF-8"?>
<Response></Response>`
	if expected != twiml {
		t.Fatalf(`\nExpected: %s\nGot:%s`, expected, twiml)
	}
}

func TestProviderEnteredDigits_EnteredOtherDigit(t *testing.T) {

	patientPhone := "+12061111111"
	practicePhone := "+12062222222"
	orgID := "o1"

	// the params are intended to simulate the dial leg of the call
	// where the call shows up as originating from the practice phone to
	// the number of the provider in the forwarding list
	params := &rawmsg.TwilioParams{
		From:          "+14151111111",
		To:            "+14152222222",
		Digits:        "2",
		ParentCallSID: "12345",
	}

	mdal := dalmock.New(t)
	defer mdal.Finish()

	mdal.Expect(mock.NewExpectation(mdal.LookupIncomingCall, params.ParentCallSID).WithReturns(&models.IncomingCall{
		OrganizationID: orgID,
		Source:         phone.Number(patientPhone),
		Destination:    phone.Number(practicePhone),
		CallSID:        params.ParentCallSID,
	}, nil))

	mdirectory := directorymock.New(t)
	defer mdirectory.Finish()

	mdirectory.Expect(mock.NewExpectation(mdirectory.LookupEntitiesByContact, &directory.LookupEntitiesByContactRequest{
		ContactValue: patientPhone,
		RequestedInformation: &directory.RequestedInformation{
			Depth: 0,
			EntityInformation: []directory.EntityInformation{
				directory.EntityInformation_MEMBERSHIPS,
			},
		},
		Statuses:  []directory.EntityStatus{directory.EntityStatus_ACTIVE},
		RootTypes: []directory.EntityType{directory.EntityType_EXTERNAL},
	}).WithReturns(&directory.LookupEntitiesByContactResponse{
		Entities: []*directory.Entity{
			{
				Type: directory.EntityType_EXTERNAL,
				Info: &directory.EntityInfo{
					FirstName:   "J",
					LastName:    "S",
					DisplayName: "JS",
				},
				Memberships: []*directory.Entity{
					{
						ID:   orgID,
						Type: directory.EntityType_ORGANIZATION,
					},
				},
			},
		},
	}, nil))

	sig, err := sig.NewSigner([][]byte{[]byte("key")}, nil)
	test.OK(t, err)
	signer := urlutil.NewSigner("apiDomain", sig, clock.New())

	es := NewEventHandler(mdirectory, nil, mdal, &mockSNS_Twilio{}, clock.New(), nil, "https://test.com", "", "", "", nil, signer)
	twiml, err := providerEnteredDigits(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatal(err)
	}

	expected := `<?xml version="1.0" encoding="UTF-8"?>
<Response><Gather action="/twilio/call/provider_entered_digits" method="POST" timeout="10" numDigits="1"><Say voice="alice">You have an incoming call from JS</Say><Say voice="alice">Press 1 to answer.</Say></Gather><Hangup></Hangup></Response>`

	if expected != twiml {
		t.Fatalf(`\nExpected: %s\nGot:%s`, expected, twiml)
	}
}

func TestVoicemailTwiML(t *testing.T) {
	orgID := "12345"
	providerID := "p1"
	practicePhoneNumber := "+14152222222"
	md := &mockDirectoryService_Twilio{
		entitiesList: []*directory.Entity{
			{
				ID:   orgID,
				Type: directory.EntityType_ORGANIZATION,
				Contacts: []*directory.Contact{
					{
						ContactType: directory.ContactType_PHONE,
						Value:       practicePhoneNumber,
						Provisioned: true,
					},
				},
				Info: &directory.EntityInfo{
					DisplayName: "Dewabi Corp",
				},
				Members: []*directory.Entity{
					{
						ID: providerID,
						Contacts: []*directory.Contact{
							{
								ContactType: directory.ContactType_PHONE,
								Value:       "+14151111111",
							},
						},
					},
				},
			},
		},
	}

	params := &rawmsg.TwilioParams{
		From:    "+14151111111",
		To:      "+14152222222",
		Digits:  "2",
		CallSID: "callSID",
	}

	msettings := settingsmock.New(t)
	defer msettings.Finish()

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key:    excommsSettings.ConfigKeyVoicemailOption,
				Subkey: practicePhoneNumber,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Key: &settings.ConfigKey{
					Key: excommsSettings.ConfigKeyVoicemailOption,
				},
				Type: settings.ConfigType_SINGLE_SELECT,
				Value: &settings.Value_SingleSelect{
					SingleSelect: &settings.SingleSelectValue{
						Item: &settings.ItemValue{
							ID: excommsSettings.VoicemailOptionDefault,
						},
					},
				},
			},
		},
	}, nil))

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key: excommsSettings.ConfigKeyTranscribeVoicemail,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Key: &settings.ConfigKey{
					Key: excommsSettings.ConfigKeyTranscribeVoicemail,
				},
				Type: settings.ConfigType_BOOLEAN,
				Value: &settings.Value_Boolean{
					Boolean: &settings.BooleanValue{
						Value: true,
					},
				},
			},
		},
	}, nil))

	mdal := dalmock.New(t)
	defer mdal.Finish()

	mdal.Expect(mock.NewExpectation(mdal.UpdateIncomingCall, params.CallSID, &dal.IncomingCallUpdate{
		SentToVoicemail: ptr.Bool(true),
	}).WithReturns(int64(1), nil))

	sig, err := sig.NewSigner([][]byte{[]byte("key")}, nil)
	test.OK(t, err)
	signer := urlutil.NewSigner("apiDomain", sig, clock.New())

	es := NewEventHandler(md, msettings, mdal, &mockSNS_Twilio{}, clock.New(), nil, "https://test.com", "", "", "", nil, signer)

	twiml, err := voicemailTWIML(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatal(err)
	}
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<Response><Say voice="alice">You have reached Dewabi Corp. Please leave a message after the tone. Speak slowly and clearly as your message will be transcribed.</Say><Record action="/twilio/call/no_op" timeout="60" maxLength="3600" transcribeCallback="/twilio/call/process_voicemail" playBeep="true"></Record></Response>`

	if expected != twiml {
		t.Fatalf("\nExpected: %s\nGot: %s", expected, twiml)
	}
}

func TestVoicemailTwiML_Custom(t *testing.T) {
	orgID := "12345"
	providerID := "p1"
	practicePhoneNumber := "+14152222222"
	md := &mockDirectoryService_Twilio{
		entitiesList: []*directory.Entity{
			{
				ID:   orgID,
				Type: directory.EntityType_ORGANIZATION,
				Contacts: []*directory.Contact{
					{
						ContactType: directory.ContactType_PHONE,
						Value:       practicePhoneNumber,
						Provisioned: true,
					},
				},
				Info: &directory.EntityInfo{
					DisplayName: "Dewabi Corp",
				},
				Members: []*directory.Entity{
					{
						ID: providerID,
						Contacts: []*directory.Contact{
							{
								ContactType: directory.ContactType_PHONE,
								Value:       "+14151111111",
							},
						},
					},
				},
			},
		},
	}

	params := &rawmsg.TwilioParams{
		From:    "+14151111111",
		To:      "+14152222222",
		Digits:  "2",
		CallSID: "callSID",
	}

	msettings := settingsmock.New(t)
	defer msettings.Finish()

	customVoicemailMediaID := "123456789"

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key:    excommsSettings.ConfigKeyVoicemailOption,
				Subkey: practicePhoneNumber,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Key: &settings.ConfigKey{
					Key: excommsSettings.ConfigKeyVoicemailOption,
				},
				Type: settings.ConfigType_SINGLE_SELECT,
				Value: &settings.Value_SingleSelect{
					SingleSelect: &settings.SingleSelectValue{
						Item: &settings.ItemValue{
							ID:               excommsSettings.VoicemailOptionCustom,
							FreeTextResponse: customVoicemailMediaID,
						},
					},
				},
			},
		},
	}, nil))

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key: excommsSettings.ConfigKeyTranscribeVoicemail,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Key: &settings.ConfigKey{
					Key: excommsSettings.ConfigKeyTranscribeVoicemail,
				},
				Type: settings.ConfigType_BOOLEAN,
				Value: &settings.Value_Boolean{
					Boolean: &settings.BooleanValue{
						Value: true,
					},
				},
			},
		},
	}, nil))

	mdal := dalmock.New(t)
	defer mdal.Finish()
	mdal.Expect(mock.NewExpectation(mdal.UpdateIncomingCall, params.CallSID, &dal.IncomingCallUpdate{
		SentToVoicemail: ptr.Bool(true),
	}).WithReturns(int64(1), nil))

	mc := clock.NewManaged(time.Now())
	sig, err := sig.NewSigner([][]byte{[]byte("key")}, nil)
	test.OK(t, err)
	signer := urlutil.NewSigner("apiDomain", sig, clock.New())

	expectedURL, err := signer.SignedURL(fmt.Sprintf("/media/%s", customVoicemailMediaID), url.Values{}, ptr.Time(mc.Now().Add(time.Hour)))
	test.OK(t, err)

	es := NewEventHandler(md, msettings, mdal, &mockSNS_Twilio{}, mc, nil, "https://test.com", "", "", "", nil, signer)

	twiml, err := voicemailTWIML(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatal(err)
	}
	expected := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response><Play>%s</Play><Record action="/twilio/call/no_op" timeout="60" maxLength="3600" transcribeCallback="/twilio/call/process_voicemail" playBeep="true"></Record></Response>`, html.EscapeString(expectedURL))

	if expected != twiml {
		t.Fatalf("\nExpected: %s\nGot: %s", expected, twiml)
	}
}

func TestVoicemailTwiML_NoTranscription(t *testing.T) {
	orgID := "12345"
	providerID := "p1"
	practicePhoneNumber := "+14152222222"
	md := &mockDirectoryService_Twilio{
		entitiesList: []*directory.Entity{
			{
				ID:   orgID,
				Type: directory.EntityType_ORGANIZATION,
				Contacts: []*directory.Contact{
					{
						ContactType: directory.ContactType_PHONE,
						Value:       practicePhoneNumber,
						Provisioned: true,
					},
				},
				Info: &directory.EntityInfo{
					DisplayName: "Dewabi Corp",
				},
				Members: []*directory.Entity{
					{
						ID: providerID,
						Contacts: []*directory.Contact{
							{
								ContactType: directory.ContactType_PHONE,
								Value:       "+14151111111",
							},
						},
					},
				},
			},
		},
	}

	params := &rawmsg.TwilioParams{
		From:    "+14151111111",
		To:      "+14152222222",
		Digits:  "2",
		CallSID: "callSID",
	}

	msettings := settingsmock.New(t)
	defer msettings.Finish()

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key:    excommsSettings.ConfigKeyVoicemailOption,
				Subkey: practicePhoneNumber,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Key: &settings.ConfigKey{
					Key: excommsSettings.ConfigKeyVoicemailOption,
				},
				Type: settings.ConfigType_SINGLE_SELECT,
				Value: &settings.Value_SingleSelect{
					SingleSelect: &settings.SingleSelectValue{
						Item: &settings.ItemValue{
							ID: excommsSettings.VoicemailOptionDefault,
						},
					},
				},
			},
		},
	}, nil))

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key: excommsSettings.ConfigKeyTranscribeVoicemail,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Key: &settings.ConfigKey{
					Key: excommsSettings.ConfigKeyTranscribeVoicemail,
				},
				Type: settings.ConfigType_BOOLEAN,
				Value: &settings.Value_Boolean{
					Boolean: &settings.BooleanValue{
						Value: false,
					},
				},
			},
		},
	}, nil))

	mdal := dalmock.New(t)
	defer mdal.Finish()

	mdal.Expect(mock.NewExpectation(mdal.UpdateIncomingCall, params.CallSID, &dal.IncomingCallUpdate{
		SentToVoicemail: ptr.Bool(true),
	}).WithReturns(int64(1), nil))

	sig, err := sig.NewSigner([][]byte{[]byte("key")}, nil)
	test.OK(t, err)
	signer := urlutil.NewSigner("apiDomain", sig, clock.New())

	es := NewEventHandler(md, msettings, mdal, &mockSNS_Twilio{}, clock.New(), nil, "https://test.com", "", "", "", nil, signer)

	twiml, err := voicemailTWIML(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatal(err)
	}
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<Response><Say voice="alice">You have reached Dewabi Corp. Please leave a message after the tone.</Say><Record action="/twilio/call/process_voicemail" timeout="60" maxLength="3600" transcribeCallback="/twilio/call/no_op" playBeep="true"></Record></Response>`

	if expected != twiml {
		t.Fatalf("\nExpected: %s\nGot: %s", expected, twiml)
	}
}

func TestIncomingCallStatus_CallCompleted(t *testing.T) {
	conc.Testing = true
	ms := &mockSNS_Twilio{}
	params := &rawmsg.TwilioParams{
		From:       "+12068773590",
		To:         "+17348465522",
		CallStatus: rawmsg.TwilioParams_COMPLETED,
		CallSID:    "12345",
	}

	md := dalmock.New(t)
	defer md.Finish()

	md.Expect(mock.NewExpectation(md.LookupIncomingCall, params.CallSID).WithReturns(&models.IncomingCall{
		OrganizationID: "o1",
		Source:         phone.Number(params.From),
		Destination:    phone.Number(params.To),
		Answered:       true,
	}, nil))

	md.Expect(mock.NewExpectation(md.LookupIncomingCall, params.CallSID).WithReturns(&models.IncomingCall{
		OrganizationID: "o1",
		Source:         phone.Number(params.From),
		Destination:    phone.Number(params.To),
		Answered:       true,
	}, nil))

	mdir := directorymock.New(t)
	defer mdir.Finish()

	mdir.Expect(mock.NewExpectation(mdir.LookupEntities, &directory.LookupEntitiesRequest{
		LookupKeyType: directory.LookupEntitiesRequest_ENTITY_ID,
		LookupKeyOneof: &directory.LookupEntitiesRequest_EntityID{
			EntityID: "o1",
		},
		RequestedInformation: &directory.RequestedInformation{
			Depth: 1,
			EntityInformation: []directory.EntityInformation{
				directory.EntityInformation_EXTERNAL_IDS,
				directory.EntityInformation_MEMBERS,
			},
		},
		Statuses: []directory.EntityStatus{directory.EntityStatus_ACTIVE},
	}).WithReturns(&directory.LookupEntitiesResponse{
		Entities: []*directory.Entity{
			{
				ID:   "o1",
				Type: directory.EntityType_ORGANIZATION,
				Members: []*directory.Entity{
					{
						ID:   "p1",
						Type: directory.EntityType_INTERNAL,
					},
				},
				ExternalIDs: []string{"account_1"},
			},
		},
	}, nil))

	mclock := clock.NewManaged(time.Now())
	sig, err := sig.NewSigner([][]byte{[]byte("key")}, nil)
	test.OK(t, err)
	signer := urlutil.NewSigner("apiDomain", sig, mclock)

	es := NewEventHandler(mdir, nil, md, ms, mclock, nil, "", "", "", "", nil, signer)

	twiml, err := processIncomingCallStatus(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatal(err)
	} else if twiml != "" {
		t.Fatalf("Expected %s but got %s", "", twiml)
	}

	// ensure that item was published
	if len(ms.published) != 1 {
		t.Fatalf("Expected %d got %d", 1, len(ms.published))
	}

	pem, err := parsePublishedExternalMessage(*ms.published[0].Message)
	if err != nil {
		t.Fatal(err)
	}

	test.Equals(t, &excomms.PublishedExternalMessage{
		FromChannelID: params.From,
		ToChannelID:   params.To,
		Timestamp:     uint64(mclock.Now().Unix()),
		Direction:     excomms.PublishedExternalMessage_INBOUND,
		Type:          excomms.PublishedExternalMessage_INCOMING_CALL_EVENT,
		Item: &excomms.PublishedExternalMessage_Incoming{
			Incoming: &excomms.IncomingCallEventItem{
				Type:              excomms.IncomingCallEventItem_ANSWERED,
				DurationInSeconds: params.CallDuration,
			},
		},
	}, pem)
}

func TestIncomingCallStatus_MissedCall(t *testing.T) {
	conc.Testing = true
	ms := &mockSNS_Twilio{}
	params := &rawmsg.TwilioParams{
		From:       "+12068773590",
		To:         "+17348465522",
		CallStatus: rawmsg.TwilioParams_COMPLETED,
		CallSID:    "12345",
	}

	md := dalmock.New(t)
	defer md.Finish()

	md.Expect(mock.NewExpectation(md.LookupIncomingCall, params.CallSID).WithReturns(&models.IncomingCall{
		OrganizationID: "o1",
		Source:         phone.Number(params.From),
		Destination:    phone.Number(params.To),
	}, nil))

	md.Expect(mock.NewExpectation(md.LookupIncomingCall, params.CallSID).WithReturns(&models.IncomingCall{
		OrganizationID: "o1",
		Source:         phone.Number(params.From),
		Destination:    phone.Number(params.To),
	}, nil))

	mdir := directorymock.New(t)
	defer mdir.Finish()

	mdir.Expect(mock.NewExpectation(mdir.LookupEntities, &directory.LookupEntitiesRequest{
		LookupKeyType: directory.LookupEntitiesRequest_ENTITY_ID,
		LookupKeyOneof: &directory.LookupEntitiesRequest_EntityID{
			EntityID: "o1",
		},
		RequestedInformation: &directory.RequestedInformation{
			Depth: 1,
			EntityInformation: []directory.EntityInformation{
				directory.EntityInformation_EXTERNAL_IDS,
				directory.EntityInformation_MEMBERS,
			},
		},
		Statuses: []directory.EntityStatus{directory.EntityStatus_ACTIVE},
	}).WithReturns(&directory.LookupEntitiesResponse{
		Entities: []*directory.Entity{
			{
				ID:   "o1",
				Type: directory.EntityType_ORGANIZATION,
				Members: []*directory.Entity{
					{
						ID:   "p1",
						Type: directory.EntityType_INTERNAL,
					},
				},
				ExternalIDs: []string{"account_1"},
			},
		},
	}, nil))

	mclock := clock.NewManaged(time.Now())
	sig, err := sig.NewSigner([][]byte{[]byte("key")}, nil)
	test.OK(t, err)
	signer := urlutil.NewSigner("apiDomain", sig, mclock)

	es := NewEventHandler(mdir, nil, md, ms, mclock, nil, "", "", "", "", nil, signer)

	twiml, err := processIncomingCallStatus(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatal(err)
	} else if twiml != "" {
		t.Fatalf("Expected %s but got %s", "", twiml)
	}

	// ensure that item was published
	if len(ms.published) != 1 {
		t.Fatalf("Expected %d got %d", 1, len(ms.published))
	}

	pem, err := parsePublishedExternalMessage(*ms.published[0].Message)
	if err != nil {
		t.Fatal(err)
	}

	test.Equals(t, &excomms.PublishedExternalMessage{
		FromChannelID: params.From,
		ToChannelID:   params.To,
		Timestamp:     uint64(mclock.Now().Unix()),
		Direction:     excomms.PublishedExternalMessage_INBOUND,
		Type:          excomms.PublishedExternalMessage_INCOMING_CALL_EVENT,
		Item: &excomms.PublishedExternalMessage_Incoming{
			Incoming: &excomms.IncomingCallEventItem{
				Type:              excomms.IncomingCallEventItem_UNANSWERED,
				DurationInSeconds: params.CallDuration,
			},
		},
	}, pem)
}

func TestIncomingCallStatus_SentToVoicemail(t *testing.T) {
	conc.Testing = true
	ms := &mockSNS_Twilio{}
	params := &rawmsg.TwilioParams{
		From:       "+12068773590",
		To:         "+17348465522",
		CallStatus: rawmsg.TwilioParams_COMPLETED,
		CallSID:    "12345",
	}

	md := dalmock.New(t)
	defer md.Finish()

	md.Expect(mock.NewExpectation(md.LookupIncomingCall, params.CallSID).WithReturns(&models.IncomingCall{
		OrganizationID:  "o1",
		Source:          phone.Number(params.From),
		Destination:     phone.Number(params.To),
		SentToVoicemail: true,
	}, nil))

	mclock := clock.NewManaged(time.Now())
	es := NewEventHandler(nil, nil, md, ms, mclock, nil, "", "", "", "", nil, nil)

	twiml, err := processIncomingCallStatus(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatal(err)
	} else if twiml != "" {
		t.Fatalf("Expected %s but got %s", "", twiml)
	}

	// ensure that item was published
	if len(ms.published) != 0 {
		t.Fatalf("Expected %d got %d", 0, len(ms.published))
	}
}

func TestIncomingCallStatus_OtherCallStatus(t *testing.T) {
	testIncomingCallStatus_Other(t, rawmsg.TwilioParams_FAILED)
	testIncomingCallStatus_Other(t, rawmsg.TwilioParams_NO_ANSWER)
	testIncomingCallStatus_Other(t, rawmsg.TwilioParams_IN_PROGRESS)
	testIncomingCallStatus_Other(t, rawmsg.TwilioParams_QUEUED)
	testIncomingCallStatus_Other(t, rawmsg.TwilioParams_INITIATED)
	testIncomingCallStatus_Other(t, rawmsg.TwilioParams_BUSY)
	testIncomingCallStatus_Other(t, rawmsg.TwilioParams_CANCELED)
	testIncomingCallStatus_Other(t, rawmsg.TwilioParams_RINGING)
}

func testIncomingCallStatus_Other(t *testing.T, incomingStatus rawmsg.TwilioParams_CallStatus) {
	conc.Testing = true
	ms := &mockSNS_Twilio{}
	params := &rawmsg.TwilioParams{
		From:           "+12068773590",
		To:             "+17348465522",
		DialCallStatus: incomingStatus,
		CallSID:        "callSID12345",
		ParentCallSID:  "parentCallSID12345",
	}

	orgID := "12345"
	providerID := "p1"
	providerPersonalPhone := "+14152222222"
	practicePhoneNumber := "+17348465522"

	md := directorymock.New(t)
	defer md.Finish()

	md.Expect(mock.NewExpectation(md.LookupEntities, &directory.LookupEntitiesRequest{
		LookupKeyType: directory.LookupEntitiesRequest_ENTITY_ID,
		LookupKeyOneof: &directory.LookupEntitiesRequest_EntityID{
			EntityID: orgID,
		},
		RequestedInformation: &directory.RequestedInformation{
			Depth: 0,
		},
		Statuses: []directory.EntityStatus{directory.EntityStatus_ACTIVE},
	}).WithReturns(&directory.LookupEntitiesResponse{
		Entities: []*directory.Entity{
			{
				ID:   orgID,
				Type: directory.EntityType_ORGANIZATION,
				Contacts: []*directory.Contact{
					{
						ContactType: directory.ContactType_PHONE,
						Value:       practicePhoneNumber,
						Provisioned: true,
					},
				},
				Info: &directory.EntityInfo{
					DisplayName: "Dewabi Corp",
				},
				Members: []*directory.Entity{
					{
						ID: providerID,
						Contacts: []*directory.Contact{
							{
								ContactType: directory.ContactType_PHONE,
								Value:       providerPersonalPhone,
							},
						},
					},
				},
			},
		},
	}, nil))

	md.Expect(mock.NewExpectation(md.LookupEntities, &directory.LookupEntitiesByContactRequest{
		ContactValue: practicePhoneNumber,

		RequestedInformation: &directory.RequestedInformation{
			Depth: 0,
		},
		Statuses:  []directory.EntityStatus{directory.EntityStatus_ACTIVE},
		RootTypes: []directory.EntityType{directory.EntityType_ORGANIZATION},
	}).WithReturns(&directory.LookupEntitiesByContactResponse{
		Entities: []*directory.Entity{
			{
				ID:   orgID,
				Type: directory.EntityType_ORGANIZATION,
				Contacts: []*directory.Contact{
					{
						ContactType: directory.ContactType_PHONE,
						Value:       practicePhoneNumber,
						Provisioned: true,
					},
				},
				Info: &directory.EntityInfo{
					DisplayName: "Dewabi Corp",
				},
				Members: []*directory.Entity{
					{
						ID: providerID,
						Contacts: []*directory.Contact{
							{
								ContactType: directory.ContactType_PHONE,
								Value:       providerPersonalPhone,
							},
						},
					},
				},
			},
		},
	}, nil))

	mdal := dalmock.New(t)
	defer mdal.Finish()

	mdal.Expect(mock.NewExpectation(mdal.LookupIncomingCall, params.CallSID).WithReturns(&models.IncomingCall{
		OrganizationID: orgID,
	}, nil))

	mdal.Expect(mock.NewExpectation(mdal.UpdateIncomingCall, params.ParentCallSID, &dal.IncomingCallUpdate{
		SentToVoicemail: ptr.Bool(true),
	}).WithReturns(int64(1), nil))

	msettings := settingsmock.New(t)
	defer msettings.Finish()

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key:    excommsSettings.ConfigKeyAfterHoursVociemailEnabled,
				Subkey: practicePhoneNumber,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Type: settings.ConfigType_BOOLEAN,
				Value: &settings.Value_Boolean{
					Boolean: &settings.BooleanValue{
						Value: false,
					},
				},
			},
		},
	}, nil))

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key:    excommsSettings.ConfigKeyVoicemailOption,
				Subkey: practicePhoneNumber,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Key: &settings.ConfigKey{
					Key: excommsSettings.ConfigKeyVoicemailOption,
				},
				Type: settings.ConfigType_SINGLE_SELECT,
				Value: &settings.Value_SingleSelect{
					SingleSelect: &settings.SingleSelectValue{
						Item: &settings.ItemValue{
							ID: excommsSettings.VoicemailOptionDefault,
						},
					},
				},
			},
		},
	}, nil))

	msettings.Expect(mock.NewExpectation(msettings.GetValues, &settings.GetValuesRequest{
		Keys: []*settings.ConfigKey{
			{
				Key: excommsSettings.ConfigKeyTranscribeVoicemail,
			},
		},
		NodeID: orgID,
	}).WithReturns(&settings.GetValuesResponse{
		Values: []*settings.Value{
			{
				Key: &settings.ConfigKey{
					Key: excommsSettings.ConfigKeyTranscribeVoicemail,
				},
				Type: settings.ConfigType_BOOLEAN,
				Value: &settings.Value_Boolean{
					Boolean: &settings.BooleanValue{
						Value: true,
					},
				},
			},
		},
	}, nil))

	sig, err := sig.NewSigner([][]byte{[]byte("key")}, nil)
	test.OK(t, err)
	signer := urlutil.NewSigner("apiDomain", sig, clock.New())

	es := NewEventHandler(md, msettings, mdal, ms, clock.New(), nil, "https://test.com", "", "", "", nil, signer)

	twiml, err := processDialedCallStatus(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatal(err)
	}

	expected := `<?xml version="1.0" encoding="UTF-8"?>
<Response><Say voice="alice">You have reached Dewabi Corp. Please leave a message after the tone. Speak slowly and clearly as your message will be transcribed.</Say><Record action="/twilio/call/no_op" timeout="60" maxLength="3600" transcribeCallback="/twilio/call/process_voicemail" playBeep="true"></Record></Response>`
	if expected != twiml {
		t.Fatalf("\nExpected: %s\nGot: %s", expected, twiml)
	}

	// ensure that item was published
	if len(ms.published) != 0 {
		t.Fatalf("Expected %d got %d", 0, len(ms.published))
	}
}

func TestProcessVoicemail(t *testing.T) {
	conc.Testing = true
	params := &rawmsg.TwilioParams{
		From:              "+12068773590",
		To:                "+17348465522",
		RecordingDuration: 10,
		RecordingURL:      "http://google.com",
	}

	ms := &mockSNS_Twilio{}
	md := dalmock.New(t)
	defer md.Finish()

	md.Expect(mock.NewExpectation(md.StoreIncomingRawMessage, &rawmsg.Incoming{
		Type: rawmsg.Incoming_TWILIO_VOICEMAIL,
		Message: &rawmsg.Incoming_Twilio{
			Twilio: params,
		},
	}))

	md.Expect(mock.NewExpectation(md.LookupIncomingCall, params.CallSID).WithReturns(&models.IncomingCall{
		OrganizationID: "o1",
		Source:         phone.Number(params.From),
		Destination:    phone.Number(params.To),
	}, nil))

	mdir := directorymock.New(t)
	defer mdir.Finish()

	mdir.Expect(mock.NewExpectation(mdir.LookupEntities, &directory.LookupEntitiesRequest{
		LookupKeyType: directory.LookupEntitiesRequest_ENTITY_ID,
		LookupKeyOneof: &directory.LookupEntitiesRequest_EntityID{
			EntityID: "o1",
		},
		RequestedInformation: &directory.RequestedInformation{
			Depth: 1,
			EntityInformation: []directory.EntityInformation{
				directory.EntityInformation_EXTERNAL_IDS,
				directory.EntityInformation_MEMBERS,
			},
		},
		Statuses: []directory.EntityStatus{directory.EntityStatus_ACTIVE},
	}).WithReturns(&directory.LookupEntitiesResponse{
		Entities: []*directory.Entity{
			{
				ID:   "o1",
				Type: directory.EntityType_ORGANIZATION,
				Members: []*directory.Entity{
					{
						ID:   "p1",
						Type: directory.EntityType_INTERNAL,
					},
				},
				ExternalIDs: []string{"account_1"},
			},
		},
	}, nil))

	sig, err := sig.NewSigner([][]byte{[]byte("key")}, nil)
	test.OK(t, err)
	signer := urlutil.NewSigner("apiDomain", sig, clock.New())

	es := NewEventHandler(mdir, nil, md, ms, clock.New(), nil, "", "", "", "", nil, signer)

	twiml, err := processVoicemail(context.Background(), params, es.(*eventsHandler))
	if err != nil {
		t.Fatal(err)
	} else if twiml != "" {
		t.Fatalf("Expected %s got %s", "", twiml)
	}

	if len(ms.published) != 1 {
		t.Fatalf("Expected 1 but got %d", len(ms.published))
	}

	md.Finish()
}
