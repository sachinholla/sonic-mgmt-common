////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or //
//  its subsidiaries.                                                         //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//     http://www.apache.org/licenses/LICENSE-2.0                             //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

/*
Package translib implements APIs like Create, Get, Subscribe etc.

to be consumed by the north bound management server implementations

This package takes care of translating the incoming requests to

Redis ABNF format and persisting them in the Redis DB.

It can also translate the ABNF format to YANG specific JSON IETF format

This package can also talk to non-DB clients.
*/

package translib

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/Azure/sonic-mgmt-common/translib/db"
	"github.com/Azure/sonic-mgmt-common/translib/tlerr"
	"github.com/Workiva/go-datastructures/queue"
	log "github.com/golang/glog"
	"github.com/openconfig/ygot/ygot"
)

// Write lock for all write operations to be synchronized
var writeMutex = &sync.Mutex{}

// Interval value for interval based subscription needs to be within the min and max
// minimum global interval for interval based subscribe in secs
var minSubsInterval = 20

// maximum global interval for interval based subscribe in secs
var maxSubsInterval = 600

type ErrSource int

const (
	ProtoErr ErrSource = iota
	AppErr
)

const (
	TRANSLIB_FMT_IETF_JSON = iota
	TRANSLIB_FMT_YGOT
)

type TranslibFmtType int

type UserRoles struct {
	Name  string
	Roles []string
}

type SetRequest struct {
	Path             string
	Payload          []byte
	User             UserRoles
	AuthEnabled      bool
	ClientVersion    Version
	DeleteEmptyEntry bool
}

type SetResponse struct {
	ErrSrc ErrSource
	Err    error
}

type GetRequest struct {
	Path          string
	FmtType       TranslibFmtType
	Ctxt          context.Context
	User          UserRoles
	AuthEnabled   bool
	ClientVersion Version

	// Depth limits the depth of data subtree in the response
	// payload. Default value 0 indicates there is no limit.
	Depth uint
}

type GetResponse struct {
	Payload   []byte
	ValueTree *ygot.ValidatedGoStruct
	ErrSrc    ErrSource
}

type ActionRequest struct {
	Path          string
	Payload       []byte
	User          UserRoles
	AuthEnabled   bool
	ClientVersion Version
}

type ActionResponse struct {
	Payload []byte
	ErrSrc  ErrSource
}

type BulkRequest struct {
	DeleteRequest  []SetRequest
	ReplaceRequest []SetRequest
	UpdateRequest  []SetRequest
	CreateRequest  []SetRequest
	User           UserRoles
	AuthEnabled    bool
	ClientVersion  Version
}

type BulkResponse struct {
	DeleteResponse  []SetResponse
	ReplaceResponse []SetResponse
	UpdateResponse  []SetResponse
	CreateResponse  []SetResponse
}

// SubscribeRequest holds the request data for Subscribe and Stream APIs.
type SubscribeRequest struct {
	Paths         []string
	Q             *queue.PriorityQueue
	Stop          chan struct{}
	User          UserRoles
	AuthEnabled   bool
	ClientVersion Version
	Session       *SubscribeSession
	Ctx           context.Context
}

type SubscribeResponse struct {
	Path         string
	Update       ygot.ValidatedGoStruct // updated values
	Delete       []string               // deleted paths - relative to Path
	Timestamp    int64
	SyncComplete bool
	IsTerminated bool
}

type NotificationType int

const (
	TargetDefined NotificationType = iota
	Sample
	OnChange
)

type IsSubscribeRequest struct {
	Paths         []IsSubscribePath
	User          UserRoles
	AuthEnabled   bool
	ClientVersion Version
	Session       *SubscribeSession
}

type IsSubscribePath struct {
	ID   uint32           // Path ID for correlating with IsSubscribeResponse
	Path string           // Subscribe path
	Mode NotificationType // Requested subscribe mode
}

type IsSubscribeResponse struct {
	ID                  uint32 // Path ID
	Path                string
	IsSubPath           bool // Subpath of the requested path
	IsOnChangeSupported bool
	IsWildcardSupported bool // true if wildcard keys are supported in the path
	MinInterval         int
	Err                 error
	PreferredType       NotificationType
}

// SubscribeSession is used to share session data between subscription
// related APIs - IsSubscribeSupported, Subscribe and Stream.
type SubscribeSession struct {
	ID string
	translatedPathCache
}

