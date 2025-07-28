package domain

type Webhook struct {
	ID     string      `json:"id" dynamodbav:"ID"`
	URL    string      `json:"url" dynamodbav:"URL"`
	Events []EventType `json:"events" dynamodbav:"Events"` // Lista de eventos a los que está suscrito
}
