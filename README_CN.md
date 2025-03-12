# confx

confx 是一个功能丰富的 Go 语言配置管理库，提供了全面的解决方案来简化应用程序配置的处理。它结合了命令行参数、环境变量、配置文件和默认值，帮助开发者高效地管理应用程序配置。

## 功能特点

- **统一配置管理**：自动绑定命令行参数、环境变量和配置文件
- **强类型支持**：使用结构体定义配置，支持类型安全的配置访问
- **丰富的数据类型**：支持基本类型、切片、映射、嵌套结构体等
- **指针类型支持**：自动处理 nil 指针，确保配置加载后所有字段都有可用值
- **标签驱动**：通过结构体标签定义配置项的键名、用法说明等
- **完整验证支持**：集成 go-playground/validator，支持其全部验证规则和特性
- **增强的条件验证**：在支持标准 validator 的基础上，提供增强型的嵌套结构体条件验证功能
- **结构体嵌入**：支持使用`squash`标签扁平化嵌套结构体
- **自定义选项**：提供灵活的选项来自定义配置加载行为
- **通用配置读取**：支持从多种格式的配置文件中读取配置

## 安装

```bash
go get github.com/qor5/confx
```

## 快速开始

### 创建默认配置文件

首先创建一个包含默认配置的 YAML 文件，命名为`default-config.yaml`：

```yaml
# default-config.yaml
server:
  host: localhost
  port: 8080
  timeout: 30s

database:
  host: localhost
  port: 5432
  username: user
  password: password
  database: myapp

logLevel: info
```

### 定义配置结构体

```go
package main

import (
    "context"
    "embed"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/qor5/confx"
)

//go:embed default-config.yaml
var defaultConfigYaml string

// 定义配置结构体
type ServerConfig struct {
    Host    string `confx:"host" usage:"服务器主机地址" validate:"required"`
    Port    int    `confx:"port" usage:"服务器端口" validate:"gte=1,lte=65535"`
    Timeout time.Duration `confx:"timeout" usage:"请求超时时间" validate:"gte=0"`
}

type DatabaseConfig struct {
    Host     string `confx:"host" usage:"数据库主机地址" validate:"required"`
    Port     int    `confx:"port" usage:"数据库端口" validate:"gte=1,lte=65535"`
    Username string `confx:"username" usage:"数据库用户名"`
    Password string `confx:"password" usage:"数据库密码"`
    Database string `confx:"database" usage:"数据库名称" validate:"required"`
}

type Config struct {
    Server   ServerConfig   `confx:"server" validate:"required"`
    Database DatabaseConfig `confx:"database" validate:"required"`
    LogLevel string         `confx:"logLevel" usage:"日志级别" validate:"oneof=debug info warn error"`
}

func main() {
    // 从嵌入的YAML字符串中读取默认配置
    // 通常我们会将默认配置嵌入到二进制中，这样有三个好处：
    // 1. CLI可以独立运行，不依赖外部配置文件
    // 2. 可以将此文件交付给用户，让用户了解所有可用的配置项及其细节
    // 3. 用户可以复制此文件并根据需要修改，简化自定义配置的过程

    // 注意：如果你的配置足够简单，也可以直接构造Config对象，无需从文件读取
    // 例如：defaultConfig := Config{Server: ServerConfig{Host: "localhost", Port: 8080}, ...}
    defaultConfig, err := confx.Read[Config]("yaml", strings.NewReader(defaultConfigYaml))
    if err != nil {
        log.Fatalf("读取默认配置失败: %v", err)
    }

    // 初始化配置加载器
    loader, err := confx.Initialize(defaultConfig)
    if err != nil {
        log.Fatalf("初始化配置失败: %v", err)
    }

    // 加载配置（可选择性地指定配置文件路径）
    config, err := loader(context.Background(), "config.yaml")
    if err != nil {
        log.Fatalf("加载配置失败: %v", err)
    }

    // 使用配置
    fmt.Printf("服务器配置: %s:%d\n", config.Server.Host, config.Server.Port)
    fmt.Printf("数据库配置: %s:%d/%s\n", config.Database.Host, config.Database.Port, config.Database.Database)
    fmt.Printf("日志级别: %s\n", config.LogLevel)
}
```

### 命令行参数

confx 会自动为配置结构体中的每个字段生成对应的命令行参数：

```bash
./myapp --server-host=127.0.0.1 --server-port=9090 --log-level=debug
```

### 环境变量

confx 也会绑定环境变量到配置字段：

```bash
SERVER_HOST=127.0.0.1 SERVER_PORT=9090 LOG_LEVEL=debug ./myapp
```

可以通过`WithEnvPrefix`选项自定义环境变量前缀：

```go
loader, err := confx.Initialize(defaultConfig, confx.WithEnvPrefix("APP_"))
```

然后使用带前缀的环境变量：

