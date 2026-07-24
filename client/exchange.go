package client

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/bloom-kit/mappers"
	"github.com/slipe-fun/skid-v4/pkg/identity"
	"github.com/slipe-fun/skid-v4/pkg/messages"
)

type ExchangeEvent struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
}

type ExchangeInit struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
	domain.PublicKeys
}

type ExchangeHandshake struct {
	Type             string `json:"type"`
	UserID           string `json:"user_id"`
	domain.Handshake `json:"handshake"`
}

type ExchangeMessage struct {
	Type       string `json:"type"`
	UserID     string `json:"user_id"`
	Ciphertext string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
	Salt       string `json:"salt"`
}

func (c *BloomClient) SendExchangeEvent(eventType string, conn *websocket.Conn) error {
	askForHandshakeMessage := ExchangeEvent{
		Type:   eventType,
		UserID: c.exchangeUserIdentity.ID,
	}

	eventBytes, err := json.Marshal(askForHandshakeMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, eventBytes); err != nil {
		return fmt.Errorf("failed to send: %w", err)
	}

	return nil
}

func (c *BloomClient) GenerateExchangeSession(exchangeType string) error {
	userIdentity, secretKeys, err := identity.GenerateIdentity()
	if err != nil {
		return err
	}

	c.exchangeMu.Lock()
	if c.exchangeSecretKeys != nil {
		c.exchangeSecretKeys.Wipe()
		c.exchangeSecretKeys = nil
	}
	c.exchangeSecretKeys = secretKeys
	c.exchangeUserIdentity = userIdentity
	if exchangeType == "push" {
		c.exchangeUserIdentity.ID = c.credentials.UserID
	}
	c.exchangeMu.Unlock()

	return nil
}

func (c *BloomClient) GenerateExchangeFingerprint() string {
	c.exchangeMu.Lock()
	defer c.exchangeMu.Unlock()

	if c.exchangeUserIdentity == nil {
		return ""
	}

	fingerprintBytes := sha256.Sum256(c.exchangeUserIdentity.PublicKeys.MlKem768)
	return base64.StdEncoding.EncodeToString(fingerprintBytes[:])
}

func (c *BloomClient) StartExchangeSession(exchangeType string) (string, string, error) {
	roomID, err := c.exchangeManager.StartSession(context.Background())
	if err != nil {
		return "", "", err
	}

	c.GenerateExchangeSession(exchangeType)

	fingerprint := c.GenerateExchangeFingerprint()

	return roomID, fingerprint, nil
}

func (c *BloomClient) CancelExchange() {
	c.exchangeMu.Lock()
	defer c.exchangeMu.Unlock()
	if c.exchangeCancel != nil {
		c.exchangeCancel()
		c.exchangeCancel = nil
	}
}

