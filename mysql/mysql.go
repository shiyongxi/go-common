package mysql

import (
	"fmt"
	"github.com/shiyongxi/go-common/common"
	"github.com/shiyongxi/go-common/tracer"
	"golang.org/x/net/context"
	"log"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type (
	Model struct {
		CreatedAt int64 `json:"createdAt" gorm:"column:created_at"`
		UpdatedAt int64 `json:"updatedAt" gorm:"column:updated_at"`
		Valid     int32 `json:"valid" gorm:"column:valid"`
	}

	MysqlConf struct {
		Dialect           string `yaml:"dialect"`
		Host              string `yaml:"host"`
		Port              int64  `yaml:"port"`
		DbName            string `yaml:"dbname"`
		User              string `yaml:"user"`
		Password          string `yaml:"password"`
		Charset           string `yaml:"charset"`
		ParseTime         bool   `yaml:"parseTime"`
		MaxIdle           int    `yaml:"maxIdle"`
		MaxOpen           int    `yaml:"maxOpen"`
		Debug             bool   `yaml:"debug"`
		InterpolateParams bool   `yaml:"interpolateParams"`
		MultiStatements   bool   `yaml:"multiStatements"`
	}
)

var (
	connMap sync.Map
	err     error
)

func NewMysql(conf *MysqlConf) *MysqlConf {
	return conf
}

const (
	connectionDefault = "default"
)

func NewDBClient(ctx context.Context, key ...string) *gorm.DB {
	conn, ok := connMap.Load(getConnKey(key))
	if !ok {
		log.Fatal("no mysql connection found")
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return SetSpanToGorm(ctx, conn.(*gorm.DB))
}

func (m *MysqlConf) Connection(key ...string) *gorm.DB {
	conn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=%s&parseTime=%t&loc=Local",
		m.User,
		m.Password,
		m.Host,
		m.Port,
		m.DbName,
		m.Charset,
		m.ParseTime)

	if m.InterpolateParams {
		conn = fmt.Sprintf("%s&interpolateParams=%t", conn, m.InterpolateParams)
	}

	if m.MultiStatements {
		conn = fmt.Sprintf("%s&multiStatements=%t", conn, m.MultiStatements)
	}

	db, err := gorm.Open(m.Dialect, conn)
	if err != nil {
		log.Fatal(err)
	}

	db.Debug()
	db.DB().SetMaxIdleConns(m.MaxIdle)
	db.DB().SetMaxOpenConns(m.MaxOpen)
	db.DB().SetConnMaxLifetime(time.Second * 14400)

	db.SingularTable(true)

	db.LogMode(m.Debug)
	db.Set("gorm:table_options", "ENGINE=InnoDB")

	AddGormCallbacks(db, tracer.GetTracerClient())
	connMap.LoadOrStore(getConnKey(key), db)
	return db
}

func (m *Model) BeforeSave(scope *gorm.Scope) {
	m.UpdatedAt = common.NewTools().GetNowMillisecond()
}

func (m *Model) BeforeCreate(scope *gorm.Scope) {
	m.CreatedAt = common.NewTools().GetNowMillisecond()
	m.UpdatedAt = common.NewTools().GetNowMillisecond()
}

func (m *Model) BeforeUpdate(scope *gorm.Scope) {
	m.UpdatedAt = common.NewTools().GetNowMillisecond()
}

func getConnKey(key []string) string {
	if len(key) == 1 {
		return key[0]
	}

	return connectionDefault
}
