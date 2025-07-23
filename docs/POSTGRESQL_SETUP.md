# PostgreSQL 数据库搭建指南

## 概述

本文档描述了为 OpenChain 项目搭建 PostgreSQL 数据库的完整过程，包括安装、配置、用户创建和权限设置。

## 系统要求

- Ubuntu/Debian 系统
- 管理员权限
- 网络连接（用于安装包）

## 安装步骤

### 1. 更新系统包

```bash
sudo apt update
sudo apt upgrade -y
```

### 2. 安装 PostgreSQL

```bash
# 安装 PostgreSQL 14
sudo apt install postgresql postgresql-contrib -y

# 验证安装
psql --version
```

### 3. 启动 PostgreSQL 服务

```bash
# 启动 PostgreSQL 服务
sudo systemctl start postgresql

# 设置开机自启
sudo systemctl enable postgresql

# 检查服务状态
sudo systemctl status postgresql
```

### 4. 配置 PostgreSQL

#### 4.1 切换到 postgres 用户

```bash
sudo -i -u postgres
```

#### 4.2 创建数据库用户

```bash
# 创建 ethereum 用户
createuser --interactive ethereum

# 设置密码
psql -c "ALTER USER ethereum PASSWORD 'ethereum';"
```

#### 4.3 创建数据库

```bash
# 创建 ethereum 数据库
createdb -O ethereum ethereum
```

#### 4.4 配置访问权限

编辑 PostgreSQL 配置文件：

```bash
# 编辑 pg_hba.conf
sudo nano /etc/postgresql/16/main/pg_hba.conf
```

添加以下行（允许本地连接）：

```
# IPv4 local connections:
host    all             all             127.0.0.1/32            md5
# IPv6 local connections:
host    all             all             ::1/128                 md5
```

#### 4.5 重启 PostgreSQL 服务

```bash
sudo systemctl restart postgresql
```

### 5. 验证数据库连接

#### 5.1 使用 psql 连接

```bash
# 使用 ethereum 用户连接
psql -h localhost -U ethereum -d ethereum
```

#### 5.2 测试连接

```sql
-- 在 psql 中执行
SELECT version();
SELECT current_user;
SELECT current_database();
```

### 6. 创建必要的表结构

#### 6.1 连接到数据库

```bash
psql -h localhost -U ethereum -d ethereum
```

#### 6.2 创建基础表

```sql
-- 创建用户表
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建交易表
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    tx_hash VARCHAR(66) UNIQUE NOT NULL,
    chain VARCHAR(20) NOT NULL,
    block_number BIGINT,
    from_address VARCHAR(42),
    to_address VARCHAR(42),
    value NUMERIC,
    gas_used BIGINT,
    gas_price NUMERIC,
    status INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建合约表
CREATE TABLE contracts (
    id SERIAL PRIMARY KEY,
    address VARCHAR(42) UNIQUE NOT NULL,
    chain VARCHAR(20) NOT NULL,
    name VARCHAR(100),
    abi TEXT,
    bytecode TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建编译记录表
CREATE TABLE compilation_records (
    id SERIAL PRIMARY KEY,
    contract_id INTEGER REFERENCES contracts(id),
    compiler_version VARCHAR(20),
    source_code TEXT,
    compilation_result TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 6.3 创建索引

```sql
-- 为交易表创建索引
CREATE INDEX idx_transactions_tx_hash ON transactions(tx_hash);
CREATE INDEX idx_transactions_chain ON transactions(chain);
CREATE INDEX idx_transactions_block_number ON transactions(block_number);

-- 为合约表创建索引
CREATE INDEX idx_contracts_address ON contracts(address);
CREATE INDEX idx_contracts_chain ON contracts(chain);

-- 为编译记录表创建索引
CREATE INDEX idx_compilation_records_contract_id ON compilation_records(contract_id);
```

### 7. 配置应用程序连接

#### 7.1 环境变量设置

在应用程序中设置以下环境变量：

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=ethereum
export DB_USER=ethereum
export DB_PASSWORD=ethereum
```

#### 7.2 连接字符串格式

```
postgresql://ethereum:ethereum@localhost:5432/ethereum
```

### 8. 权限管理

#### 8.1 授予必要权限

