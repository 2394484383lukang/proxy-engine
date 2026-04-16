# Proxy Engine

一个用 Go 实现的开源代理引擎，类似 Clash 的规则化代理客户端。

## 功能特性

- 🚀 **高性能** - Go 语言实现，原生并发支持
- 🔌 **多协议** - 支持 SOCKS5、HTTP、Shadowsocks、VMess/VLESS、Trojan
- 📋 **规则引擎** - 域名/IP/GeoIP 分流，支持规则订阅
- 🔐 **DNS 分流** - 国内/海外 DNS 分流，防止泄漏
- 🌐 **TUN 模式** - 接管系统全部流量
- 🖥️ **跨平台** - macOS、Windows、Linux、Android、iOS

## 快速开始

### 构建

```bash
go build -o proxy-engine ./cmd/proxy-engine
```

### 运行

```bash
./proxy-engine -c config.yaml
```

### 配置示例

```yaml
port: 7890              # HTTP 代理端口
socks-port: 7891        # SOCKS5 代理端口
mode: rule              # 运行模式
log-level: info          # 日志级别
```

## 测试

```bash
go test ./... -v
```

## 开发状态

✅ Phase 1: 核心骨架 + SOCKS5/HTTP 入站（已完成）
⏳ Phase 2: Shadowsocks 协议支持（待开发）
⏳ Phase 3-9: 更多功能规划中

详细开发文档见 [CLAUDE.md](CLAUDE.md)

## License

MIT License