```bash
APP_SERVER_HOST=127.0.0.1 APP_SERVER_PORT=9090 APP_LOG_LEVEL=debug ./myapp
```

### 配置文件

confx 支持从各种格式的配置文件中加载配置：

```yaml
# config.yaml
server:
  host: 127.0.0.1
  port: 9090
  timeout: 60s

database:
  host: db.example.com
  port: 5432
  username: admin
  password: secret
  database: production

logLevel: debug
```

## 高级特性

### 验证功能

confx 完全集成了 go-playground/validator，支持其所有内置验证规则和特性。此外，confx 还提供了额外的增强功能。

#### 标准验证器功能

以下是 go-playground/validator 提供的常见验证功能示例：

```go
type Config struct {
    // 基本验证规则
    Port      int       `validate:"required,gte=1,lte=65535"`
    Email     string    `validate:"required,email"`
    URL       string    `validate:"url"`
    CreatedAt time.Time `validate:"required"`

    // 条件验证
    OutputPath string `validate:"required_if=OutputType file"` // 当OutputType为"file"时必填

    // 切片验证
    Tags []string `validate:"required,min=1,dive,required"`
}
```

#### confx 增强的条件验证

confx 在标准 validator 基础上为嵌套结构体增加了`skip_nested_unless`验证规则，用于条件性地验证整个嵌套结构：

```go
type AuthConfig struct {
    Provider string    `confx:"provider" validate:"required,oneof=jwt oauth basic"`
    // 仅当Provider为"jwt"时才验证JWT配置 - 这是confx的增强功能
    JWT      JWTConfig `confx:"jwt" validate:"skip_nested_unless=Provider jwt"`
    // 仅当Provider为"oauth"时才验证OAuth配置 - 这是confx的增强功能
    OAuth    OAuthConfig `confx:"oauth" validate:"skip_nested_unless=Provider oauth"`
}
```

### 结构体嵌入

可以使用`squash`标签将嵌套结构体的字段扁平化到父结构体中：

```go
type CommonDBConfig struct {
    Name     string `confx:"name" validate:"required"`
    Username string `confx:"username"`
    Password string `confx:"password"`
}

type DatabaseConfig struct {
    Type string `confx:"type" validate:"required,oneof=postgres sqlite"`
    // 扁平化嵌入CommonDBConfig的字段
    CommonDBConfig `confx:",squash"`
    // 其他数据库特定字段
    Host string `confx:"host"`
    Port int    `confx:"port" validate:"omitempty,gte=1,lte=65535"`
}
```

### 忽略字段

使用 `confx:"-"` 标签可以让 confx 完全忽略结构体中的某些字段，这些字段不会被映射、不会生成标志，也不会被环境变量覆盖：

```go
type Config struct {
    // 正常处理的字段
    Database DatabaseConfig `confx:"database"`

    // 被忽略的字段 - 不会被confx处理
    InternalState string `confx:"-"`

    // 私有字段自动被忽略（无需显式标记）
    internalCache map[string]interface{}

    // 即使是导出字段，也可以通过"-"标签被忽略
    HelperFunction func() `confx:"-"`
}
```

### 自定义选项

confx 提供了多种选项来自定义配置加载行为：

```go
loader, err := confx.Initialize(defaultConfig,
    confx.WithEnvPrefix("APP_"),           // 设置环境变量前缀
    confx.WithFlagSet(customFlagSet),      // 使用自定义的FlagSet
    confx.WithViper(customViper),          // 使用自定义的Viper实例
    confx.WithValidator(customValidator),  // 使用自定义的验证器
    confx.WithTagName("custom"),           // 使用自定义的结构体标签名
    confx.WithUsageTagName("description"), // 使用自定义的用法标签名
    confx.WithFieldHook(customFieldHook),  // 自定义字段处理
)
```

## 实用工具

### 直接从文件加载配置

除了使用 Initialize 方法外，confx 还提供了简单的函数来直接从文件或读取器加载配置：

```go
// 从配置文件加载
config, err := confx.Read[Config]("yaml", configFile)

// 使用自定义标签名从配置文件加载
config, err := confx.ReadWithTagName[Config]("custom", "yaml", configFile)
```

## 与 Viper 和 Cobra 的集成

confx 与流行的 Viper 和 Cobra 库无缝集成：

- **Viper**：confx 使用 Viper 作为底层配置管理引擎，可以通过`WithViper`选项使用自定义的 Viper 实例
- **Cobra**：查看`examples/cobra`目录了解如何将 confx 与 Cobra 命令行框架集成

## 示例

查看`examples`目录获取更多示例：

- `examples/basic`：基本使用示例
- `examples/cobra`：与 Cobra 集成的示例
- `examples/config`：示例共享的配置包，定义了示例使用的配置结构

## 贡献

欢迎贡献代码、报告问题或提供改进建议！请提交 issue 或 pull request 到此仓库。

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件
