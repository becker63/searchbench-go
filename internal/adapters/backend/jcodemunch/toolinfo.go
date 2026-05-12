package jcodemunch

import (
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolInfoFromMCP converts an MCP tool definition into Eino [schema.ToolInfo].
func ToolInfoFromMCP(t *mcp.Tool) (*schema.ToolInfo, error) {
	if t == nil {
		return nil, fmt.Errorf("nil tool")
	}
	params, err := paramsFromInputSchema(t.InputSchema)
	if err != nil {
		return nil, fmt.Errorf("tool %q input schema: %w", t.Name, err)
	}
	desc := t.Description
	return &schema.ToolInfo{
		Name:        t.Name,
		Desc:        desc,
		ParamsOneOf: params,
	}, nil
}

func paramsFromInputSchema(input any) (*schema.ParamsOneOf, error) {
	if input == nil {
		return schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}), nil
	}
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	var js jsonschema.Schema
	if err := json.Unmarshal(raw, &js); err != nil {
		// Last resort: accept arbitrary JSON object arguments as a single string field.
		// This keeps marginal servers usable even when their schema is not draft 2020-12 compatible.
		return schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"arguments": {
				Desc: "Opaque JSON object arguments for the tool.",
				Type: schema.String,
			},
		}), nil
	}
	return schema.NewParamsOneOfByJSONSchema(&js), nil
}
