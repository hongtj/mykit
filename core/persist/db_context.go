package persist

import (
	"context"
	"fmt"
	. "mykit/core/dsp"
	. "mykit/core/types"
	"reflect"
	"strings"
	"time"
)

func ParseDbContextList(raw []DbContext) []map[string]interface{} {
	res := []map[string]interface{}{}

	for _, v := range raw {
		if len(v) == 0 {
			continue
		}

		res = append(res, v)
	}

	return res
}

func (t DbContext) Set(f string, v interface{}) DbContext {
	if f == "" {
		return t
	}

	t[f] = v

	return t
}

func (t DbContext) SetStr(f, v string) DbContext {
	if f == "" || v == "" {
		return t
	}

	t[f] = v

	return t
}

func (t DbContext) SetInt64(f string, v int64) DbContext {
	if f == "" || v == 0 {
		return t
	}

	t[f] = v

	return t
}

func (t DbContext) SetBool(f string, v bool) DbContext {
	if f == "" {
		return t
	}

	if v {
		t[f] = 1
	} else {
		t[f] = 0
	}

	return t
}

func (t DbContext) SetSome(f string, v interface{}) DbContext {
	if f == "" {
		return t
	}

	vt := reflect.TypeOf(v)
	vt = DeType(vt)
	if vt.Kind() != reflect.Slice {
		return t
	}

	vv := reflect.ValueOf(v)
	vv = DeValue(vv)
	if vv.Len() == 0 {
		return t
	}

	t[f] = v

	return t
}

func (t DbContext) SetSomeStr(f string, v ...string) DbContext {
	if len(v) == 0 || len(f) == 0 {
		return t
	}

	v = OpSetStrList(v)

	t[f] = v
	return t
}

func (t DbContext) SetSomeInt64(f string, v ...int64) DbContext {
	if len(v) == 0 || len(f) == 0 {
		return t
	}

	v = OpSetInt64List(v)

	t[f] = v
	return t
}

func (t DbContext) Raw(f string, v interface{}) DbContext {
	if f == "" {
		return t
	}

	t[f] = Raw(v)

	return t
}

func (t DbContext) Rawf(f, pat string, v ...interface{}) DbContext {
	if f == "" {
		return t
	}

	t[f] = Raw(fmt.Sprintf(pat, v...))

	return t
}

func (t DbContext) Clone() DbContext {
	res := DbContext{}

	for k, v := range t {
		res[k] = v
	}

	return res
}

func (t DbContext) Values() []interface{} {
	res := []interface{}{}

	for k, v := range t {
		res = append(res, k, v)
	}

	return res
}

func (t DbContext) FilterBlackFiled() DbContext {
	res := DbContext{}

	for k, v := range t {
		if k == "" {
			continue
		}

		res[k] = v
	}

	return res
}

func (t DbContext) Merge(raw ...DbContext) DbContext {
	for _, v := range raw {
		for k1, v1 := range v {
			_, ok := t[k1]
			if !ok {
				t[k1] = v1
			}
		}
	}

	return t
}

func (t DbContext) MergeSetItem(ctx context.Context, raw ...SetItem) DbContext {
	for _, v := range raw {
		for k1, v1 := range v.ToSet(ctx) {
			t[k1] = v1
		}
	}

	return t
}

func (t DbContext) MergeGeoPoint(raw GeoPoint) DbContext {
	t["longitude"] = raw.Longitude
	t["latitude"] = raw.Latitude
	t["height"] = raw.Height

	return t
}

func (t DbContext) MergeDbItem(raw interface{}, ignoreFiled ...string) DbContext {
	return t.Merge(DbContextFromDbItem(raw, ignoreFiled...))
}

func (t DbContext) MergeBffItem(raw interface{}, ignoreFiled ...string) DbContext {
	return t.Merge(DbContextFromBff(raw, ignoreFiled...))
}

func (t DbContext) QueryByUk(uk ...string) DbContext {
	res := DbContext{}

	m := ToStrListFlag(uk)
	for k, v := range t {
		if m[k] {
			res[k] = v
		}
	}

	return res
}

func (t DbContext) CreateAt(tick ...int64) DbContext {
	_, ok := t[CreatedAtDb]
	if !ok {
		t[CreatedAtDb] = ParseTick(tick)
	}

	return t
}

