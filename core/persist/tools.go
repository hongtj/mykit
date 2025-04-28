package persist

import (
	"fmt"
	. "mykit/core/types"

	"github.com/didi/gendry/builder"
)

func (t *SqlxContext) DropTable(n ...string) error {
	table := ParseStrParam(n, t.table)
	sqlStr := fmt.Sprintf("DROP TABLE IF EXISTS `%v`;", table)
	_, err := t.exec(1, sqlStr)
	return err
}

func (t *SqlxContext) DeleteSomeTable(table ...string) {
	for _, v := range table {
		sqlStr, args, err := builder.BuildDelete(v, DbContext{})
		if err != nil {
			continue
		}

		t.exec(1, sqlStr, args...)
	}
}

func (t *SqlxContext) DeleteAll() error {
	_, err := t.Delete(DbContext{})
	return err
}

func (t *SqlxContext) FindAll(data interface{}, selectField ...string) (err error) {
	sqlStr, args, err := t.BuildSelect(DbContext{}, selectField...)
	if err != nil {
		t.Errorf("Find all, select %v BuildSelect err, %v", selectField, err)
		return err
	}

	if !IsSliceElem(data) {
		_, err = t.getRow(1, data, sqlStr, args...)
		return
	}

	err = t.getList(1, data, sqlStr, args...)

	return
}

func (t *SqlxContext) ExistsByUuid(id string) bool {
	var n int64
	has, _ := t.FindRes(&n, DbContext{UuidDb: id}, IdDb)
	return has
}

func (t *SqlxContext) MustFindId(where DbContext, selectField ...string) (id int64) {
	if len(where) == 0 {
		return -1
	}

	idFiled := IdDb
	if len(selectField) > 0 {
		idFiled = selectField[0]
	}

	sqlStr, args, err := t.BuildSelectForUpdate(where, idFiled)
	if err != nil {
		return -2
	}

	_, err = t.getRow(1, &id, sqlStr, args...)
	if err != nil {
		return -3
	}

	return
}

func (t *SqlxContext) DelBy(k string, v interface{}) (err error) {
	where := DbContext{k: v}
	sqlStr, args, err := t.BuildDelete(where)
	if err != nil {
		return
	}

	_, err = t.exec(1, sqlStr, args...)

	return
}

func (t *SqlxContext) DelById(id ...int64) (err error) {
	if len(id) == 0 {
		return
	}

	where := DbContext{IdDb: id}
	sqlStr, args, err := t.BuildDelete(where)
	if err != nil {
		return
	}

	_, err = t.exec(1, sqlStr, args...)

	return err
}

func (t *SqlxContext) DelByUuid(id ...string) (err error) {
	if len(id) == 0 {
		return
	}

	where := DbContext{UuidDb: id}
	sqlStr, args, err := t.BuildDelete(where)
	if err != nil {
		return
	}

	_, err = t.exec(1, sqlStr, args...)

	return err
}

func (t *SqlxContext) DeleteBy(k string, v interface{}) (rowsAffected int64, err error) {
	where := DbContext{
		k: v,
	}
	sqlStr, args, err := t.BuildDelete(where)
	if err != nil {
		return
	}

	result, err := t.exec(1, sqlStr, args...)
	if err == nil {
		rowsAffected, _ = result.RowsAffected()
	}

	return
}

func (t *SqlxContext) DeleteById(id ...int64) (rowsAffected int64, err error) {
	if len(id) == 0 {
		return
	}

	where := DbContext{IdDb: id}
	sqlStr, args, err := t.BuildDelete(where)
	if err != nil {
		return
	}

	result, err := t.exec(1, sqlStr, args...)
	if err == nil {
		rowsAffected, _ = result.RowsAffected()
	}

	return
}

func (t *SqlxContext) DeleteByUuid(id ...string) (rowsAffected int64, err error) {
	if len(id) == 0 {
		return
	}

	where := DbContext{UuidDb: id}
	sqlStr, args, err := t.BuildDelete(where)
	if err != nil {
		return
	}

	result, err := t.exec(1, sqlStr, args...)
	if err == nil {
		rowsAffected, _ = result.RowsAffected()
	}

	return
}

func (t *SqlxContext) FindById(data interface{}, id interface{}, selectField ...string) (err error) {
	if !IsInt64OrInt64Slice(id) {
		return ErrInvalidParam
	}

	where := DbContext{IdDb: id}
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

func (t *SqlxContext) FindByUuid(data interface{}, uuid interface{}, selectField ...string) (err error) {
	if !IsStrOrStrSlice(uuid) {
		return ErrInvalidParam
	}

	where := DbContext{UuidDb: uuid}
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

func (t *SqlxContext) FindId(where DbContext) (res int64, err error) {
	sqlStr, args, err := t.BuildSelect(where, IdDb)
	if err != nil {
		return
	}

	_, err = t.getRow(1, &res, sqlStr, args...)

	return
}

func (t *SqlxContext) FindIdByUuid(uuid string) (res int64, err error) {
	where := DbContext{UuidDb: uuid}
	sqlStr, args, err := t.BuildSelect(where, IdDb)
	if err != nil {
		return
	}

	_, err = t.getRow(1, &res, sqlStr, args...)

	return
}

func (t *SqlxContext) FindUuid(where DbContext) (res string, err error) {
	sqlStr, args, err := t.BuildSelect(where, UuidDb)
	if err != nil {
		return
	}

	_, err = t.getRow(1, &res, sqlStr, args...)

	return
}

func (t *SqlxContext) FindUuidById(id int64) (res string, err error) {
	where := DbContext{IdDb: id}
	sqlStr, args, err := t.BuildSelect(where, UuidDb)
	if err != nil {
		return
	}

	_, err = t.getRow(1, &res, sqlStr, args...)

	return
}

func (t *SqlxContext) UpdateById(id interface{}, update DbContext) (rowsAffected int64, err error) {
	if !IsInt64OrInt64Slice(id) {
		err = ErrInvalidParam
		return
	}

	where := DbContext{IdDb: id}
	sqlStr, args, err := t.BuildUpdate(where, update)
	if err != nil {
		return
	}

	result, err := t.exec(1, sqlStr, args...)
	if err == nil {
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			rowsAffected = -3
		}
	}

	return
}

func (t *SqlxContext) UpdateByUuid(id interface{}, update DbContext) (rowsAffected int64, err error) {
	if !IsStrOrStrSlice(id) {
		err = ErrInvalidParam
		return
	}

	where := DbContext{UuidDb: id}
	sqlStr, args, err := t.BuildUpdate(where, update)
	if err != nil {
		return
	}

	result, err := t.exec(1, sqlStr, args...)
	if err == nil {
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			rowsAffected = -3
		}
	}

	return
}