func (c *BloomClient) Exchange(exchangeType, roomID, fingerprint string) (string, error) {
	isScanner := c.GenerateExchangeFingerprint() != fingerprint

	if isScanner {
		c.GenerateExchangeSession(exchangeType)
	}

	ctx, cancel := context.WithCancel(context.Background())

	c.exchangeMu.Lock()
	c.exchangeCancel = cancel
	c.exchangeMu.Unlock()

	defer func() {
		c.exchangeMu.Lock()
		if c.exchangeSecretKeys != nil {
			c.exchangeSecretKeys.Wipe()
			c.exchangeSecretKeys = nil
		}
		c.exchangeUserIdentity = nil
		if c.exchangeCancel != nil {
			c.exchangeCancel = nil
		}
		c.exchangeMu.Unlock()
		cancel()
	}()

	wsURL := strings.Replace(c.wsURL, "/ws", "/exchange/ws", 1)
	url := fmt.Sprintf("%s?room_id=%s", wsURL, roomID)

	var creds *domain.SavedCredentials
	if exchangeType == "push" {
		var err error
		creds, err = c.loadCredentials()
		if err != nil {
			return "", fmt.Errorf("failed to load credentials for push exchange: %w", err)
		}
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to dial exchange websocket: %w", err)
	}
	defer conn.Close()

	identityMessage := &ExchangeInit{
		Type:   "exchange.identity",
		UserID: c.exchangeUserIdentity.ID,
		PublicKeys: domain.PublicKeys{
			MlKem768: base64.StdEncoding.EncodeToString(c.exchangeUserIdentity.PublicKeys.MlKem768),
			X448:     base64.StdEncoding.EncodeToString(c.exchangeUserIdentity.PublicKeys.X448),
			Ed448:    base64.StdEncoding.EncodeToString(c.exchangeUserIdentity.PublicKeys.Ed448),
		},
	}

	identityBytes, err := json.Marshal(identityMessage)
	if err != nil {
		return "", fmt.Errorf("failed to marshal: %w", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, identityBytes); err != nil {
		return "", fmt.Errorf("failed to send: %w", err)
	}

	_ = c.SendExchangeEvent("exchange.ask_for_identity", conn)

	var (
		recipientIdentity       *identity.User
		handshake               *identity.HandshakePayload
		chatKey                 []byte
		syncKey                 []byte
		pendingHandshakeRequest bool
		pendingRecoveryKey      bool
	)

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				return "", fmt.Errorf("websocket read error: %w", err)
			}

			var event WsEvent
			if err := json.Unmarshal(message, &event); err != nil {
				continue
			}

			switch event.Type {
			case "exchange.identity":
				var event ExchangeInit
				if err := json.Unmarshal(message, &event); err != nil {
					continue
				}

				if event.UserID == c.exchangeUserIdentity.ID {
					continue
				}

				if isScanner {
					decodedMlKem, err := base64.StdEncoding.DecodeString(event.PublicKeys.MlKem768)
					if err != nil {
						return "", fmt.Errorf("failed to decode incoming ml_kem public key: %w", err)
					}

					incomingFingerprintBytes := sha256.Sum256(decodedMlKem)
					incomingFingerprint := base64.StdEncoding.EncodeToString(incomingFingerprintBytes[:])

					if incomingFingerprint != fingerprint {
						return "", fmt.Errorf("incoming public key fingerprint does not match QR code")
					}
				}

				receiverPublicKeys, err := mappers.UnmapPublicKeys(&event.PublicKeys)
				if err != nil {
					return "", err
				}

				recipientIdentity = &identity.User{
					ID:         event.UserID,
					PublicKeys: *receiverPublicKeys,
				}

				if exchangeType == "push" && pendingHandshakeRequest {
					pendingHandshakeRequest = false

					handshake, chatKey, syncKey, err = identity.InitiateKeyExchange(c.exchangeUserIdentity, c.exchangeSecretKeys, recipientIdentity)
					if err != nil {
						return "", err
					}

					mappedHandshake := mappers.MapHandshake(handshake)

					exchangeHandshakeMessage := ExchangeHandshake{
						Type:      "exchange.handshake",
						UserID:    c.exchangeUserIdentity.ID,
						Handshake: *mappedHandshake,
					}

					handshakeBytes, err := json.Marshal(exchangeHandshakeMessage)
					if err != nil {
						return "", fmt.Errorf("failed to marshal: %w", err)
					}

					if err := conn.WriteMessage(websocket.TextMessage, handshakeBytes); err != nil {
						return "", fmt.Errorf("failed to send: %w", err)
					}
				}

				if exchangeType == "pull" {
					if err := c.SendExchangeEvent("exchange.ask_for_handshake", conn); err != nil {
						return "", err
					}
				}

			case "exchange.handshake":
				var event ExchangeHandshake
				if err := json.Unmarshal(message, &event); err != nil {
					continue
				}

				if event.UserID == c.exchangeUserIdentity.ID {
					continue
				}

				handshake, err = mappers.DecodeHandshake(&event.Handshake)
				if err != nil {
					return "", err
				}

				if recipientIdentity == nil {
					if err := c.SendExchangeEvent("exchange.ask_for_identity", conn); err != nil {
						return "", err
					}
					continue
				}

				if exchangeType == "pull" {
					if err := c.SendExchangeEvent("exchange.ask_for_recovery_key", conn); err != nil {
						return "", err
					}
				}

				chatKey, syncKey, err = c.chatManager.FinalizeHandshake(handshake, c.exchangeUserIdentity, c.exchangeSecretKeys, recipientIdentity)
				if err != nil {
					return "", err
				}

				if exchangeType == "push" && pendingRecoveryKey {
					pendingRecoveryKey = false

					message, err := messages.Encrypt(chatKey, creds.RecoveryKey, syncKey, c.exchangeUserIdentity, recipientIdentity)
					if err != nil {
						return "", err
					}

					exchangeEncryptedRecoveryKeyMessage := ExchangeMessage{
						Type:       "exchange.recovery_key",
						Ciphertext: base64.StdEncoding.EncodeToString(message.Ciphertext),
						Nonce:      base64.StdEncoding.EncodeToString(message.Nonce),
						Salt:       base64.StdEncoding.EncodeToString(message.Salt),
					}

					recoveryKeyBytes, err := json.Marshal(exchangeEncryptedRecoveryKeyMessage)
					if err != nil {
						return "", fmt.Errorf("failed to marshal: %w", err)
					}

					if err := conn.WriteMessage(websocket.TextMessage, recoveryKeyBytes); err != nil {
						return "", fmt.Errorf("failed to send: %w", err)
					}
				}

			case "exchange.recovery_key":
				var event ExchangeMessage
				if err := json.Unmarshal(message, &event); err != nil {
					continue
				}

				if event.UserID == c.exchangeUserIdentity.ID {
					continue
				}

				if recipientIdentity == nil {
					if err := c.SendExchangeEvent("exchange.ask_for_identity", conn); err != nil {
						return "", err
					}
					continue
				}

				if chatKey == nil || syncKey == nil {
					if err := c.SendExchangeEvent("exchange.ask_for_handshake", conn); err != nil {
						return "", err
					}
					continue
				}

				ciphertext, err := base64.StdEncoding.DecodeString(event.Ciphertext)
				if err != nil {
					return "", err
				}

				nonce, err := base64.StdEncoding.DecodeString(event.Nonce)
				if err != nil {
					return "", err
				}

				salt, err := base64.StdEncoding.DecodeString(event.Salt)
				if err != nil {
					return "", err
				}

				decryptedMessage, err := messages.Decrypt(chatKey, messages.EncryptedMessage{
					Ciphertext: ciphertext,
					Nonce:      nonce,
					Salt:       salt,
				}, syncKey, c.exchangeUserIdentity, recipientIdentity)
				if err != nil {
					return "", fmt.Errorf("failed to decrypt recovery key: %w", err)
				}

				_ = c.SendExchangeEvent("exchange.confirm", conn)

				return hex.EncodeToString(decryptedMessage.Content), nil

			case "exchange.confirm":
				var event ExchangeEvent
				if err := json.Unmarshal(message, &event); err != nil {
					continue
				}

				if event.UserID == c.exchangeUserIdentity.ID {
					continue
				}

				return hex.EncodeToString(creds.RecoveryKey), nil

			case "exchange.ask_for_identity":
				var event ExchangeEvent
				if err := json.Unmarshal(message, &event); err != nil {
					continue
				}

				if event.UserID == c.exchangeUserIdentity.ID {
					continue
				}

				if err := conn.WriteMessage(websocket.TextMessage, identityBytes); err != nil {
					return "", fmt.Errorf("failed to send: %w", err)
				}

			case "exchange.ask_for_handshake":
				var event ExchangeEvent
				if err := json.Unmarshal(message, &event); err != nil {
					continue
				}

				if event.UserID == c.exchangeUserIdentity.ID {
					continue
				}

				if recipientIdentity == nil {
					pendingHandshakeRequest = true
					if err := c.SendExchangeEvent("exchange.ask_for_identity", conn); err != nil {
						return "", err
					}
					continue
				}

				handshake, chatKey, syncKey, err = identity.InitiateKeyExchange(c.exchangeUserIdentity, c.exchangeSecretKeys, recipientIdentity)
				if err != nil {
					return "", err
				}

				mappedHandshake := mappers.MapHandshake(handshake)

				exchangeHandshakeMessage := ExchangeHandshake{
					Type:      "exchange.handshake",
					UserID:    c.exchangeUserIdentity.ID,
					Handshake: *mappedHandshake,
				}

				handshakeBytes, err := json.Marshal(exchangeHandshakeMessage)
				if err != nil {
					return "", fmt.Errorf("failed to marshal: %w", err)
				}

				if err := conn.WriteMessage(websocket.TextMessage, handshakeBytes); err != nil {
					return "", fmt.Errorf("failed to send: %w", err)
				}

			case "exchange.ask_for_recovery_key":
				var event ExchangeEvent
				if err := json.Unmarshal(message, &event); err != nil {
					continue
				}

				if event.UserID == c.exchangeUserIdentity.ID {
					continue
				}

				if recipientIdentity == nil {
					if err := c.SendExchangeEvent("exchange.ask_for_identity", conn); err != nil {
						return "", err
					}
					continue
				}

				if chatKey == nil || syncKey == nil {
					pendingRecoveryKey = true
					if err := c.SendExchangeEvent("exchange.ask_for_handshake", conn); err != nil {
						return "", err
					}
					continue
				}

				message, err := messages.Encrypt(chatKey, creds.RecoveryKey, syncKey, c.exchangeUserIdentity, recipientIdentity)
				if err != nil {
					return "", err
				}

				exchangeEncryptedRecoveryKeyMessage := ExchangeMessage{
					Type:       "exchange.recovery_key",
					Ciphertext: base64.StdEncoding.EncodeToString(message.Ciphertext),
					Nonce:      base64.StdEncoding.EncodeToString(message.Nonce),
					Salt:       base64.StdEncoding.EncodeToString(message.Salt),
				}

				recoveryKeyBytes, err := json.Marshal(exchangeEncryptedRecoveryKeyMessage)
				if err != nil {
					return "", fmt.Errorf("failed to marshal: %w", err)
				}

				if err := conn.WriteMessage(websocket.TextMessage, recoveryKeyBytes); err != nil {
					return "", fmt.Errorf("failed to send: %w", err)
				}
			}
		}
	}
}
