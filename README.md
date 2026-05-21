<div align="center">

# 📂 aifiler

> [!WARNING] **This repository is permanently archived.** It is no longer maintained and will not receive any further updates or bug fixes.

Your AI-powered local filesystem assistant. Instead of manual sorting and naming, simply describe your intent and let `aifiler` handle the planning, approval, and execution.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%7CmacOS%7CLinux-6E40C9)](#quick-start)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

![aifiler preview](preview.png)

</div>

---

## ✨ Features

- 🧠 **Dynamic Planning**: Translates natural language into structured operations.
- 🗂️ **Context Awareness**: Intelligently scans your workspace for relevant suggestions.
- ✅ **Safety First**: Every action is staged for your approval before execution.
- 🔌 **Provider Agnostic**: OpenAI, Anthropic, Gemini, Ollama, Vercel AI Gateway.
- 🎨 **Modern CLI**: Clean output with interactive menus and colorful feedback.

---

## 🚀 Quick Start

Run `aifiler` followed by your request in quotes. It scans the root directory for context.

```bash
aifiler "organize my images into folders by year"
```

Download the latest binary from [Releases](https://github.com/joshiminh/aifiler/releases). The app uses a `config.yaml` next to the executable.

### Installation

**Windows**
```batch
run.bat
```

**macOS / Linux**
```bash
go mod tidy
go build -o aifiler ./cmd/aifiler
./aifiler
```

## ⚙️ Configuration

Configure providers interactively or edit `config.yaml` directly:

```bash
aifiler provider
```

## 📄 License

Licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.