```sql
-- 连接到数据库
psql -h localhost -U ethereum -d ethereum

-- 授予所有权限给 ethereum 用户
GRANT ALL PRIVILEGES ON DATABASE ethereum TO ethereum;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ethereum;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ethereum;

-- 设置默认权限
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO ethereum;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO ethereum;
```

#### 8.2 验证权限

```sql
-- 检查用户权限
\du ethereum

-- 检查表权限
\dp
```

### 9. 性能优化

#### 9.1 调整 PostgreSQL 配置

编辑 `postgresql.conf`：

```bash
sudo nano /etc/postgresql/16/main/postgresql.conf
```

添加以下配置：

```ini
# 内存配置
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB

# 连接配置
max_connections = 100

# 日志配置
log_statement = 'all'
log_duration = on
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
```

#### 9.2 重启服务

```bash
sudo systemctl restart postgresql
```

### 10. 备份和恢复

#### 10.1 创建备份

```bash
# 创建完整备份
pg_dump -h localhost -U ethereum -d ethereum > ethereum_backup.sql

# 创建压缩备份
pg_dump -h localhost -U ethereum -d ethereum | gzip > ethereum_backup.sql.gz
```

#### 10.2 恢复备份

```bash
# 恢复完整备份
psql -h localhost -U ethereum -d ethereum < ethereum_backup.sql

# 恢复压缩备份
gunzip -c ethereum_backup.sql.gz | psql -h localhost -U ethereum -d ethereum
```

### 11. 监控和维护

#### 11.1 查看数据库状态

```bash
# 查看连接数
psql -h localhost -U ethereum -d ethereum -c "SELECT count(*) FROM pg_stat_activity;"

# 查看表大小
psql -h localhost -U ethereum -d ethereum -c "
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
"
```

#### 11.2 定期维护

```bash
# 分析表统计信息
psql -h localhost -U ethereum -d ethereum -c "ANALYZE;"

# 清理过期日志
sudo find /var/log/postgresql -name "*.log" -mtime +30 -delete
```

### 12. 故障排除

#### 12.1 常见问题

**问题**: 连接被拒绝

```bash
# 检查服务状态
sudo systemctl status postgresql

# 检查端口监听
sudo netstat -tlnp | grep 5432
```

**问题**: 权限错误

```sql
-- 重新授予权限
GRANT ALL PRIVILEGES ON DATABASE ethereum TO ethereum;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ethereum;
```

**问题**: 内存不足

```bash
# 调整系统内存
echo 'vm.swappiness=10' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

#### 12.2 日志查看

```bash
# 查看 PostgreSQL 日志
sudo tail -f /var/log/postgresql/postgresql-16-main.log

# 查看系统日志
sudo journalctl -u postgresql -f
```

### 13. 安全配置

#### 13.1 网络安全

```bash
# 配置防火墙
sudo ufw allow 5432/tcp
sudo ufw enable
```

#### 13.2 SSL 配置

```bash
# 生成 SSL 证书
sudo openssl req -new -x509 -days 365 -nodes -text -out /etc/ssl/certs/server.crt -keyout /etc/ssl/private/server.key -subj "/CN=localhost"
```

编辑 `postgresql.conf`：

```ini
ssl = on
ssl_cert_file = '/etc/ssl/certs/server.crt'
ssl_key_file = '/etc/ssl/private/server.key'
```

### 14. 测试连接

#### 14.1 使用 psql 测试

```bash
psql -h localhost -U ethereum -d ethereum -c "SELECT 'Database connection successful' as status;"
```

#### 14.2 使用应用程序测试

```bash
# 测试数据库连接
curl -X GET "http://localhost:8083/api/v1/health" | jq
```

## 总结

通过以上步骤，我们成功搭建了一个完整的 PostgreSQL 数据库环境，包括：

- ✅ PostgreSQL 16 安装和配置
- ✅ 数据库用户和数据库创建
- ✅ 表结构和索引创建
- ✅ 权限管理和安全配置
- ✅ 性能优化和监控
- ✅ 备份和恢复策略
- ✅ 故障排除指南

数据库现在可以支持 OpenChain 项目的所有微服务，包括：

- signature-database-srv
- vyper-compiler-srv
- solidity-compiler-srv
- tx-tracer-srv

所有服务都可以通过配置的环境变量连接到数据库，实现数据的持久化存储。
