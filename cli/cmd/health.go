package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "检查服务健康状态",
	Long:  `检查解析 API 服务是否正常运行。`,
	RunE:  runHealth,
}

func init() {
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
	baseURL := getBaseURL()
	url := baseURL + "/metrics"

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("✗ 服务不可用: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if status, ok := result["status"].(string); ok && status == "success" {
		fmt.Printf("✓ 服务正常: %s\n", result["message"])
		fmt.Printf("  Base URL: %s\n", baseURL)
	} else {
		fmt.Printf("⚠ 服务异常: %v\n", result)
	}

	return nil
}
