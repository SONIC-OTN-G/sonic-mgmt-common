//////////////////////////////////////////////////////////////////////////
//
// OTN 告警上报(方案一)—— alarm_app
// 注册 /openconfig-alarms:alarms 路径;isNative 模式(raw JSON payload),
// 不依赖 ygot binding(ocbinds 重生成验证在 deb 构建,见 plan Task 5)。
// GET 实现见 Task 4;订阅实现见 Task 8。
//
//////////////////////////////////////////////////////////////////////////

package translib

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/Azure/sonic-mgmt-common/translib/alarm"
	"github.com/Azure/sonic-mgmt-common/translib/db"
	log "github.com/golang/glog"
)

const alarmPath = "/openconfig-alarms:alarms"

type alarmApp struct {
	path     *PathInfo
	appDb    *db.DB
	respJSON interface{}
}

func init() {
	log.Info("Init called for alarm module")
	err := register(alarmPath,
		&appInfo{appType: reflect.TypeOf(alarmApp{}),
			isNative: true})
	if err != nil {
		log.Fatal("Register alarm app module with App Interface failed with error=", err)
	}

	err = addModel(&ModelData{Name: "openconfig-alarms",
		Org: "OpenConfig working group",
		Ver: "0.3.2"})
	if err != nil {
		log.Fatal("Adding alarm model data to appinterface failed with error=", err)
	}
}

func (app *alarmApp) initialize(data appData) {
	log.Info("initialize:alarm:path =", data.path)
	*app = alarmApp{path: NewPathInfo(data.path)}
}

func (app *alarmApp) translateCreate(d *db.DB) ([]db.WatchKeys, error) {
	return nil, errors.New("Not implemented")
}

func (app *alarmApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error) {
	return nil, errors.New("Not implemented")
}

func (app *alarmApp) translateReplace(d *db.DB) ([]db.WatchKeys, error) {
	return nil, errors.New("Not implemented")
}

func (app *alarmApp) translateDelete(d *db.DB) ([]db.WatchKeys, error) {
	return nil, errors.New("Not implemented")
}

func (app *alarmApp) translateGet(dbs [db.MaxDB]*db.DB) error {
	return nil
}

func (app *alarmApp) translateAction(dbs [db.MaxDB]*db.DB) error {
	return errors.New("Not implemented")
}

func (app *alarmApp) translateSubscribe(req translateSubRequest) (translateSubResponse, error) {
	// Task 8 实现订阅
	return translateSubResponse{}, errors.New("Not implemented")
}

func (app *alarmApp) processCreate(d *db.DB) (SetResponse, error) {
	return SetResponse{}, errors.New("Not implemented")
}

func (app *alarmApp) processUpdate(d *db.DB) (SetResponse, error) {
	return SetResponse{}, errors.New("Not implemented")
}

func (app *alarmApp) processReplace(d *db.DB) (SetResponse, error) {
	return SetResponse{}, errors.New("Not implemented")
}

func (app *alarmApp) processDelete(d *db.DB) (SetResponse, error) {
	return SetResponse{}, errors.New("Not implemented")
}

// dbCuralarmStore 把 translib/db 的 STATE_DB 访问包装成 alarm.CuralarmStore
type dbCuralarmStore struct {
	dbh *db.DB
}

func (s *dbCuralarmStore) Keys() ([]string, error) {
	tbl, err := s.dbh.GetTable(&db.TableSpec{Name: "CURALARM"})
	if err != nil {
		return nil, err
	}
	dbKeys, err := tbl.GetKeys()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(dbKeys))
	for _, k := range dbKeys {
		out = append(out, k.Get(0))
	}
	return out, nil
}

func (s *dbCuralarmStore) HGetAll(key string) (map[string]string, error) {
	v, err := s.dbh.GetEntry(&db.TableSpec{Name: "CURALARM"}, db.Key{Comp: []string{key}})
	if err != nil {
		return nil, err
	}
	return v.Field, nil
}

// NewCuralarmStore 供外部进程(netconfd)构造 STATE_DB CURALARM 读取器。
// 自建 db 句柄(getDBOptions 同款 Options:StateDB 分隔符必须是 "|",
// 裸 Options 会导致 redis2key 分裂出垃圾键);调用方负责不复用时关闭?
// —— db 句柄生命周期 = 进程级,无需显式关闭。
func NewCuralarmStore() (alarm.CuralarmStore, error) {
	dbh, err := db.NewDB(db.Options{DBNo: db.StateDB, TableNameSeparator: "|", KeySeparator: "|"})
	if err != nil {
		return nil, err
	}
	return &dbCuralarmStore{dbh: dbh}, nil
}

func (app *alarmApp) processGet(dbs [db.MaxDB]*db.DB, fmtType TranslibFmtType) (GetResponse, error) {
	st := &dbCuralarmStore{dbh: dbs[db.StateDB]}
	entries, err := alarm.GetAlarmEntries(st)
	if err != nil {
		return GetResponse{}, err
	}
	payload := map[string]interface{}{
		"openconfig-alarms:alarms": map[string]interface{}{"alarm": entries},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return GetResponse{}, err
	}
	return GetResponse{Payload: b}, nil
}

func (app *alarmApp) processAction(dbs [db.MaxDB]*db.DB) (ActionResponse, error) {
	return ActionResponse{}, errors.New("Not implemented")
}

func (app *alarmApp) processSubscribe(req processSubRequest) (processSubResponse, error) {
	// Task 8 实现
	return processSubResponse{}, errors.New("Not implemented")
}
