package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	showConfig bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置 API 凭证",
	Long: `配置 API Key 和 Base URL。

配置会保存到 ~/.config/jklz-parse/ 目录。

示例：
  jklz-parse config --api-key YOUR_API_KEY
  jklz-parse config --base-url http://192.168.42.15:15216
  jklz-parse config --show`,
	RunE: runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "设置 API Key")
	configCmd.Flags().StringVar(&baseURLFlag, "base-url", "", "设置 Base URL")
	configCmd.Flags().BoolVar(&showConfig, "show", false, "显示当前配置")

	// 初始化 viper
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	configDir := filepath.Join(home, ".config", "jklz-parse")
	os.MkdirAll(configDir, 0755)

	viper.AddConfigPath(configDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.ReadInConfig()
}

func runConfig(cmd *cobra.Command, args []string) error {
	if showConfig {
		fmt.Println("当前配置：")

		apiKey := viper.GetString("api_key")
		if apiKey != "" {
			fmt.Printf("  API Key: %s...\n", apiKey[:min(10, len(apiKey))])
		} else {
			fmt.Println("  API Key: 未配置")
		}

		baseURL := viper.GetString("base_url")
		if baseURL != "" {
			fmt.Printf("  Base URL: %s\n", baseURL)
		} else {
			fmt.Println("  Base URL: 未配置（将使用默认值）")
		}

		home, _ := os.UserHomeDir()
		configFile := filepath.Join(home, ".config", "jklz-parse", "config.yaml")
		fmt.Printf("\n配置文件: %s\n", configFile)

		return nil
	}

	if apiKeyFlag != "" {
		viper.Set("api_key", apiKeyFlag)
		fmt.Println("✓ API Key 已保存")
	}

	if baseURLFlag != "" {
		viper.Set("base_url", baseURLFlag)
		fmt.Println("✓ Base URL 已保存")
	}

	if apiKeyFlag != "" || baseURLFlag != "" {
		if err := viper.WriteConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				if err := viper.SafeWriteConfig(); err != nil {
					return fmt.Errorf("保存配置失败: %w", err)
				}
			} else {
				return fmt.Errorf("保存配置失败: %w", err)
			}
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
