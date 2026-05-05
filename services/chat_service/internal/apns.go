package chat

import (
	"crypto/ecdsa"
	"encoding/base64"
	"log"
	"os"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"
)

var apnsClient *apns2.Client

func InitAPNs() {
	keyID := os.Getenv("APNS_KEY_ID")
	teamID := os.Getenv("APNS_TEAM_ID")
	apnsEnv := os.Getenv("APNS_ENVIRONMENT") // "production" or "development"

	if keyID == "" || teamID == "" {
		log.Println("APNs: APNS_KEY_ID or APNS_TEAM_ID not set, push notifications disabled")
		return
	}

	var authKey *ecdsa.PrivateKey
	var err error

	// Prefer loading the key from a base64-encoded environment variable (Docker-friendly).
	// Fall back to file path for local development.
	if keyB64 := os.Getenv("APNS_KEY_BASE64"); keyB64 != "" {
		keyBytes, decErr := base64.StdEncoding.DecodeString(keyB64)
		if decErr != nil {
			log.Printf("APNs: failed to decode APNS_KEY_BASE64: %v", decErr)
			return
		}
		authKey, err = token.AuthKeyFromBytes(keyBytes)
	} else {
		authKeyPath := os.Getenv("APNS_KEY_PATH")
		if authKeyPath == "" {
			authKeyPath = "../../AuthKey_" + keyID + ".p8"
		}
		authKey, err = token.AuthKeyFromFile(authKeyPath)
	}

	if err != nil {
		log.Printf("APNs: failed to load auth key: %v", err)
		return
	}

	authToken := &token.Token{
		AuthKey: authKey,
		KeyID:   keyID,
		TeamID:  teamID,
	}

	if apnsEnv == "production" {
		apnsClient = apns2.NewTokenClient(authToken).Production()
	} else {
		apnsClient = apns2.NewTokenClient(authToken).Development()
	}
	log.Printf("APNs Client initialized (%s)", apnsEnv)
}

func SendPushNotification(deviceTokens []string, title, body string, metadata map[string]interface{}, badgeCount int) {
	if apnsClient == nil {
		return
	}

	bundleID := os.Getenv("APNS_BUNDLE_ID")
	if bundleID == "" {
		bundleID = "com.danilo.GoMessengeriOS"
	}

	for _, deviceToken := range deviceTokens {
		notification := &apns2.Notification{
			DeviceToken: deviceToken,
			Topic:       bundleID,
			Payload: map[string]interface{}{
				"aps": map[string]interface{}{
					"alert": map[string]interface{}{
						"title": title,
						"body":  body,
					},
					"sound": "default",
					"badge": badgeCount,
				},
				"metadata": metadata,
			},
		}

		res, err := apnsClient.Push(notification)
		if err != nil {
			log.Printf("APNs Push Error to %s: %v\n", deviceToken, err)
			continue
		}
		if res.StatusCode != 200 {
			log.Printf("APNs Rejected token %s: %v %v\n", deviceToken, res.StatusCode, res.Reason)
		} else {
			log.Printf("APNs Success sending to %s", deviceToken)
		}
	}
}
