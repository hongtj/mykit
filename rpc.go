package core

import (
	"mykit/core/smarter"
	"mykit/core/transfer"
	"mykit/core/types"
)

var (
	rpcSection = "rpc"
	rpcRoot    = "/" + rpcSection
)

const (
	RpcMethodHeartbeat        = "heartbeat"
	RpcMethodGetTenant        = "getTenant"
	RpcMethodGrantRole        = "grantRole"
	RpcMethodSetRoleAuthority = "setRoleAuthority"
)

const (
	RpcMethodGetFreshAlert     = "getFreshAlert"
	RpcMethodGetHistoryAlert   = "getHistoryAlert"
	RpcMethodGetHistoryProduce = "getHistoryProduce"
	RpcMethodLookOverAlert     = "lookOverAlert"
)

const (
	RpcMethodGetDayReport       = "getDayReport"
	RpcMethodGetMonthReport     = "getMonthReport"
	RpcMethodGetSomeDayReport   = "getSomeDayReport"
	RpcMethodGetSomeMonthReport = "getSomeMonthReport"
)

func RpcRoot() string {
	return rpcRoot
}

var (
	RpcPrefix = rpcSection + ".lixx."
	RpcScada  = "scada"
	RpcMes    = "mes"
	RpcDpc    = "dpc"
)

func RpcEndpoint(srv string) string {
	return Project + rpcDelimiter + Tenant + rpcDelimiter + srv
}

//reg rpc

func InitRpc(prefix string) {
	types.SetStrValue(&RpcPrefix, prefix)

	transfer.InitRpcPrefix(RpcPrefix,
		&RpcScada,
		&RpcMes,
		&RpcDpc,
		&RpcWms,
	)

	initDispatch() //业务app和rpc节点映射关系
}

func initDispatch() {
	RegRpcScada()
	RegRpcMes()
	RegRpcDpc()
	RegRpcWms()
}

func RegRpcScada() {
	smarter.RegRpc(RpcScada,
		AppAuth,
		AppCamera,
		AppChart,
		AppMd,
		AppMonitor,
		AppResource,
		AppStock,
		AppTopology,
		AppUser,
	)
}

func RegRpcMes() {
	smarter.RegRpc(RpcMes,
		AppMes,
		AppCollect,
		AppQuality,
		AppForm,
		AppDict,
		AppCategory,
		AppLog,
		AppView,
	)
}

func RegRpcDpc() {
	smarter.RegRpc(RpcDpc)
}

//reg handler

var AddAuthHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppAuth, f...)
}

var AddCameraHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppCamera, f...)
}

var AddChartHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppChart, f...)
}

var AddCollectHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppCollect, f...)
}

var AddLogHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppLog, f...)
}

var AddMdHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppMd, f...)
}

var AddMesHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppMes, f...)
}

var AddViewHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppView, f...)
}

var AddFormHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppForm, f...)
}

var AddQualityHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppQuality, f...)
}

var AddMonitorHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppMonitor, f...)
}

var AddResourceHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppResource, f...)
}

var AddStockHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppStock, f...)
}

var AddTopologyHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppTopology, f...)
}

var AddUserHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppUser, f...)
}

var AddDictHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppDict, f...)
}

var AddCategoryHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppCategory, f...)
}

var AddStorehouseHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppStorehouse, f...)
}

var AddMsgHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppMessage, f...)
}

var AddOssHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppOss, f...)
}

var AddProcurementHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppProcurement, f...)
}

var AddProduceHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppProduce, f...)
}

var AddApprovalHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppApproval, f...)
}

var AddSalesHandler = func(f ...transfer.Handler) {
	smarter.AddHandler(AppSales, f...)
}

var (
	RpcWms = "wms"
)

func RegRpcWms() {
	smarter.RegRpc(RpcWms,
		AppInventory,
	)
}
