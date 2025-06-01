package model

type Record struct {
	Value interface{}
}
type StringRecord struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}
