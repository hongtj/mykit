package persist

import (
	"context"
	"database/sql"
	"encoding/json"
	. "mykit/core/dsp"
	. "mykit/core/types"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"github.com/didi/gendry/builder"
	"github.com/google/uuid"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func DeleteHelpTable(raw string) string {
	return "zzz_" + raw
}

func ParseInfluxPoint(table string, data ...InfluxPoint) []*client.Point {
	res := []*client.Point{}

	t := time.Now()

	for _, v := range data {
		t = t.Add(time.Nanosecond)
		item := v.ToPoint(table, t)

		res = append(res, item)
	}

	return res
}

func DbContextFromDbItem(raw interface{}, ignoreFiled ...string) DbContext {
	var res DbContext
	if v, ok := raw.(DbContext); ok {
		res = v
	} else {
		res = TransStructToDbContextByTag(raw, TagDb)
	}

	delete(res, IdDb)
	for _, v := range ignoreFiled {
		delete(res, v)
	}

	return res
}

func DbContextFromBff(raw interface{}, ignoreFiled ...string) DbContext {
	var res DbContext
	if v, ok := raw.(DbContext); ok {
		res = v
	} else {
		res = TransStructToDbContext(raw, TagSet, TagDb, TagJson)
	}

	ignoreFiled = SetStrList(ignoreFiled)
	for _, v := range ignoreFiled {
		delete(res, v)
	}

	return res
}

func TransStructToDbContextByTag(raw interface{}, tag string) DbContext {
	return TransStructToMapStrInterfaceByTag(raw, tag)
}

func TransStructToDbContext(raw interface{}, tag ...string) DbContext {
	if len(tag) == 0 {
		panic("need tag")
	}

	res := map[string]interface{}{}

	t := GetTypeOfStruct(raw)
	v := GetValueOfStruct(raw)

	var k string
	var obj interface{}

	n := v.NumField()
	for i := 0; i < n; i++ {
		obj = v.Field(i).Interface()
		if obj == nil {
			continue
		}

		f := t.Field(i)
		k = GetTag(f.Tag, tag...)
		if k == "" || k == "-" {
			continue
		}

		r, ok := obj.(json.RawMessage)
		if ok {
			obj = EnsureJsonStrFromByte(r)
		}

		res[k] = obj
	}

	return res
}

func TransToDbContext(raw interface{}) (res DbContext) {
	res, ok := raw.(DbContext)
	if ok {
		return res
	}

	tmp := MustJsonMarshal(raw)
	if len(tmp) == 0 {
		return
	}

	dc := DbContext{}
	UnmarshalJson(tmp, &dc)

	res = DbContext{}
	for k, v := range dc {
		if v == nil {
			continue
		}

		res[k] = v
	}

	return
}

func ParseDbContextParam(raw []DbContext) DbContext {
	if len(raw) > 0 && len(raw[0]) > 0 {
		return raw[0]
	}

	return DbContext{}
}

func TransDbItemToSchema(raw interface{}) (schema []string, meta map[int]bool) {
	schema, meta = TransStructToSchema(raw, TagDb)

	return
}

func TransStructToSchema(raw interface{}, tag string) (schema []string, meta map[int]bool) {
	if tag == "" {
		panic("need tag")
	}

	schema = []string{}
	meta = map[int]bool{}

	t := GetTypeOfStruct(raw)
	v := GetValueOfStruct(raw)

	var k string
	var obj interface{}

	n := v.NumField()
	for i := 0; i < n; i++ {
		obj = v.Field(i).Interface()
		if obj == nil {
			continue
		}

		k = t.Field(i).Tag.Get(tag)
		if strings.Contains(k, ",") {
			k = strings.Split(k, ",")[0]
		}

		schema = append(schema, k)

		_, meta[i] = obj.(string)
	}

	return
}

func OffsetInt32(n int32) uint {
	if n < 0 {
		return 0
	}

	return uint(n)
}

func OffsetInt(n int) uint {
	if n < 0 {
		return 0
	}

	return uint(n)
}

func OffsetInt64(n int64) uint {
	if n < 0 {
		return 0
	}

	return uint(n)
}

func SqlAddInt64(n int64) string {
	opp := "+"
	if n < 0 {
		opp = "-"
	}

	return opp + ParseInt64ToStr(AbsInt64(n))
}

func SqlMulInt64(n int64) string {
	return "*" + ParseInt64ToStr(n)
}

func SqlAddFloat64(n float64) string {
	opp := "+"
	if n < 0 {
		opp = "-"
	}

	return opp + ParseFloat64ToStr(AbsFloat64(n))
}

func SqlMulFloat64(n float64) string {
	return "*" + ParseFloat64ToStr(n)
}

func PageCount(count, pagesize int) int {
	if count%pagesize > 0 {
		return count/pagesize + 1
	} else {
		return count / pagesize
	}
}

func StartIndex(page, pagesize int) int {
	if page > 1 {
		return (page - 1) * pagesize
	}
	return 0
}

func CountInOneDayFilter(t time.Time, f string) DbContext {
	t0 := GetZeroTime(t)
	t1 := t0.Add(time.Hour * 24)

	res := DbContext{
		f + " >=": t0.Unix(),
		f + " <":  t1.Unix(),
	}

	return res
}

func (t *SqlxContext) output(k int, cost time.Duration, err error, f ...zap.Field) {
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

func (t *SqlxContext) SetUnsafe(u ...bool) {
	t.unsafe = ParseBool(u)
}

func (t *SqlxContext) isUnsafe() bool {
	return t.unsafe || sqlxGlobalUnsafe
}

func (t *SqlxContext) SetCtx(ctx context.Context) {
	t.ctx = ctx
}

func (t *SqlxContext) Ctx() context.Context {
	return t.ctx
}

func (t *SqlxContext) Step() uint64 {
	return atomic.LoadUint64(t.step)
}

func (t *SqlxContext) Forward() {
	atomic.AddUint64(t.step, 1)
}

func (t *SqlxContext) Reset() {
	t.msg = t.initMsg

	t.transaction = false
	t.tx = nil

	atomic.StoreUint64(t.step, 0)
}

func (t *SqlxContext) BuildInsertSqlNamedStr(uMap DbContext, ig ...bool) (sqlStr string) {
	return BuildInsertSqlNamedStr(t.table, uMap, ig...)
}

func (t *SqlxContext) BuildDelete(where DbContext) (string, []interface{}, error) {
	where = where.FilterBlackFiled()
	sqlStr, args, err := builder.BuildDelete(t.table, where)
	if err != nil {
		t.Errorf("Find %v, BuildDelete err, %v", where, err)
		return "", nil, err
	}

	return sqlStr, args, err
}

func (t *SqlxContext) BuildSelect(where DbContext, selectField ...string) (string, []interface{}, error) {
	where = where.FilterBlackFiled()
	sqlStr, args, err := builder.BuildSelect(t.table, where, selectField)
	if err != nil {
		t.Errorf("Find %v, Select %v, BuildSelect err, %v", where, selectField, err)
		return "", nil, err
	}

	return sqlStr, args, err
}

func (t *SqlxContext) BuildSelectForUpdate(where DbContext, selectField ...string) (string, []interface{}, error) {
	where = where.FilterBlackFiled()
	sqlStr, args, err := t.BuildSelect(where, selectField...)
	if err != nil {
		t.Errorf("Find %v, Select %v, BuildSelect ForUpdate err, %v", where, selectField, err)
		return "", nil, err
	}

	var b strings.Builder
	b.WriteString(sqlStr)
	b.WriteString(" FOR UPDATE")

	return b.String(), args, err
}

func (t *SqlxContext) BuildQuery(sql string, where DbContext) (string, []interface{}, error) {
	where = where.FilterBlackFiled()
	sqlStr, args, err := builder.NamedQuery(sql, where)
	if err != nil {
		t.Errorf("Query %v, BuildQuery err, %v", where, err)
		return "", nil, err
	}

	return sqlStr, args, err
}

func (t *SqlxContext) BuildUpdate(where, update DbContext) (string, []interface{}, error) {
	where = where.FilterBlackFiled()

	update = update.FilterBlackFiled()
	ParseToSet(update)

	sqlStr, args, err := builder.BuildUpdate(t.table, where, update)
	if err != nil {
		t.Errorf("Find %v, Update %v, BuildUpdate err, %v", where, update, err)
		return "", nil, err
	}

	return sqlStr, args, err
}

func (t *SqlxContext) ExecAndGetResult(sqlStr string, args ...interface{}) (result sql.Result, err error) {
	return t.exec(1, sqlStr, args...)
}

func (t *SqlxContext) ExecAndGetLastInsertId(sqlStr string, args ...interface{}) (mySqlNumber DbOperateResult, lastId int64, err error) {
	result, err := t.exec(1, sqlStr, args...)
	if err != nil {
		mySqlNumber = MatchMySqlError(err)
		return mySqlNumber, lastId, err
	}

	lastId, err = result.LastInsertId()
	if err != nil {
		return DbOperateFail, lastId, err
	}

	return DbOperateSuccess, lastId, nil
}

func (t *SqlxContext) NamedExec(sqlStr string, uMap DbContext) (rowsAffected int64, err error) {
	result, err := t.namedExec(1, "namedExec", sqlStr, uMap)
	if err != nil {
		return -2, err
	}

	rowsAffected, _ = result.RowsAffected()

	return
}

func (t *SqlxContext) NamedExecAndGetResult(sqlStr string, uMap DbContext) (result sql.Result, err error) {
	return t.namedExec(1, "namedExecAndGetResult", sqlStr, uMap)
}

func (t *SqlxContext) insertOneAndGetResult(k int, data interface{}, ig ...bool) (result sql.Result, err error) {
	k++

	uMap := DbContextFromDbItem(data)
	sqlStr := t.BuildInsertSqlNamedStr(uMap, ig...)

	return t.namedExec(k, "insertOneAndGetResult", sqlStr, uMap)
}

func (t *SqlxContext) insertManyAndGetResult(k int, dataList interface{}, ig ...bool) (result sql.Result, err error) {
	k++

	s := reflect.ValueOf(dataList)
	s = DeValue(s)

	l := s.Len()
	if l == 0 {
		return
	}

	uMap := DbContextFromDbItem(s.Index(0).Interface())
	sqlStr := t.BuildInsertSqlNamedStr(uMap, ig...)

	var cost time.Duration
	var rowsAffected int64 = -2

	defer func() {
		detail := map[string]interface{}{
			logFiledSql:          sqlStr,
			logFiledArgNum:       l,
			logFiledLike:         uMap,
			logFiledRowsAffected: rowsAffected,
		}

		f := []zap.Field{
			LogProcessor("insertManyAndGetResult"),
			LogDetail(detail),
		}

		t.output(k, cost, err, f...)
	}()

	/*
			t0 := time.Now()

			batchSize := 1500 // 每个批次的数据量

			for i := 0; i < l; i += batchSize {
			   end := i + batchSize
			   if end > l {
				   end = l
			   }
			   batchData := s.Slice(i, end)
		       result, err = t.Db().NamedExec(sqlStr, batchData.Interface())
			}
	*/

	t0 := time.Now()
	result, err = t.Db().NamedExec(sqlStr, dataList)
	cost = time.Now().Sub(t0)

	if err != nil {
		return
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		rowsAffected = -3
	}

	return
}

func (t *SqlxContext) getRow(k int, data interface{}, sqlStr string, args ...interface{}) (hasData bool, err error) {
	k++

	t0 := time.Now()
	err = t.Db().Get(data, sqlStr, args...)
	cost := time.Now().Sub(t0)

	defer func() {
		l := len(args)
		if l == 0 {
			args = []interface{}{}
		}

		//n := MinInt(l, maxArgNum)

		detail := map[string]interface{}{
			logFiledSql:    sqlStr,
			logFiledArgNum: l,
			logFiledArgs:   args,
			logFiledHas:    hasData,
		}

		f := []zap.Field{
			LogProcessor("getRow"),
			LogDetail(detail),
		}

		t.output(k, cost, err, f...)
	}()

	if err != nil {
		if IgnoreMysqlErr(err) {
			return hasData, nil
		}

		return hasData, err
	}

	hasData = true

	return hasData, nil
}

func (t *SqlxContext) getList(k int, dataList interface{}, sqlStr string, raw ...interface{}) (err error) {
	k++

	l := len(raw)
	//n := MinInt(l, maxArgNum)

	sqlStr, args, err := sqlx.In(sqlStr, raw...)
	if l == 0 {
		args = []interface{}{}
	}

	if err != nil {
		detail := map[string]interface{}{
			logFiledSql:    sqlStr,
			logFiledArgNum: l,
			logFiledArgs:   args,
		}

		f := []zap.Field{
			LogProcessor("getList"),
			LogDetail(detail),
		}

		t.output(k, 0, err, f...)
		return
	}

	t0 := time.Now()
	err = t.Db().Select(dataList, sqlStr, args...)
	cost := time.Now().Sub(t0)

	l = len(args)
	//n = MinInt(l, maxArgNum)

	s := reflect.ValueOf(dataList)
	s = DeValue(s)

	detail := map[string]interface{}{
		logFiledSql:    sqlStr,
		logFiledArgNum: l,
		logFiledArgs:   args,
		logFiledResNum: s.Len(),
	}

	f := []zap.Field{
		LogProcessor("getList"),
		LogDetail(detail),
	}

	t.output(k, cost, err, f...)

	return err
}

func (t *SqlxContext) namedExec(k int, event, sqlStr string, uMap DbContext) (result sql.Result, err error) {
	k++

	uMap = uMap.FilterBlackFiled()

	t0 := time.Now()
	result, err = t.Db().NamedExec(sqlStr, uMap)
	cost := time.Now().Sub(t0)

	var rowsAffected int64 = -2
	if err == nil {
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			rowsAffected = -3
		}
	}

	detail := map[string]interface{}{
		logFiledSql:          sqlStr,
		logFiledUmap:         uMap,
		logFiledRowsAffected: rowsAffected,
	}

	f := []zap.Field{
		LogProcessor(event),
		LogDetail(detail),
	}

	t.output(k, cost, err, f...)

	return result, err
}

func (t *SqlxContext) exec(k int, sqlStr string, args ...interface{}) (result sql.Result, err error) {
	k++

	t0 := time.Now()
	result, err = t.Db().ExecContext(t.Ctx(), sqlStr, args...)
	cost := time.Now().Sub(t0)

	var rowsAffected int64 = -2
	if err == nil {
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			rowsAffected = -3
		}
	}

	l := len(args)
	if l == 0 {
		args = []interface{}{}
	}
	//n := MinInt(l, maxArgNum)

	detail := map[string]interface{}{
		logFiledSql:          sqlStr,
		logFiledArgNum:       l,
		logFiledArgs:         args,
		logFiledRowsAffected: rowsAffected,
	}

	f := []zap.Field{
		LogProcessor("exec"),
		LogDetail(detail),
	}

	t.output(k, cost, err, f...)

	return result, err
}