// Counter is a monotonically increasing unsigned integer.
type Counter uint64

type ModelData struct {
	Name string
	Org  string
	Ver  string
}

type notificationOpts struct {
	isOnChangeSupported bool
	mInterval           int
	pType               NotificationType // for TARGET_DEFINED
}

// initializes logging and app modules
func init() {
	log.Flush()
}

// Create - Creates entries in the redis DB pertaining to the path and payload
func Create(req SetRequest) (SetResponse, error) {
	var keys []db.WatchKeys
	var resp SetResponse
	path := req.Path
	payload := req.Payload
	if !isAuthorizedForSet(req) {
		return resp, tlerr.AuthorizationError{
			Format: "User is unauthorized for Create Operation",
			Path:   path,
		}
	}

	log.Info("Create request received with path =", path)
	log.Info("Create request received with payload =", string(payload))

	app, appInfo, err := getAppModule(path, req.ClientVersion)

	if err != nil {
		resp.ErrSrc = ProtoErr
		return resp, err
	}

	err = appInitialize(app, appInfo, path, &payload, nil, CREATE)

	if err != nil {
		resp.ErrSrc = AppErr
		return resp, err
	}

	writeMutex.Lock()
	defer writeMutex.Unlock()

	d, err := db.NewDB(getDBOptions(db.ConfigDB))

	if err != nil {
		resp.ErrSrc = ProtoErr
		return resp, err
	}

	defer d.DeleteDB()

	keys, err = (*app).translateCreate(d)

	if err != nil {
		resp.ErrSrc = AppErr
		return resp, err
	}

	err = d.StartTx(keys, appInfo.tablesToWatch)

	if err != nil {
		resp.ErrSrc = AppErr
		return resp, err
	}

	resp, err = (*app).processCreate(d)

	if err != nil {
		d.AbortTx()
		resp.ErrSrc = AppErr
		return resp, err
	}

	err = d.CommitTx()

	if err != nil {
		resp.ErrSrc = AppErr
	}

	return resp, err
}

// Update - Updates entries in the redis DB pertaining to the path and payload
func Update(req SetRequest) (SetResponse, error) {
	var keys []db.WatchKeys
	var resp SetResponse
	path := req.Path
	payload := req.Payload
	if !isAuthorizedForSet(req) {
		return resp, tlerr.AuthorizationError{
			Format: "User is unauthorized for Update Operation",
			Path:   path,
		}
	}

	log.Info("Update request received with path =", path)
	log.Info("Update request received with payload =", string(payload))

	app, appInfo, err := getAppModule(path, req.ClientVersion)

	if err != nil {
		resp.ErrSrc = ProtoErr
		return resp, err
	}

	err = appInitialize(app, appInfo, path, &payload, nil, UPDATE)

	if err != nil {
		resp.ErrSrc = AppErr
		return resp, err
	}

	writeMutex.Lock()
	defer writeMutex.Unlock()

	d, err := db.NewDB(getDBOptions(db.ConfigDB))

	if err != nil {
		resp.ErrSrc = ProtoErr
		return resp, err
	}

	defer d.DeleteDB()

	keys, err = (*app).translateUpdate(d)

	if err != nil {
		resp.ErrSrc = AppErr
		return resp, err
	}

	err = d.StartTx(keys, appInfo.tablesToWatch)

	if err != nil {
		resp.ErrSrc = AppErr
		return resp, err
	}

	resp, err = (*app).processUpdate(d)

	if err != nil {
		d.AbortTx()
		resp.ErrSrc = AppErr
		return resp, err
	}

	err = d.CommitTx()

	if err != nil {
		resp.ErrSrc = AppErr
	}

	return resp, err
}

