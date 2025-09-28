package tools

import (
	"uf/mcp/pkg/common"

	"github.com/tmc/langchaingo/llms/openai"
)

var (
	model *openai.LLM
)

func init() {
	model = common.GetModel()
}
