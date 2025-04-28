package persist

import (
	"context"
	"database/sql"
	"fmt"
	. "mykit/core/dsp"
	. "mykit/core/types"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/go-xorm/xorm"
	"go.uber.org/zap"
)

type XormTransactionParticipant interface {
	With(ctx context.Context, tx *xorm.Session, step *uint64)
}

type XormContext struct {
	ctx  context.Context
	step *uint64

	table string
	*xorm.Engine

	transaction bool
	*xorm.Session

	msg     string
	initMsg string
	log     bool
	*ZLogger
}

func NewXormContext(ctx context.Context, e *xorm.Engine, table string) *XormContext {
	res := &XormContext{
		ctx:         ctx,
		step:        new(uint64),
		table:       table,
		Engine:      e,
		transaction: false,
		Session:     nil,
		msg:         DeStrParam(GetMethod(ctx), "XormContext"),
		log:         true,
	}

	res.initMsg = res.msg

	res.ZLogger = NewZLoggerWithFields(ctx, res.msg, LogEventSql())

	return res
}

func (t *XormContext) IgnoreLog() {
	t.log = false
}

func (t *XormContext) output(k int, cost time.Duration, err error, f ...zap.Field) {
	k++

	f = append(f, LogDuration(cost))

	logger := t.Skip(k)
	if err != nil {
		f = append(f, LogError(err))
		logger.Error(t.msg, f...)
		return
	}

	if cost > sqlQuerySlowThreshold {
		ZapSlow(logger, t.msg, f...)

	} else {
		if !sqlxGlobalIgnoreLog && t.log {
			logger.Info(t.msg, f...)
		}
	}
}

func (t *XormContext) session() *xorm.Session {
	t.Forward()

	if t.Session == nil {
		t.Session = t.Engine.Table(t.table)
	} else {
		t.Session.Table(t.table)
	}

	return t.Session
}

func (t *XormContext) Ctx() context.Context {
	return t.ctx
}

func (t *XormContext) SetCtx(ctx context.Context) {
	t.ctx = ctx
}

func (t *XormContext) Step() uint64 {
	return atomic.LoadUint64(t.step)
}

func (t *XormContext) Forward() {
	atomic.AddUint64(t.step, 1)
}

func (t *XormContext) NewSession() *XormContext {
	t.Session = t.Engine.Table(t.table)

	return t
}

func (t *XormContext) Close() {
	t.session().Close()
}

func (t *XormContext) Begin(msg string, n ...XormTransactionParticipant) error {
	t.msg = msg
	t.transaction = true

	t.Session = t.Engine.NewSession().Table(t.table)
	err := t.Session.Begin()
	if err != nil {
		return err
	}

	t.Next(n...)

	return nil
}

func (t *XormContext) Rollback() error {
	if !t.transaction {
		return nil
	}

	return t.session().Rollback()
}

func (t *XormContext) Commit(err *error) {
	var panicErr error
	if r := recover(); r != nil {
		var ok bool
		panicErr, ok = r.(error)
		if !ok {
			panicErr = fmt.Errorf("%v", r)
		}

		t.Failed(
			LogEvent("panic occur @ xorm commit"),
			LogProcessor(t.msg),
			LogStep(t.Step()),
			LogError(panicErr),
		)
	}

	if *err != nil {
		t.Failed(
			LogEvent("error occur @ xorm commit"),
			LogProcessor(t.msg),
			LogStep(t.Step()),
			LogError(*err),
		)
	}

	if !t.transaction {
		return
	}

	if *err != nil || panicErr != nil {
		t.Failed(
			LogEvent("rollback @ xorm commit"),
			LogProcessor(t.msg),
			LogStep(t.Step()),
		)
		_ = t.session().Rollback()
	} else {
		_ = t.session().Commit()
	}

	return
}

func (t *XormContext) Next(n ...XormTransactionParticipant) {
	if t.transaction {
		for _, v := range n {
			v.With(t.ctx, t.Session, t.step)
		}
	}
}

func (t *XormContext) With(ctx context.Context, tx *xorm.Session, step *uint64) {
	t.ctx = ctx
	t.Session = tx
	t.transaction = true
	t.step = step
}