// Replace - Replaces entries in the redis DB pertaining to the path and payload
func Replace(req SetRequest) (SetResponse, error) {
	var err error
	var keys []db.WatchKeys
	var resp SetResponse
	path := req.Path
	payload := req.Payload
	if !isAuthorizedForSet(req) {
		return resp, tlerr.AuthorizationError{
			Format: "User is unauthorized for Replace Operation",
			Path:   path,
		}
	}

	log.Info("Replace request received with path =", path)
	log.Info("Replace request received with payload =", string(payload))

	app, appInfo, err := getAppModule(path, req.ClientVersion)

	if err != nil {
		resp.ErrSrc = ProtoErr
		return resp, err
	}

	err = appInitialize(app, appInfo, path, &payload, nil, REPLACE)

	if err != nil {
		resp.ErrSrc = AppErr
		return resp, err
	}

	writeMutex.Lock()
	defer writeMutex.Unlock()

	d, err := db.NewDB(getDBOptions(db.ConfigDB))

	if err != nil {
		resp.ErrSrc = ProtoErr
		return resp, err
	}

	defer d.DeleteDB()

	keys, err = (*app).translateReplace(d)

	if err != nil {
		resp.ErrSrc = AppErr
		return resp, err
	}

	err = d.StartTx(keys, appInfo.tablesToWatch)

	if err != nil {
		resp.ErrSrc = AppErr
		return resp, err
	}

	resp, err = (*app).processReplace(d)

	if err != nil {
		d.AbortTx()
		resp.ErrSrc = AppErr
		return resp, err
	}

	err = d.CommitTx()

	if err != nil {
		resp.ErrSrc = AppErr
	}

	return resp, err
}

// Delete - Deletes entries in the redis DB pertaining to the path
func Delete(req SetRequest) (SetResponse, error) {
	var err error
	var keys []db.WatchKeys
	var resp SetResponse
	path := req.Path
	if !isAuthorizedForSet(req) {
		return resp, tlerr.AuthorizationError{
			Format: "User is unauthorized for Delete Operation",
			Path:   path,
		}
	}

	log.Info("Delete request received with path =", path)

	app, appInfo, err := getAppModule(path, req.ClientVersion)

	if err != nil {
		resp.ErrSrc = ProtoErr
		return resp, err
	}

	opts := appOptions{deleteEmptyEntry: req.DeleteEmptyEntry}
	err = appInitialize(app, appInfo, path, nil, &opts, DELETE)

	if err != nil {
		resp.ErrSrc = AppErr
		return resp, err
	}

	writeMutex.Lock()
	defer writeMutex.Unlock()

	d, err := db.NewDB(getDBOptions(db.ConfigDB))

	if err != nil {
		resp.ErrSrc = ProtoErr
		return resp, err
	}

	defer d.DeleteDB()

	keys, err = (*app).translateDelete(d)

	if err != nil {
		resp.ErrSrc = AppErr
		return resp, err
	}

	err = d.StartTx(keys, appInfo.tablesToWatch)

	if err != nil {
		resp.ErrSrc = AppErr
		return resp, err
	}

	resp, err = (*app).processDelete(d)

	if err != nil {
		d.AbortTx()
		resp.ErrSrc = AppErr
		return resp, err
	}

	err = d.CommitTx()

	if err != nil {
		resp.ErrSrc = AppErr
	}

	return resp, err
}

// Get - Gets data from the redis DB and converts it to northbound format
func Get(req GetRequest) (GetResponse, error) {
	var payload []byte
	var resp GetResponse
	path := req.Path
	if !isAuthorizedForGet(req) {
		return resp, tlerr.AuthorizationError{
			Format: "User is unauthorized for Get Operation",
			Path:   path,
		}
	}

	log.Info("Received Get request for path = ", path)

	app, appInfo, err := getAppModule(path, req.ClientVersion)

	if err != nil {
		resp = GetResponse{Payload: payload, ErrSrc: ProtoErr}
		return resp, err
	}

	opts := appOptions{depth: req.Depth}
	err = appInitialize(app, appInfo, path, nil, &opts, GET)

	if err != nil {
		resp = GetResponse{Payload: payload, ErrSrc: AppErr}
		return resp, err
	}

	dbs, err := getAllDbs(withWriteDisable)

	if err != nil {
		resp = GetResponse{Payload: payload, ErrSrc: ProtoErr}
		return resp, err
	}

	defer closeAllDbs(dbs[:])

	err = (*app).translateGet(dbs)

	if err != nil {
		resp = GetResponse{Payload: payload, ErrSrc: AppErr}
		return resp, err
	}

	resp, err = (*app).processGet(dbs, req.FmtType)

	return resp, err
}

