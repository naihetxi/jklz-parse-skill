package cmd

import (
	"bufio"
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
	returnType   string
	imageMode    string
	pageRange    string
	output       string
	tableFormat  string
	apiKeyFlag   string
	baseURLFlag  string
)

var parseCmd = &cobra.Command{
	Use:   "parse <file>",
	Short: "解析文档",
	Long: `解析 PDF、Word、Excel 等文档，提取文本、表格等内容。

示例：
  jklz-parse parse document.pdf
  jklz-parse parse report.pdf --return content#toc#table
  jklz-parse parse data.xlsx --return table --table-format markdown
  jklz-parse parse doc.pdf --output result.md
  jklz-parse parse large.pdf --page-range "1-5,10"`,
	Args: cobra.ExactArgs(1),
	RunE: runParse,
}

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().StringVar(&returnType, "return", "content", "返回类型：content/html/toc/table/slice (可用#分隔)")
	parseCmd.Flags().StringVar(&imageMode, "image-mode", "vl", "图像解析模式：vl(高精度) 或 cv(高性能)")
	parseCmd.Flags().StringVar(&pageRange, "page-range", "", "页面范围，如 \"1-5,10\"")
	parseCmd.Flags().StringVarP(&output, "output", "o", "", "输出文件路径")
	parseCmd.Flags().StringVar(&tableFormat, "table-format", "markdown", "表格格式：html 或 markdown")
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
	baseURL := getBaseURL()

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
	result, err := callParseAPI(baseURL, apiKey, filePath)
	if err != nil {
		// 如果是 502/503 错误且使用 vl 模式，尝试切换到 cv 模式
		if strings.Contains(err.Error(), "502") || strings.Contains(err.Error(), "503") {
			if imageMode == "vl" {
				fmt.Fprintf(os.Stderr, "服务暂时不可用，尝试切换到 CV 模式...\n")
				imageMode = "cv"
				result, err = callParseAPI(baseURL, apiKey, filePath)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			return err
		}
	}

	// 输出结果
	if output != "" {
		if err := saveOutput(result, output); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "✓ 已保存到 %s\n", output)
	} else {
		printResult(result)
	}

	return nil
}

func callParseAPI(baseURL, apiKey, filePath string) (map[string]interface{}, error) {
	url := baseURL + "/service/document/parse/stream/v1"

	// 创建 multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加表单字段
	writer.WriteField("api_key", apiKey)
	writer.WriteField("stream_type", "lz")
	writer.WriteField("return", returnType)
	writer.WriteField("image_parse_mode", imageMode)

	if tableFormat != "" {
		writer.WriteField("table_format", tableFormat)
	}

	if pageRange != "" {
		writer.WriteField("page_selecte2parse", pageRange)
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

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 错误 %d", resp.StatusCode)
	}

	// 解析 SSE 响应
	return parseSSEResponse(resp.Body)
}

func parseSSEResponse(body io.Reader) (map[string]interface{}, error) {
	scanner := bufio.NewScanner(body)
	var result map[string]interface{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if !strings.HasPrefix(line, "data:") {
			continue
		}

		jsonStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			continue
		}

		if data["code"] == "200" {
			dataObj, ok := data["data"].(map[string]interface{})
			if !ok {
				continue
			}

			dataType, _ := dataObj["type"].(string)

			switch dataType {
			case "parse_return":
				value, ok := dataObj["value"].(map[string]interface{})
				if ok {
					result = value
				}
			case "error":
				value, _ := dataObj["value"].(map[string]interface{})
				errorMsg, _ := value["error"].(string)
				return nil, fmt.Errorf("解析错误: %s", errorMsg)
			case "stop":
				return result, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return result, nil
}

func printResult(result map[string]interface{}) {
	types := strings.Split(returnType, "#")

	for _, t := range types {
		switch t {
		case "content":
			if content, ok := result["content"].(string); ok {
				fmt.Println(content)
			}
		case "html":
			if html, ok := result["html"].(string); ok {
				fmt.Println(html)
			}
		case "toc", "table", "slice":
			if data, ok := result[t]; ok {
				jsonData, _ := json.MarshalIndent(data, "", "  ")
				fmt.Println(string(jsonData))
			}
		}
	}
}

func saveOutput(result map[string]interface{}, outputPath string) error {
	var content string

	if strings.Contains(returnType, "content") {
		if c, ok := result["content"].(string); ok {
			content = c
		}
	} else if strings.Contains(returnType, "html") {
		if h, ok := result["html"].(string); ok {
			content = h
		}
	} else {
		jsonData, _ := json.MarshalIndent(result, "", "  ")
		content = string(jsonData)
	}

	return os.WriteFile(outputPath, []byte(content), 0644)
}

func getAPIKey() string {
	if apiKeyFlag != "" {
		return apiKeyFlag
	}

	if key := viper.GetString("api_key"); key != "" {
		return key
	}

	if key := os.Getenv("JKLZ_PARSE_APIKEY"); key != "" {
		return key
	}

	return ""
}

func getBaseURL() string {
	if baseURLFlag != "" {
		return baseURLFlag
	}

	if url := viper.GetString("base_url"); url != "" {
		return url
	}

	if url := os.Getenv("JKLZ_PARSE_BASEURL"); url != "" {
		return url
	}

	return "http://192.168.42.15:15216"
}
