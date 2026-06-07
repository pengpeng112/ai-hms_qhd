// 命令 server:连接数据库 → 迁移 Schedule_* 表 → 启动 Gin HTTP 服务。
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/sdsph/dialysis-scheduling/internal/api"
	"github.com/sdsph/dialysis-scheduling/internal/db"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("请设置环境变量 DATABASE_URL,例如:\n" +
			`  host=localhost user=postgres password=xxx dbname=aihms port=5432 sslmode=disable TimeZone=Asia/Shanghai`)
	}

	g, err := db.Open(dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	if err := db.Migrate(g); err != nil {
		log.Fatalf("迁移失败: %v", err)
	}
	log.Println("✅ Schedule_* 表迁移完成")

	r := gin.Default()
	(&api.Server{DB: g}).Register(r)

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("🚀 透析排班服务监听 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