type SimplePropertyItem struct {
	Status int32
	Sort   int64
	A      map[string]interface{}
	P      map[string]interface{}
}

func (t SimplePropertyItem) GetStatus() int32 {
	return t.Status
}

func (t SimplePropertyItem) GetSort(id int64) int64 {
	return t.Sort
}

func (t SimplePropertyItem) GetAttribute() DbContext {
	return t.A
}

func RadiusQuery(c GeoPoint, radius float64) DbContext {
	e, s, w, n := c.Radius(radius)

	res := DbContext{
		"longitude <=": e.Longitude,
		"latitude >=":  s.Latitude,
		"longitude >=": w.Longitude,
		"latitude <=":  n.Latitude,
	}

	return res
}

func ParseCrudDecoratorArgs(args ...interface{}) (g int, obj []string, f []CrudDecorator) {
	g = len(args) / 2

	obj = []string{}
	f = []CrudDecorator{}

	for i := 0; i < g; i++ {
		obj = append(obj, args[i*2].(string))
		f = append(f, args[i*2+1].(CrudDecorator))
	}

	return
}

func RegCrudDecoratorMap(m map[string]CrudDecorator, args ...interface{}) {
	g, obj, f := ParseCrudDecoratorArgs(args...)

	for i := 0; i < g; i++ {
		_, ok := m[obj[i]]
		if !ok {
			m[obj[i]] = f[i]
		}
	}
}

