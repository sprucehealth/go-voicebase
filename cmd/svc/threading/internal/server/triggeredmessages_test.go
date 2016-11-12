package server

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"

	"github.com/sprucehealth/backend/cmd/svc/threading/internal/dal"
	"github.com/sprucehealth/backend/cmd/svc/threading/internal/models"

	"github.com/sprucehealth/backend/libs/test"
	"github.com/sprucehealth/backend/libs/testhelpers/mock"
	"github.com/sprucehealth/backend/svc/threading"
)

func TestCreateTriggeredMessage(t *testing.T) {
	ctx := context.Background()
	tmID, err := models.NewTriggeredMessageID()
	test.OK(t, err)
	tmiID1, err := models.NewTriggeredMessageItemID()
	test.OK(t, err)
	tmiID2, err := models.NewTriggeredMessageItemID()
	test.OK(t, err)
	message1 := &threading.MessagePost{Summary: "Summary"}
	message2 := &threading.MessagePost{Summary: "Summary"}
	tmModel := &models.TriggeredMessage{
		ID:                   tmID,
		ActorEntityID:        "ActorEntityID",
		OrganizationEntityID: "OrganizationEntityID",
		TriggerKey:           "NEW_PATIENT",
		TriggerSubkey:        "subkey",
	}
	tmiModel1 := &models.TriggeredMessageItem{
		ID:                 tmiID1,
		TriggeredMessageID: tmID,
		Ordinal:            0,
		ActorEntityID:      "ActorEntityID",
		Data:               &models.Message{},
	}
	tmiModel2 := &models.TriggeredMessageItem{
		ID:                 tmiID2,
		TriggeredMessageID: tmID,
		Ordinal:            0,
		ActorEntityID:      "ActorEntityID",
		Data:               &models.Message{},
	}
	t.Run("Error-KeyRequired", func(t *testing.T) {
		st := newServerTest(t)
		testCreateTriggeredMessage(t, st, &threading.CreateTriggeredMessageRequest{}, nil, grpcErrorf(codes.InvalidArgument, "Key is required"))
	})
	t.Run("Error-UnknownKey", func(t *testing.T) {
		st := newServerTest(t)
		testCreateTriggeredMessage(t, st, &threading.CreateTriggeredMessageRequest{
			Key: &threading.TriggeredMessageKey{},
		}, nil, grpcErrorf(codes.InvalidArgument, "Invalid triggered message key %s", threading.TRIGGERED_MESSAGE_KEY_INVALID))
	})
	t.Run("Error-ActorEntityIDRequired", func(t *testing.T) {
		st := newServerTest(t)
		testCreateTriggeredMessage(t, st, &threading.CreateTriggeredMessageRequest{
			Key: &threading.TriggeredMessageKey{
				Key: threading.TRIGGERED_MESSAGE_KEY_NEW_PATIENT,
			},
		}, nil, grpcErrorf(codes.InvalidArgument, "ActorEntityID is required"))
	})
	t.Run("Error-OrganizationEntityIDRequired", func(t *testing.T) {
		st := newServerTest(t)
		testCreateTriggeredMessage(t, st, &threading.CreateTriggeredMessageRequest{
			Key: &threading.TriggeredMessageKey{
				Key: threading.TRIGGERED_MESSAGE_KEY_NEW_PATIENT,
			},
			ActorEntityID: "ActorEntityID",
		}, nil, grpcErrorf(codes.InvalidArgument, "OrganizationEntityID is required"))
	})
	t.Run("Error-NoMessages", func(t *testing.T) {
		st := newServerTest(t)
		testCreateTriggeredMessage(t, st, &threading.CreateTriggeredMessageRequest{
			Key: &threading.TriggeredMessageKey{
				Key: threading.TRIGGERED_MESSAGE_KEY_NEW_PATIENT,
			},
			ActorEntityID:        "ActorEntityID",
			OrganizationEntityID: "OrganizationEntityID",
		}, nil, grpcErrorf(codes.InvalidArgument, "At least 1 Message is required"))
	})
	t.Run("Success-NoExistingTriggeredMessage", func(t *testing.T) {
		st := newServerTest(t)
		st.dal.Expect(
			mock.NewExpectation(st.dal.TriggeredMessageForKeys, "NEW_PATIENT", "subkey", []interface{}{}).WithReturns(
				(*models.TriggeredMessage)(nil), dal.ErrNotFound))
		st.dal.Expect(
			mock.NewExpectation(st.dal.CreateTriggeredMessage, &models.TriggeredMessage{
				ActorEntityID:        "ActorEntityID",
				OrganizationEntityID: "OrganizationEntityID",
				TriggerKey:           "NEW_PATIENT",
				TriggerSubkey:        "subkey",
			}).WithReturns(tmID, nil))

		req1, err := createPostMessageRequest(ctx, models.EmptyThreadID(), "ActorEntityID", message1)
		test.OK(t, err)
		threadItem1, err := dal.ThreadItemFromPostMessageRequest(ctx, req1, st.clk)
		test.OK(t, err)
		threadItem1.ID = models.EmptyThreadItemID()
		data1 := threadItem1.Data.(*models.Message)
		req2, err := createPostMessageRequest(context.Background(), models.EmptyThreadID(), "ActorEntityID", message2)
		test.OK(t, err)
		threadItem2, err := dal.ThreadItemFromPostMessageRequest(ctx, req2, st.clk)
		test.OK(t, err)
		threadItem2.ID = models.EmptyThreadItemID()
		data2 := threadItem2.Data.(*models.Message)

		st.dal.Expect(
			mock.NewExpectation(st.dal.CreateTriggeredMessageItem, &models.TriggeredMessageItem{
				TriggeredMessageID: tmID,
				Ordinal:            0,
				Internal:           false,
				Data:               data1,
			}))
		st.dal.Expect(
			mock.NewExpectation(st.dal.CreateTriggeredMessageItem, &models.TriggeredMessageItem{
				TriggeredMessageID: tmID,
				Ordinal:            1,
				Internal:           false,
				Data:               data2,
			}))

		st.dal.Expect(mock.NewExpectation(st.dal.TriggeredMessage, tmID, []interface{}{}).WithReturns(tmModel, nil))
		st.dal.Expect(mock.NewExpectation(st.dal.TriggeredMessageItemsForTriggeredMessageID, tmID, []interface{}{}).WithReturns(
			[]*models.TriggeredMessageItem{tmiModel1, tmiModel2}, nil))

		rtm, err := transformTriggeredMessageToResponse(tmModel, []*models.TriggeredMessageItem{tmiModel1, tmiModel2})
		test.OK(t, err)

		testCreateTriggeredMessage(t, st, &threading.CreateTriggeredMessageRequest{
			Key: &threading.TriggeredMessageKey{
				Key:    threading.TRIGGERED_MESSAGE_KEY_NEW_PATIENT,
				Subkey: "subkey",
			},
			ActorEntityID:        "ActorEntityID",
			OrganizationEntityID: "OrganizationEntityID",
			Messages:             []*threading.MessagePost{message1, message2},
		}, &threading.CreateTriggeredMessageResponse{
			TriggeredMessage: rtm,
		}, nil)
	})
	t.Run("Success-ExistingTriggeredMessage", func(t *testing.T) {
		st := newServerTest(t)
		st.dal.Expect(
			mock.NewExpectation(st.dal.TriggeredMessageForKeys, "NEW_PATIENT", "subkey", []interface{}{}).WithReturns(
				tmModel, nil))
		st.dal.Expect(mock.NewExpectation(st.dal.DeleteTriggeredMessageItemsForTriggeredMessage, tmID))
		st.dal.Expect(mock.NewExpectation(st.dal.DeleteTriggeredMessage, tmID))
		st.dal.Expect(
			mock.NewExpectation(st.dal.CreateTriggeredMessage, &models.TriggeredMessage{
				ActorEntityID:        "ActorEntityID",
				OrganizationEntityID: "OrganizationEntityID",
				TriggerKey:           "NEW_PATIENT",
				TriggerSubkey:        "subkey",
			}).WithReturns(tmID, nil))

		req1, err := createPostMessageRequest(ctx, models.EmptyThreadID(), "ActorEntityID", message1)
		test.OK(t, err)
		threadItem1, err := dal.ThreadItemFromPostMessageRequest(ctx, req1, st.clk)
		test.OK(t, err)
		threadItem1.ID = models.EmptyThreadItemID()
		data1 := threadItem1.Data.(*models.Message)
		req2, err := createPostMessageRequest(context.Background(), models.EmptyThreadID(), "ActorEntityID", message2)
		test.OK(t, err)
		threadItem2, err := dal.ThreadItemFromPostMessageRequest(ctx, req2, st.clk)
		test.OK(t, err)
		threadItem2.ID = models.EmptyThreadItemID()
		data2 := threadItem2.Data.(*models.Message)

		st.dal.Expect(
			mock.NewExpectation(st.dal.CreateTriggeredMessageItem, &models.TriggeredMessageItem{
				TriggeredMessageID: tmID,
				Ordinal:            0,
				Internal:           false,
				Data:               data1,
			}))
		st.dal.Expect(
			mock.NewExpectation(st.dal.CreateTriggeredMessageItem, &models.TriggeredMessageItem{
				TriggeredMessageID: tmID,
				Ordinal:            1,
				Internal:           false,
				Data:               data2,
			}))

		st.dal.Expect(mock.NewExpectation(st.dal.TriggeredMessage, tmID, []interface{}{}).WithReturns(tmModel, nil))
		st.dal.Expect(mock.NewExpectation(st.dal.TriggeredMessageItemsForTriggeredMessageID, tmID, []interface{}{}).WithReturns(
			[]*models.TriggeredMessageItem{tmiModel1, tmiModel2}, nil))

		rtm, err := transformTriggeredMessageToResponse(tmModel, []*models.TriggeredMessageItem{tmiModel1, tmiModel2})
		test.OK(t, err)

		testCreateTriggeredMessage(t, st, &threading.CreateTriggeredMessageRequest{
			Key: &threading.TriggeredMessageKey{
				Key:    threading.TRIGGERED_MESSAGE_KEY_NEW_PATIENT,
				Subkey: "subkey",
			},
			ActorEntityID:        "ActorEntityID",
			OrganizationEntityID: "OrganizationEntityID",
			Messages:             []*threading.MessagePost{message1, message2},
		}, &threading.CreateTriggeredMessageResponse{
			TriggeredMessage: rtm,
		}, nil)
	})
}

