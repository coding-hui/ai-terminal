# AI-终端

[🇺🇸 English](./README.md) | [🇨🇳 中文](./README_zh.md) | [🇯🇵 日本語](./README_ja.md)

AI-终端是一款AI驱动的命令行工具，通过智能自动化和优化提升终端工作效率。

## 主要功能

- **智能助手：** 上下文感知的命令建议和补全
- **任务自动化：** 使用AI生成的快捷方式自动化重复任务
- **智能搜索：** 高级文件和内容搜索功能
- **错误处理：** 命令纠正和替代方案建议
- **可扩展性：** 支持自定义集成的插件系统

## 快速开始

### 环境要求

- Go 1.22.0 或更高版本

### 安装

**Homebrew:**
```bash
brew install coding-hui/tap/ai-terminal
```

**直接下载：**
- [软件包][releases] (Debian/RPM格式)
- [二进制文件][releases] (Linux/macOS/Windows)

[releases]: https://github.com/coding-hui/ai-terminal/releases

**源码编译：**
```bash
make build
```

**初始化配置：**
```bash
ai configure
```

### Shell 自动补全

包含 Bash、ZSH、Fish 和 PowerShell 的自动补全文件。手动生成：
```bash
ai completion [bash|zsh|fish|powershell] -h
```

## 使用示例

### 聊天与助手
```bash
ai ask "如何优化Docker性能？"
ai ask --file prompt.txt
echo "代码内容" | ai ask "分析这段代码"
```

### 代码生成
```bash
# 交互式模式
ai coder

# 批量处理
ai ctx load context.txt
ai coder -c 会话ID -p "添加错误处理"
```

### 代码审查
```bash
ai review --exclude-list "*.md,*.txt"
```

### 命令执行
```bash
ai exec "查找上周的大文件"
ai exec --yes "docker ps -a"
ai exec --interactive
```

### 提交信息
```bash
ai commit --diff-unified 3 --lang zh
```

## 贡献

查看 [CONTRIBUTING_zh.md](CONTRIBUTING_zh.md) 了解贡献指南。

**更新日志：** [CHANGELOG.md](CHANGELOG.md)  
**许可证：** [MIT](LICENSE) © 2024 coding-hui
