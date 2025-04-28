package persist

import (
	"context"
	"encoding/json"
	"fmt"
	. "mykit/core/dsp"
	. "mykit/core/types"
	"reflect"
	"strings"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/didi/gendry/builder"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type EtcdEventHandler func(*mvccpb.KeyValue)

type RouteMgr interface {
	RegNode(ctx context.Context, node, ip string) error
}

type SqlModel interface {
	TableName() string
}

type DbItem interface {
	Select() []string
}

type ObjModel interface {
	Obj() string
}

type CtlModel interface {
	CtlWhere() DbContext
}

type SqlObj interface {
	SqlModel
	ObjModel
}

type CrudDecorator func(ctx context.Context, operator DbContext, tick ...int64) error

type ToAdd interface {
	ToAdd(ctx context.Context) Dao
}

type Dao interface {
	SqlModel
	Create(tick ...int64)
}

type DelItem interface {
	DelKey() string
	Target() string
	SetDeleteAt(...int64)
}

type SetItem interface {
	ToSet(ctx context.Context) DbContext
}

type ToSet interface {
	Where(ctx context.Context) DbContext
	ToSet(ctx context.Context) DbContext
}

type ToPut interface {
	ToAdd
	ToSet
}

type UuidItem interface {
	SetUuid(string)
	GetUuid() string
}

type TreeElement interface {
	UuidItem
	ToAdd
	ToSet
	SetPid(string)
}

type TreeElementMgr interface {
	New(ctx context.Context) TreeElementMgr
	Category() string
	Element() TreeElement
	Add(data ToAdd, tick ...int64) (id int64, err error)
	AfterAdd() CrudDecorator
	DelByUuid(id ...string) (err error)
	AfterDel() CrudDecorator
	ExistsByUuid(id string) bool
	GetPropertyItemMap(id []string, selectField ...string) (res PropertyItemMap)
	Set(data ToSet, tick ...int64) (rowsAffected int64, err error)
	AfterSet() CrudDecorator
	TransactionParticipant
}

type PropertyItem interface {
	GetStatus() int32
	GetSort(int64) int64
	GetAttribute() DbContext
}

type PropertyItemMap map[string]PropertyItem

type TenantDbRouter func(ctx context.Context, useMaster ...bool) *sqlx.DB

type DbContext map[string]interface{}

type DbContextList []DbContext

type DeleteContext [3]string

type UpdateContext struct {
	Query  DbContext
	Update DbContext
}

type UpdateContextList []UpdateContext

type AttributeRes struct {
	Data DbContextList `json:"data"`
}

type PageQueryReq struct {
	Page  int64  `json:"page" validate:"gte=0"`
	Size  int64  `json:"size" validate:"gte=1,lte=200"`
	Model int64  `json:"model" validate:"oneof=0 1"`
	Desc  string `json:"desc"`
}

func (t PageQueryReq) DbContext() DbContext {
	if t.Model == 1 {
		return DbContext{}
	}

	res := DbContext{}
	if t.Desc != "" {
		res.Desc(t.Desc)
	}

	return res.PageLimit(t.Page, t.Size)
}

func (t PageQueryReq) Query(sqlStr string) string {
	return SqlPageQueryStr(t, sqlStr)
}

func (t PageQueryReq) LimitStr() string {
	q := t.DbContext()
	if len(q) == 0 {
		return ""
	}

	limit := SqlLimitUint(PageLimit(t.Page, t.Size)...)

	var b strings.Builder
	b.WriteString("LIMIT ")
	b.WriteString(ParseIntToStr(int(limit[0])))
	b.WriteString(",")
	b.WriteString(ParseIntToStr(int(limit[1])))

	res := b.String()

	return res
}

type PageQueryRes struct {
	Result interface{} `json:"result"`
	Total  int64       `json:"total"`
	Err    error       `json:"-"`
}

func NewPageQueryRes() *PageQueryRes {
	res := &PageQueryRes{
		Result: RawMessageOfNullList,
	}

	return res
}

func (t PageQueryRes) Value() reflect.Value {
	return reflect.ValueOf(t.Result).Elem()
}

type PullPageReq struct {
	Id   int64  `json:"id"`
	Uuid string `json:"uuid"`
	Size int32  `json:"size" validate:"gte=1,lte=20"`
}

type RadiusQueryReq struct {
	GeoPoint
	R float64 `json:"radius"`
}

func (t RadiusQueryReq) Where() DbContext {
	return RadiusQuery(t.GeoPoint, t.R)
}

func (t RadiusQueryReq) Contains(p GeoPoint) bool {
	return t.DistanceFrom(p) <= t.R
}

func Raw(raw interface{}) builder.Raw {
	str, ok := raw.(string)
	if ok {
		return builder.Raw(str)
	}

	bytes, ok := raw.([]byte)
	if ok {
		return builder.Raw(BytesToString(bytes))
	}

	return builder.Raw(fmt.Sprint(raw))
}

type BaseModel struct {
	Id        int64 `xorm:"id BIGINT(20)" db:"id" json:"id,omitempty" set:"-"`
	CreatedAt int64 `xorm:"INT(20) 'created_at'" db:"created_at" json:"created_at" set:"-"`
	UpdatedAt int64 `xorm:"INT(20) 'updated_at'" db:"updated_at" json:"updated_at" set:"-"`
}

func (t BaseModel) ParseOperator(raw Operator) BffOperator {
	res := BffOperator{
		CreatedBy:  raw.CreatedBy,
		CreatedAt:  t.CreatedAt,
		OperatedBy: raw.OperatedBy,
		OperatedAt: raw.OperatedAt,
	}

	return res
}

func CreateItem(tick ...int64) BaseModel {
	res := BaseModel{CreatedAt: ParseTick(tick)}

	return res
}

func CreateItemMs(tick ...int64) BaseModel {
	res := BaseModel{CreatedAt: ParseTickMs(tick)}

	return res
}

type PoModel struct {
	Id        int64  `db:"id" json:"id,omitempty" set:"-"`
	Uuid      string `xorm:"VARCHAR(36) 'uuid'" db:"uuid" json:"uuid,omitempty" set:"-"`
	CreatedAt int64  `xorm:"INT(20) 'created_at'" db:"created_at" json:"created_at" set:"-"`
	UpdatedAt int64  `xorm:"INT(20) 'updated_at'" db:"updated_at" json:"updated_at" set:"-"`
}

func (t PoModel) ParseOperator(raw Operator) BffOperator {
	res := BffOperator{
		CreatedBy:  raw.CreatedBy,
		CreatedAt:  t.CreatedAt,
		OperatedBy: raw.OperatedBy,
		OperatedAt: raw.OperatedAt,
	}

	return res
}

func NewPoModel(id ...string) PoModel {
	if len(id) > 0 && id[0] != "" {
		return PoModel{Uuid: id[0]}
	}

	return PoModel{Uuid: uuid.NewString()}
}

func NewPoModelWithTick(tick int64, raw ...string) PoModel {
	if len(raw) > 0 && raw[0] != "" {
		return PoModel{Uuid: raw[0], CreatedAt: tick}
	}

	return PoModel{Uuid: uuid.NewString(), CreatedAt: tick}
}

func CreatePo(tick ...int64) PoModel {
	res := PoModel{Uuid: uuid.NewString(), CreatedAt: ParseTick(tick)}

	return res
}

func CreatePoMs(tick ...int64) PoModel {
	res := PoModel{Uuid: uuid.NewString(), CreatedAt: ParseTickMs(tick)}

	return res
}

type DelModel struct {
	IsDel     int64 `xorm:"TINYINT(4)" db:"is_del" json:"is_del"`
	DeletedAt int64 `xorm:"INT(20) 'deleted_at'" db:"deleted_at" json:"deleted_at"`
}

func (t DelModel) ToDel() DeleteContext {
	return [3]string{IsDelDb, "1", "0"}
}

type BffOperator struct {
	CreatedBy  string `json:"created_by" set:"-"`
	CreatedAt  int64  `json:"created_at" set:"-"`
	OperatedBy string `json:"operated_by"`
	OperatedAt int64  `json:"operated_at"`
}

func (t BffOperator) ExecuteAt() int64 {
	return DeInt64Param(t.OperatedAt, t.CreatedAt)
}

type Operator struct {
	CreatedBy  string `xorm:"VARCHAR(64) 'created_by'" db:"created_by" json:"created_by"`
	OperatedBy string `xorm:"VARCHAR(64) 'operated_by'" db:"operated_by" json:"operated_by"`
	OperatedAt int64  `xorm:"VARCHAR(64) 'operated_at'" db:"operated_at" json:"operated_at"`
}

func NewOperator(ctx context.Context) Operator {
	return Creator(GetUser(ctx))
}

func Creator(user string) Operator {
	res := Operator{
		CreatedBy: user,
	}

	return res
}

func OneOperator(ctx context.Context) Operator {
	res := Operator{
		OperatedBy: GetUser(ctx),
	}

	return res
}

func (t Operator) Account() []string {
	return []string{t.CreatedBy, t.OperatedBy}
}

type Confirm struct {
	ConfirmedBy string `xorm:"VARCHAR(64) 'confirmed_by'" db:"confirmed_by" json:"confirmed_by"`
	ConfirmAt   int64  `xorm:"VARCHAR(64) 'confirmed_at'" db:"confirmed_at" json:"confirmed_at"`
}

func NewConfirm(ctx context.Context, tick ...int64) Confirm {
	res := Confirm{
		ConfirmedBy: GetUser(ctx),
		ConfirmAt:   ParseTick(tick),
	}

	return res
}

func Confirmer(user string, tick ...int64) Confirm {
	res := Confirm{
		ConfirmedBy: user,
		ConfirmAt:   ParseTick(tick),
	}

	return res
}

type BffProperty struct {
	Name        string          `json:"name" validate:"required"`     //名称
	Description string          `json:"description"`                  //描述
	Property    json.RawMessage `json:"property" validate:"required"` //属性
	Status      int32           `json:"status"`                       //状态
	Sort        int64           `json:"sort"`                         //排序
}

func (t BffProperty) BaseProperty() BaseProperty {
	res := BaseProperty{
		Name:        t.Name,
		Description: t.Description,
		Property:    t.JsonProperty(),
		Status:      t.Status,
		Sort:        t.Sort,
	}

	return res
}

func (t BffProperty) PropertyContext() DbContext {
	res := DbContext{
		"name":        t.Name,
		"description": t.Description,
		"property":    t.JsonProperty(),
		"status":      t.Status,
		"sort":        t.Sort,
	}

	return res
}

type BaseProperty struct {
	Name        string `db:"name"`        //名称
	Description string `db:"description"` //描述
	Property    string `db:"property"`    //属性
	Status      int32  `db:"status"`      //状态
	Sort        int64  `db:"sort"`        //排序
}

func (t BaseProperty) BffProperty(id int64) BffProperty {
	res := BffProperty{
		Name:        t.Name,
		Description: t.Description,
		Property:    t.JsonProperty(),
		Status:      t.Status,
		Sort:        DeInt64Param(t.Sort, id),
	}

	return res
}

func (t BaseProperty) PropertyContext(id int64, ig ...string) DbContext {
	res := DbContext{
		"name":        t.Name,
		"description": t.Description,
		"property":    t.JsonProperty(),
		"status":      t.Status,
		"sort":        DeInt64Param(t.Sort, id),
	}

	for _, v := range ig {
		delete(res, v)
	}

	return res
}

type ResourceBff struct {
	UUID
	Pid         string          `json:"pid" set:"-"`
	Name        string          `json:"name" validate:"required"`     //名称
	Description string          `json:"description"`                  //描述
	Status      int32           `json:"status"`                       //状态
	Sort        int64           `json:"sort"`                         //排序
	Property    json.RawMessage `json:"property" validate:"required"` //属性
}

func (t *ResourceBff) SetPid(p string) {
	t.Pid = p
}

func (t ResourceBff) ResourceItem() ResourceItem {
	res := ResourceItem{
		PoModel:     t.PoModel(),
		Pid:         t.Pid,
		Name:        t.Name,
		Description: t.Description,
		Status:      t.Status,
		Sort:        t.Sort,
		Property:    t.JsonProperty(),
	}

	return res
}

func (t ResourceBff) ToUpdate(raw interface{}, ignoreFiled ...string) DbContext {
	res := t.ToSetContext().
		MergeBffItem(raw, ignoreFiled...)

	return res
}

func (t ResourceBff) ToSetContext() DbContext {
	res := DbContext{
		"name":        t.Name,
		"description": t.Description,
		"status":      t.Status,
		"sort":        t.Sort,
		"property":    t.JsonProperty(),
	}

	return res
}

type ResourceItem struct {
	PoModel
	Pid         string `db:"pid"`
	Name        string `db:"name"`        //名称
	Description string `db:"description"` //描述
	Status      int32  `db:"status"`      //状态
	Sort        int64  `db:"sort"`        //排序
	Property    string `db:"property"`    //属性
}

func (t ResourceItem) Bff(id int64) ResourceBff {
	res := ResourceBff{
		Name:        t.Name,
		Description: t.Description,
		Status:      t.Status,
		Sort:        DeInt64Param(t.Sort, id),
		Property:    t.JsonProperty(),
	}

	return res
}

func (t ResourceItem) PropertyContext(id int64, ig ...string) DbContext {
	res := DbContext{
		"pid":         t.Pid,
		"name":        t.Name,
		"description": t.Description,
		"status":      t.Status,
		"sort":        DeInt64Param(t.Sort, id),
		"property":    t.JsonProperty(),
	}

	for _, v := range ig {
		delete(res, v)
	}

	return res
}

func (t ResourceItem) Fields() []string {
	return []string{"pid", "name", "description", "status", "sort", "property"}
}

func (t ResourceItem) GetStatus() int32 {
	return t.Status
}

func (t ResourceItem) GetSort(id int64) int64 {
	return DeInt64Param(t.Sort, id)
}
