# Proxy Engine

一个类似 Clash 的规则化代理客户端，用 Go 语言实现的开源项目。

## 项目概述

**目标：** 构建跨平台代理引擎（桌面 + 移动端），支持多种代理协议和高级规则系统。

**架构：** Go 核心引擎 + 各平台原生 UI 壳

**当前状态：** Phase 1 已完成（核心骨架 + SOCKS5/HTTP 入站）

## 技术栈

- **语言：** Go 1.21+
- **依赖：**
  - `gopkg.in/yaml.v3` - YAML 配置解析
  - `github.com/stretchr/testify` - 测试断言

## 项目结构

```
proxy-engine/
├── cmd/proxy-engine/    # CLI 入口
├── internal/
│   ├── proxy/          # 代理协议实现
│   │   ├── proxy.go    # Proxy 接口定义
│   │   ├── direct.go   # DIRECT 出站（直连）
│   │   ├── reject.go   # REJECT 出站（拒绝）
│   │   ├── socks5/     # SOCKS5 入站服务器
│   │   └── http/       # HTTP CONNECT 入站
│   ├── hub/            # 连接调度器（核心分发）
│   ├── config/         # YAML 配置管理
│   └── api/            # REST API 服务器
└── docs/superpowers/  # 设计文档和实现计划
```

## 核心设计

### Proxy 接口

所有出站代理实现统一的 `Proxy` 接口：

```go
type Proxy interface {
    Dial(ctx context.Context, target string) (net.Conn, error)
    Type() ProxyType
}
```

### Hub 调度器

Hub 是中央调度器，负责：
1. 接收入站连接（来自 SOCKS5/HTTP 服务器）
2. 匹配规则确定目标策略
3. 选择对应的出站代理
4. 建立连接并双向转发数据

### 配置格式

兼容 Clash YAML 格式：

```yaml
port: 7890              # HTTP 代理端口
socks-port: 7891        # SOCKS5 代理端口
mixed-port: 7892        # 混合端口
mode: rule              # rule / global / direct
log-level: info          # debug / info / warning / error

# 未来将添加：
# dns, tun, proxies, proxy-groups, rules, rule-providers
```

## 开发模式

**TDD 驱动：** 所有功能先写测试，再实现代码

**提交规范：**
- `feat:` 新功能
- `fix:` Bug 修复
- `test:` 测试相关
- `refactor:` 重构

**构建命令：**
```bash
# 构建
go build -o proxy-engine ./cmd/proxy-engine

# 运行
./proxy-engine -c config.yaml

# 测试
go test ./... -v
```

## 已完成功能（Phase 1）

✅ Proxy 接口 + DIRECT/REJECT 出站
✅ SOCKS5 入站服务器（RFC 1928 子集）
✅ HTTP CONNECT 入站服务器
✅ Hub 连接调度器（双向数据转发）
✅ YAML 配置加载器
✅ REST API 服务器
✅ 端到端集成测试（SOCKS5）

## 待实现功能

### Phase 2: Shadowsocks 协议支持
- [ ] SS AEAD 加密（AES-GCM, Chacha20-Poly1305）
- [ ] 密钥派生（HKDF-SHA1）
- [ ] 数据包编/解码
- [ ] 与 Hub 集成

### Phase 3: VMess/VLESS + Trojan
- [ ] VMess 协议（TCP/mTLS/WS/gRPC）
- [ ] VLESS 协议
- [ ] Trojan 协议

### Phase 4: 规则引擎
- [ ] 域名/IP/GeoIP 匹配
- [ ] 正则表达式支持
- [ ] 规则集订阅（Rule Provider）

### Phase 5: DNS 分流
- [ ] 分流 DNS（国内/海外）
- [ ] fake-ip / redir-host 模式
- [ ] DNS 缓存

### Phase 6: TUN 模式
- [ ] macOS/Linux TUN 设备
- [ ] Windows WinTun
- [ ] Android/iOS VPN API

### Phase 7: 平台 UI
- [ ] gRPC 绑定层
- [ ] macOS UI (SwiftUI)
- [ ] Android UI (Kotlin)
- [ ] iOS UI (Swift)

## 测试策略

- **单元测试：** 每个包独立测试
- **集成测试：** 端到端流程测试
- **基准测试：** 性能关键路径

运行测试：
```bash
# 所有测试
go test ./... -v

# 覆盖率
go test -cover ./...
```

## 构建 & 发布

```bash
# 跨平台编译
GOOS=linux GOARCH=amd64 go build -o proxy-engine-linux ./cmd/proxy-engine
GOOS=windows GOARCH=amd64 go build -o proxy-engine.exe ./cmd/proxy-engine
GOOS=darwin GOARCH=arm64 go build -o proxy-engine-mac ./cmd/proxy-engine
```

## 依赖管理

添加新依赖：
```bash
go get <package>
go mod tidy
```

## 相关资源

- 设计文档：`docs/superpowers/specs/2026-04-14-proxy-engine-design.md`
- 实现计划：`docs/superpowers/plans/2026-04-14-proxy-engine-phase1.md`
- SOCKS5 协议：RFC 1928
- HTTP CONNECT：RFC 7231 Section 4.3.6