func (t *XormContext) Where(query interface{}, args ...interface{}) *XormContext {
	t.Session = t.session().Where(query, args...)

	return t
}

func (t *XormContext) Sql(query interface{}, args ...interface{}) *XormContext {
	t.Session = t.session().SQL(query, args)

	return t
}

func (t *XormContext) And(query interface{}, args ...interface{}) *XormContext {
	t.Session = t.session().And(query, args...)

	return t
}

func (t *XormContext) Or(query interface{}, args ...interface{}) *XormContext {
	t.Session = t.session().Or(query, args...)

	return t
}

func (t *XormContext) Id(id interface{}) *XormContext {
	t.Session = t.session().ID(id)

	return t
}

func (t *XormContext) In(column string, args ...interface{}) *XormContext {
	t.Session = t.session().In(column, args...)

	return t
}

func (t *XormContext) NotIn(column string, args ...interface{}) *XormContext {
	t.Session = t.session().NotIn(column, args...)

	return t
}

func (t *XormContext) Asc(colNames ...string) *XormContext {
	t.Session = t.session().Asc(colNames...)

	return t
}

func (t *XormContext) Desc(colNames ...string) *XormContext {
	t.Session = t.session().Desc(colNames...)

	return t
}

func (t *XormContext) Limit(limit int, start ...int) *XormContext {
	t.Session = t.session().Limit(limit, start...)

	return t
}

func (t *XormContext) OrderBy(order string) *XormContext {
	t.Session = t.session().OrderBy(order)

	return t
}

func (t *XormContext) Insert(beans ...interface{}) (rowsAffected int64, err error) {
	l := len(beans)
	if l == 0 {
		return
	}

	sess := t.session()

	t0 := time.Now()
	rowsAffected, err = sess.Insert(beans...)
	cost := time.Now().Sub(t0)
	sqlStr, _ := sess.LastSQL()

	detail := map[string]interface{}{
		logFiledSql:          sqlStr,
		logFiledArgNum:       l,
		logFiledRowsAffected: rowsAffected,
	}

	f := []zap.Field{
		LogProcessor("Xorm.Insert"),
		LogDetail(detail),
	}

	t.output(1, cost, err, f...)

	return
}

func (t *XormContext) InsertOne(bean interface{}) (rowsAffected int64, err error) {
	sess := t.session()

	t0 := time.Now()
	rowsAffected, err = sess.InsertOne(bean)
	cost := time.Now().Sub(t0)

	sqlStr, _ := sess.LastSQL()

	detail := map[string]interface{}{
		logFiledSql:          sqlStr,
		logFiledRowsAffected: rowsAffected,
	}

	f := []zap.Field{
		LogProcessor("Xorm.InsertOne"),
		LogDetail(detail),
	}

	t.output(1, cost, err, f...)

	return
}

func (t *XormContext) Delete(bean interface{}) (n int64, err error) {
	sess := t.session()

	t0 := time.Now()
	n, err = sess.Delete(bean)
	cost := time.Now().Sub(t0)

	sqlStr, _ := sess.LastSQL()

	detail := map[string]interface{}{
		logFiledSql:          sqlStr,
		logFiledRowsAffected: n,
	}

	f := []zap.Field{
		LogProcessor("Xorm.Delete"),
		LogDetail(detail),
	}

	t.output(1, cost, err, f...)

	return
}

func (t *XormContext) Get(bean interface{}) (hasData bool, err error) {
	sess := t.session()

	t0 := time.Now()
	hasData, err = sess.Get(bean)
	cost := time.Now().Sub(t0)

	sqlStr, args := sess.LastSQL()
	l := len(args)
	//n := MinInt(l, maxArgNum)

	detail := map[string]interface{}{
		logFiledSql:    sqlStr,
		logFiledArgNum: l,
		logFiledArgs:   args,
		logFiledHas:    hasData,
	}

	f := []zap.Field{
		LogProcessor("Xorm.Get"),
		LogDetail(detail),
	}

	t.output(1, cost, err, f...)

	return
}