func testCreateTriggeredMessage(
	t *testing.T,
	st *serverTest,
	in *threading.CreateTriggeredMessageRequest,
	exp *threading.CreateTriggeredMessageResponse,
	expErr error) {
	resp, err := st.server.CreateTriggeredMessage(st.ctx, in)
	test.Equals(t, expErr, err)
	test.Equals(t, exp, resp)
	st.Finish()
}

func TestTriggeredMessages(t *testing.T) {
	tmID, err := models.NewTriggeredMessageID()
	test.OK(t, err)
	tmiID1, err := models.NewTriggeredMessageItemID()
	test.OK(t, err)
	tmiID2, err := models.NewTriggeredMessageItemID()
	test.OK(t, err)
	tmModel := &models.TriggeredMessage{
		ID:                   tmID,
		ActorEntityID:        "ActorEntityID",
		OrganizationEntityID: "OrganizationEntityID",
		TriggerKey:           "NEW_PATIENT",
		TriggerSubkey:        "subkey",
	}
	tmiModel1 := &models.TriggeredMessageItem{
		ID:                 tmiID1,
		TriggeredMessageID: tmID,
		Ordinal:            0,
		ActorEntityID:      "ActorEntityID",
		Data:               &models.Message{},
	}
	tmiModel2 := &models.TriggeredMessageItem{
		ID:                 tmiID2,
		TriggeredMessageID: tmID,
		Ordinal:            0,
		ActorEntityID:      "ActorEntityID",
		Data:               &models.Message{},
	}
	t.Run("Error-KeyLookup-NotFound", func(t *testing.T) {
		st := newServerTest(t)
		st.dal.Expect(
			mock.NewExpectation(st.dal.TriggeredMessageForKeys, "NEW_PATIENT", "subkey", []interface{}{}).WithReturns(
				(*models.TriggeredMessage)(nil), dal.ErrNotFound))
		testTriggeredMessages(t, st, &threading.TriggeredMessagesRequest{
			LookupKey: &threading.TriggeredMessagesRequest_Key{
				Key: &threading.TriggeredMessageKey{
					Key:    threading.TRIGGERED_MESSAGE_KEY_NEW_PATIENT,
					Subkey: "subkey",
				},
			},
		}, nil, grpcErrorf(codes.NotFound, "TriggeredMessage not found for key(s) %s %s", "NEW_PATIENT", "subkey"))
	})
	t.Run("Success-KeyLookup", func(t *testing.T) {
		st := newServerTest(t)
		st.dal.Expect(
			mock.NewExpectation(st.dal.TriggeredMessageForKeys, "NEW_PATIENT", "subkey", []interface{}{}).WithReturns(
				tmModel, nil))
		st.dal.Expect(mock.NewExpectation(st.dal.TriggeredMessageItemsForTriggeredMessageID, tmID, []interface{}{}).WithReturns(
			[]*models.TriggeredMessageItem{tmiModel1, tmiModel2}, nil))
		rtm, err := transformTriggeredMessageToResponse(tmModel, []*models.TriggeredMessageItem{tmiModel1, tmiModel2})
		test.OK(t, err)
		testTriggeredMessages(t, st, &threading.TriggeredMessagesRequest{
			LookupKey: &threading.TriggeredMessagesRequest_Key{
				Key: &threading.TriggeredMessageKey{
					Key:    threading.TRIGGERED_MESSAGE_KEY_NEW_PATIENT,
					Subkey: "subkey",
				},
			},
		}, &threading.TriggeredMessagesResponse{
			TriggeredMessages: []*threading.TriggeredMessage{rtm},
		}, nil)
	})
}