func (t DbContext) UpdateAt(tick ...int64) DbContext {
	_, ok := t[UpdatedAtDb]
	if !ok {
		t[UpdatedAtDb] = ParseTick(tick)
	}

	return t
}

func (t DbContext) UpdateAtMs(tick ...int64) DbContext {
	_, ok := t[UpdatedAtDb]
	if !ok {
		t[UpdatedAtDb] = ParseTickMs(tick)
	}

	return t
}

func (t DbContext) DeletedBy(user string, tick ...int64) DbContext {
	_, ok := t[DeletedAtDb]
	if !ok {
		t[DeletedAtDb] = ParseTick(tick)
	}

	_, ok = t[DeletedByDb]
	if !ok {
		t[DeletedByDb] = user
	}

	return t
}

func (t DbContext) DeletedAtMs(user string, tick ...int64) DbContext {
	_, ok := t[DeletedAtDb]
	if !ok {
		t[DeletedAtDb] = ParseTickMs(tick)
	}

	_, ok = t[DeletedByDb]
	if !ok {
		t[DeletedByDb] = user
	}

	return t
}

func (t DbContext) Id(id int64) DbContext {
	t[IdDb] = id
	return t
}

func (t DbContext) Ids(id []int64) DbContext {
	if len(id) == 0 {
		return t
	}

	t[IdDb] = id
	return t
}

func (t DbContext) SetId(id ...int64) DbContext {
	return t.SetSomeInt64(IdDb, id...)
}

func (t DbContext) Uuid(id string) DbContext {
	t[UuidDb] = id
	return t
}

func (t DbContext) Uuids(id []string) DbContext {
	if len(id) == 0 {
		return t
	}

	t[UuidDb] = id
	return t
}

func (t DbContext) SetUuid(id ...string) DbContext {
	return t.SetSomeStr(UuidDb, id...)
}

func (t DbContext) GeoPoint(raw GeoPoint) DbContext {
	t["longitude"] = raw.Longitude
	t["latitude"] = raw.Latitude
	t["height"] = raw.Height

	return t
}

func (t DbContext) GeoDbContext(raw DbContext) DbContext {
	t["longitude"] = raw["longitude"]
	t["latitude"] = raw["latitude"]
	t["height"] = raw["height"]

	return t
}

func (t DbContext) Limit(n ...int32) DbContext {
	t[TagLimit] = SqlLimit(n...)
	return t
}

func (t DbContext) PageLimit(page, size int64) DbContext {
	if size == 0 {
		return t
	}

	t[TagLimit] = SqlLimitUint(PageLimit(page, size)...)
	return t
}

func (t DbContext) ClearTag(f ...string) DbContext {
	for _, v := range f {
		delete(t, v)
	}

	return t
}

func (t DbContext) ClearTagLimit() DbContext {
	return t.ClearTag(TagLimit)
}

func (t DbContext) PageQuery() DbContext {
	sf, ok := t[tagSize]
	if !ok {
		return t
	}

	s, ok := sf.(int64)
	if !ok {
		return t
	}

	p, ok := t[tagPage].(int64)
	if !ok {
		return t
	}

	delete(t, tagPage)
	delete(t, tagSize)

	return t.PageLimit(p, s)
}

func (t DbContext) OrderBy(raw string) DbContext {
	t[tagOrderBy] = raw

	return t
}

func (t DbContext) Asc(f ...string) DbContext {
	t[tagOrderBy] = joinStr(ParseStrParam(f, IdDb), tagAsc)
	return t
}

func (t DbContext) Desc(f ...string) DbContext {
	t[tagOrderBy] = joinStr(ParseStrParam(f, IdDb), tagDesc)
	return t
}

func (t DbContext) First() DbContext {
	t[TagLimit] = SqlLimit(1)
	return t
}

func (t DbContext) Last(f ...string) DbContext {
	return t.Desc(f...).Limit()
}

func (t DbContext) Like(f, v string) DbContext {
	if f == "" || v == "" {
		return t
	}

	t[joinStr(f, tagLike)] = "%" + v + "%"

	return t
}

func (t DbContext) Or(raw ...DbContext) DbContext {
	if len(raw) == 0 {
		return t
	}

	parsed := ParseDbContextList(raw)
	if len(parsed) == 0 {
		return t
	}

	t[tagOr] = parsed

	return t
}