func Action(req ActionRequest) (ActionResponse, error) {
	var payload []byte
	var resp ActionResponse
	path := req.Path

	if !isAuthorizedForAction(req) {
		return resp, tlerr.AuthorizationError{
			Format: "User is unauthorized for Action Operation",
			Path:   path,
		}
	}

	log.Info("Received Action request for path = ", path)

	app, appInfo, err := getAppModule(path, req.ClientVersion)

	if err != nil {
		resp = ActionResponse{Payload: payload, ErrSrc: ProtoErr}
		return resp, err
	}

	aInfo := *appInfo

	aInfo.isNative = true

	err = appInitialize(app, &aInfo, path, &req.Payload, nil, GET)

	if err != nil {
		resp = ActionResponse{Payload: payload, ErrSrc: AppErr}
		return resp, err
	}

	writeMutex.Lock()
	defer writeMutex.Unlock()

	dbs, err := getAllDbs()

	if err != nil {
		resp = ActionResponse{Payload: payload, ErrSrc: ProtoErr}
		return resp, err
	}

	defer closeAllDbs(dbs[:])

	err = (*app).translateAction(dbs)

	if err != nil {
		resp = ActionResponse{Payload: payload, ErrSrc: AppErr}
		return resp, err
	}

	resp, err = (*app).processAction(dbs)

	return resp, err
}

func Bulk(req BulkRequest) (BulkResponse, error) {
	var err error
	var keys []db.WatchKeys
	var errSrc ErrSource

	delResp := make([]SetResponse, len(req.DeleteRequest))
	replaceResp := make([]SetResponse, len(req.ReplaceRequest))
	updateResp := make([]SetResponse, len(req.UpdateRequest))
	createResp := make([]SetResponse, len(req.CreateRequest))

	resp := BulkResponse{DeleteResponse: delResp,
		ReplaceResponse: replaceResp,
		UpdateResponse:  updateResp,
		CreateResponse:  createResp}

	if !isAuthorizedForBulk(req) {
		return resp, tlerr.AuthorizationError{
			Format: "User is unauthorized for Action Operation",
		}
	}

	writeMutex.Lock()
	defer writeMutex.Unlock()

	d, err := db.NewDB(getDBOptions(db.ConfigDB))

	if err != nil {
		return resp, err
	}

	defer d.DeleteDB()

	//Start the transaction without any keys or tables to watch will be added later using AppendWatchTx
	err = d.StartTx(nil, nil)

	if err != nil {
		return resp, err
	}

	for i := range req.DeleteRequest {
		path := req.DeleteRequest[i].Path
		opts := appOptions{deleteEmptyEntry: req.DeleteRequest[i].DeleteEmptyEntry}

		log.Info("Delete request received with path =", path)

		app, appInfo, err := getAppModule(path, req.DeleteRequest[i].ClientVersion)

		if err != nil {
			errSrc = ProtoErr
			goto BulkDeleteError
		}

		err = appInitialize(app, appInfo, path, nil, &opts, DELETE)

		if err != nil {
			errSrc = AppErr
			goto BulkDeleteError
		}

		keys, err = (*app).translateDelete(d)

		if err != nil {
			errSrc = AppErr
			goto BulkDeleteError
		}

		err = d.AppendWatchTx(keys, appInfo.tablesToWatch)

		if err != nil {
			errSrc = AppErr
			goto BulkDeleteError
		}

		resp.DeleteResponse[i], err = (*app).processDelete(d)

		if err != nil {
			errSrc = AppErr
		}

	BulkDeleteError:

		if err != nil {
			d.AbortTx()
			resp.DeleteResponse[i].ErrSrc = errSrc
			resp.DeleteResponse[i].Err = err
			return resp, err
		}
	}

	for i := range req.ReplaceRequest {
		path := req.ReplaceRequest[i].Path
		payload := req.ReplaceRequest[i].Payload

		log.Info("Replace request received with path =", path)

		app, appInfo, err := getAppModule(path, req.ReplaceRequest[i].ClientVersion)

		if err != nil {
			errSrc = ProtoErr
			goto BulkReplaceError
		}

		log.Info("Bulk replace request received with path =", path)
		log.Info("Bulk replace request received with payload =", string(payload))

		err = appInitialize(app, appInfo, path, &payload, nil, REPLACE)

		if err != nil {
			errSrc = AppErr
			goto BulkReplaceError
		}

		keys, err = (*app).translateReplace(d)

		if err != nil {
			errSrc = AppErr
			goto BulkReplaceError
		}

		err = d.AppendWatchTx(keys, appInfo.tablesToWatch)

		if err != nil {
			errSrc = AppErr
			goto BulkReplaceError
		}

		resp.ReplaceResponse[i], err = (*app).processReplace(d)

		if err != nil {
			errSrc = AppErr
		}

	BulkReplaceError:

		if err != nil {
			d.AbortTx()
			resp.ReplaceResponse[i].ErrSrc = errSrc
			resp.ReplaceResponse[i].Err = err
			return resp, err
		}
	}

	for i := range req.UpdateRequest {
		path := req.UpdateRequest[i].Path
		payload := req.UpdateRequest[i].Payload

		log.Info("Update request received with path =", path)

		app, appInfo, err := getAppModule(path, req.UpdateRequest[i].ClientVersion)

		if err != nil {
			errSrc = ProtoErr
			goto BulkUpdateError
		}

		err = appInitialize(app, appInfo, path, &payload, nil, UPDATE)

		if err != nil {
			errSrc = AppErr
			goto BulkUpdateError
		}

		keys, err = (*app).translateUpdate(d)

		if err != nil {
			errSrc = AppErr
			goto BulkUpdateError
		}

		err = d.AppendWatchTx(keys, appInfo.tablesToWatch)

		if err != nil {
			errSrc = AppErr
			goto BulkUpdateError
		}

		resp.UpdateResponse[i], err = (*app).processUpdate(d)

		if err != nil {
			errSrc = AppErr
		}

	BulkUpdateError:

		if err != nil {
			d.AbortTx()
			resp.UpdateResponse[i].ErrSrc = errSrc
			resp.UpdateResponse[i].Err = err
			return resp, err
		}
	}

	for i := range req.CreateRequest {
		path := req.CreateRequest[i].Path
		payload := req.CreateRequest[i].Payload

		log.Info("Create request received with path =", path)

		app, appInfo, err := getAppModule(path, req.CreateRequest[i].ClientVersion)

		if err != nil {
			errSrc = ProtoErr
			goto BulkCreateError
		}

		err = appInitialize(app, appInfo, path, &payload, nil, CREATE)

		if err != nil {
			errSrc = AppErr
			goto BulkCreateError
		}

		keys, err = (*app).translateCreate(d)

		if err != nil {
			errSrc = AppErr
			goto BulkCreateError
		}

		err = d.AppendWatchTx(keys, appInfo.tablesToWatch)

		if err != nil {
			errSrc = AppErr
			goto BulkCreateError
		}

		resp.CreateResponse[i], err = (*app).processCreate(d)

		if err != nil {
			errSrc = AppErr
		}

	BulkCreateError:

		if err != nil {
			d.AbortTx()
			resp.CreateResponse[i].ErrSrc = errSrc
			resp.CreateResponse[i].Err = err
			return resp, err
		}
	}

	err = d.CommitTx()

	return resp, err
}

