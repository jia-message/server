package services

import (
	"context"
	"log"
	"sync"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
	"google.golang.org/api/option"
	"jia/server/internal/models"
	"jia/server/internal/repositories"
)

type PushService struct {
	pushRepo     *repositories.PushRepository
	settingsRepo *repositories.SettingsRepository

	fcmEnabled bool
	fcmClient  *messaging.Client

	apnsEnabled bool
	apnsClient  *apns2.Client
	apnsTopic   string

	mu sync.RWMutex
}

func NewPushService(
	pushRepo *repositories.PushRepository,
	settingsRepo *repositories.SettingsRepository,
) *PushService {
	ps := &PushService{
		pushRepo:     pushRepo,
		settingsRepo: settingsRepo,
	}
	ps.ReloadConfig()
	return ps
}

func (s *PushService) ReloadConfig() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.settingsRepo.IsSetupCompleted() {
		log.Println("Push service bypass: setup is not yet completed")
		return
	}

	// 1. Load FCM Configuration
	var fcmEnabled bool
	_ = s.settingsRepo.Get("push.fcm_enabled", &fcmEnabled)
	s.fcmEnabled = fcmEnabled

	if fcmEnabled {
		var credentialsJSON string
		err := s.settingsRepo.Get("push.fcm_credentials", &credentialsJSON)
		if err == nil && credentialsJSON != "" {
			opts := option.WithCredentialsJSON([]byte(credentialsJSON))
			app, err := firebase.NewApp(context.Background(), nil, opts)
			if err != nil {
				log.Printf("FCM initialization error: %v", err)
				s.fcmEnabled = false
			} else {
				client, err := app.Messaging(context.Background())
				if err != nil {
					log.Printf("FCM client initialization error: %v", err)
					s.fcmEnabled = false
				} else {
					s.fcmClient = client
					log.Println("FCM messaging client successfully loaded")
				}
			}
		} else {
			log.Println("FCM enabled but push.fcm_credentials setting is empty")
			s.fcmEnabled = false
		}
	}

	// 2. Load APNs Configuration
	var apnsEnabled bool
	_ = s.settingsRepo.Get("push.apns_enabled", &apnsEnabled)
	s.apnsEnabled = apnsEnabled

	if apnsEnabled {
		var keyContent, keyID, teamID, topic string
		_ = s.settingsRepo.Get("push.apns_key", &keyContent)
		_ = s.settingsRepo.Get("push.apns_key_id", &keyID)
		_ = s.settingsRepo.Get("push.apns_team_id", &teamID)
		_ = s.settingsRepo.Get("push.apns_topic", &topic)

		s.apnsTopic = topic

		if keyContent != "" && keyID != "" && teamID != "" {
			authKey, err := token.AuthKeyFromBytes([]byte(keyContent))
			if err != nil {
				log.Printf("APNs private key parsing error: %v", err)
				s.apnsEnabled = false
			} else {
				t := &token.Token{
					AuthKey: authKey,
					KeyID:   keyID,
					TeamID:  teamID,
				}
				s.apnsClient = apns2.NewTokenClient(t).Development() // default to sandbox, could config
				log.Println("APNs client successfully loaded")
			}
		} else {
			log.Println("APNs enabled but credentials settings are incomplete")
			s.apnsEnabled = false
		}
	}
}

func (s *PushService) SendNotification(userID uuid.UUID, senderName string, conversationID uuid.UUID, messageID uuid.UUID, contentType string) {
	subs, err := s.pushRepo.GetSubscriptionsByUserID(userID)
	if err != nil || len(subs) == 0 {
		return
	}

	for _, sub := range subs {
		go func(sub models.PushSubscription) {
			s.mu.RLock()
			fcmEnabled := s.fcmEnabled
			fcmClient := s.fcmClient
			apnsEnabled := s.apnsEnabled
			apnsClient := s.apnsClient
			apnsTopic := s.apnsTopic
			s.mu.RUnlock()

			if sub.Platform == "fcm" && fcmEnabled && fcmClient != nil {
				s.sendFCM(fcmClient, sub.DeviceToken, senderName, conversationID, messageID, contentType)
			} else if sub.Platform == "apns" && apnsEnabled && apnsClient != nil {
				s.sendAPNs(apnsClient, apnsTopic, sub.DeviceToken, senderName, conversationID, messageID, contentType)
			}
		}(sub)
	}
}

func (s *PushService) sendFCM(client *messaging.Client, token string, senderName string, convID, msgID uuid.UUID, contentType string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: senderName,
			Body:  "Sent you a message", // Do not send E2E ciphertext in push notification
		},
		Data: map[string]string{
			"conversation_id": convID.String(),
			"message_id":      msgID.String(),
			"content_type":    contentType,
			"click_action":    "FLUTTER_NOTIFICATION_CLICK",
		},
	}

	_, err := client.Send(ctx, message)
	if err != nil {
		log.Printf("FCM send error: %v", err)
		// If token is invalid, deactivate it
		if messaging.IsRegistrationTokenNotRegistered(err) {
			_ = s.pushRepo.DeactivateToken(token)
		}
	}
}

func (s *PushService) sendAPNs(client *apns2.Client, topic string, token string, senderName string, convID, msgID uuid.UUID, contentType string) {
	p := payload.NewPayload().
		AlertTitle(senderName).
		AlertBody("Sent you a message").
		MutableContent().
		Custom("conversation_id", convID.String()).
		Custom("message_id", msgID.String()).
		Custom("content_type", contentType)

	notification := &apns2.Notification{
		DeviceToken: token,
		Topic:       topic,
		Payload:     p,
	}

	res, err := client.Push(notification)
	if err != nil {
		log.Printf("APNs send error: %v", err)
		return
	}

	if res != nil && res.StatusCode == 410 {
		// Device token is no longer active
		_ = s.pushRepo.DeactivateToken(token)
	}
}

func (s *PushService) RegisterToken(userID uuid.UUID, platform, token, name string) error {
	sub := &models.PushSubscription{
		UserID:      userID,
		Platform:    platform,
		DeviceToken: token,
		DeviceName:  &name,
		IsActive:    true,
	}
	return s.pushRepo.Subscribe(sub)
}

func (s *PushService) UnregisterToken(userID uuid.UUID, token string) error {
	return s.pushRepo.Unsubscribe(userID, token)
}