func SqlCountStrForSub(sqlStr string) (countStr string) {
	/* countStr is like
	SELECT
		 COALESCE(COUNT(*), 0)
	FROM (
		$sqlStr
	) AS sub;
	*/

	var bc strings.Builder
	bc.WriteString("SELECT COALESCE(COUNT(*), 0) FROM (")
	bc.WriteString(sqlStr)
	bc.WriteString(") AS sub;")

	countStr = bc.String()

	return
}

func SqlLimitStr(limit []uint) (limitStr string) {
	limit = SqlLimitUint(limit...)

	var b strings.Builder
	b.WriteString("LIMIT ")
	b.WriteString(ParseIntToStr(int(limit[0])))
	b.WriteString(",")
	b.WriteString(ParseIntToStr(int(limit[1])))

	limitStr = b.String()

	return limitStr
}

func SqlPageQueryStr(page PageQueryReq, sqlStr string) (pageQueryStr string) {
	/* pageQueryStr is like
	$sqlStr LIMIT $page.Page,$page.Size
	*/

	var b strings.Builder
	b.WriteString(sqlStr)
	b.WriteString(" ")
	b.WriteString(page.LimitStr())

	pageQueryStr = b.String()

	return
}

func (t BaseModel) Fields() []string {
	return []string{IdDb, CreatedAtDb, UpdatedAtDb}
}

