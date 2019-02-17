package gokasa

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

type KasaError struct {
	Code    int    `json:"err_code"`
	Message string `json:"err_msg"`
}

func (e KasaError) Error() string {
	return fmt.Sprintf("Kasa Error(%d): %s", e.Code, e.Message)
}

type KasaResponse struct {
	Results map[string]KasaError `json:"system"`
}

func (kr *KasaResponse) GetErrors() error {
	var err *multierror.Error
	for _, v := range kr.Results {
		if v.Code != 0 {
			err = multierror.Append(err, v)
		}
	}

	return err.ErrorOrNil()
}