// NewSubscribeSession creates a new SubscribeSession. Caller
// MUST close the session object through CloseSubscribeSession
// call at the end.
func NewSubscribeSession() *SubscribeSession {
	return &SubscribeSession{
		ID: fmt.Sprintf("%d", subscribeCounter.Next()),
	}
}

// Close a SubscribeSession and release all resources it held by it.
// API client MUST close the sessions it creates; and not reuse the session after closing.
func (ss *SubscribeSession) Close() {
	if ss != nil {
		ss.reset()
	}
}

// Subscribe - Subscribes to the paths requested and sends notifications when the data changes in DB
func Subscribe(req SubscribeRequest) error {
	paths := req.Paths
	q := req.Q
	stop := req.Stop

	dbs, err := getAllDbs(withWriteDisable, withOnChange)

	if err != nil {
		return err
	}

	sInfo := &subscribeInfo{
		id:   subscribeCounter.Next(),
		q:    q,
		stop: stop,
		dbs:  dbs,
	}

	sCtx := subscribeContext{
		sInfo:   sInfo,
		dbs:     dbs,
		version: req.ClientVersion,
		session: req.Session,
		recurse: true,
	}

	for _, path := range paths {
		err = sCtx.translateAndAddPath(path, OnChange)
		if err != nil {
			closeAllDbs(dbs[:])
			return err
		}
	}

	// Start db subscription and exit. DB objects will be
	// closed automatically when the subscription ends.
	err = sCtx.startSubscribe()

	return err
}

