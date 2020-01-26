package common

import (
	"fmt"

	"github.com/root-gg/utils"
)

// Result object
type Result struct {
	Message string      `json:"message"`
	Value   interface{} `json:"value"`
}

// NewResult create a new Result instance
func NewResult(message string, value interface{}) (r *Result) {
	r = new(Result)
	r.Message = message
	r.Value = value
	return
}

// ToJSON serialize result object to JSON
func (result *Result) ToJSON() []byte {
	j, err := utils.ToJson(result)
	if err != nil {
		msg := fmt.Sprintf("Unable to serialize result %s to json : %s", result.Message, err)
		return []byte("{message:\"" + msg + "\"}")
	}

	return j
}

// ToJSONString is the same as ToJson but it returns
// a string instead of a byte array
func (result *Result) ToJSONString() string {
	return string(result.ToJSON())
}
