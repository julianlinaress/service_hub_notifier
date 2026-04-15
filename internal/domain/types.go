package domain

// DeliveryRequest is the normalized payload accepted by the delivery endpoint.
type DeliveryRequest struct {
	DeliveryAttemptKey string            `json:"delivery_attempt_key"`
	Provider           string            `json:"provider"`
	Destination        map[string]any    `json:"destination"`
	Notification       NotificationInput `json:"notification"`
	Event              EventInput        `json:"event"`
}

// NotificationInput carries notification details used to render provider messages.
type NotificationInput struct {
	EventName    string         `json:"event_name"`
	CheckType    string         `json:"check_type"`
	Severity     string         `json:"severity"`
	Message      string         `json:"message"`
	ServiceID    any            `json:"service_id"`
	DeploymentID any            `json:"deployment_id"`
	Metadata     map[string]any `json:"metadata"`
}

// EventInput contains metadata about the originating event.
type EventInput struct {
	ID   string         `json:"id"`
	Name string         `json:"name"`
	Tags map[string]any `json:"tags"`
}

// DeliveryResponse is the normalized result returned by provider adapters.
type DeliveryResponse struct {
	Status               string         `json:"status"`
	ProviderMessageID    string         `json:"provider_message_id,omitempty"`
	ProviderResponseCode string         `json:"provider_response_code,omitempty"`
	ProviderResponse     map[string]any `json:"provider_response"`
	Retryable            bool           `json:"retryable,omitempty"`
	ErrorCode            string         `json:"error_code,omitempty"`
	ErrorMessage         string         `json:"error_message,omitempty"`
}

// SuccessResponse builds a successful delivery response payload.
func SuccessResponse(messageID string, responseCode string, providerResponse map[string]any) DeliveryResponse {
	if providerResponse == nil {
		providerResponse = map[string]any{}
	}

	return DeliveryResponse{
		Status:               StatusDelivered,
		ProviderMessageID:    messageID,
		ProviderResponseCode: responseCode,
		ProviderResponse:     providerResponse,
	}
}

// FailureResponse builds a failed delivery response payload.
func FailureResponse(retryable bool, errorCode string, errorMessage string, responseCode string, providerResponse map[string]any) DeliveryResponse {
	if providerResponse == nil {
		providerResponse = map[string]any{}
	}

	return DeliveryResponse{
		Status:               StatusFailed,
		Retryable:            retryable,
		ErrorCode:            errorCode,
		ErrorMessage:         errorMessage,
		ProviderResponseCode: responseCode,
		ProviderResponse:     providerResponse,
	}
}
