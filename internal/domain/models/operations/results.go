package operations

import "encoding/json"

type OperationType string

const (
	Created OperationType = "created"
	Updated OperationType = "updated"
	Error   OperationType = "error"
)

type Result struct {
	Key   string        `json:"key"`
	Type  OperationType `json:"action"`
	Error error         `json:"-"`
}

type Results map[string]Result

// MarshalJSON custom marshaler to properly serialize errors
func (r Result) MarshalJSON() ([]byte, error) {
	type Alias Result
	if r.Error != nil {
		return json.Marshal(&struct {
			*Alias
			Error string `json:"error"`
		}{
			Alias: (*Alias)(&r),
			Error: r.Error.Error(),
		})
	}
	return json.Marshal(&struct {
		*Alias
		Error *string `json:"error"`
	}{
		Alias: (*Alias)(&r),
		Error: nil,
	})
}