// Stream function streams the value for requested paths through a queue.
// Unlike Get, this function can return smaller chunks of response separately.
// Individual chunks are packed in a SubscribeResponse object and pushed to the req.Q.
// Pushes a SubscribeResponse with SyncComplete=true after data are pushed.
// Function will block until all values are returned. This can be used for
// handling "Sample" subscriptions (NotificationType.Sample).
// Client should be authorized to perform "subscribe" operation.
func Stream(req SubscribeRequest) error {
	sid := subscribeCounter.Next()
	log.Infof("[%v] Stream request rcvd for paths %v", sid, req.Paths)

	dbs, err := getAllDbs(withWriteDisable)
	if err != nil {
		return err
	}
	defer closeAllDbs(dbs[:])

	sc := subscribeContext{
		id:      sid,
		dbs:     dbs,
		version: req.ClientVersion,
		session: req.Session,
	}

	for _, path := range req.Paths {
		err := sc.translateAndAddPath(path, Sample)
		if err != nil {
			return err
		}
	}

	sInfo := &subscribeInfo{
		id:  sid,
		q:   req.Q,
		dbs: dbs,
		ctx: req.Ctx,
	}

	for _, nInfo := range sc.tgtInfos {
		err = sendInitialUpdate(sInfo, nInfo)
		if err != nil {
			return err
		}
	}

	// Push a SyncComplete message at the end
	sInfo.syncDone = true
	sendSyncNotification(sInfo, false)
	return nil
}

// IsSubscribeSupported - Check if subscribe is supported on the given paths
func IsSubscribeSupported(req IsSubscribeRequest) ([]*IsSubscribeResponse, error) {
	reqID := subscribeCounter.Next()
	paths := req.Paths
	resp := make([]*IsSubscribeResponse, len(paths))

	for i := range resp {
		resp[i] = newIsSubscribeResponse(paths[i].ID, paths[i].Path)
	}

	log.Infof("[%v] IsSubscribeSupported: %v", reqID, paths)

	dbs, err := getAllDbs(withWriteDisable)

	if err != nil {
		return resp, err
	}

	defer closeAllDbs(dbs[:])

	sc := subscribeContext{
		id:      reqID,
		dbs:     dbs,
		version: req.ClientVersion,
		session: req.Session,
		recurse: true,
	}

	for i, p := range paths {
		trInfo, errApp := sc.translateSubscribe(p.Path, p.Mode)
		if errApp != nil {
			resp[i].Err = errApp
			err = errApp
			continue
		}

		// Split target_defined request into separate on_change and sample
		// sub-requests if required.
		if p.Mode == TargetDefined {
			for _, xInfo := range trInfo.segregateSampleSubpaths() {
				xr := newIsSubscribeResponse(p.ID, xInfo.path)
				xr.IsSubPath = true
				resp = append(resp, xr)
				collectNotificationPreferences(xInfo.response.ntfAppInfoTrgt, xr)
				collectNotificationPreferences(xInfo.response.ntfAppInfoTrgtChlds, xr)
				xInfo.saveToSession()
			}
		}

		r := resp[i]
		collectNotificationPreferences(trInfo.response.ntfAppInfoTrgt, r)
		collectNotificationPreferences(trInfo.response.ntfAppInfoTrgtChlds, r)
		trInfo.saveToSession()
	}

	log.Infof("[%v] IsSubscribeSupported: returning %d IsSubscribeResponse; err=%v", reqID, len(resp), err)
	if log.V(1) {
		for i, r := range resp {
			log.Infof("[%v] IsSubscribeResponse[%d]: path=%s, onChg=%v, pref=%v, minInt=%d, err=%v",
				reqID, i, r.Path, r.IsOnChangeSupported, r.PreferredType, r.MinInterval, r.Err)
		}
	}

	return resp, err
}

func newIsSubscribeResponse(id uint32, path string) *IsSubscribeResponse {
	return &IsSubscribeResponse{
		ID:                  id,
		Path:                path,
		IsOnChangeSupported: true,
		IsWildcardSupported: true,
		MinInterval:         minSubsInterval,
		PreferredType:       OnChange,
	}
}

