package persist

import (
	"database/sql"
	"errors"
	. "mykit/core/types"
	"strings"

	"github.com/go-sql-driver/mysql"
)

type DbOperateResult uint16

const (
	DbOperateFail    DbOperateResult = 0
	DbOperateSuccess DbOperateResult = 1
	DuplicateKey     DbOperateResult = 1062 //唯一索引重复
	TableNotExist    DbOperateResult = 1146 //表不存在
)

func IsMysqlDuplicateKey(err error) bool {
	if mySqlErr, ok := err.(*mysql.MySQLError); ok {
		return mySqlErr.Number == uint16(DuplicateKey)
	}

	return false
}

func NotMysqlDuplicateKey(err error, dbOperateResult DbOperateResult) bool {
	return err != nil && dbOperateResult != DuplicateKey
}

func IsMysqlTableNotExist(err error) bool {
	if mySqlErr, ok := err.(*mysql.MySQLError); ok {
		return mySqlErr.Number == uint16(TableNotExist)
	}

	return false
}

func MatchMySqlError(err error) DbOperateResult {
	if mySqlErr, ok := err.(*mysql.MySQLError); ok {
		switch mySqlErr.Number {
		case uint16(DuplicateKey):
			return DuplicateKey
		case uint16(TableNotExist):
			return TableNotExist
		}
	}

	return DbOperateFail
}

func IgnoreMysqlErr(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func SqlLimit(n ...int32) []uint {
	l := len(n)
	if l == 0 {
		return []uint{0, 1}
	}

	if l == 1 {
		return []uint{0, OffsetInt32(n[0])}
	}

	return []uint{OffsetInt32(n[0]), OffsetInt32(n[1])}
}

func SqlLimitInt(n ...int) []uint {
	l := len(n)
	if l == 0 {
		return []uint{0, 1}
	}

	if l == 1 {
		return []uint{0, OffsetInt(n[0])}
	}

	return []uint{OffsetInt(n[0]), OffsetInt(n[1])}
}

func SqlLimitUint(n ...uint) []uint {
	l := len(n)
	if l == 0 {
		return []uint{0, 1}
	}

	if l == 1 {
		return []uint{0, n[0]}
	}

	return []uint{n[0], n[1]}
}

func SqlLimitInt64(n ...int64) []uint {
	l := len(n)
	if l == 0 {
		return []uint{0, 1}
	}

	if l == 1 {
		return []uint{0, OffsetInt64(n[0])}
	}

	return []uint{OffsetInt64(n[0]), OffsetInt64(n[1])}
}

func PageLimit(page, size int64) []uint {
	return SqlLimitInt64((page-1)*size, size)
}

func BuildInsertSql(table string, m interface{}, ig ...bool) (sqlStr string, uMap DbContext) {
	uMap = TransStructToDbContextByTag(m, TagDb)
	sqlStr = BuildInsertSqlNamedStr(table, uMap, ig...)

	return
}

func BuildInsertSqlFromJson(table string, m interface{}, ig ...bool) (sqlStr string, uMap DbContext) {
	uMap = TransStructToDbContextByTag(m, TagJson)
	sqlStr = BuildInsertSqlNamedStr(table, uMap, ig...)

	return
}

func BuildInsertSqlNamedStr(table string, uMap DbContext, ig ...bool) string {
	var filedBuilder strings.Builder
	var valuesBuilder strings.Builder

	uMap = uMap.FilterBlackFiled()
	for k := range uMap {
		filedBuilder.WriteString("`")
		filedBuilder.WriteString(k)
		filedBuilder.WriteString("`,")

		valuesBuilder.WriteString(":")
		valuesBuilder.WriteString(k)
		valuesBuilder.WriteString(",")
	}

	intoStr := filedBuilder.String()
	intoStr = intoStr[:len(intoStr)-1]

	valuesStr := valuesBuilder.String()
	valuesStr = valuesStr[:len(valuesStr)-1]

	var b strings.Builder
	b.WriteString("INSERT ")

	ignore := ParseBoolParam(ig, false)
	if ignore {
		b.WriteString("Ignore ")
	}

	b.WriteString("INTO ")
	b.WriteString(table)
	b.WriteString(" (")
	b.WriteString(intoStr)
	b.WriteString(") VALUES(")
	b.WriteString(valuesStr)
	b.WriteString(")")

	res := b.String()

	return res
}

func coalesce(method, filed string) string {
	var b strings.Builder
	b.WriteString("COALESCE(")
	b.WriteString(method)
	b.WriteString("(")
	b.WriteString(filed)
	b.WriteString("), 0)")

	res := b.String()

	return res
}

func SelectCount(raw ...string) string {
	return coalesce("COUNT", ParseStrParam(raw, IdDb))
}

func SelectMax(raw string) string {
	return coalesce("MAX", raw)
}

func SelectMin(raw string) string {
	return coalesce("MIN", raw)
}

func SelectSum(raw string) string {
	return coalesce("SUM", raw)
}

func selectMethod(method, filed string) string {
	var b strings.Builder
	b.WriteString(method)
	b.WriteString("(")
	b.WriteString(filed)
	b.WriteString(")")

	res := b.String()

	return res
}

func SelectDistinct(raw string) string {
	return selectMethod("DISTINCT", raw)
}
