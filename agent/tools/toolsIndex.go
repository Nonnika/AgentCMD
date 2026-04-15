// This package is used for tools index
package tools

var ToolsIndex = []string{
	`{
		"type": "function",
		"function": {
			"name": "CreateFile",
			"description": "Create a file at the specified path.",
			"parameters": {
				"type": "object",
				"properties": {
					"fp": {
						"type": "string",
						"description": "The path to create the file with fileName."
					}
				},
				"required": ["fp"]
			}
		}
	}`,
}