func (t BaseModel) I64() I64 {
	return I64{Id: t.Id}
}

func (t BaseModel) IdStr() string {
	return ParseInt64ToStr(t.Id)
}

func (t BaseModel) Where(ctx context.Context) DbContext {
	if t.Id == 0 {
		return DbContext{}
	}

	return DbContext{IdDb: t.Id}
}

func (t *BaseModel) Create(tick ...int64) {
	if t.CreatedAt == 0 {
		t.CreatedAt = ParseTick(tick)
	}
}

func (t *BaseModel) CreateMs(tick ...int64) {
	if t.CreatedAt == 0 {
		t.CreatedAt = ParseTickMs(tick)
	}
}

func (t PoModel) Fields() []string {
	return []string{IdDb, UuidDb, CreatedAtDb, UpdatedAtDb}
}

func (t PoModel) Bff() PoModel {
	return PoModel{Uuid: t.Uuid, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt}
}

func (t PoModel) I64() I64 {
	return I64{Id: t.Id}
}

func (t PoModel) UUID() UUID {
	return UUID{Uuid: t.Uuid}
}

func (t PoModel) Where(ctx context.Context) DbContext {
	if t.Uuid == "" {
		return DbContext{}
	}

	return DbContext{UuidDb: t.Uuid}
}

