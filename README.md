# UBNT DisGOvery 🚀

A cross-platform, fast, and feature-rich network discovery tool for Ubiquiti devices, built with Go and Fyne. 

UBNT DisGOvery is a complete GUI replacement for the legacy Ubiquiti Discovery Tool, offering advanced features like automatic temporary IP assignment for accessing devices on different subnets, multi-interface scanning, and support for both Ubiquiti Discovery Protocol V1 and V2.

## ✨ Features

- 🔎 **Ubiquiti Device Discovery**: Supports both V1 and V2 discovery protocols.
- 🖥️ **Cross-Platform GUI**: Built with [Fyne](https://fyne.io/), providing a native and modern UI for Windows and Linux.
- 🌐 **Smart Web Access**: Double-click any discovered device to open its web interface. If the device is on a different subnet, UBNT DisGOvery can automatically assign a temporary IP to your network interface to enable direct access.
- 🔌 **Custom Port Support**: Specify custom HTTP and HTTPS ports for web access.
- 🌍 **Bilingual Interface**: Seamlessly switch between English and German.
- 🎨 **Theming**: Built-in support for both Light and Dark modes.
- 🛠 **Zero Dependencies**: Compiled as a single executable binary.

## 📥 Installation

Ensure you have Go 1.21+ installed on your system.

```bash
git clone https://github.com/ExtrudeEmpty/ubnt-disgovery.git
cd ubnt-disgovery
go mod tidy
go build -o ubnt-disgovery main.go
```

## 🚀 Usage

1. Launch the application.
2. Select the network interface you want to scan (or choose "All").
3. Select the discovery protocols (V1 and/or V2).
4. Click **Start** to begin scanning.
5. Double-click a device in the list to open its management UI in your default web browser.

## 🛠 Building for Windows (with icon)

UBNT DisGOvery includes a Windows manifest and icon (`app.ico`). To build a Windows executable with the icon:

```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o ubnt-disgovery.exe main.go
```

## 🙏 Acknowledgements

A special thanks goes to **Carlos Guerrero** and his open-source NodeJS project [`ubnt-discover`](https://github.com/guerrerocarlos/ubnt-discover). 
His work in analyzing the V1 and V2 discovery protocols served as the fundamental inspiration and technical blueprint for **UBNT DisGOvery**. While this tool is a complete rewrite in Go with a native desktop GUI, the core discovery logic builds upon the ideas first shared in his repository.

## 📄 License

This software is released under the **MIT License**. See the [LICENSE](LICENSE) file for more information.
