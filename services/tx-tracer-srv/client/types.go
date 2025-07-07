package client

// AddressInfo 对应前端的 AddressInfo 类型
type AddressInfo struct {
	Label     string                 `json:"label"`
	Functions map[string]interface{} `json:"functions"`
	Events    map[string]interface{} `json:"events"`
	Errors    map[string]interface{} `json:"errors"`
	Fragments []interface{}          `json:"fragments"`
}

// TraceEntryCall 对应前端的 TraceEntryCall 类型
type TraceEntryCall struct {
	Path         string       `json:"path"`
	Type         string       `json:"type"`
	Variant      string       `json:"variant"`
	Gas          int          `json:"gas"`
	IsPrecompile bool         `json:"isPrecompile"`
	From         string       `json:"from"`
	To           string       `json:"to"`
	Input        string       `json:"input"`
	Output       string       `json:"output"`
	GasUsed      int          `json:"gasUsed"`
	Value        string       `json:"value"`
	Status       int          `json:"status"`
	Codehash     string       `json:"codehash"`
	Children     []TraceEntry `json:"children"`
}

// TraceEntryLog 对应前端的 TraceEntryLog 类型
type TraceEntryLog struct {
	Path   string   `json:"path"`
	Type   string   `json:"type"`
	Topics []string `json:"topics"`
	Data   string   `json:"data"`
}

// TraceEntrySload 对应前端的 TraceEntrySload 类型
type TraceEntrySload struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Slot  string `json:"slot"`
	Value string `json:"value"`
}

// TraceEntrySstore 对应前端的 TraceEntrySstore 类型
type TraceEntrySstore struct {
	Path     string `json:"path"`
	Type     string `json:"type"`
	Slot     string `json:"slot"`
	OldValue string `json:"oldValue"`
	NewValue string `json:"newValue"`
}

// TraceEntry 对应前端的 TraceEntry 联合类型
type TraceEntry interface{}

// TraceResponse 对应前端的 TraceResponse 类型
type TraceResponse struct {
	Chain      string                            `json:"chain"`
	Txhash     string                            `json:"txhash"`
	Preimages  map[string]string                 `json:"preimages"`
	Addresses  map[string]map[string]AddressInfo `json:"addresses"`
	Entrypoint TraceEntryCall                    `json:"entrypoint"`
}

// TypeDescriptions 对应前端的 TypeDescriptions 类型
type TypeDescriptions struct {
	TypeIdentifier string `json:"typeIdentifier"`
	TypeString     string `json:"typeString"`
}

// TypeName 对应前端的 TypeName 类型
type TypeName struct {
	NodeType         string           `json:"nodeType"`
	TypeDescriptions TypeDescriptions `json:"typeDescriptions"`
	KeyType          *TypeName        `json:"keyType,omitempty"`
	ValueType        *TypeName        `json:"valueType,omitempty"`
}

// VariableInfo 对应前端的 VariableInfo 类型
type VariableInfo struct {
	Name     string   `json:"name"`
	FullName string   `json:"fullName"`
	TypeName TypeName `json:"typeName"`
	Bits     int      `json:"bits"`
}

// BaseSlotInfo 对应前端的 BaseSlotInfo 类型
type BaseSlotInfo struct {
	Resolved  bool                 `json:"resolved"`
	Variables map[int]VariableInfo `json:"variables"`
}

// RawSlotInfo 对应前端的 RawSlotInfo 类型
type RawSlotInfo struct {
	BaseSlotInfo
	Type string `json:"type"` // "raw"
}

// DynamicSlotInfo 对应前端的 DynamicSlotInfo 类型
type DynamicSlotInfo struct {
	BaseSlotInfo
	Type     string `json:"type"` // "dynamic"
	BaseSlot string `json:"baseSlot"`
	Key      string `json:"key"`
	Offset   int    `json:"offset"`
}

// MappingSlotInfo 对应前端的 MappingSlotInfo 类型
type MappingSlotInfo struct {
	BaseSlotInfo
	Type       string `json:"type"` // "mapping"
	BaseSlot   string `json:"baseSlot"`
	MappingKey string `json:"mappingKey"`
	Offset     int    `json:"offset"`
}

// ArraySlotInfo 对应前端的 ArraySlotInfo 类型
type ArraySlotInfo struct {
	BaseSlotInfo
	Type     string `json:"type"` // "array"
	BaseSlot string `json:"baseSlot"`
	Offset   int    `json:"offset"`
}

// StructSlotInfo 对应前端的 StructSlotInfo 类型
type StructSlotInfo struct {
	BaseSlotInfo
	Type   string `json:"type"` // "struct"
	Offset int    `json:"offset"`
}

// SlotInfo 对应前端的 SlotInfo 联合类型
type SlotInfo interface{}

// StorageResponse 对应前端的 StorageResponse 类型
type StorageResponse struct {
	AllStructs []interface{}       `json:"allStructs"`
	Arrays     []interface{}       `json:"arrays"`
	Structs    []interface{}       `json:"structs"`
	Slots      map[string]SlotInfo `json:"slots"`
}
