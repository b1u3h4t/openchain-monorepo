package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/openchainxyz/openchainxyz-monorepo/internal/ethclient"
	"github.com/openchainxyz/openchainxyz-monorepo/services/tx-tracer-srv/client"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	HttpPort int `def:"8083" env:"HTTP_PORT"`
}

type Service struct {
	config *Config
}

func New(config *Config) (*Service, error) {
	return &Service{
		config: config,
	}, nil
}

func (s *Service) Start() error {
	s.startServer()
	return nil
}

func (s *Service) serveTrace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chain := vars["chain"]
	txhash := vars["txhash"]

	// 验证交易哈希格式
	if !common.IsHexAddress(txhash) && len(txhash) != 66 { // 66 = 0x + 64 hex chars
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": "invalid transaction hash format",
		})
		return
	}

	// 连接到以太坊节点
	cli, err := ethclient.Dial("https://rpc.ankr.com/avalanche/")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": fmt.Sprintf("failed to connect to ethereum node: %v", err),
		})
		return
	}

	// 使用 TraceTransaction 方法
	traceConfig := &tracers.TraceConfig{
		Tracer: stringPtr("callTracer"),
	}

	hash := common.HexToHash(txhash)
	result, err := cli.TraceTransaction(context.Background(), hash, traceConfig)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": fmt.Sprintf("failed to trace transaction: %v", err),
		})
		return
	}

	// 检查结果是否为空
	if len(result) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": "transaction not found or no trace available",
		})
		return
	}

	// 解析 trace 结果
	var traceResult map[string]interface{}
	if err := json.Unmarshal(result, &traceResult); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": fmt.Sprintf("failed to parse trace result: %v", err),
		})
		return
	}

	// 转换为 TraceResponse 格式
	response := s.convertTraceResultToResponse(chain, txhash, traceResult)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"ok":     true,
		"error":  "",
		"result": response,
	})
}

// convertTraceResultToResponse 将 trace 结果转换为 TraceResponse 格式
func (s *Service) convertTraceResultToResponse(chain, txhash string, traceResult map[string]interface{}) client.TraceResponse {
	// 初始化响应
	response := client.TraceResponse{
		Chain:     chain,
		Txhash:    txhash,
		Preimages: make(map[string]string),
		Addresses: make(map[string]map[string]client.AddressInfo),
	}

	// 转换主调用并收集所有地址
	addresses := make(map[string]bool)
	if entrypoint, ok := s.convertCallToEntry(traceResult, "0", addresses); ok {
		response.Entrypoint = entrypoint
	}

	// 为每个地址生成基本的address信息
	log.Printf("Collected addresses: %v", addresses)
	for address := range addresses {
		if address != "" {
			log.Printf("Adding address: %s", address)
			response.Addresses[address] = map[string]client.AddressInfo{
				"0x": {
					Label:     "Contract",
					Functions: make(map[string]interface{}),
					Events:    make(map[string]interface{}),
					Errors:    make(map[string]interface{}),
					Fragments: []interface{}{},
				},
			}
		}
	}

	return response
}

