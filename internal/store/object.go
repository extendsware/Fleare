package store

type Kind int

const (
	Invalid Kind = iota
	Int
	Float
	String
	Bool
)

type Object struct {
	Type  Kind
	Value interface{}
}
