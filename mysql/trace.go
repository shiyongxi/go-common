package mysql

import (
	"fmt"
	"golang.org/x/net/context"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	parentSpanGormKey = "opentracingParentSpan"
	spanGormKey       = "opentracingSpan"
)

func SetSpanToGorm(ctx context.Context, db *gorm.DB) *gorm.DB {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		return db.Set(parentSpanGormKey, span)
	}

	return db
}

func AddGormCallbacks(db *gorm.DB, trace opentracing.Tracer) {
	callbacks := newCallbacks(trace)

	registerCallbacks(db, "create", callbacks)
	registerCallbacks(db, "query", callbacks)
	registerCallbacks(db, "update", callbacks)
	registerCallbacks(db, "delete", callbacks)
	registerCallbacks(db, "row_query", callbacks)
}

type callbacks struct {
	trace opentracing.Tracer
}

func newCallbacks(trace opentracing.Tracer) *callbacks {
	return &callbacks{
		trace: trace,
	}
}

func (c *callbacks) beforeCreate(scope *gorm.Scope) { c.before(scope) }

func (c *callbacks) afterCreate(scope *gorm.Scope) { c.after(scope, "INSERT") }

func (c *callbacks) beforeQuery(scope *gorm.Scope) { c.before(scope) }

func (c *callbacks) afterQuery(scope *gorm.Scope) { c.after(scope, "SELECT") }

func (c *callbacks) beforeUpdate(scope *gorm.Scope) { c.before(scope) }

func (c *callbacks) afterUpdate(scope *gorm.Scope) { c.after(scope, "UPDATE") }

func (c *callbacks) beforeDelete(scope *gorm.Scope) { c.before(scope) }

func (c *callbacks) afterDelete(scope *gorm.Scope) { c.after(scope, "DELETE") }

func (c *callbacks) beforeRowQuery(scope *gorm.Scope) { c.before(scope) }

func (c *callbacks) afterRowQuery(scope *gorm.Scope) { c.after(scope, "") }

func (c *callbacks) before(scope *gorm.Scope) {
	val, ok := scope.Get(parentSpanGormKey)
	if !ok {
		return
	}

	span := c.trace.StartSpan("SQL", opentracing.ChildOf(val.(opentracing.Span).Context()))
	ext.DBType.Set(span, "sql")

	scope.Set(spanGormKey, span)
}

func (c *callbacks) after(scope *gorm.Scope, operation string) {
	val, ok := scope.Get(spanGormKey)
	if !ok {
		return
	}

	span := val.(opentracing.Span)
	if operation == "" {
		operation = strings.ToUpper(strings.Split(scope.SQL, " ")[0])
	}

	ext.Error.Set(span, scope.HasError())
	ext.DBStatement.Set(span, scope.SQL)
	span.SetTag("db.table", scope.TableName())
	span.SetTag("db.method", operation)
	span.SetTag("db.err", scope.HasError())
	span.SetTag("db.count", scope.DB().RowsAffected)

	span.Finish()
}

func registerCallbacks(db *gorm.DB, name string, c *callbacks) {
	beforeName := fmt.Sprintf("tracing:%v_before", name)
	afterName := fmt.Sprintf("tracing:%v_after", name)
	gormCallbackName := fmt.Sprintf("gorm:%v", name)

	switch name {
	case "create":
		db.Callback().Create().Before(gormCallbackName).Register(beforeName, c.beforeCreate)
		db.Callback().Create().After(gormCallbackName).Register(afterName, c.afterCreate)

	case "query":
		db.Callback().Query().Before(gormCallbackName).Register(beforeName, c.beforeQuery)
		db.Callback().Query().After(gormCallbackName).Register(afterName, c.afterQuery)

	case "update":
		db.Callback().Update().Before(gormCallbackName).Register(beforeName, c.beforeUpdate)
		db.Callback().Update().After(gormCallbackName).Register(afterName, c.afterUpdate)

	case "delete":
		db.Callback().Delete().Before(gormCallbackName).Register(beforeName, c.beforeDelete)
		db.Callback().Delete().After(gormCallbackName).Register(afterName, c.afterDelete)

	case "row_query":
		db.Callback().RowQuery().Before(gormCallbackName).Register(beforeName, c.beforeRowQuery)
		db.Callback().RowQuery().After(gormCallbackName).Register(afterName, c.afterRowQuery)
	}
}
