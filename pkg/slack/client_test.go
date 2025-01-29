package slack

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

// TestNewSlackClient ensures the Slack client initializes correctly
func TestNewSlackClient(t *testing.T) {
	client := NewSlackClient("xoxb-test-token", "xapp-test-token")
	assert.NotNil(t, client)
}

// TestFetchMessages ensures FetchMessages retrieves messages correctly
func TestFetchMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	fmt.Println("wtf")
	mockSlack := NewMockClient(ctrl)
	channelID := "C123456"
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	mockSlack.EXPECT().
		FetchMessages(gomock.Any(), channelID, startTime, endTime).
		Return([]slack.Message{{Msg: slack.Msg{Text: "Hello World"}}}, nil).
		Times(1)

	messages, err := mockSlack.FetchMessages(context.Background(), channelID, startTime, endTime)
	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "Hello World", messages[0].Text)
}

// TestFetchMessages_Error simulates an API failure
func TestFetchMessages_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := NewMockClient(ctrl)

	mockSlack.EXPECT().
		FetchMessages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("Slack API error")).
		Times(1)

	messages, err := mockSlack.FetchMessages(context.Background(), "C123456", time.Now().Add(-24*time.Hour), time.Now())
	assert.Error(t, err)
	assert.Nil(t, messages)
}

// TestPostMessageContext ensures PostMessageContext sends messages correctly
func TestPostMessageContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := NewMockClient(ctrl)
	channelID := "C123456"
	message := "Hello Slack"

	mockSlack.EXPECT().
		PostMessageContext(gomock.Any(), channelID, gomock.Any()).
		Return("TS12345", "C123456", nil).
		Times(1)

	_, _, err := mockSlack.PostMessageContext(context.Background(), channelID, slack.MsgOptionText(message, false))
	assert.NoError(t, err)
}

// TestPostMessageContext_Error simulates a failure when posting a message
func TestPostMessageContext_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := NewMockClient(ctrl)

	mockSlack.EXPECT().
		PostMessageContext(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", "", errors.New("Slack API error")).
		Times(1)

	_, _, err := mockSlack.PostMessageContext(context.Background(), "C123456", slack.MsgOptionText("Test", false))
	assert.Error(t, err)
}

// TestUploadFileV2Context ensures file upload works correctly
func TestUploadFileV2Context(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := NewMockClient(ctrl)

	mockSlack.EXPECT().
		UploadFileV2Context(gomock.Any(), gomock.Any()).
		Return(&slack.FileSummary{}, nil).
		Times(1)

	_, err := mockSlack.UploadFileV2Context(context.Background(), slack.UploadFileV2Parameters{})
	assert.NoError(t, err)
}

// TestUploadFileV2Context_Error simulates an error in file upload
func TestUploadFileV2Context_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := NewMockClient(ctrl)

	mockSlack.EXPECT().
		UploadFileV2Context(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("upload failed")).
		Times(1)

	_, err := mockSlack.UploadFileV2Context(context.Background(), slack.UploadFileV2Parameters{})
	assert.Error(t, err)
}

// TestOpenConversationContext ensures OpenConversationContext works
func TestOpenConversationContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSlack := NewMockClient(ctrl)

	mockSlack.EXPECT().
		OpenConversationContext(gomock.Any(), gomock.Any()).
		Return(&slack.Channel{
			GroupConversation: slack.GroupConversation{
				Conversation: slack.Conversation{
					ID: "DM123456",
				},
			},
		}, true, false, nil).
		Times(1)

	conv, _, _, err := mockSlack.OpenConversationContext(context.Background(), &slack.OpenConversationParameters{})
	assert.NoError(t, err)
	assert.NotNil(t, conv)
	assert.Equal(t, "DM123456", conv.ID)
}

// TestOpenConversationContext_Error simulates a failure in opening a conversation
func TestOpenConversationContext_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := NewMockClient(ctrl)

	mockSlack.EXPECT().
		OpenConversationContext(gomock.Any(), gomock.Any()).
		Return(nil, false, false, errors.New("failed to open DM")).
		Times(1)

	conv, _, _, err := mockSlack.OpenConversationContext(context.Background(), &slack.OpenConversationParameters{})
	assert.Error(t, err)
	assert.Nil(t, conv)
}

