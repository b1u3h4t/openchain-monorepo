package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/openchainxyz/openchainxyz-monorepo/internal/compiler"
	"github.com/openchainxyz/openchainxyz-monorepo/services/vyper-compiler-srv/client"
	log "github.com/sirupsen/logrus"
)

func (s *Service) serveCompile(w http.ResponseWriter, r *http.Request) {
	var request client.CompileRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	vyper, err := compiler.NewVyperCompiler(request.Version)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	output, err := vyper.CompileFromString(request.Code)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	obj := output["/dev/stdin"]

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Type assertion for ABI
	abi, ok := obj.Info.AbiDefinition.([]any)
	if !ok {
		// If AbiDefinition is not []any, try to convert it
		if abiBytes, ok := obj.Info.AbiDefinition.([]byte); ok {
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

	json.NewEncoder(w).Encode(&client.CompileResponse{
		Status:          client.StatusSuccess,
		ABI:             abi,
		Bytecode:        obj.Code,
		BytecodeRuntime: obj.RuntimeCode,
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
