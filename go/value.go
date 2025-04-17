package value

type ValueType int8

const (
	None        ValueType = 0
	Boolean     ValueType = 1
	Int64       ValueType = 2
	String      ValueType = 3
	Dictionnary ValueType = 4
	List        ValueType = 5
)
