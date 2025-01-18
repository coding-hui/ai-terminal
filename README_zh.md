# AI-终端

AI-终端是一个先进的AI驱动的CLI，通过AI驱动的自动化和优化来增强终端工作流程。它有效地管理任务，如文件管理、数据处理和系统诊断。

## 主要特点

- **上下文帮助：** 从用户命令中学习，提供语法建议。
- **自动化任务：** 识别重复任务模式并创建快捷方式。
- **智能搜索：** 在文件、目录和特定文件类型中进行搜索。
- **错误纠正：** 纠正不正确的命令并提供替代方案。
- **自定义集成：** 通过插件或API支持与各种工具和服务集成。

## 入门

### 先决条件

- Go版本v1.22.0或更高。

### 安装

使用 Homebrew 安装：

```bash
brew install coding-hui/tap/ai-terminal
```

或者直接下载：

- [软件包][releases] 提供 Debian 和 RPM 格式
- [二进制文件][releases] 适用于 Linux、macOS 和 Windows

[releases]: https://github.com/coding-hui/ai-terminal/releases

或者从源码编译（需要 Go 1.22+）：

```sh
make build
```

然后初始化配置：
```sh
ai configure
```

<details>
<summary>Shell 自动补全</summary>

所有软件包和压缩包都包含预生成的 Bash、ZSH、Fish 和 PowerShell 的自动补全文件。

如果从源码构建，可以使用以下命令生成：

```bash
ai completion bash -h
ai completion zsh -h
ai completion fish -h
ai completion powershell -h
```

如果使用软件包（如 Homebrew、Debs 等），只要 shell 配置正确，自动补全应该会自动设置。

</details>

### 使用

以下是一些使用AI-终端的示例，按功能分组：

#### 聊天

- **启动聊天：**
  ```sh
  ai ask "管理Docker容器的最佳方式是什么？"
  ```

- **使用提示文件：**
  ```sh
  ai ask --file /path/to/prompt_file.txt
  ```

- **管道输入：**
  ```sh
  cat some_script.go | ai ask generate unit tests
  ```

#### 代码生成

- **自动生成代码：**
  ```sh
  ai coder
  ```
  这个命令启动自动编码功能，允许你根据提示交互式地生成代码。

- **带上下文生成代码：**
  ```sh
  ai ctx load /path/to/context_file
  ai coder "实现功能xxx"
  ```
  先加载上下文文件为代码生成提供额外信息。

#### 代码审查

- **审查代码更改：**
  ```sh
  ai review --exclude-list "*.md,*.txt"
  ```

#### 提交消息

- **生成提交消息：**
  ```sh
  ai commit --diff-unified 3 --lang zh
  ```

## 贡献

我们欢迎贡献！请参阅我们的[贡献指南](CONTRIBUTING_zh.md)以获取更多详细信息。

## 许可证

版权所有 2024 coding-hui。根据[MIT许可证](LICENSE)授权。
