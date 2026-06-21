package repository

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func sqlMeta(rawSQL string) string {
	return regexp.QuoteMeta(rawSQL)
}

func newMockDb(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("create sqlmock failed:", err)
	}

	// postgres://admin:123456@192.168.1.110:5432/olixops
	dialector := postgres.New(postgres.Config{
		Conn: sqlDB,
	})

	gdb, err := gorm.Open(dialector, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})

	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}

	return gdb, mock
}

//func Test_ClusterRepo_Create(t *testing.T) {
//	gdb, mock := newMockDb(t)
//	repo := NewClusterRepo(gdb)
//	ctx := context.Background()
//
//	new := time.Now()
//	cluster := &domain.Cluster{
//		ID:          "cls_0001",
//		TenantID:    "tenant_0001",
//		Name:        "测试集群",
//		Environment: "dev",
//		CreatorID:   "creator_0001",
//	}
//
//}
