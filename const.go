package core

import (
	"mykit/core/persist"
	"mykit/core/types"
)

const (
	ConfigFile = "config.yml"
	IniFile    = "config.ini"
)

const (
	dbDelimiter       = "_"
	rpcDelimiter      = "."
	RedisKeyDelimiter = types.RedisKeyDelimiter
	NumDelimiter      = types.NumDelimiter
	CacheKeyStr       = persist.CacheKeyStr
	CacheKeyHash      = persist.CacheKeyHash
	CacheKeyList      = persist.CacheKeyList
	CacheKeySet       = persist.CacheKeySet
	CacheKeyZset      = persist.CacheKeyZset
)

const (
	ConfigKeyPlatform = "platform"
)

const (
	PlatForPc          = "1"
	PlatForAdmin       = "2"
	PlatForWorkStation = "3"
	Plat4              = "4"
)

const (
	Delta1M = types.SecondOf1M //1分钟的秒数
	Delta3M = Delta1M * 3      //3分钟的秒数
	Delta1H = Delta1M * 60
	Delta1D = Delta1H * 24
)

const (
	MsDelta1M = Delta1M * 1000
	MsDelta3M = MsDelta1M * 3
)

const (
	AppAuth       = "auth"
	AppCamera     = "camera"
	AppChart      = "chart"
	AppCollect    = "collect"
	AppLog        = "log"
	AppMd         = "md"
	AppMes        = "mes"
	AppQuality    = "quality"
	AppMonitor    = "monitor"
	AppResource   = "resource"
	AppStock      = "stock"
	AppTopology   = "topology"
	AppTenant     = "tenant"
	AppUser       = "user"
	AppDict       = "dict"
	AppCategory   = "category"
	AppStorehouse = "storehouse"
	AppMessage    = "msg"
	AppOss        = "oss"
	AppForm       = "form"
	AppView       = "view"
)

const (
	AppInventory = "inventory"
)

const (
	AppProcurement = "procurement"
	AppProduce     = "produce"
	AppSales       = "sales"
	AppApproval    = "approval"
)

const (
	UserStateNormal     = 1
	UserStateForbidden  = -1
	UserStateDeleted    = -2
	UserStateNormalStr  = "1"
	UserStateDeletedStr = "-2"
)

const (
	LoginSuccess  = 1
	LoginFailPwd  = -11
	LoginFailPlat = -21
)

const (
	MdDictSensor = 1
)

const (
	ResourceStateStop  = 0 //生产线停工
	ResourceStateWork  = 1 //生产线开工
	ResourceStateFault = 2 //生产线故障
)

const (
	PlanStatePrepare      = 0 //生产线准备开工
	PlanStateNormal       = 1 //生产线进度正常
	PlanStateLag          = 2 //生产线进度滞后
	PlanStateFault        = 3 //生产线进度因故障停止
	PlanStateComplete     = 4 //生产线计划完成
	PlanStateDescPrepare  = "准备"
	PlanStateDescNormal   = "正常"
	PlanStateDescLag      = "滞后"
	PlanStateDescFault    = "故障未完成"
	PlanStateDescComplete = "完成"
	//计划整体状态
	PlanWholeUndone = 0 //未完成
	PlanWholeDone   = 1 //已完成
)

const (
	ToSendStatus = 0
	SentStatus   = 1
)

const (
	NoAlert           = 0
	AlertOpenStatus   = 1
	AlertClosedStatus = 2
	RecoverStatus     = 9
)

const (
	NotLookOveredStatus = 0
	LookOveredStatus    = 1
)

const (
	ProductPlanStateNotDo    = 0 //未执行
	ProductPlanStateRunning  = 1 //正在执行
	ProductPlanStateComplete = 2 //完成
	ProductPlanStatePause    = 3 //暂停
)

const (
	ResourceTypeGroup    = "jt"
	ResourceTypeCompany  = "gs"
	ResourceTypeWorkshop = "cj"
	ResourceTypeScx      = "scx"
	ResourceTypeDevice   = "sb"
)

const (
	KindResource = "r"
	KindTeam     = "t"
)