func (t PoModel) ToUpdate(raw interface{}, ignoreFiled ...string) DbContext {
	return DbContextFromBff(raw, ignoreFiled...)
}

func (t *PoModel) NewUuid(id ...string) {
	if len(id) > 0 && id[0] != "" {
		return
	}

	t.Uuid = uuid.NewString()
}

func (t *PoModel) Create(tick ...int64) {
	if t.Uuid == "" {
		t.Uuid = uuid.NewString()
	}

	if t.CreatedAt == 0 {
		t.CreatedAt = ParseTick(tick)
	}
}

func (t *PoModel) CreateMs(tick ...int64) {
	if t.Uuid == "" {
		t.Uuid = uuid.NewString()
	}

	if t.CreatedAt == 0 {
		t.CreatedAt = ParseTickMs(tick)
	}
}

func (t BaseProperty) Fields() []string {
	return []string{"name", "description", "property", "status", "sort"}
}

func (t BaseProperty) GetStatus() int32 {
	return t.Status
}

func (t BaseProperty) GetSort(id int64) int64 {
	return DeInt64Param(t.Sort, id)
}

func (t BaseProperty) BaseMeta(id int64, uuid string) BaseMeta {
	res := BaseMeta{
		Id:          uuid,
		Name:        t.Name,
		Description: t.Description,
		Status:      t.Status,
		Sort:        t.GetSort(id),
	}

	return res
}

