package persist

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	. "mykit/core/dsp"
	. "mykit/core/types"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

var (
	sqlxGlobalUnsafe    = false
	sqlxGlobalIgnoreLog = false
	sqlxDebugSwitch     = false
)

func SetSqlxGlobalUnsafe(raw bool) {
	sqlxGlobalUnsafe = raw
}

func SetSqlxGlobalIgnoreLog(raw bool) {
	sqlxGlobalIgnoreLog = raw
}

func SetSqlxDebug(raw ...bool) {
	sqlxDebugSwitch = ParseBool(raw)
}

type SqlInterface interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	NamedExec(query string, arg interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type TransactionParticipant interface {
	TX(...DbContext) DbContext
	Begin(msg string, n ...TransactionParticipant)
	Commit(err *error)
	Next(n ...TransactionParticipant)
	With(ctx context.Context, tx *sqlx.Tx, step *uint64)
}

type SqlxContext struct {
	ctx  context.Context
	step *uint64

	table  string
	unsafe bool
	db     *sqlx.DB

	ctl DbContext

	transaction bool
	tx          *sqlx.Tx

	msg     string
	initMsg string
	log     bool
	*ZLogger
}

func NewSqlxContext(ctx context.Context, db *sqlx.DB, item SqlModel, table ...string) *SqlxContext {
	tableName := ParseStrParam(table, item.TableName())
	if tableName == "" {
		msg := fmt.Sprintf("%v table name is empty", item)
		HandleInitErr("NewSqlxContext", errors.New(msg))
	}

	res := &SqlxContext{}
	res.Init(ctx, db, item, tableName)

	return res
}

func (t *SqlxContext) Init(ctx context.Context, db *sqlx.DB, item interface{}, table string) {
	t.ctx = ctx
	t.step = new(uint64)

	t.table = table
	t.unsafe = false
	t.db = db

	ctl, ok := item.(CtlModel)
	if ok {
		t.ctl = ctl.CtlWhere()
	} else {
		t.ctl = DbContext{}
	}

	t.transaction = false
	t.tx = nil

	t.msg = DeStrParam(GetMethod(ctx), "SqlxContext")
	t.initMsg = t.msg
	t.log = true

	t.ZLogger = NewZLoggerWithFields(ctx, t.msg, LogEventSql())
}

func (t *SqlxContext) IgnoreLog() {
	t.log = false
}

func (t *SqlxContext) DevDebug() {
	t.log = sqlxDebugSwitch
}

func (t *SqlxContext) SetTable(table string) {
	t.table = table
}

func (t *SqlxContext) Table() string {
	return t.table
}

func (t *SqlxContext) SetCtl(where ...DbContext) *SqlxContext {
	t.ctl.Merge(where...)

	return t
}

func (t *SqlxContext) Where(where ...DbContext) DbContext {
	return t.ctl.Clone().Merge(where...)
}

func (t *SqlxContext) Transaction() TransactionParticipant {
	return TransactionParticipant(t)
}

func (t *SqlxContext) Db() SqlInterface {
	t.Forward()

	if !t.transaction {
		if t.isUnsafe() {
			return t.db.Unsafe()
		}

		return t.db
	}

	return t.tx
}

func (t *SqlxContext) TX(raw ...DbContext) DbContext {
	if len(raw) == 0 {
		return DbContext{tagTx: t.Transaction()}
	}

	return raw[0].Set(tagTx, t.Transaction())
}

func (t *SqlxContext) Begin(msg string, n ...TransactionParticipant) {
	t.Reset()

	t.msg = msg
	t.transaction = true

	if t.isUnsafe() {
		t.tx = t.db.Unsafe().MustBegin()
	} else {
		t.tx = t.db.MustBegin()
	}

	t.Next(n...)
}

func (t *SqlxContext) Commit(err *error) {
	var panicErr error
	if r := recover(); r != nil {
		var ok bool
		panicErr, ok = r.(error)
		if !ok {
			panicErr = fmt.Errorf("%v", r)
		}

		t.Failed(
			LogEvent("panic occur @ commit"),
			LogProcessor(t.msg),
			LogStep(t.Step()),
			LogError(panicErr),
		)
	}

	if *err != nil {
		t.Failed(
			LogEvent("error occur @ commit"),
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
			LogEvent("rollback @ commit"),
			LogProcessor(t.msg),
			LogStep(t.Step()),
		)
		_ = t.tx.Rollback()
	} else {
		_ = t.tx.Commit()
	}

	t.Reset()

	return
}

func (t *SqlxContext) Next(n ...TransactionParticipant) {
	if t.transaction {
		for _, v := range n {
			v.With(t.ctx, t.tx, t.step)
		}
	}
}

func (t *SqlxContext) With(ctx context.Context, tx *sqlx.Tx, step *uint64) {
	t.ctx = ctx
	t.tx = tx
	t.transaction = true
	t.step = step
}

func (t *SqlxContext) Insert(data interface{}, ig ...bool) (err error) {
	if !IsSlice(data) {
		_, err = t.insertOneAndGetResult(1, data, ig...)
		return err
	}

	_, err = t.insertManyAndGetResult(1, data, ig...)

	return err
}

func (t *SqlxContext) InsertAndGetResult(data interface{}, ig ...bool) (result sql.Result, err error) {
	if !IsSlice(data) {
		result, err = t.insertOneAndGetResult(1, data, ig...)
		return
	}

	result, err = t.insertManyAndGetResult(1, data, ig...)
	return
}

func (t *SqlxContext) InsertOneAndGetLastInsertId(data interface{}, ig ...bool) (Id int64, err error) {
	res, err := t.insertOneAndGetResult(1, data, ig...)
	if err != nil {
		return
	}

	Id, err = res.LastInsertId()

	return
}

func (t *SqlxContext) Delete(where DbContext) (rowsAffected int64, err error) {
	sqlStr, args, err := t.BuildDelete(where)
	if err != nil {
		return -1, err
	}

	result, err := t.exec(1, sqlStr, args...)
	if err != nil {
		return -2, err
	}

	rowsAffected, _ = result.RowsAffected()

	return
}

func (t *SqlxContext) Find(data interface{}, where DbContext, selectField ...string) (err error) {
	sqlStr, args, err := t.BuildSelect(where, selectField...)
	if err != nil {
		return err
	}

	if !IsSliceElem(data) {
		_, err = t.getRow(1, data, sqlStr, args...)
		return err
	}

	err = t.getList(1, data, sqlStr, args...)

	return
}

func (t *SqlxContext) FindRes(data interface{}, where DbContext, selectField ...string) (hasData bool, err error) {
	sqlStr, args, err := t.BuildSelect(where, selectField...)
	if err != nil {
		return false, err
	}

	if !IsSliceElem(data) {
		hasData, err = t.getRow(1, data, sqlStr, args...)
		return
	}

	err = t.getList(1, data, sqlStr, args...)
	if err == nil {
		s := reflect.ValueOf(data)
		for s.Kind() == reflect.Ptr {
			s = s.Elem()
		}
		hasData = s.Len() > 0
	}

	return
}

func (t *SqlxContext) FindItem(data DbItem, where ...DbContext) (hasData bool, err error) {
	if data == nil {
		err = ErrInvalidParam
		return
	}

	q := ParseDbContextParam(where)
	sqlStr, args, err := t.BuildSelect(q, data.Select()...)
	if err != nil {
		return false, err
	}

	if !IsSliceElem(data) {
		hasData, err = t.getRow(1, data, sqlStr, args...)
		return
	}

	err = t.getList(1, data, sqlStr, args...)
	if err == nil {
		s := reflect.ValueOf(data)
		for s.Kind() == reflect.Ptr {
			s = s.Elem()
		}
		hasData = s.Len() > 0
	}

	return
}

func (t *SqlxContext) FindForUpdate(data interface{}, where DbContext, selectField ...string) (err error) {
	if len(selectField) == 0 {
		selectField = []string{IdDb}
	}

	sqlStr, args, err := t.BuildSelect(where, selectField...)
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString(sqlStr)
	b.WriteString(" For Update")

	if !IsSliceElem(data) {
		_, err = t.getRow(1, data, b.String(), args...)
		return err
	}

	err = t.getList(1, data, b.String(), args...)

	return
}

func (t *SqlxContext) Get(data interface{}, sqlStr string, args ...interface{}) (err error) {
	if !IsSliceElem(data) {
		_, err = t.getRow(1, data, sqlStr, args...)
		return
	}

	err = t.getList(1, data, sqlStr, args...)

	return
}

func (t *SqlxContext) GetRes(data interface{}, sqlStr string, args ...interface{}) (hasData bool, err error) {
	if !IsSliceElem(data) {
		return t.getRow(1, data, sqlStr, args...)
	}

	err = t.getList(1, data, sqlStr, args...)
	s := reflect.ValueOf(data)
	s = DeValue(s)
	hasData = s.Len() > 0

	return
}

func (t *SqlxContext) Update(where, update DbContext) (rowsAffected int64, err error) {
	q := where.FilterBlackFiled()
	if len(q) == 0 {
	}

	sqlStr, args, err := t.BuildUpdate(where, update)
	if err != nil {
		return -1, err
	}

	result, err := t.exec(1, sqlStr, args...)
	if err != nil {
		return -2, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		rowsAffected = -3
	}

	return
}

func (t *SqlxContext) DoUpdate(c ...UpdateContext) (rowsAffected int64, err error) {
	sqlStr := ""
	var args []interface{}
	var result sql.Result
	var n int64

	for _, v := range c {
		q := v.Query.FilterBlackFiled()
		if len(q) == 0 {
			continue
		}

		sqlStr, args, err = t.BuildUpdate(q, v.Update.UpdateAt())
		if err != nil {
			return
		}

		result, err = t.exec(1, sqlStr, args...)
		if err != nil {
			return
		}

		n, err = result.RowsAffected()
		if err == nil {
			rowsAffected += n
		}
	}

	return
}

func (t *SqlxContext) Add(data ToAdd, tick ...int64) (id int64, err error) {
	dao := data.ToAdd(t.Ctx())
	if dao == nil {
		err = ErrInvalidParam
		return
	}

	dao.Create(tick...)

	res, err := t.insertOneAndGetResult(1, dao)
	if err != nil {
		if IsMysqlDuplicateKey(err) {
			err = NewError("重复数据冲突")
		}

		return
	}

	id, err = res.LastInsertId()

	return
}

func (t *SqlxContext) AddSome(raw interface{}, tick ...int64) (id int64, err error) {
	toAdd := []Dao{}

	data := InterfaceToSliceInterface(raw)
	for _, v := range data {
		err = ValidateStruct(v)
		if err != nil {
			err = NewErrorCode(
				err.Error(),
				-36,
			)
			return
		}

		addItem, ok := v.(ToAdd)
		if !ok {
			continue
		}

		dao := addItem.ToAdd(t.Ctx())
		if dao == nil {
			continue
		}

		dao.Create(tick...)
		toAdd = append(toAdd, dao)
	}

	if len(toAdd) == 0 {
		return -1, ErrInvalidParam
	}

	res, err := t.insertManyAndGetResult(1, toAdd)
	if err != nil {
		if IsMysqlDuplicateKey(err) {
			err = NewError("重复数据冲突")
		}

		return
	}

	id, err = res.LastInsertId()

	return
}

func (t *SqlxContext) AfterAdd() CrudDecorator {
	return nil
}

func (t *SqlxContext) Create(data Dao, tick ...int64) (Id int64, err error) {
	if data == nil {
		err = ErrInvalidParam
		return
	}

	data.Create(tick...)

	res, err := t.insertOneAndGetResult(1, data)
	if err != nil {
		return
	}

	Id, err = res.LastInsertId()

	return
}

func (t *SqlxContext) Set(data ToSet, tick ...int64) (rowsAffected int64, err error) {
	if data == nil {
		return -11, ErrorCodeInvalidParam
	}

	q := data.Where(t.Ctx()).FilterBlackFiled()
	if len(q) == 0 {
		return -12, ErrorCodeInvalidParam
	}

	u := data.
		ToSet(t.Ctx()).
		UpdateAt(tick...)

	sqlStr, args, err := t.BuildUpdate(q, u)
	if err != nil {
		return -1, err
	}

	result, err := t.exec(1, sqlStr, args...)
	if err != nil {
		return -2, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		rowsAffected = -3
	}

	return
}

func (t *SqlxContext) AfterSet() CrudDecorator {
	return nil
}

func (t *SqlxContext) Put(data ToPut, tick ...int64) (err error) {
	ctx := t.Ctx()
	q := data.Where(ctx)
	if len(q) == 0 {
		dao := data.ToAdd(ctx)
		dao.Create(tick...)

		return t.Insert(dao)
	}

	id, err := t.FindId(q)
	if err != nil {
		return
	}

	if id == 0 {
		dao := data.ToAdd(ctx)
		dao.Create(tick...)

		return t.Insert(dao)
	}

	_, err = t.Set(data, tick...)

	return
}

func (t *SqlxContext) PageQuery(data interface{}, where DbContext, selectField ...string) (result *PageQueryRes) {
	result = NewPageQueryRes()

	var countStr string
	var queryStr string
	var args []interface{}

	where = where.FilterBlackFiled().PageQuery()
	countWhere := where.Clone()

	v, ok := countWhere[TagLimit]
	if ok {
		limit, toCount := v.([]uint)
		if toCount && len(limit) > 0 {
			delete(countWhere, TagLimit)

			countStr, args, result.Err = t.BuildSelect(countWhere, SelectCount())
			if result.Err != nil {
				return
			}

			_, result.Err = t.getRow(1, &result.Total, countStr, args...)
			if result.Err != nil || result.Total == 0 {
				return
			}
		}
	}

	queryStr, args, result.Err = t.BuildSelect(where, selectField...)
	if result.Err != nil {
		return
	}

	result.Err = t.getList(0, data, queryStr, args...)
	if result.Err != nil {
		return
	}

	if len(where) == 0 {
		s := reflect.ValueOf(data)
		s = DeValue(s)
		result.Total = int64(s.Len())
	}

	result.Result = data

	return
}

func (t *SqlxContext) PageGet(data interface{}, page PageQueryReq, sqlStr string, args ...interface{}) (result *PageQueryRes) {
	result = NewPageQueryRes()

	countStr := SqlCountStrForSub(sqlStr)
	_, result.Err = t.getRow(1, &result.Total, countStr, args...)
	if result.Err != nil || result.Total == 0 {
		return
	}

	queryStr := page.Query(sqlStr)
	result.Err = t.getList(0, data, queryStr, args...)
	if result.Err != nil {
		return
	}

	result.Result = data

	return
}

func (t *SqlxContext) PullPage(data interface{}, req PullPageReq, where DbContext, selectField ...string) (err error) {
	var id int64
	if req.Uuid != "" {
		id, err = t.FindIdByUuid(req.Uuid)
		if err != nil {
			return
		}
	}

	q := DbContext{
		"id >": id,
	}.Limit(req.Size)

	q = t.ctl.Merge(q, where)

	err = t.Find(data, q, selectField...)

	return
}

func (t *SqlxContext) Query(data interface{}, sql string, q DbContext) (err error) {
	sqlStr, args, err := t.BuildQuery(sql, q)
	if err != nil {
		return err
	}

	if !IsSliceElem(data) {
		_, err = t.getRow(1, data, sqlStr, args...)
		return err
	}

	err = t.getList(1, data, sqlStr, args...)

	return
}

func (t *SqlxContext) AfterDel() CrudDecorator {
	return nil
}
