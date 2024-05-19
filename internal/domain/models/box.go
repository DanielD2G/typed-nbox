package models

type Box struct {
	Service string           `json:"service"`
	Stage   map[string]Stage `json:"stage"`
}

type Stage struct {
	Template Template `json:"template"`
}

type Template struct {
	Name  string `json:"name" dynamodbav:"path"` // s3 path
	Value string `json:"value" dynamodbav:"value"`
}
