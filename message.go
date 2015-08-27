package service

// Message provides a simple JSON struct for serialising string messages as
// responses to calls that don't have a complex Type
type Message struct {
	Message string `json:"message"`
}
