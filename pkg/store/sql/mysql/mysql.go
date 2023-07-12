package mysql

import (
	"context"
	"fmt"
	"time"

	gomysql "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	sqlstore "github.com/Chamindu36/organization-name-registry-service/pkg/store/sql"
)


type Config struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Protocol string `yaml:"protocol"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`

	Collation        string `yaml:"collation"`
	Loc              string `yaml:"loc"`
	MaxAllowedPacket int    `yaml:"maxAllowedPacket"`
	ServerPubKey     string `yaml:"serverPubKey"`
	TLSConfig        string `yaml:"tlsConfig"`

	AllowAllFiles           bool `yaml:"allowAllFiles"`
	AllowCleartextPasswords bool `yaml:"allowCleartextPasswords"`
	AllowNativePasswords    bool `yaml:"allowNativePasswords"`
	AllowOldPasswords       bool `yaml:"allowOldPasswords"`
	CheckConnLiveness       bool `yaml:"checkConnLiveness"`
	ClientFoundRows         bool `yaml:"clientFoundRows"`
	ColumnsWithAlias        bool `yaml:"columnsWithAlias"`
	InterpolateParams       bool `yaml:"interpolateParams"`
	MultiStatements         bool `yaml:"multiStatements"`
	ParseTime               bool `yaml:"parseTime"`
	RejectReadOnly          bool `yaml:"rejectReadOnly"`

	Params map[string]string `yaml:"additionalParams"`
}

type mysql struct {
	db *sqlx.DB
}

func MustConnect(cfg *Config) *sqlx.DB {
	return sqlx.MustConnect("mysql", makeDsn(cfg))
}

func MustOpen(cfg *Config) *sqlx.DB {
	return sqlx.MustOpen("mysql", makeDsn(cfg))
}

func makeDsn(cfg *Config) string {
	loc, err := time.LoadLocation(cfg.Loc)
	if err != nil {
		panic(err)
	}
	mysqlCfg := gomysql.Config{
		User:   cfg.Username,
		Passwd: cfg.Password,
		Net:    cfg.Protocol,
		Addr:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		DBName: cfg.Name,

		Collation:        cfg.Collation,
		Loc:              loc,
		MaxAllowedPacket: cfg.MaxAllowedPacket,
		ServerPubKey:     cfg.ServerPubKey,
		TLSConfig:        cfg.TLSConfig,

		AllowAllFiles:           cfg.AllowAllFiles,
		AllowCleartextPasswords: cfg.AllowCleartextPasswords,
		AllowNativePasswords:    cfg.AllowNativePasswords,
		AllowOldPasswords:       cfg.AllowOldPasswords,
		CheckConnLiveness:       cfg.CheckConnLiveness,
		ClientFoundRows:         cfg.ClientFoundRows,
		ColumnsWithAlias:        cfg.ColumnsWithAlias,
		InterpolateParams:       cfg.InterpolateParams,
		MultiStatements:         cfg.MultiStatements,
		ParseTime:               cfg.ParseTime,
		RejectReadOnly:          cfg.RejectReadOnly,

		Params: cfg.Params,
	}
	return mysqlCfg.FormatDSN()
}

func NewMySql(db *sqlx.DB) *mysql {
	return &mysql{
		db: db,
	}
}

func (m *mysql) BeginTransaction() (sqlstore.Transaction, error) {
	return m.db.Beginx()
}

func (m *mysql) InTransaction(ctx context.Context, txFunc sqlstore.TransactionFunc) (err error) {
	tx, err := m.db.Beginx()
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = txFunc(sqlstore.NewContext(ctx, tx))
	return
}

func (m *mysql) GetQuerier() sqlstore.Querier {
	return m.db
}
