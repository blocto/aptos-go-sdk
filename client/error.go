package client

import "fmt"

const ErrTableItemNotFound = "table_item_not_found"

type Error struct {
	StatusCode  int    `json:"status_code"`
	Message     string `json:"message"`
	ErrorCode   string `json:"error_code"`
	VMErrorCode int    `json:"vm_error_code"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode, e.Message)
}

func (e Error) IsTableItemNotFound() bool {
	return e.ErrorCode == ErrTableItemNotFound
}
