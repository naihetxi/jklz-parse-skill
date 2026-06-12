package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Add new API commands to root
func init() {
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(cancelCmd)
	rootCmd.AddCommand(modifyCmd)
	rootCmd.AddCommand(cleanupCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(exportCmd)

	// Common flags
	for _, c := range []*cobra.Command{getCmd, historyCmd, cancelCmd, modifyCmd, cleanupCmd, searchCmd, exportCmd} {
		c.Flags().StringVar(&baseURLFlag, "base-url", "", "Base URL")
	}

	// get
	getCmd.Flags().StringSliceP("return-types", "r", []string{"content"}, "返回类型列表 (如 content,html,toc)")

	// modify
	modifyCmd.Flags().StringP("chunk-id", "c", "", "Chunk ID")
	modifyCmd.Flags().StringP("text", "t", "", "修改后的文本内容")
	modifyCmd.MarkFlagRequired("chunk-id")
	modifyCmd.MarkFlagRequired("text")

	// search
	searchCmd.Flags().StringSliceP("keywords", "k", []string{}, "搜索关键词列表")
	searchCmd.MarkFlagRequired("keywords")

	// export
	exportCmd.Flags().StringP("type", "t", "md", "导出文件类型: md/html/docx/xlsx")
	exportCmd.Flags().StringP("output", "o", "", "输出文件路径")
	exportCmd.Flags().String("chunk-id", "", "只导出指定 chunkId")
	exportCmd.Flags().String("chunk-type", "", "只导出指定 chunkType")
}

// callJSONAPI is a helper for simple JSON POST APIs
func callJSONAPI(endpoint string, payload interface{}) error {
	baseURL := getBaseURL()
	url := baseURL + endpoint

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("API请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP错误 %d: %s", resp.StatusCode, string(body))
	}

	// Pretty print JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println(string(body)) // fallback to raw string
		return nil
	}

	pretty, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(pretty))
	return nil
}

func exportByID(baseURL, userID, jobID, fileID, fileType, outputPath, chunkID, chunkType string) error {
	if fileType == "" {
		fileType = "md"
	}
	switch fileType {
	case "md", "html", "docx", "xlsx":
	default:
		return fmt.Errorf("不支持的导出类型: %s，可选 md/html/docx/xlsx", fileType)
	}

	if outputPath == "" {
		outputPath = fmt.Sprintf("%s.%s", fileID, fileType)
	}

	url := baseURL + "/service/document/export/v2"
	payload := map[string]string{
		"userId":   userID,
		"jobId":    jobID,
		"fileId":   fileID,
		"fileType": fileType,
	}
	if chunkID != "" {
		payload["chunkId"] = chunkID
	}
	if chunkType != "" {
		payload["chunkType"] = chunkType
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

	client := &http.Client{Timeout: 60 * time.Second}
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

	codeVal := fmt.Sprintf("%v", exportRes.Code)
	if codeVal == "200.0" {
		codeVal = "200"
	}
	if codeVal != "200" {
		return fmt.Errorf("导出失败: code=%s, message=%s", codeVal, exportRes.Message)
	}
	if exportRes.Data.URL == "" {
		return fmt.Errorf("导出接口未返回下载链接")
	}

	fmt.Fprintf(os.Stderr, "正在下载文件: %s\n", exportRes.Data.URL)
	getResp, err := client.Get(exportRes.Data.URL)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载文件返回 HTTP %d", getResp.StatusCode)
	}

	if err := saveDownloadedExport(exportRes.Data.URL, getResp.Body, outputPath); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "✓ 已导出并保存到 %s\n", outputPath)
	return nil
}

func normalizeStringSlice(values []string) []string {
	var out []string
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				out = append(out, item)
			}
		}
	}
	return out
}

var getCmd = &cobra.Command{
	Use:   "get <userId> <jobId> <fileId>",
	Short: "获取指定解析任务的结果",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		returnTypes, _ := cmd.Flags().GetStringSlice("return-types")
		return callJSONAPI("/service/document/parse/get/v2", map[string]interface{}{
			"userId":         args[0],
			"jobId":          args[1],
			"fileId":         args[2],
			"returnTypeList": returnTypes,
		})
	},
}

var historyCmd = &cobra.Command{
	Use:   "history <userId>",
	Short: "查询历史解析记录",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return callJSONAPI("/service/document/parse/history/v2", map[string]interface{}{
			"userId": args[0],
		})
	},
}

var cancelCmd = &cobra.Command{
	Use:   "cancel <userId> <jobId>",
	Short: "停止正在运行的解析任务",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return callJSONAPI("/service/document/parse/cancel/v2", map[string]interface{}{
			"userId": args[0],
			"jobId":  args[1],
		})
	},
}

var modifyCmd = &cobra.Command{
	Use:   "modify <userId> <jobId> <fileId>",
	Short: "修改解析后的Chunk内容",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		chunkID, _ := cmd.Flags().GetString("chunk-id")
		text, _ := cmd.Flags().GetString("text")
		return callJSONAPI("/service/document/parse/modify/v2", map[string]interface{}{
			"userId":  args[0],
			"jobId":   args[1],
			"fileId":  args[2],
			"chunkId": chunkID,
			"text":    text,
		})
	},
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup <userId> <time>",
	Short: "清理历史解析文件 (time格式如: 7d, 24h, 30m)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return callJSONAPI("/service/document/parse/cleanup/v2", map[string]interface{}{
			"user_id": args[0],
			"time":    args[1],
		})
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <userId> <jobId> <fileId>",
	Short: "在解析结果中搜索关键词",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		keywords, _ := cmd.Flags().GetStringSlice("keywords")
		keywords = normalizeStringSlice(keywords)
		return callJSONAPI("/service/document/search/v2", map[string]interface{}{
			"userId":   args[0],
			"jobId":    args[1],
			"fileId":   args[2],
			"keywords": keywords,
		})
	},
}

var exportCmd = &cobra.Command{
	Use:   "export <userId> <jobId> <fileId>",
	Short: "导出已解析结果为文件",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		fileType, _ := cmd.Flags().GetString("type")
		outputPath, _ := cmd.Flags().GetString("output")
		chunkID, _ := cmd.Flags().GetString("chunk-id")
		chunkType, _ := cmd.Flags().GetString("chunk-type")
		return exportByID(getBaseURL(), args[0], args[1], args[2], fileType, outputPath, chunkID, chunkType)
	},
}