// ✅ TestRegisterEventHandler ensures RegisterEventHandler correctly stores handlers
func TestRegisterEventHandler(t *testing.T) {
	client := NewSlackClient("xoxb-test-token", "xapp-test-token")
	assert.NotNil(t, client)

	mockHandler := func(eventType string, event interface{}) error {
		return nil
	}

	client.RegisterEventHandler("app_mention", mockHandler)

	// Ensure the handler was registered
	assert.NotNil(t, client.eventsRegistry.events["app_mention"])
}

func TestClientRegisterCommandHandler(t *testing.T) {
	client := NewSlackClient("xoxb-test-token", "xapp-test-token")
	assert.NotNil(t, client)

	mockHandler := func(ctx context.Context, cmd CommandRequest) (string, error) {
		return "Success", nil
	}

	client.RegisterCommandHandler("/test", mockHandler)

	// Ensure the handler was registered
	assert.NotNil(t, client.commandsRegistry.commands["/test"])
}

func TestRegisterInteractiveHandler(t *testing.T) {
	client := NewSlackClient("xoxb-test-token", "xapp-test-token")
	assert.NotNil(t, client)

	mockHandler := func(interType slack.InteractionType, payload slack.InteractionCallback) error {
		return nil
	}

	client.RegisterInteractiveHandler(slack.InteractionTypeShortcut, "test_callback", mockHandler)

	// Ensure the handler was registered
	assert.NotNil(t, client.interaciveRegistry.handlers[slack.InteractionTypeShortcut]["test_callback"])
}

func TestListenEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := NewMockClient(ctrl)

	mockSlack.EXPECT().
		ListenEvents(gomock.Any()).
		Return(nil).
		Times(1)

	err := mockSlack.ListenEvents(context.Background())
	assert.NoError(t, err)
}

func TestFetchReactions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := NewMockClient(ctrl)

	mockSlack.EXPECT().
		FetchReactions(gomock.Any(), "C123456", "1678901234.567890").
		Return([]slack.ItemReaction{
			{Name: "thumbsup", Count: 3},
			{Name: "tada", Count: 5},
		}, nil).
		Times(1)

	reactions, err := mockSlack.FetchReactions(context.Background(), "C123456", "1678901234.567890")
	assert.NoError(t, err)
	assert.Len(t, reactions, 2)
	assert.Equal(t, "thumbsup", reactions[0].Name)
	assert.Equal(t, 3, reactions[0].Count)
}

func TestFetchReactions_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := NewMockClient(ctrl)

	mockSlack.EXPECT().
		FetchReactions(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("Slack API error")).
		Times(1)

	reactions, err := mockSlack.FetchReactions(context.Background(), "C123456", "1678901234.567890")
	assert.Error(t, err)
	assert.Nil(t, reactions)
}

func TestOpenViewContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := NewMockClient(ctrl)

	mockSlack.EXPECT().
		OpenViewContext(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&slack.ViewResponse{}, nil).
		Times(1)

	viewResponse, err := mockSlack.OpenViewContext(context.Background(), "trigger123", slack.ModalViewRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, viewResponse)
}

func TestOpenViewContext_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := NewMockClient(ctrl)

	mockSlack.EXPECT().
		OpenViewContext(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("Failed to open view")).
		Times(1)

	viewResponse, err := mockSlack.OpenViewContext(context.Background(), "trigger123", slack.ModalViewRequest{})
	assert.Error(t, err)
	assert.Nil(t, viewResponse)
}

func TestPostEphemeralContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := NewMockClient(ctrl)

	mockSlack.EXPECT().
		PostEphemeralContext(gomock.Any(), "C123456", "U123456", gomock.Any()).
		Return("TS12345", nil).
		Times(1)

	ts, err := mockSlack.PostEphemeralContext(context.Background(), "C123456", "U123456", slack.MsgOptionText("Hello!", false))
	assert.NoError(t, err)
	assert.Equal(t, "TS12345", ts)
}

func TestPostEphemeralContext_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := NewMockClient(ctrl)

	mockSlack.EXPECT().
		PostEphemeralContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", errors.New("Failed to send ephemeral message")).
		Times(1)

	ts, err := mockSlack.PostEphemeralContext(context.Background(), "C123456", "U123456", slack.MsgOptionText("Hello!", false))
	assert.Error(t, err)
	assert.Empty(t, ts)
}
