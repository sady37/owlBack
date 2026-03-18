package observation

// DataType for field definitions.
type DataType uint8

const (
	TInt   DataType = iota // integer
	TFloat                 // float64
	TBool                  // 0/1
	TEnum                  // integer index into Enum list
	TStr                   // string
)

var dataTypeNames = [...]string{"int", "float", "bool", "enum", "string"}

func (t DataType) String() string {
	if int(t) < len(dataTypeNames) {
		return dataTypeNames[t]
	}
	return "unknown"
}
