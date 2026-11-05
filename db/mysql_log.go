package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

	"github.com/chuccp/go-web-frame/log"
)

// loggedMySQLDriverName 是带 SQL 日志的 MySQL 驱动注册名。
const loggedMySQLDriverName = "web-frame-mysql-logged"

var registerLoggedMySQLOnce sync.Once

// openLoggedMySQL 使用包装后的 MySQL driver 打开连接池。
// database/sql 层执行的每条语句（包括 GORM 和直接使用 *sql.DB 的业务代码）
// 都会以 debug 级别写入框架日志，日志级别由 web.log.level 控制。
func openLoggedMySQL(dsn string) (*sql.DB, error) {
	registerLoggedMySQLOnce.Do(func() {
		sql.Register(loggedMySQLDriverName, &loggingDriver{inner: &mysql.MySQLDriver{}})
	})
	return sql.Open(loggedMySQLDriverName, dsn)
}

type loggingDriver struct {
	inner driver.Driver
}

func (d *loggingDriver) Open(dsn string) (driver.Conn, error) {
	conn, err := d.inner.Open(dsn)
	if err != nil {
		return nil, err
	}
	return &loggingConn{Conn: conn}, nil
}

type loggingConn struct {
	driver.Conn
}

func (c *loggingConn) Prepare(query string) (driver.Stmt, error) {
	start := time.Now()
	stmt, err := c.Conn.Prepare(query)
	logSQL(query, nil, time.Since(start), err)
	if err != nil {
		return nil, err
	}
	return &loggingStmt{Stmt: stmt, query: query}, nil
}

func (c *loggingConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	start := time.Now()
	var (
		stmt driver.Stmt
		err  error
	)
	if preparer, ok := c.Conn.(driver.ConnPrepareContext); ok {
		stmt, err = preparer.PrepareContext(ctx, query)
	} else {
		stmt, err = c.Conn.Prepare(query)
	}
	logSQL(query, nil, time.Since(start), err)
	if err != nil {
		return nil, err
	}
	return &loggingStmt{Stmt: stmt, query: query}, nil
}

func (c *loggingConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	execer, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	start := time.Now()
	result, err := execer.ExecContext(ctx, query, args)
	logSQL(query, namedToAny(args), time.Since(start), err)
	return result, err
}

func (c *loggingConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	queryer, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	start := time.Now()
	rows, err := queryer.QueryContext(ctx, query, args)
	logSQL(query, namedToAny(args), time.Since(start), err)
	return rows, err
}

func (c *loggingConn) Begin() (driver.Tx, error) {
	start := time.Now()
	tx, err := c.Conn.Begin()
	logSQL("BEGIN", nil, time.Since(start), err)
	if err != nil {
		return nil, err
	}
	return &loggingTx{Tx: tx}, nil
}

func (c *loggingConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	start := time.Now()
	var (
		tx  driver.Tx
		err error
	)
	if beginner, ok := c.Conn.(driver.ConnBeginTx); ok {
		tx, err = beginner.BeginTx(ctx, opts)
	} else {
		tx, err = c.Conn.Begin()
	}
	logSQL("BEGIN", nil, time.Since(start), err)
	if err != nil {
		return nil, err
	}
	return &loggingTx{Tx: tx}, nil
}

func (c *loggingConn) CheckNamedValue(value *driver.NamedValue) error {
	if checker, ok := c.Conn.(driver.NamedValueChecker); ok {
		return checker.CheckNamedValue(value)
	}
	return driver.ErrSkip
}

func (c *loggingConn) Ping(ctx context.Context) error {
	if pinger, ok := c.Conn.(driver.Pinger); ok {
		return pinger.Ping(ctx)
	}
	return nil
}

func (c *loggingConn) ResetSession(ctx context.Context) error {
	if resetter, ok := c.Conn.(driver.SessionResetter); ok {
		return resetter.ResetSession(ctx)
	}
	return nil
}

