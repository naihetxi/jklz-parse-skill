package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	returnType  string
	imageMode   string
	pageRange   string
	output      string
	apiKeyFlag  string
	baseURLFlag string
)

var parseCmd = &cobra.Command{
	Use:   "parse <file>",
	Short: "解析文档",
	Long: `解析 PDF、Word、Excel 等文档，提取文本、表格等内容。

示例：
  jklz-parse parse document.pdf
  jklz-parse parse report.pdf --return content#toc#table
  jklz-parse parse data.xlsx --return table
  jklz-parse parse doc.pdf --output result.md
  jklz-parse parse large.pdf --page-range "1-5,10"`,
	Args: cobra.ExactArgs(1),
	RunE: runParse,
}

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().StringVar(&returnType, "return", "content", "返回类型：content/html/toc/table/slice (可用#分隔)")
	parseCmd.Flags().StringVar(&imageMode, "image-mode", "cv", "图像解析模式：vl(高精度) 或 cv(高性能)")
	parseCmd.Flags().StringVar(&pageRange, "page-range", "", "页面范围，如 \"1-5,10\"")
	parseCmd.Flags().StringVarP(&output, "output", "o", "", "输出文件路径（支持 .md/.html/.docx/.xlsx）")
	parseCmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "API Key（覆盖配置）")
	parseCmd.Flags().StringVar(&baseURLFlag, "base-url", "", "Base URL（覆盖配置）")
}

func runParse(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}

	// 获取配置
	apiKey := getAPIKey()
	apiBaseURL := getBaseURL()

	if apiKey == "" {
		return fmt.Errorf("未配置 API Key\n请运行：jklz-parse config --api-key YOUR_KEY")
	}

	// 检查文件大小
	fileInfo, _ := os.Stat(filePath)
	if fileInfo.Size() > 200*1024*1024 {
		fmt.Fprintf(os.Stderr, "⚠️  文件较大 (%.1f MB)，建议使用 --return slice\n", float64(fileInfo.Size())/1024/1024)
	}

	fmt.Fprintf(os.Stderr, "正在解析 %s...\n", filepath.Base(filePath))

	// 调用 API
	result, err := callParseAPI(apiBaseURL, apiKey, filePath)
	if err != nil {
		// 增强鲁棒性：遇到 502/503 等服务侧异常时，等待 5 秒重试一次，而非立刻失败
		if strings.Contains(err.Error(), "502") || strings.Contains(err.Error(), "503") {
			fmt.Fprintf(os.Stderr, "服务暂时不可用或触发限流 (502/503)，等待 5 秒后自动重试...\n")
			time.Sleep(5 * time.Second)
			result, err = callParseAPI(apiBaseURL, apiKey, filePath)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// 输出结果
	if output != "" {
		ext := strings.ToLower(filepath.Ext(output))
		var fileType string
		switch ext {
		case ".docx":
			fileType = "docx"
		case ".xlsx":
			fileType = "xlsx"
		case ".html":
			fileType = "html"
		case ".md":
			fileType = "md"
		}

		if fileType != "" {
			fmt.Fprintf(os.Stderr, "检测到输出格式 %s，正在请求服务导出格式化文件...\n", fileType)
			if err := exportAndSave(apiBaseURL, result, fileType, output); err != nil {
				return fmt.Errorf("导出文件失败: %w", err)
			}
		} else {
			if err := saveOutput(result, output); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "✓ 已保存到 %s\n", output)
		}
	} else {
		printResult(result)
	}

	return nil
}

