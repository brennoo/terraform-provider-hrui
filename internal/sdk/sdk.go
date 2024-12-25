package sdk

// SDK wraps the underlying client to provide high-level functionality for CRUD operations.
type SDK struct {
	client *HRUIClient
}

// NewSDK creates a new SDK instance with the provided client.
func NewSDK(client *HRUIClient) *SDK {
	return &SDK{
		client: client,
	}
}
