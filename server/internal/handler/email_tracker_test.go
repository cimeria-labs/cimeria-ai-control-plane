package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func signedResendRequest(t *testing.T, body []byte, secret []byte) *http.Request {
	t.Helper()
	req := httptest.NewRequest("POST", "/api/webhooks/resend", nil)
	msgID := "msg_test"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(msgID + "." + timestamp + "."))
	mac.Write(body)
	signature := "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
	req.Header.Set("svix-id", msgID)
	req.Header.Set("svix-timestamp", timestamp)
	req.Header.Set("svix-signature", signature)
	return req
}

func TestVerifyResendWebhookSignature(t *testing.T) {
	body := []byte(`{"type":"email.delivered"}`)
	secret := []byte("test-secret")
	t.Setenv("RESEND_WEBHOOK_SECRET", "whsec_"+base64.StdEncoding.EncodeToString(secret))

	req := signedResendRequest(t, body, secret)
	if err := verifyResendWebhookSignature(req, body); err != nil {
		t.Fatalf("expected valid signature, got %v", err)
	}

	req.Header.Set("svix-signature", "v1,invalid")
	if err := verifyResendWebhookSignature(req, body); err == nil {
		t.Fatal("expected invalid signature error")
	}
}

func TestVerifyResendWebhookSignatureRequiresSecret(t *testing.T) {
	t.Setenv("RESEND_WEBHOOK_SECRET", "")
	req := signedResendRequest(t, []byte(`{}`), []byte("test-secret"))
	if err := verifyResendWebhookSignature(req, []byte(`{}`)); err != errWebhookSecretMissing {
		t.Fatalf("expected missing secret error, got %v", err)
	}
}
