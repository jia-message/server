package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"jia/server/internal/models"
	"jia/server/internal/repositories"
)

type KeyService struct {
	keyRepo *repositories.KeyRepository
}

func NewKeyService(keyRepo *repositories.KeyRepository) *KeyService {
	return &KeyService{
		keyRepo: keyRepo,
	}
}

func (s *KeyService) RegisterDeviceBundle(
	userID uuid.UUID,
	deviceID string,
	identityKey string,
	signedPrekey string,
	prekeySignature string,
	oneTimePrekeys []string,
) error {
	if deviceID == "" || identityKey == "" || signedPrekey == "" || prekeySignature == "" {
		return errors.New("missing required key bundle fields")
	}

	dk := &models.DeviceKey{
		UserID:          userID,
		DeviceID:        deviceID,
		IdentityKey:     identityKey,
		SignedPrekey:    signedPrekey,
		PrekeySignature: prekeySignature,
		CreatedAt:       time.Now(),
	}

	if err := s.keyRepo.SaveDeviceKey(dk); err != nil {
		return err
	}

	// Save one-time prekeys
	otps := make([]models.OneTimePrekey, len(oneTimePrekeys))
	now := time.Now()
	for i, pubKey := range oneTimePrekeys {
		otps[i] = models.OneTimePrekey{
			DeviceKeyID: dk.ID,
			PublicKey:   pubKey,
			IsUsed:      false,
			CreatedAt:   now,
		}
	}

	return s.keyRepo.SaveOneTimePrekeys(otps)
}

type DeviceKeyBundle struct {
	DeviceID        string  `json:"device_id"`
	IdentityKey     string  `json:"identity_key"`
	SignedPrekey    string  `json:"signed_prekey"`
	PrekeySignature string  `json:"prekey_signature"`
	OneTimePrekey   *string `json:"one_time_prekey,omitempty"` // Consumed OTP
}

func (s *KeyService) FetchKeyBundle(targetUserID uuid.UUID) ([]DeviceKeyBundle, error) {
	devices, err := s.keyRepo.GetDeviceKeysByUserID(targetUserID)
	if err != nil {
		return nil, err
	}

	bundles := make([]DeviceKeyBundle, len(devices))
	for i, dev := range devices {
		otp, err := s.keyRepo.GetUnusedOneTimePrekey(dev.ID)
		var otpPub *string
		if err == nil && otp != nil {
			otpPub = &otp.PublicKey
		}

		bundles[i] = DeviceKeyBundle{
			DeviceID:        dev.DeviceID,
			IdentityKey:     dev.IdentityKey,
			SignedPrekey:    dev.SignedPrekey,
			PrekeySignature: dev.PrekeySignature,
			OneTimePrekey:   otpPub,
		}
	}

	return bundles, nil
}

func (s *KeyService) ReplenishPrekeys(userID uuid.UUID, deviceID string, oneTimePrekeys []string) error {
	dev, err := s.keyRepo.GetDeviceKey(userID, deviceID)
	if err != nil {
		return errors.New("device not registered")
	}

	otps := make([]models.OneTimePrekey, len(oneTimePrekeys))
	now := time.Now()
	for i, pubKey := range oneTimePrekeys {
		otps[i] = models.OneTimePrekey{
			DeviceKeyID: dev.ID,
			PublicKey:   pubKey,
			IsUsed:      false,
			CreatedAt:   now,
		}
	}

	return s.keyRepo.SaveOneTimePrekeys(otps)
}

func (s *KeyService) GetUnusedPrekeysCount(userID uuid.UUID, deviceID string) (int64, error) {
	dev, err := s.keyRepo.GetDeviceKey(userID, deviceID)
	if err != nil {
		return 0, errors.New("device not registered")
	}
	return s.keyRepo.CountUnusedPrekeys(dev.ID)
}
