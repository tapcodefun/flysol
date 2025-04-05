# README

## 配置

你可以通过编辑 `wails.json` 来配置项目。更多关于项目设置的信息可以在这里找到：
https://wails.io/docs/reference/project-config

## 开发

要在实时开发模式下运行，请在项目目录中运行 `wails dev`。这将启动一个 Vite 开发服务器，提供非常快速的前端热重载。如果你想在浏览器中开发并访问你的 Go 方法，还有一个运行在 http://localhost:34115 的开发服务器。在浏览器中连接到此地址，你就可以从开发者工具中调用你的 Go 代码。

## 构建

要构建可重新分发的生产模式包，请使用 `wails build`。

## 命令
#### 打包命令
```
wails build -p darwin/amd64  # 针对 Intel Mac
wails build -p darwin/arm64  # 针对 Apple Silicon Mac
```
#### go环境异常处理
```
export PATH=$PATH:$(go env GOPATH)/bin
```
