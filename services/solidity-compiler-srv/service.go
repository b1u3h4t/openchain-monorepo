package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/openchainxyz/openchainxyz-monorepo/internal/compiler"
	solidityclient "github.com/openchainxyz/openchainxyz-monorepo/services/solidity-compiler-srv/client"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	HttpPort int `def:"8082" env:"HTTP_PORT"`
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

func (s *Service) serveCompile(w http.ResponseWriter, r *http.Request) {
	var request solidityclient.CompileRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	solidity, err := compiler.NewSolidityCompiler(request.Version)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	// Get the first source file
	var sourceContent string
	for _, source := range request.Input.Sources {
		sourceContent = source.Content
		break
	}

	output, err := solidity.CompileFromString(sourceContent)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	// Get the first contract (assuming single contract compilation)
	var contract *compiler.Contract
	for _, c := range output {
		contract = c
		break
	}

	if contract == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "failed",
			"message": "no contract found in compilation output",
		})
		return
	}

	// Type assertion for ABI
	abi, ok := contract.Info.AbiDefinition.([]any)
	if !ok {
		// If AbiDefinition is not []any, try to convert it
		if abiBytes, ok := contract.Info.AbiDefinition.([]byte); ok {
			var abiData []any
			if err := json.Unmarshal(abiBytes, &abiData); err == nil {
				abi = abiData
			}
		}
		// If still not ok, use empty slice
		if !ok {
			abi = []any{}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&solidityclient.CompileResponse{
		Ok:    true,
		Error: "",
		Result: &solidityclient.SolcStandardOutput{
			Contracts: map[string]map[string]*solidityclient.SolcContract{
				"contract.sol": {
					"Contract": {
						ABI: abi,
						EVM: solidityclient.SolcEVM{
							Bytecode: solidityclient.SolcBytecode{
								Object: contract.Code,
							},
							DeployedBytecode: solidityclient.SolcBytecode{
								Object: contract.RuntimeCode,
							},
						},
					},
				},
			},
		},
	})
}

func (s *Service) startServer() {
	m := mux.NewRouter()
	m.HandleFunc("/v1/compile", s.serveCompile).Methods("POST")

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