func (t *XormContext) Find(rowsSlicePtr interface{}, condiBean ...interface{}) (err error) {
	sess := t.session()

	t0 := time.Now()
	err = sess.Find(rowsSlicePtr, condiBean...)
	cost := time.Now().Sub(t0)

	sqlStr, args := sess.LastSQL()
	l := len(args)
	//n := MinInt(l, maxArgNum)

	s := reflect.ValueOf(rowsSlicePtr)
	s = DeValue(s)

	detail := map[string]interface{}{
		logFiledSql:    sqlStr,
		logFiledArgNum: l,
		logFiledArgs:   args,
		logFiledHas:    s.Len() > 0,
		logFiledResNum: s.Len(),
	}

	f := []zap.Field{
		LogProcessor("Xorm.Find"),
		LogDetail(detail),
	}

	t.output(1, cost, err, f...)

	return err
}

func (t *XormContext) FindAndCount(rowsSlicePtr interface{}, condiBean ...interface{}) (count int64, err error) {
	sess := t.session()

	t0 := time.Now()
	count, err = sess.FindAndCount(rowsSlicePtr, condiBean...)
	cost := time.Now().Sub(t0)

	sqlStr, args := sess.LastSQL()
	l := len(args)
	//n := MinInt(l, maxArgNum)

	detail := map[string]interface{}{
		logFiledSql:    sqlStr,
		logFiledArgNum: l,
		logFiledArgs:   args,
		logFiledHas:    count > 0,
		logFiledResNum: count,
	}

	f := []zap.Field{
		LogProcessor("Xorm.FindAndCount"),
		LogDetail(detail),
	}

	t.output(1, cost, err, f...)

	return
}

func (t *XormContext) Count(bean ...interface{}) (count int64, err error) {
	sess := t.session()

	t0 := time.Now()
	count, err = sess.Count(bean...)
	cost := time.Now().Sub(t0)

	sqlStr, args := sess.LastSQL()
	l := len(args)
	//n := MinInt(l, maxArgNum)

	detail := map[string]interface{}{
		logFiledSql:    sqlStr,
		logFiledArgNum: l,
		logFiledArgs:   args,
		logFiledHas:    count > 0,
		logFiledResNum: count,
	}

	f := []zap.Field{
		LogProcessor("Xorm.Count"),
		LogDetail(detail),
	}

	t.output(1, cost, err, f...)

	return
}

func (t *XormContext) Update(bean interface{}, condiBean ...interface{}) (rowsAffected int64, err error) {
	sess := t.session()

	t0 := time.Now()
	rowsAffected, err = sess.Update(bean, condiBean...)
	cost := time.Now().Sub(t0)

	sqlStr, args := sess.LastSQL()
	l := len(args)
	//n := MinInt(l, maxArgNum)

	detail := map[string]interface{}{
		logFiledSql:          sqlStr,
		logFiledArgNum:       l,
		logFiledArgs:         args,
		logFiledRowsAffected: rowsAffected,
	}

	f := []zap.Field{
		LogProcessor("Xorm.Update"),
		LogDetail(detail),
	}

	t.output(1, cost, err, f...)

	return
}

func (t *XormContext) Exec(sqlOrArgs ...interface{}) (res sql.Result, err error) {
	sess := t.session()

	t0 := time.Now()
	res, err = sess.Exec(sqlOrArgs...)
	cost := time.Now().Sub(t0)

	sqlStr, args := sess.LastSQL()
	l := len(args)
	//n := MinInt(l, maxArgNum)

	detail := map[string]interface{}{
		logFiledSql:    sqlStr,
		logFiledArgNum: l,
		logFiledArgs:   args,
	}

	f := []zap.Field{
		LogProcessor("Xorm.Exec"),
		LogDetail(detail),
	}

	t.output(1, cost, err, f...)

	return
}

func (t *XormContext) Cols(columns ...string) *XormContext {
	t.Session = t.session().Cols(columns...)

	return t
}

func (t *XormContext) Exist(bean ...interface{}) (bool, error) {
	return t.session().Exist(bean...)
}

func (t *XormContext) SQL(query interface{}, args ...interface{}) *XormContext {
	t.Session = t.session().SQL(query, args...)
	return t
}