type I64 struct {
	Id int64 `json:"id,omitempty"`
}

func NewI64(id int64) I64 {
	res := I64{
		Id: id,
	}

	return res
}

func (t I64) Where(ctx context.Context) DbContext {
	if t.Id == 0 {
		return DbContext{}
	}

	return DbContext{IdDb: t.Id}
}

func (t I64) ToUpdate(raw interface{}, ignoreFiled ...string) DbContext {
	return DbContextFromBff(raw, ignoreFiled...)
}

type UUID struct {
	Uuid string `json:"uuid,omitempty" set:"-"`
}

func NewUuid(id string) UUID {
	res := UUID{
		Uuid: id,
	}

	return res
}

func (t UUID) Where(ctx context.Context) DbContext {
	if t.Uuid == "" {
		return DbContext{}
	}

	return DbContext{UuidDb: t.Uuid}
}

func (t UUID) ToUpdate(raw interface{}, ignoreFiled ...string) DbContext {
	return DbContextFromBff(raw, ignoreFiled...)
}

func (t *UUID) NewUuid(id ...string) {
	if len(id) > 0 && id[0] != "" {
		_, err := uuid.Parse(id[0])

		if err == nil {
			t.Uuid = id[0]
		} else {
			t.Uuid = uuid.NewString()
		}
		return
	}

	t.Uuid = uuid.NewString()
}

func (t *UUID) SetUuid(id string) {
	if t.Uuid == "" {
		t.Uuid = id
	}
}

func (t UUID) GetUuid() string {
	return t.Uuid
}

func (t *UUID) PoModel() PoModel {
	if t.Uuid == "" {
		t.Uuid = uuid.NewString()
	}

	return PoModel{Uuid: t.Uuid}
}

type PK string

func (t PK) DeStrParam(raw string) string {
	if t == "" {
		return raw
	}

	return string(t)
}

func (t PK) Q(f string) func(ctx context.Context) DbContext {
	if t == "" {
		var res = func(ctx context.Context) DbContext {
			return DbContext{}
		}

		return res
	}

	var res = func(ctx context.Context) DbContext {
		return DbContext{f: t}
	}

	return res
}

func (t BffProperty) JsonProperty() string {
	return EnsureJsonStrFromByte(t.Property)
}

func (t BaseProperty) JsonProperty() json.RawMessage {
	return EnsureJsonRawMessage(t.Property)
}

func (t ResourceBff) JsonProperty() string {
	return EnsureJsonStrFromByte(t.Property)
}

func (t ResourceItem) JsonProperty() json.RawMessage {
	return EnsureJsonRawMessage(t.Property)
}
