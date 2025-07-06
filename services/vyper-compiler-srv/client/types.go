package client

type CompileRequest struct {
	Version    string `json:"version"`
	Code       string `json:"code"`
	EVMVersion string `json:"evm_version"`
}

type Status string

const (
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

type CompileResponse struct {
	Status          Status `json:"status"`
	Message         string `json:"message,omitempty"`
	ABI             []any  `json:"abi"`
	Bytecode        string `json:"bytecode,omitempty"`
	BytecodeRuntime string `json:"bytecode_runtime,omitempty"`
}
