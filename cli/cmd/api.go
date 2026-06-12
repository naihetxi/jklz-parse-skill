package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	// Common flags
	for _, c := range []*cobra.Command{getCmd, historyCmd, cancelCmd, modifyCmd, cleanupCmd, searchCmd} {
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
		return callJSONAPI("/service/document/search/v2", map[string]interface{}{
			"userId":   args[0],
			"jobId":    args[1],
			"fileId":   args[2],
			"keywords": keywords,
		})
	},
}
