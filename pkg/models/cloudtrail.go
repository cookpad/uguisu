package models

// CloudTrailLogObject represents S3 object data of CloudTrail log
type CloudTrailLogObject struct {
	Records []*CloudTrailRecord `json:"Records"`
}

// CloudTrailRecord represents one event log of CloudTrail
type CloudTrailRecord struct {
	EventTime       string                 `json:"eventTime"`
	EventVersion    string                 `json:"eventVersion"`
	UserIdentity    CloudTrailUserIdentity `json:"userIdentity"`
	EventSource     string                 `json:"eventSource"`
	EventName       string                 `json:"eventName"`
	AwsRegion       string                 `json:"awsRegion"`
	SourceIPAddress string                 `json:"sourceIPAddress"`
	UserAgent       string                 `json:"userAgent"`

	ErrorCode    *string `json:"errorCode,omitempty"`
	ErrorMessage *string `json:"errorMessage,omitempty"`

	RequestParameters   map[string]interface{}         `json:"requestParameters,omitempty"`
	ResponseElements    map[string]interface{}         `json:"responseElements,omitempty"`
	AdditionalEventData *CloudTrailAdditionalEventData `json:"additionalEventData,omitempty"`
	RequestID           string                         `json:"requestID"`
	EventID             string                         `json:"eventID"`
	EventType           string                         `json:"eventType"`
	APIVersion          string                         `json:"apiVersion"`
	ManagementEvent     bool                           `json:"managementEvent"`
	ReadOnly            bool                           `json:"readOnly"`
	Resources           interface{}                    `json:"resources,omitempty"`
	RecipientAccountID  string                         `json:"recipientAccountId"`
	ServiceEventDetails map[string]interface{}         `json:"serviceEventDetails"`
	SharedEventID       string                         `json:"sharedEventID"`
	VpcEndpointID       string                         `json:"vpcEndpointId"`
	EventCategory       string                         `json:"eventCategory"`
}

// CloudTrailUserIdentity represents userIdentity field in CloudTrail record
type CloudTrailUserIdentity struct {
	AccountID      string                    `json:"accountId"`
	ARN            string                    `json:"arn"`
	InvokedBy      *string                   `json:"invokedBy,omitempty"`
	SessionContext *CloudTrailSessionContext `json:"sessionContext,omitempty"`
	Type           string                    `json:"type"`
}

type CloudTrailAdditionalEventData struct {
	MFAUsed         string
	SamlProviderArn *string `json:"SamlProviderArn,omitempty"`
}

type CloudTrailSessionContext struct {
	SessionIssuer *CloudTrailSessionIssuer `json:"sessionIssuer,omitempty"`
}

type CloudTrailSessionIssuer struct {
	Type string `json:"type"`
}
