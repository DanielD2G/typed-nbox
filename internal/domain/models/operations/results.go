package operations

type OperationType string

const (
	Created OperationType = "created"
	Updated OperationType = "updated"
)

type Result struct {
	Key   string        `json:"key"`
	Type  OperationType `json:"action"`
	Error error         `json:"error"`
}

type Results map[string]Result
