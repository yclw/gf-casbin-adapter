# GF Casbin Adapter

[![Go](https://img.shields.io/badge/Go-1.24.2+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![GoFrame](https://img.shields.io/badge/GoFrame-v2.9.0+-00ADD8?style=flat)](https://github.com/gogf/gf)
[![Casbin](https://img.shields.io/badge/Casbin-v2.115.0+-FF6B6B?style=flat)](https://github.com/casbin/casbin)

[English](README.md) |  [中文](README.zh.md)

GoFrame 适配器是基于 [GoFrame](https://github.com/gogf/gf) 框架的 [Casbin](https://github.com/casbin/casbin) 适配器。通过这个库，Casbin 可以从 GoFrame 支持的数据库中加载策略或将策略保存到数据库中。支持或计划支持casbin adapter的所有接口。

基于 GoFrame 数据库驱动支持，当前支持的数据库有：

* MySQL
* PostgreSQL
* 以及其他 GoFrame 支持的数据库

model与dao均由gf gen生成，符合 gf 的 orm 规范

## 功能特性

* 该适配器实现了Casbin的所有Adapter接口，但功能有待测试。

## 快速使用

> 请先完成goframe数据库配置。并引入相关驱动包，如：
> `import _ "github.com/gogf/gf/contrib/drivers/mysql/v2"`

### 安装

```bash
go get github.com/yclw/gf-casbin-adapter
```

### 导入

```go
import (
 gfadapter "github.com/yclw/gf-casbin-adapter"
 "github.com/casbin/casbin/v2"
 "github.com/casbin/casbin/v2/model"
)
```

### 创建Adapter

```go
// 创建Adapter
adapter, err := gfadapter.NewAdapter()
```

### 创建Enforcer

```go
// 根据casbin需求创建model

// 使用model和adapter创建enforcer
enforcer, err := casbin.NewEnforcer(model, adapter)
```

## 注意事项

1. 确保 GoFrame 数据库配置正确。

## 数据库表结构

```sql
CREATE TABLE casbin_rule (
    id    BIGINT AUTO_INCREMENT PRIMARY KEY,
    ptype VARCHAR(100) DEFAULT '' NOT NULL,
    v0    VARCHAR(100) DEFAULT '' NOT NULL,
    v1    VARCHAR(100) DEFAULT '' NOT NULL,
    v2    VARCHAR(100) DEFAULT '' NOT NULL,
    v3    VARCHAR(100) DEFAULT '' NOT NULL,
    v4    VARCHAR(100) DEFAULT '' NOT NULL,
    v5    VARCHAR(100) DEFAULT '' NOT NULL,
    INDEX idx_ptype (ptype),
    INDEX idx_v0 (v0),
    INDEX idx_v1 (v1),
    INDEX idx_v2 (v2),
    INDEX idx_v3 (v3),
    INDEX idx_v4 (v4),
    INDEX idx_v5 (v5),
    UNIQUE INDEX uniq_ptype_v0_v1_v2_v3_v4_v5 (ptype, v0, v1, v2, v3, v4, v5)
) COMMENT 'Casbin';
```

## 获取帮助

* [Casbin 官方文档](https://casbin.org/)
* [GoFrame 官方文档](https://goframe.org/)  
* [Casbin 中文文档](https://casbin.org/zh/)
* [GoFrame 中文文档](https://goframe.org/pages/viewpage.action?pageId=1114119)
* [Casbin GitHub](https://github.com/casbin/casbin)
* [hailaz/gf-casbin-adapter](https://github.com/hailaz/gf-casbin-adapter)
* [casbin/gorm-adapter](https://github.com/casbin/gorm-adapter)
* [casbin/xorm-adapter](https://github.com/casbin/xorm-adapter)

## 许可证

本项目采用 Apache 2.0 许可证。详见 LICENSE 文件。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。