// convertCallToEntry 将 call 对象转换为 TraceEntryCall
func (s *Service) convertCallToEntry(call map[string]interface{}, path string, addresses map[string]bool) (client.TraceEntryCall, bool) {
	entry := client.TraceEntryCall{
		Path:         path,
		Type:         "call",
		Variant:      "call",
		IsPrecompile: false,
		Children:     []client.TraceEntry{},
	}

	// 转换基本字段
	if from, ok := call["from"].(string); ok {
		entry.From = from
		addresses[from] = true
	}
	if to, ok := call["to"].(string); ok {
		entry.To = to
		addresses[to] = true
	}
	if input, ok := call["input"].(string); ok {
		entry.Input = input
	}
	if output, ok := call["output"].(string); ok {
		entry.Output = output
	}
	if value, ok := call["value"].(string); ok {
		entry.Value = value
	}

	// 转换 gas 相关字段
	if gas, ok := call["gas"].(string); ok {
		if len(gas) >= 2 && gas[:2] == "0x" {
			if gasInt, err := strconv.ParseInt(gas[2:], 16, 64); err == nil {
				entry.Gas = int(gasInt)
			}
		}
	}
	if gasUsed, ok := call["gasUsed"].(string); ok {
		if len(gasUsed) >= 2 && gasUsed[:2] == "0x" {
			if gasUsedInt, err := strconv.ParseInt(gasUsed[2:], 16, 64); err == nil {
				entry.GasUsed = int(gasUsedInt)
			}
		}
	}

	// 设置状态为成功（1表示成功）
	entry.Status = 1

	// 转换子调用
	if calls, ok := call["calls"].([]interface{}); ok {
		for i, callInterface := range calls {
			if callMap, ok := callInterface.(map[string]interface{}); ok {
				childPath := fmt.Sprintf("%s.%d", path, i)
				if childEntry, ok := s.convertCallToEntry(callMap, childPath, addresses); ok {
					entry.Children = append(entry.Children, childEntry)
				}
			}
		}
	}

	return entry, true
}

// stringPtr 返回字符串指针
func stringPtr(v string) *string {
	return &v
}

func (s *Service) serveStorage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chain := vars["chain"]
	address := vars["address"]
	codehash := vars["codehash"]

	_ = chain
	_ = address
	_ = codehash

	// 构造完整的 StorageResponse，包含所有必要字段
	response := client.StorageResponse{
		AllStructs: []interface{}{
			map[string]interface{}{
				"name": "ExampleStruct",
				"members": []interface{}{
					map[string]interface{}{
						"name": "value",
						"type": "uint256",
					},
				},
			},
		},
		Arrays: []interface{}{
			map[string]interface{}{
				"name": "ExampleArray",
				"type": "uint256[]",
			},
		},
		Structs: []interface{}{
			map[string]interface{}{
				"name": "ExampleStruct",
				"members": []interface{}{
					map[string]interface{}{
						"name": "value",
						"type": "uint256",
					},
				},
			},
		},
		Slots: map[string]client.SlotInfo{
			"0x0000000000000000000000000000000000000000000000000000000000000000": client.RawSlotInfo{
				BaseSlotInfo: client.BaseSlotInfo{
					Resolved: true,
					Variables: map[int]client.VariableInfo{
						0: {
							Name:     "value",
							FullName: "ExampleStruct.value",
							TypeName: client.TypeName{
								NodeType: "ElementaryTypeName",
								TypeDescriptions: client.TypeDescriptions{
									TypeIdentifier: "t_uint256",
									TypeString:     "uint256",
								},
							},
							Bits: 256,
						},
					},
				},
				Type: "raw",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"ok":     true,
		"error":  "",
		"result": response,
	})
}

func (s *Service) startServer() {
	m := mux.NewRouter()
	m.HandleFunc("/api/v1/trace/{chain}/{txhash}", s.serveTrace).Methods("GET")
	m.HandleFunc("/api/v1/storage/{chain}/{address}/{codehash}", s.serveStorage).Methods("GET")

	// 添加OPTIONS请求处理
	m.HandleFunc("/api/v1/trace/{chain}/{txhash}", s.handleOptions).Methods("OPTIONS")
	m.HandleFunc("/api/v1/storage/{chain}/{address}/{codehash}", s.handleOptions).Methods("OPTIONS")

	cors := handlers.CORS(
		handlers.AllowedMethods([]string{"OPTIONS", "HEAD", "GET", "POST"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		handlers.ExposedHeaders([]string{"Content-Length"}),
		handlers.AllowCredentials(),
	)(m)

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", s.config.HttpPort), cors); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Errorf("failed to listen and server")
		}
	}()
}

// handleOptions 处理OPTIONS预检请求
func (s *Service) handleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With, Authorization")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.WriteHeader(http.StatusOK)
}