func (t DbContext) In(f string, s interface{}) DbContext {
	vt := reflect.TypeOf(s)
	if vt.Kind() != reflect.Slice {
		return t
	}

	vs := reflect.ValueOf(s)
	if vs.Len() == 0 {
		return t
	}

	t[f] = s

	return t
}

func (t DbContext) Group(v string) DbContext {
	if v == "" {
		return t
	}

	t[TagGroupBy] = v

	return t
}

func (t DbContext) Having(f, v string) DbContext {
	if v == "" {
		return t
	}

	t[TagHaving] = v

	return t
}

func (t DbContext) FromTo(f string, from, to int64) DbContext {
	t.Set(joinStr(f, ">="), from)

	if to > from {
		t.Set(joinStr(f, "<"), to)
	}

	return t
}

func (t DbContext) From(f string, tick int64) DbContext {
	if tick > 0 {
		t.Set(joinStr(f, ">="), tick)
	}

	return t
}

func (t DbContext) To(f string, tick int64) DbContext {
	if tick > 0 {
		t.Set(joinStr(f, "<="), tick)
	}

	return t
}

func (t DbContext) After(f string, tick int64) DbContext {
	t.Set(joinStr(f, ">"), tick)

	return t
}

func (t DbContext) Before(f string, tick int64) DbContext {
	t.Set(joinStr(f, "<"), tick)

	return t
}

func (t DbContext) Between(f string, v1, v2 interface{}) DbContext {
	if f == "" {
		return t
	}

	f = joinStr(f, "BETWEEN")
	t[f] = []interface{}{v1, v2}

	return t
}

func (t DbContext) Compute(f, opp string) DbContext {
	if f == "" {
		return t
	}

	t[f] = Raw(f + opp)

	return t
}

func (t DbContext) AddInt64(f string, n int64) DbContext {
	opp := SqlAddInt64(n)
	return t.Compute(f, opp)
}

func (t DbContext) MulInt64(f string, n int64) DbContext {
	opp := SqlMulInt64(n)
	return t.Compute(f, opp)
}

func (t DbContext) AddFloat64(f string, n float64) DbContext {
	opp := SqlAddFloat64(n)
	return t.Compute(f, opp)
}

func (t DbContext) MulFloat64(f string, n float64) DbContext {
	opp := SqlMulFloat64(n)
	return t.Compute(f, opp)
}

func (t DbContext) SetTx(c *SqlxContext) DbContext {
	t[tagTx] = c

	return t
}

func (t DbContext) TxNext(n ...TransactionParticipant) {
	v, ok := t[tagTx].(TransactionParticipant)
	if ok && v != nil {
		v.Next(n...)
	}
}

func (t DbContext) UpdateBy(ctx context.Context) DbContext {
	t[UpdatedByDb] = GetUser(ctx)

	return t
}

func (t DbContext) OperatedBy(ctx context.Context, tick ...int64) DbContext {
	t[OperatedByDb] = GetUser(ctx)
	t[OperatedAtDb] = ParseTick(tick)

	return t
}

func (t DbContext) CountInOneDay(t0 time.Time, f string) DbContext {
	t0 = GetZeroTime(t0)
	t1 := t0.Add(time.Hour * 24)

	return t.CountBetweenTime(t0, t1, f)
}

func (t DbContext) CountBetweenTime(t0, t1 time.Time, f string) DbContext {
	res := DbContext{
		f + " >=": t0.Unix(),
		f + " <":  t1.Unix(),
	}

	return res
}

func (t DbContext) Cols(raw interface{}, cols ...string) DbContext {
	d := DbContextFromDbItem(raw)

	for _, v := range cols {
		item, ok := d[v]
		if ok {
			t[v] = item
		}
	}

	return t
}

func (t DbContext) IgnoreCols(raw interface{}, ignoreFiled ...string) DbContext {
	return t.MergeDbItem(raw, ignoreFiled...)
}

func joinStr(s1, s2 string) string {
	if s1 == "" {
		return ""
	}

	var b strings.Builder

	b.WriteString(s1)
	b.WriteString(tagDelimiter)
	b.WriteString(s2)

	res := b.String()

	return res
}

func Like(k, v string) DbContext {
	return DbContext{}.Like(k, v)
}

func (t DbContext) GetUuid() string {
	v, _ := t[UuidDb].(string)

	return v
}

func ParseToSet(raw DbContext) {
	for k, v := range raw {
		item, ok := v.(NamedValue)
		if ok {
			raw[k] = item.Value
		}
	}
}
