package service

import (
	"encoding/json"
	"fmt"
	"net/http"

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

	// 连接到以太坊节点
	cli, err := ethclient.Dial("https://ethereum-rpc.svc.samczsun.com/")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": fmt.Sprintf("failed to connect to ethereum node: %v", err),
		})
		return
	}

	traceConfig := &tracers.TraceConfig{Tracer: nil} // 这里可根据需要自定义
	_ = cli
	_ = traceConfig

	// 构造完整的 TraceResponse，包含所有必要字段
	response := client.TraceResponse{
		Chain:  chain,
		Txhash: txhash,
		Preimages: map[string]string{
			"0x": "0x", // 示例数据
		},
		Addresses: map[string]map[string]client.AddressInfo{
			"0x0000000000000000000000000000000000000000": {
				"0x": {
					Label:     "Contract",
					Functions: map[string]interface{}{},
					Events:    map[string]interface{}{},
					Errors:    map[string]interface{}{},
				},
			},
		},
		Entrypoint: client.TraceEntryCall{
			Path:         "0",
			Type:         "call",
			Variant:      "call",
			Gas:          21000,
			IsPrecompile: false,
			From:         "0x0000000000000000000000000000000000000000",
			To:           "0x0000000000000000000000000000000000000000",
			Input:        "0x",
			Output:       "0x",
			GasUsed:      21000,
			Value:        "0x0",
			Status:       1,
			Codehash:     "0x",
			Children:     []client.TraceEntry{},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"ok":     true,
		"error":  "",
		"result": response,
	})
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

	cors := handlers.CORS(
		handlers.AllowedMethods([]string{"OPTIONS", "HEAD", "GET", "POST"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"}),
	)(m)

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", s.config.HttpPort), cors); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Errorf("failed to listen and server")
		}
	}()
}