func callParseAPI(baseURL, apiKey, filePath string) (map[string]interface{}, error) {
	url := baseURL + "/service/document/parse/stream/v2"

	// 创建 multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加表单字段
	writer.WriteField("apiKey", apiKey)
	writer.WriteField("streamType", "lz")
	writer.WriteField("return", returnType)
	writer.WriteField("imageParseMode", imageMode)

	if pageRange != "" {
		writer.WriteField("pageSelect", pageRange)
	}

	// 添加文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("创建表单文件失败: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("复制文件失败: %w", err)
	}

	writer.Close()

	// 发送请求
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 不设置总超时，只设置连接超时
	client := &http.Client{
		Timeout: 0, // 无限制，因为流式响应可能很长
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 错误 %d", resp.StatusCode)
	}

	// 解析 JSON Stream 响应
	result, err := parseJSONStream(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("API 返回了空结果")
	}

	return result, nil
}

// parseJSONStream statefully parses a stream of concatenated JSON objects.
// It handles streams where multiple JSON objects appear concatenated on a
// single line (e.g. {"code":"200",...}{"code":"200",...}) and strips any
// non-JSON prefixes such as "data: ".
func parseJSONStream(body io.Reader) (map[string]interface{}, error) {
	buf := make([]byte, 4096)
	var jsonBuf bytes.Buffer
	inString := false
	isEscaped := false
	depth := 0
	started := false
	var result map[string]interface{}

	for {
		n, err := body.Read(buf)
		if n > 0 {
			for i := 0; i < n; i++ {
				b := buf[i]
				if !started {
					if b == '{' {
						started = true
						depth = 1
						jsonBuf.WriteByte(b)
					}
					continue
				}

				jsonBuf.WriteByte(b)

				if isEscaped {
					isEscaped = false
					continue
				}

				if b == '\\' {
					isEscaped = true
					continue
				}

				if b == '"' {
					inString = !inString
					continue
				}

				if !inString {
					if b == '{' {
						depth++
					} else if b == '}' {
						depth--
						if depth == 0 {
							// Complete JSON object found.
							var data map[string]interface{}
							if err := json.Unmarshal(jsonBuf.Bytes(), &data); err != nil {
								fmt.Fprintf(os.Stderr, "[Debug] 跳过无效 JSON: %v\n", err)
							} else {
								if data["code"] == "200" {
									dataObj, ok := data["data"].(map[string]interface{})
									if !ok {
										continue
									}
									dataType, _ := dataObj["type"].(string)
									switch dataType {
									case "parseReturn", "parse_return":
										value, ok := dataObj["value"].(map[string]interface{})
										if ok {
											result = value
										}
									case "error", "fatal":
										value, _ := dataObj["value"].(map[string]interface{})
										errorMsg, _ := value["error"].(string)
										return nil, fmt.Errorf("解析错误: %s", errorMsg)
									case "stop":
										return result, nil
									}
								}
							}
							jsonBuf.Reset()
							started = false
						}
					}
				}
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}
	}

	return result, nil
}

func printResult(result map[string]interface{}) {
	if result == nil || len(result) == 0 {
		fmt.Fprintln(os.Stderr, "⚠️  未收到解析结果")
		return
	}

	types := strings.Split(returnType, "#")

	for _, t := range types {
		switch t {
		case "content":
			if content, ok := result["content"].(string); ok {
				fmt.Println(content)
			} else {
				fmt.Fprintf(os.Stderr, "⚠️  未找到 content 字段\n")
			}
		case "html":
			if html, ok := result["html"].(string); ok {
				fmt.Println(html)
			} else {
				fmt.Fprintf(os.Stderr, "⚠️  未找到 html 字段\n")
			}
		case "toc", "table", "slice":
			if data, ok := result[t]; ok {
				jsonData, _ := json.MarshalIndent(data, "", "  ")
				fmt.Println(string(jsonData))
			} else {
				fmt.Fprintf(os.Stderr, "⚠️  未找到 %s 字段\n", t)
			}
		}
	}
}

func saveOutput(result map[string]interface{}, outputPath string) error {
	if result == nil || len(result) == 0 {
		return fmt.Errorf("无有效结果可保存")
	}

	types := strings.Split(returnType, "#")
	var outputs []string

	for _, t := range types {
		switch t {
		case "content":
			if content, ok := result["content"].(string); ok {
				outputs = append(outputs, content)
			}
		case "html":
			if html, ok := result["html"].(string); ok {
				outputs = append(outputs, html)
			}
		case "toc", "table", "slice":
			if data, ok := result[t]; ok {
				jsonData, _ := json.MarshalIndent(data, "", "  ")
				outputs = append(outputs, string(jsonData))
			}
		}
	}

	var finalContent string
	if len(outputs) > 0 {
		finalContent = strings.Join(outputs, "\n\n")
	} else {
		jsonData, _ := json.MarshalIndent(result, "", "  ")
		finalContent = string(jsonData)
	}

	if finalContent == "" {
		return fmt.Errorf("内容为空")
	}

	return os.WriteFile(outputPath, []byte(finalContent), 0644)
}

func getAPIKey() string {
	if apiKeyFlag != "" {
		return apiKeyFlag
	}

	if key := os.Getenv("JKLZ_PARSE_APIKEY"); key != "" {
		return key
	}

	if key := viper.GetString("api_key"); key != "" {
		return key
	}

	return ""
}

func getBaseURL() string {
	if baseURLFlag != "" {
		return baseURLFlag
	}

	if url := os.Getenv("JKLZ_PARSE_BASEURL"); url != "" {
		return url
	}

	if url := viper.GetString("base_url"); url != "" {
		return url
	}

	return "http://192.168.42.15:15216"
}

func exportAndSave(baseURL string, result map[string]interface{}, fileType, outputPath string) error {
	var userId, jobId, fileId string
	if val, ok := result["userId"].(string); ok {
		userId = val
	} else if val, ok := result["user_id"].(string); ok {
		userId = val
	}
	if val, ok := result["jobId"].(string); ok {
		jobId = val
	} else if val, ok := result["job_id"].(string); ok {
		jobId = val
	}
	if val, ok := result["fileId"].(string); ok {
		fileId = val
	} else if val, ok := result["file_id"].(string); ok {
		fileId = val
	}

	if userId == "" || jobId == "" || fileId == "" {
		return fmt.Errorf("无法进行导出：未在解析结果中找到有效的 userId/jobId/fileId (userId: %q, jobId: %q, fileId: %q)", userId, jobId, fileId)
	}

	url := baseURL + "/service/document/export/v2"

	payload := map[string]string{
		"userId":   userId,
		"jobId":    jobId,
		"fileId":   fileId,
		"fileType": fileType,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建导出请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求导出接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("导出接口返回 HTTP %d: %s", resp.StatusCode, string(errBody))
	}

	var exportRes struct {
		Code    interface{} `json:"code"`
		Message string      `json:"message"`
		Data    struct {
			URL  string `json:"url"`
			Note string `json:"note"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&exportRes); err != nil {
		return fmt.Errorf("解析导出响应失败: %w", err)
	}

	var codeVal string
	switch v := exportRes.Code.(type) {
	case float64:
		codeVal = fmt.Sprintf("%.0f", v)
	case int:
		codeVal = fmt.Sprintf("%d", v)
	case string:
		codeVal = v
	}

	if codeVal != "200" {
		return fmt.Errorf("导出失败: code=%s, message=%s", codeVal, exportRes.Message)
	}

	downloadURL := exportRes.Data.URL
	if downloadURL == "" {
		return fmt.Errorf("导出接口未返回下载链接")
	}

	fmt.Fprintf(os.Stderr, "正在下载文件: %s\n", downloadURL)
	getReq, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("创建下载请求失败: %w", err)
	}

	getResp, err := client.Do(getReq)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载文件返回 HTTP %d", getResp.StatusCode)
	}

	outDir := filepath.Dir(outputPath)
	if outDir != "" {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("创建输出目录失败: %w", err)
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建本地输出文件失败: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, getResp.Body)
	if err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}

	return nil
}
