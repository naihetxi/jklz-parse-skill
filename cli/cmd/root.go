package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

var rootCmd = &cobra.Command{
	Use:   "jklz-parse",
	Short: "金科览智文档解析命令行工具",
	Long: `jklz-parse-cli - 金科览智文档解析工具

支持解析 PDF、DOC、DOCX、XLSX、PPT 等格式，
提取文本、表格、目录等结构化信息。`,
	Version: version,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("jklz-parse-cli %s\nCommit: %s\nBuilt: %s\n", version, commit, date))
}