func (c *loggingConn) IsValid() bool {
	if validator, ok := c.Conn.(driver.Validator); ok {
		return validator.IsValid()
	}
	return true
}

type loggingStmt struct {
	driver.Stmt
	query string
}

func (s *loggingStmt) Exec(args []driver.Value) (driver.Result, error) {
	start := time.Now()
	result, err := s.Stmt.Exec(args)
	logSQL(s.query, valuesToAny(args), time.Since(start), err)
	return result, err
}

func (s *loggingStmt) Query(args []driver.Value) (driver.Rows, error) {
	start := time.Now()
	rows, err := s.Stmt.Query(args)
	logSQL(s.query, valuesToAny(args), time.Since(start), err)
	return rows, err
}

func (s *loggingStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	start := time.Now()
	var (
		result driver.Result
		err    error
	)
	if execer, ok := s.Stmt.(driver.StmtExecContext); ok {
		result, err = execer.ExecContext(ctx, args)
	} else {
		result, err = s.Stmt.Exec(namedToValues(args))
	}
	logSQL(s.query, namedToAny(args), time.Since(start), err)
	return result, err
}

func (s *loggingStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	start := time.Now()
	var (
		rows driver.Rows
		err  error
	)
	if queryer, ok := s.Stmt.(driver.StmtQueryContext); ok {
		rows, err = queryer.QueryContext(ctx, args)
	} else {
		rows, err = s.Stmt.Query(namedToValues(args))
	}
	logSQL(s.query, namedToAny(args), time.Since(start), err)
	return rows, err
}

type loggingTx struct {
	driver.Tx
}

func (t *loggingTx) Commit() error {
	start := time.Now()
	err := t.Tx.Commit()
	logSQL("COMMIT", nil, time.Since(start), err)
	return err
}

func (t *loggingTx) Rollback() error {
	start := time.Now()
	err := t.Tx.Rollback()
	logSQL("ROLLBACK", nil, time.Since(start), err)
	return err
}

func logSQL(query string, args []any, cost time.Duration, err error) {
	fields := []zap.Field{
		zap.String("sql", compactSQL(query)),
		zap.String("cost", cost.Round(time.Microsecond).String()),
	}
	if len(args) > 0 {
		fields = append(fields, zap.String("args", summarizeArgs(args)))
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	log.Debug("SQL", fields...)
}

// compactSQL 把多行 SQL 压成一行，便于日志检索。
func compactSQL(query string) string {
	return strings.Join(strings.Fields(query), " ")
}

// summarizeArgs 只保留参数的可读摘要，避免项目文档等大字段刷屏。
func summarizeArgs(args []any) string {
	const maxArgBytes = 240
	const maxTotalBytes = 1200
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		var text string
		switch value := arg.(type) {
		case nil:
			text = "NULL"
		case []byte:
			text = fmt.Sprintf("[bytes %d]", len(value))
		case string:
			text = value
		default:
			text = fmt.Sprint(value)
		}
		parts = append(parts, truncateText(text, maxArgBytes))
	}
	return truncateText(strings.Join(parts, ", "), maxTotalBytes)
}

func truncateText(text string, max int) string {
	if len(text) <= max {
		return text
	}
	// 按字节截断后再转成 rune，保证不会把日志写得过长。
	return string([]rune(text[:max])) + "…(" + strconv.Itoa(len(text)) + " bytes)"
}

func valuesToAny(values []driver.Value) []any {
	out := make([]any, len(values))
	for i, value := range values {
		out[i] = value
	}
	return out
}

func namedToValues(args []driver.NamedValue) []driver.Value {
	values := make([]driver.Value, len(args))
	for i, arg := range args {
		values[i] = arg.Value
	}
	return values
}

func namedToAny(args []driver.NamedValue) []any {
	out := make([]any, len(args))
	for i, arg := range args {
		out[i] = arg.Value
	}
	return out
}
