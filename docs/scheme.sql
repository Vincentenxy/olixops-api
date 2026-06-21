
-- 创建数据库
CREATE DATABASE olixops
    OWNER admin
    ENCODING 'UTF8';

-- 切换库
\c olixops;

-- 自定义集群状态枚举类型
CREATE TYPE cluster_status AS ENUM ('unknown', 'online', 'offline', 'unreachable');

-- cluster 表
CREATE TABLE cluster (
                         id VARCHAR(64) PRIMARY KEY,
                         tenant_id VARCHAR(64) NOT NULL,
                         name VARCHAR(128) NOT NULL,
                         environment VARCHAR(32),
                         creator_id VARCHAR(64),
                         description TEXT,
                         kube_config_path TEXT NOT NULL,
                         status cluster_status NOT NULL DEFAULT 'unknown',
                         last_probe_at TIMESTAMP NOT NULL,
                         last_sync_at TIMESTAMP NULL,
                         node_count INT NOT NULL DEFAULT 0,
                         k8s_version VARCHAR(64),
                         created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                         updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                         deleted_at TIMESTAMP NULL
);

-- 唯一索引：集群名称全局唯一
CREATE UNIQUE INDEX idx_name ON cluster(name) WHERE deleted_at IS NULL;

-- 联合索引：租户+环境，高频列表查询
CREATE INDEX idx_tenant_env ON cluster(tenant_id, environment);

-- 状态过滤索引（在线/离线告警）
CREATE INDEX idx_status ON cluster(status);

-- 创建人筛选索引
CREATE INDEX idx_creator_id ON cluster(creator_id);

-- 软删除索引
CREATE INDEX idx_deleted_at ON cluster(deleted_at);

-- 注释
COMMENT ON TABLE cluster IS 'K8s集群信息主表，管理所有接入平台的K8s集群';
COMMENT ON COLUMN cluster.id IS '集群唯一主键ID，UUID/雪花字符串';
COMMENT ON COLUMN cluster.tenant_id IS '租户ID，多租户顶层数据隔离';
COMMENT ON COLUMN cluster.name IS '集群展示名称，全局唯一（未删除）';
COMMENT ON COLUMN cluster.environment IS '环境标识：prod/test/dev/stag';
COMMENT ON COLUMN cluster.creator_id IS '集群创建人用户ID';
COMMENT ON COLUMN cluster.description IS '集群备注、机房、负责人、业务说明';
COMMENT ON COLUMN cluster.kube_config_path IS 'kubeconfig 在envvault中的访问位置';
COMMENT ON COLUMN cluster.status IS '集群连通状态枚举：unknown/online/offline/unreachable';
COMMENT ON COLUMN cluster.last_probe_at IS '最后一次集群连通性探测时间，非空';
COMMENT ON COLUMN cluster.last_sync_at IS '最后同步节点数量、K8s版本缓存时间，未同步为NULL';
COMMENT ON COLUMN cluster.node_count IS '集群节点总数缓存值，定时任务刷新';
COMMENT ON COLUMN cluster.k8s_version IS '集群K8s版本号，如v1.28.2';
COMMENT ON COLUMN cluster.created_at IS '集群录入时间，自动填充';
COMMENT ON COLUMN cluster.updated_at IS '集群信息最后修改时间，自动更新';
COMMENT ON COLUMN cluster.deleted_at IS '软删除时间，NULL代表有效数据';