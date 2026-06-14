# jklz-parse-skill 优化分析

## 结论

技能可用，基础校验通过，但仍建议优化。优先级最高的问题是 Agent 入口文档偏重且和 API/CLI 辅助脚本存在命名漂移，容易在直接 API 调用或解析 SSE 响应时走错字段。

## Critical

无阻断级问题。`SKILL.md` frontmatter 可通过 quick_validate，Python CLI help 和 Node 脚本空输入均可运行。

## Warning

1. `scripts/parse-response.cjs` 只识别 `parse_return`、`job_id`、`file_id`，但 v2 文档和 CLI 主实现都使用/兼容 `parseReturn`、`jobId`、`fileId`。这会导致 Agent 用 curl 直接调用 v2 后再 pipe 到该脚本时抽不到结果。
2. `SKILL.md` 287 行，承载了架构说明、配置、工作流、return 参数、完整 CLI 命令、导出说明、评分机制和黑名单。按 skill-creator 原则，入口应更瘦，详细 API/CLI 参考应下沉到 `references/api.md` / CLI README，并在 SKILL.md 中明确何时读取。
3. `references/api.md` 标注 v2 优先，但 cleanup 示例仍使用 `user_id`，Python CLI 也发送 `user_id`。如果 v2 服务严格使用 camelCase，这里可能不一致；至少需要确认服务端兼容性并统一文档。
4. 仓库未发现测试文件。对于流式 JSON 解析、SSE 解析、导出 zip 解包、安装脚本平台选择等逻辑，缺少回归保护。
5. `cli/build/` 提交了约 80MB 多平台二进制。它有离线分发价值，但需要明确发布策略，否则源码变更和二进制容易不同步。

## Info

- `agents/openai.yaml` 与能力描述基本一致，但 default_prompt 偏泛，可加上“输出 Markdown/表格/目录”的默认意图。
- `SKILL.md` 的触发描述覆盖面较广，适合召回，但可以减少营销式能力列表，保留具体触发场景和边界。
- 当前 `.gitignore` 已有未提交修改，本次分析未触碰。

## 验证

- `python3 /Users/lizhi/.claude/skills/.system/skill-creator/scripts/quick_validate.py .`：通过。
- `python3 cli/jklz-parse.py --help`：通过。
- `node scripts/parse-response.cjs < /dev/null`：通过。
- `go test ./...`：未通过。当前沙箱不能写 `~/Library/Caches/go-build`，并且 Go 报告标准库路径异常如 `archive/zip is not in std`、`log/slog is not in std`，需要在干净 Go 环境重跑确认。

## 外部模型

按 CCG 要求尝试双模型分析。Claude 首次因沙箱禁止绑定临时端口失败，提权重跑后仍退出且无可读日志。Gemini 长时间无输出，已尝试结束 wrapper 进程。最终结论以本地源码、技能规范和验证结果为准。
