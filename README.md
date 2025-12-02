# SysChk - 简易系统状态检查工具

SysChk 是一个轻量级的 Linux 系统状态检查工具，使用 Go 语言编写。它完全基于 Go 标准库实现，不依赖任何第三方库，通过读取 `/proc` 文件系统和系统调用获取信息。

## ✨ 功能特性

- **零依赖**：仅使用 Go 标准库，编译产物单一，易于分发。
- **系统概况**：显示主机名、操作系统版本、内核版本、启动时间和运行时间。
- **CPU 状态**：显示 CPU 型号、核心数以及实时使用率。
- **内存状态**：显示总内存大小及使用率进度条。
- **磁盘空间**：监控根目录 (`/`) 的磁盘使用情况。
- **网络监控**：列出当前正在监听的 TCP 端口。

## 🚀 快速开始

### 下载二进制文件

请前往 [Releases](../../releases) 页面下载对应架构（amd64 或 arm64）的最新版本。

```bash
# 赋予执行权限
chmod +x syschk-linux-amd64

# 运行
./syschk-linux-amd64
```

### 手动编译

如果你安装了 Go 环境 (1.21+)，可以手动编译：

```bash
# 克隆代码
git clone https://github.com/your-username/syschk.git
cd syschk

# 编译
go build -ldflags="-s -w" -o syschk main.go

# 运行
./syschk
```
