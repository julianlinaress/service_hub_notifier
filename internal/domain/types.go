package domain

type DeliveryRequest struct {
	DeliveryAttemptKey string            `json:"delivery_attempt_key"`
	Provider           string            `json:"provider"`
	Destination        map[string]any    `json:"destination"`
	Notification       NotificationInput `json:"notification"`
	Event              EventInput        `json:"event"`
}

type NotificationInput struct {
	EventName    string         `json:"event_name"`
	CheckType    string         `json:"check_type"`
	Severity     string         `json:"severity"`
	Message      string         `json:"message"`
	ServiceID    any            `json:"service_id"`
	DeploymentID any            `json:"deployment_id"`
	Metadata     map[string]any `json:"metadata"`
}

type EventInput struct {
	ID   string         `json:"id"`
	Name string         `json:"name"`
	Tags map[string]any `json:"tags"`
}

type DeliveryResponse struct {
	Status               string         `json:"status"`
	ProviderMessageID    string         `json:"provider_message_id,omitempty"`
	ProviderResponseCode string         `json:"provider_response_code,omitempty"`
	ProviderResponse     map[string]any `json:"provider_response"`
	Retryable            bool           `json:"retryable,omitempty"`
	ErrorCode            string         `json:"error_code,omitempty"`
	ErrorMessage         string         `json:"error_message,omitempty"`
}

func SuccessResponse(messageID string, responseCode string, providerResponse map[string]any) DeliveryResponse {
	if providerResponse == nil {
		providerResponse = map[string]any{}
	}

	return DeliveryResponse{
		Status:               "delivered",
		ProviderMessageID:    messageID,
		ProviderResponseCode: responseCode,
		ProviderResponse:     providerResponse,
	}
}

func FailureResponse(retryable bool, errorCode string, errorMessage string, responseCode string, providerResponse map[string]any) DeliveryResponse {
	if providerResponse == nil {
		providerResponse = map[string]any{}
	}

	return DeliveryResponse{
		Status:               "failed",
		Retryable:            retryable,
		ErrorCode:            errorCode,
		ErrorMessage:         errorMessage,
		ProviderResponseCode: responseCode,
		ProviderResponse:     providerResponse,
	}
}