func testTriggeredMessages(
	t *testing.T,
	st *serverTest,
	in *threading.TriggeredMessagesRequest,
	exp *threading.TriggeredMessagesResponse,
	expErr error) {
	resp, err := st.server.TriggeredMessages(st.ctx, in)
	test.Equals(t, expErr, err)
	test.Equals(t, exp, resp)
	st.Finish()
}

func TestDeleteTriggeredMessage(t *testing.T) {
	tmID, err := models.NewTriggeredMessageID()
	test.OK(t, err)
	t.Run("Error-TriggeredMessageIDRequired", func(t *testing.T) {
		st := newServerTest(t)
		testDeleteTriggeredMessage(t, st, &threading.DeleteTriggeredMessageRequest{}, nil, grpcErrorf(codes.InvalidArgument, "TriggeredMessageID is required"))
	})
	t.Run("Success", func(t *testing.T) {
		st := newServerTest(t)
		st.dal.Expect(mock.NewExpectation(st.dal.DeleteTriggeredMessageItemsForTriggeredMessage, tmID))
		st.dal.Expect(mock.NewExpectation(st.dal.DeleteTriggeredMessage, tmID))
		test.OK(t, err)
		testDeleteTriggeredMessage(t, st, &threading.DeleteTriggeredMessageRequest{
			TriggeredMessageID: tmID.String(),
		}, &threading.DeleteTriggeredMessageResponse{}, nil)
	})
}

func testDeleteTriggeredMessage(
	t *testing.T,
	st *serverTest,
	in *threading.DeleteTriggeredMessageRequest,
	exp *threading.DeleteTriggeredMessageResponse,
	expErr error) {
	resp, err := st.server.DeleteTriggeredMessage(st.ctx, in)
	test.Equals(t, expErr, err)
	test.Equals(t, exp, resp)
	st.Finish()
}
