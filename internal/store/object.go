package store

type Kind uint32

const (
	Default Kind = iota
	String
	Number
	Map
	Set
	List
)

type Object struct {
	Type  Kind
	Value []byte
}
