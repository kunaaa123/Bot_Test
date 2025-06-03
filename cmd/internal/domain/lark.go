package domain

type LarkService interface {
	GetTenantAccessToken() (string, error)
	UploadImage(filePath string, token string) (string, error)
	SendWebhookMessage(payload any) error
}
