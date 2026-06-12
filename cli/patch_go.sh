sed -i '' '/func saveOutput(result map\[string\]interface{}, outputPath string) error {/,/^}/c\
func saveOutput(result map[string]interface{}, outputPath string) error {\
\tif result == nil || len(result) == 0 {\
\t\treturn fmt.Errorf("无有效结果可保存")\
\t}\
\
\ttypes := strings.Split(returnType, "#")\
\tvar outputs []string\
\n\tfor _, t := range types {\
\t\tswitch t {\
\t\tcase "content":\
\t\t\tif content, ok := result["content"].(string); ok {\
\t\t\t\toutputs = append(outputs, content)\
\t\t\t}\
\t\tcase "html":\
\t\t\tif html, ok := result["html"].(string); ok {\
\t\t\t\toutputs = append(outputs, html)\
\t\t\t}\
\t\tcase "toc", "table", "slice":\
\t\t\tif data, ok := result[t]; ok {\
\t\t\t\tjsonData, _ := json.MarshalIndent(data, "", "  ")\
\t\t\t\toutputs = append(outputs, string(jsonData))\
\t\t\t}\
\t\t}\
\t}\
\n\tvar finalContent string\
\tif len(outputs) > 0 {\
\t\tfinalContent = strings.Join(outputs, "\\n\\n")\
\t} else {\
\t\tjsonData, _ := json.MarshalIndent(result, "", "  ")\
\t\tfinalContent = string(jsonData)\
\t}\
\n\tif finalContent == "" {\
\t\treturn fmt.Errorf("内容为空")\
\t}\n\n\treturn os.WriteFile(outputPath, []byte(finalContent), 0644)\
}