// collectNotificationPreferences computes overall notification preferences (is on-change
// supported, min sample interval, preferred mode etc) by combining individual table preferences
// from the notificationAppInfo array. Writes them to the IsSubscribeResponse object 'resp'.
func collectNotificationPreferences(nAppInfos []*notificationAppInfo, resp *IsSubscribeResponse) {
	if len(nAppInfos) == 0 {
		return
	}

	for _, nInfo := range nAppInfos {
		if !nInfo.isOnChangeSupported {
			resp.IsOnChangeSupported = false
		}
		if nInfo.isNonDB() {
			resp.IsWildcardSupported = false
			resp.IsOnChangeSupported = false
			resp.PreferredType = Sample
		}
		if nInfo.pType == Sample {
			resp.PreferredType = Sample
		}
		if nInfo.mInterval > resp.MinInterval {
			resp.MinInterval = nInfo.mInterval
		}
	}

	if resp.MinInterval > maxSubsInterval {
		resp.MinInterval = maxSubsInterval
	}
}

// GetModels - Gets all the models supported by Translib
func GetModels() ([]ModelData, error) {
	var err error

	return getModels(), err
}

// Creates connection will all the redis DBs. To be used for get request
func getAllDbs(opts ...func(*db.Options)) ([db.MaxDB]*db.DB, error) {
	var dbs [db.MaxDB]*db.DB
	var err error
	for dbNum := db.DBNum(0); dbNum < db.MaxDB; dbNum++ {
		dbs[dbNum], err = db.NewDB(getDBOptions(dbNum, opts...))
		if err != nil {
			closeAllDbs(dbs[:])
			break
		}
	}

	return dbs, err
}

// Closes the dbs, and nils out the arr.
func closeAllDbs(dbs []*db.DB) {
	for dbsi, d := range dbs {
		if d != nil {
			d.DeleteDB()
			dbs[dbsi] = nil
		}
	}
}

// Compare - Implement Compare method for priority queue for SubscribeResponse struct
func (val SubscribeResponse) Compare(other queue.Item) int {
	o := other.(*SubscribeResponse)
	if val.Timestamp > o.Timestamp {
		return 1
	} else if val.Timestamp == o.Timestamp {
		return 0
	}
	return -1
}

func getDBOptions(dbNo db.DBNum, opts ...func(*db.Options)) db.Options {
	o := db.Options{DBNo: dbNo}
	for _, setopt := range opts {
		setopt(&o)
	}
	return o
}

func withWriteDisable(o *db.Options) {
	o.IsWriteDisabled = true
}

func withOnChange(o *db.Options) {
	o.IsEnableOnChange = true
}

func getAppModule(path string, clientVer Version) (*appInterface, *appInfo, error) {
	var app appInterface

	aInfo, err := getAppModuleInfo(path)

	if err != nil {
		return nil, aInfo, err
	}

	if err := validateClientVersion(clientVer, path, aInfo); err != nil {
		return nil, aInfo, err
	}

	app, err = getAppInterface(aInfo.appType)

	if err != nil {
		return nil, aInfo, err
	}

	return &app, aInfo, err
}

func appInitialize(app *appInterface, appInfo *appInfo, path string, payload *[]byte, opts *appOptions, opCode int) error {
	var err error
	var input []byte

	if payload != nil {
		input = *payload
	}

	if appInfo.isNative {
		log.Info("Native MSFT format")
		data := appData{path: path, payload: input}
		data.setOptions(opts)
		(*app).initialize(data)
	} else {
		ygotStruct, ygotTarget, err := getRequestBinder(&path, payload, opCode, &(appInfo.ygotRootType)).unMarshall()
		if err != nil {
			log.Info("Error in request binding: ", err)
			return err
		}

		data := appData{path: path, payload: input, ygotRoot: ygotStruct, ygotTarget: ygotTarget}
		data.setOptions(opts)
		(*app).initialize(data)
	}

	return err
}

func (data *appData) setOptions(opts *appOptions) {
	if opts != nil {
		data.appOptions = *opts
	}
}

func (nt NotificationType) String() string {
	switch nt {
	case TargetDefined:
		return "TargetDefined"
	case Sample:
		return "Sample"
	case OnChange:
		return "OnChange"
	default:
		return fmt.Sprintf("NotificationType(%d)", nt)
	}
}

// Next increments the counter and returns the new value
func (c *Counter) Next() uint64 {
	return atomic.AddUint64((*uint64)(c), 1)
}
