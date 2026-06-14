# Darwin 检测复核

## 复核结论

上一轮结论总体合理。该技能不是不可用，而是处在“可用但需要增强稳定性和可维护性”的状态。按 Darwin 9 维 rubric dry-run 评估，当前约 82/100。

## 9 维评分（dry_run）

| 维度 | 权重 | 评分 | 加权 |
|---|---:|---:|---:|
| Frontmatter 质量 | 7 | 8 | 5.6 |
| 工作流清晰度 | 12 | 8 | 9.6 |
| 失败模式编码 | 12 | 8 | 9.6 |
| 检查点设计 | 6 | 9 | 5.4 |
| 可执行具体性 | 17 | 8 | 13.6 |
| 资源整合度 | 4 | 6 | 2.4 |
| 整体架构 | 12 | 7 | 8.4 |
| 实测表现 | 23 | 8 | 18.4 |
| 反例与黑名单 | 6 | 8 | 4.8 |
| **总分** | **100** |  | **77.8** |

注：这是保守 dry-run 分。若只按结构看，可到 80+；但 dim8 没有真实文档和服务端 API full_test，所以不能给过高。

## 结论逐条复核

1. `scripts/parse-response.cjs` 需要修：合理且可复现。v2 样例 `parseReturn/jobId/fileId` 输入后输出为空；v1 `parse_return/job_id/file_id` 可解析。
2. `SKILL.md` 需要瘦身：合理。文件 287 行，入口混合了 API/CLI 详细参考，和 skill-creator / Darwin 的渐进披露原则不完全匹配。
3. cleanup 字段需要确认并统一：合理。`references/api.md` 主体说 v2 camelCase，但 cleanup 示例和 Python CLI 仍用 `user_id`。
4. 补测试：合理。未发现 test/tests/*_test.go/test_*.py/package.json 等测试文件。
5. 二进制发布策略：合理。`cli/build` 为 80M，适合离线分发但需要源码/二进制同步策略。

## 验证

- Runtime 红灯扫描：未命中。
- Frontmatter 长度：839 字符，低于 Darwin 的 1024 字符线。
- `quick_validate.py .`：通过。
- `python3 cli/jklz-parse.py --help`：通过。
- `parse-response.cjs` v2 样例：失败，确认字段漂移问题。
- `parse-response.cjs` v1 样例：通过。
- `go test ./...`：未通过。错误同时包含沙箱不能写 `~/Library/Caches/go-build`，以及当前 Go/GOROOT 标准库解析异常。

## 需要修正上一轮措辞

上一轮说“约 80MB 多平台二进制”准确；说“Go 工具链疑似过旧”不够准确。当前 `go version` 是 1.24.3，不是版本过旧，更像 GOROOT/安装路径或沙箱环境异常。
