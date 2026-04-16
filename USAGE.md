# Proxy Engine 使用指南

## 这是什么？

**Proxy Engine** 是一个代理客户端工具，类似于 Clash 或 V2Ray。

### 核心功能

1. **流量代理** - 将你的网络流量转发到远程服务器
2. **规则分流** - 根据域名/IP/规则决定哪些流量走代理，哪些直连
3. **多协议支持** - 支持 Shadowsocks、VMess、Trojan 等协议
4. **加密传输** - 通过加密的隧道连接，保护隐私

### 应用场景

- 🌍 **访问被屏蔽内容** - 访问 Google、GitHub 等被屏蔽的服务
- 🔒 **隐私保护** - 加密网络流量，防止被监控
- 🎮 **游戏加速** - 通过代理连接游戏服务器（延迟更低）
- 📺 **解锁地区内容** - 观看特定地区的视频内容

---

## 当前状态（Phase 1）

**✅ 已实现功能：**
- SOCKS5 代理服务器（端口 7891）
- HTTP CONNECT 代理（端口 7890）
- 基础配置管理（YAML）
- REST API 接口（端口 9090）

**⚠️ 当前限制：**
- 仅支持 DIRECT 模式（直连，无实际代理）
- Phase 2 将实现 Shadowsocks 协议支持

---

## 快速开始

### 1. 构建项目

```bash
# 进入项目目录
cd /Users/0du/TestCode/proxy-engine

# 构建二进制文件
go build -o proxy-engine ./cmd/proxy-engine
```

### 2. 准备配置文件

创建 `config.yaml`：

```yaml
port: 7890              # HTTP 代理端口
socks-port: 7891        # SOCKS5 代理端口
mode: direct           # 当前只有 direct 模式
log-level: info          # 日志级别
```

### 3. 启动代理引擎

```bash
./proxy-engine -c config.yaml
```

你会看到：
```
2026/04/16 17:36:18 proxy-engine v0.1.0 starting
2026/04/16 17:36:18 config: mode= port=7894 socks-port=7895
2026/04/16 17:36:18 SOCKS5 listening on 127.0.0.1:7895
2026/04/16 17:36:18 HTTP proxy listening on 127.0.0.1:7894
2026/04/16 17:36:18 API server listening on 127.0.0.1:9090
```

### 4. 配置系统代理

**macOS:**
```bash
# 设置网络代理（SOCKS5）
networksetup -setsocksfirewallproxy Wi-Fi "127.0.0.1" 7891

# 或者设置 HTTP 代理
networksetup -setwebproxy Wi-Fi 127.0.0.1 7890
```

**或在系统设置中：**
- 系统设置 → 网络 → 高级 → 代理
- SOCKS5 代理：127.0.0.1:7891
- HTTP 代理：127.0.0.1:7890

**命令行使用（推荐）：**
```bash
# 使用 SOCKS5 代理
curl -x socks5://127.0.0.1:7891 https://httpbin.org/ip

# 使用 HTTP 代理
curl -x http://127.0.0.1:7890 https://httpbin.org/ip
```

---

## 完整使用流程（Phase 2+）

当 Phase 2 完成后，你将能够：

### 配置代理节点

```yaml
proxies:
  - name: "ss-server"
    type: ss
    server: ss-server.com
    port: 8388
    cipher: aes-256-gcm
    password: "your-password"
```

### 配置规则分流

```yaml
rules:
  - DOMAIN-SUFFIX,google.com,PROXY
  - DOMAIN-KEYWORD,github,PROXY
  - GEOIP,CN,DIRECT
  - MATCH,PROXY
```

### 启动并使用

1. 启动代理引擎
2. 配置系统代理
3. 浏览器/应用自动使用代理

---

## API 接口

**健康检查：**
```bash
curl http://127.0.0.1:9090/api/health
# 返回: {"status":"ok"}
```

**查看配置：**
```bash
curl http://127.0.0.1:9090/api/configs
```

---

## 常见问题

### Q: 和 Clash 有什么区别？

**A:** 这是用 Go 从零实现的代理引擎，功能类似 Clash，但：
- 更轻量级
- 代码更易理解和定制
- 支持多平台原生 UI

### Q: 当前能做什么？

**A:** Phase 1 已实现基础代理框架，但还需要实现具体协议（Shadowsocks 等）才能作为实际代理使用。

### Q: 什么时候能用？

**A:** 
- **Phase 1** ✅ 基础框架（当前）
- **Phase 2** 🚧 Shadowsocks 协议（下一步）
- **Phase 3-9** 📋 更多协议和功能

### Q: 如何参与开发？

**A:** 
1. Fork GitHub 仓库
2. 创建功能分支
3. 提交 Pull Request
4. 参见 [CLAUDE.md](CLAUDE.md) 开发规范

---

## 安全提示

⚠️ **重要提醒：**

1. **合法使用** - 仅在允许的地区使用，遵守当地法律法规
2. **隐私保护** - 不要代理敏感信息
3. **密码安全** - 配置文件包含密码，注意保护
4. **开源贡献** - 欢迎提交 Issue 和 Pull Request

---

## 下一步

- [ ] Phase 2: 实现 Shadowsocks 协议
- [ ] 添加规则引擎
- [ ] 实现 TUN 模式
- [ ] 开发桌面 UI

**想要参与开发？** 查看 [CLAUDE.md](CLAUDE.md) 了解开发规范！
