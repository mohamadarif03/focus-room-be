package utils

import (
	"encoding/json"
	"net/http"
)

type Meta struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type APIResponse struct {
	Meta Meta        `json:"meta"`
	Data interface{} `json:"data"`
}

func Success(w http.ResponseWriter, data interface{}, message string, code int) {
	if code == 0 {
		code = http.StatusOK
	}

	response := APIResponse{
		Meta: Meta{
			Code:    code,
			Status:  "success",
			Message: message,
		},
		Data: data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

func Error(w http.ResponseWriter, data interface{}, message string, code int) {
	if code == 0 {
		code = http.StatusBadRequest
	}

	response := APIResponse{
		Meta: Meta{
			Code:    code,
			Status:  "error",
			Message: message,
		},
		Data: data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}
