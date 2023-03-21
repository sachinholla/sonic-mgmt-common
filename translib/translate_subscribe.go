////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2021 Broadcom. The term Broadcom refers to Broadcom Inc. and/or //
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

package translib

// This file contains utilities for invoking translateSubscribe and parse
// its response into notificationInfo.

import (
	"fmt"
	"strings"

	"github.com/Azure/sonic-mgmt-common/translib/internal/apis"
	"github.com/Azure/sonic-mgmt-common/translib/path"
	"github.com/golang/glog"
	"github.com/openconfig/gnmi/proto/gnmi"
)

// translatedPathInfo holds the response of translateSubscribe for a path
// along with additional context info. Provides methods for convert this into
// translatedSubData (i.e, notificationInfo) and to save to SubscribeSession.
type translatedPathInfo struct {
	path     string
	appInfo  *appInfo
	sContext *subscribeContext
	response *translateSubResponse // returned by transalateSubscribe
	trData   *translatedSubData    // notificationInfo's derived from response
}

// translatedSubData holds translated subscription data for a path.
type translatedSubData struct {
	targetInfos []*notificationInfo
	childInfos  []*notificationInfo
}

// translateSubscribe resolves app module for a given path and calls translateSubscribe.
// Returns a translatedPathInfo containing app's response and other context data.
func (sc *subscribeContext) translateSubscribe(reqPath string, mode NotificationType) (*translatedPathInfo, error) {
	sid := sc.id
	app, appInfo, err := getAppModule(reqPath, sc.version)
	if err != nil {
		return nil, err
	}

	recurseSubpaths := sc.recurse || mode != Sample // always recurse for on_change and target_defined
	glog.Infof("[%v] Calling translateSubscribe for path=\"%s\", mode=%v, recurse=%v",
		sid, reqPath, mode, recurseSubpaths)

	resp, err := (*app).translateSubscribe(
		&translateSubRequest{
			ctxID:   sid,
			path:    reqPath,
			mode:    mode,
			recurse: recurseSubpaths,
			dbs:     sc.dbs,
		})

	if err != nil {
		glog.Warningf("[%v] translateSubscribe failed for \"%s\"; err=%v", sid, reqPath, err)
		return nil, err
	}

	if resp == nil || len(resp.ntfAppInfoTrgt) == 0 {
		glog.Warningf("[%v] %T.translateSubscribe returned nil/empty response for path: %s", sid, *app, reqPath)
		var pType string
		if path.StrHasWildcardKey(reqPath) {
			pType = "wildcard "
		}
		return nil, fmt.Errorf("%spath not supported: %s", pType, reqPath)
	}

	glog.Infof("[%v] Path \"%s\" mapped to %d target and %d child notificationAppInfos",
		sid, reqPath, len(resp.ntfAppInfoTrgt), len(resp.ntfAppInfoTrgtChlds))
	for i, nAppInfo := range resp.ntfAppInfoTrgt {
		glog.Infof("[%v] targetInfo[%d] = %v", sid, i, nAppInfo)
	}
	for i, nAppInfo := range resp.ntfAppInfoTrgtChlds {
		glog.Infof("[%v] childInfo[%d] = %v", sid, i, nAppInfo)
	}

	return &translatedPathInfo{
		path:     reqPath,
		appInfo:  appInfo,
		sContext: sc,
		response: resp,
	}, nil
}

// getFromSession returns the translatedSubData for the path from the SubscribeSession.
// Returns nil if session does not exist or entry not found for the path.
func (sc *subscribeContext) getFromSession(path string, mode NotificationType) *translatedSubData {
	var trData *translatedSubData
	if sc.session != nil {
		trData = sc.session.get(path)
		glog.Infof("[%v] found trData %p from session for \"%s\"", sc.id, trData, path)
	}
	return trData
}

// getNInfos returns the translated data as a translatedSubData.
func (tpInfo *translatedPathInfo) getNInfos() *translatedSubData {
	if tpInfo.trData != nil {
		return tpInfo.trData
	}

	trResp := tpInfo.response
	targetLen := len(trResp.ntfAppInfoTrgt)
	childLen := len(trResp.ntfAppInfoTrgtChlds)
	trData := &translatedSubData{
		targetInfos: make([]*notificationInfo, targetLen),
		childInfos:  make([]*notificationInfo, childLen),
	}

	for i := 0; i < targetLen; i++ {
		trData.targetInfos[i] = tpInfo.newNInfo(trResp.ntfAppInfoTrgt[i])
	}

	for i := 0; i < childLen; i++ {
		trData.childInfos[i] = tpInfo.newNInfo(trResp.ntfAppInfoTrgtChlds[i])
	}

	tpInfo.trData = trData
	return trData
}

// newNInfo creates a new *notificationInfo from given *notificationAppInfo.
// Uses the context information from this tpInfo; but does not update it.
func (tpInfo *translatedPathInfo) newNInfo(nAppInfo *notificationAppInfo) *notificationInfo {
	nInfo := &notificationInfo{
		dbno:        nAppInfo.dbno,
		table:       nAppInfo.table,
		key:         nAppInfo.key,
		fields:      nAppInfo.dbFldYgPathInfoList,
		path:        nAppInfo.path,
		handler:     nAppInfo.handlerFunc,
		appInfo:     tpInfo.appInfo,
		sInfo:       tpInfo.sContext.sInfo,
		opaque:      nAppInfo.opaque,
		fldScanPatt: nAppInfo.fieldScanPattern,
		keyGroup:    nAppInfo.keyGroupComps,
	}

	// Make sure field prefix path has a leading and trailing "/".
	// Helps preparing full path later by joining parts
	for _, pi := range nInfo.fields {
		if path.StrHasWildcardKey(pi.rltvPath) {
			nInfo.flags.Set(niWildcardSubpath)
		}
		if len(pi.rltvPath) != 0 && pi.rltvPath[0] != '/' {
			pi.rltvPath = "/" + pi.rltvPath
		}
		// Look for fields mapped to yang key - formatted as "{xyz}"
		for _, leaf := range pi.dbFldYgPathMap {
			if len(leaf) != 0 && leaf[0] == '{' {
				nInfo.flags.Set(niKeyFields)
			}
		}
	}

	if nAppInfo.isLeafPath() {
		nInfo.flags.Set(niLeafPath)
	}
	switch nAppInfo.deleteAction {
	case apis.InspectPathOnDelete:
		nInfo.flags.Set(niDeleteAsUpdate)
	case apis.InspectLeafOnDelete:
		nInfo.flags.Set(niDeleteAsUpdate | niPartial)
	}
	if path.HasWildcardKey(nAppInfo.path) {
		nInfo.flags.Set(niWildcardPath)
	}
	if nAppInfo.isOnChangeSupported {
		nInfo.flags.Set(niOnChangeSupported)
	}
	if nAppInfo.isDataSrcDynamic {
		nInfo.flags.Set(niDynamic)
	}

	return nInfo
}

// saveToSession saves the derived notificationInfo data into the
// SubscribeSession.
func (tpInfo *translatedPathInfo) saveToSession() {
	session := tpInfo.sContext.session
	if session == nil {
		return
	}

	trData := tpInfo.getNInfos()
	path := tpInfo.path
	if trData == nil {
		glog.Infof("[%v] no trData for \"%s\"", tpInfo.sContext.id, path)
		return
	}
	glog.Infof("[%v] set trData %p in session for \"%s\"", tpInfo.sContext.id, trData, path)
	session.put(path, trData)
}

// segregateSampleSubpaths inspects the translateSubscribe response and prepares
// extra SAMPLE subscribe sub-paths for the TARGET_DEFINED subscription if needed.
// Returns nil if current path & its subtree prefers only ON_CHANGE or only SAMPLE.
// If path prefers ON_CHANGE but some subtree nodes prefer SAMPLE, only the ON_CHANGE
// related notificationAppInfo will be retained in this translatedPathInfo. Returns new
// translatedPathInfo for each of the subtree segments that prefer SAMPLE.
func (tpInfo *translatedPathInfo) segregateSampleSubpaths() []*translatedPathInfo {
	sid := tpInfo.sContext.id
	var samplePInfos []*translatedPathInfo
	var samplePaths []string
	var onchgInfos, sampleInfos []*notificationAppInfo

	glog.Infof("[%v] segregate on_change and sample segments for %s", sid, tpInfo.path)

	// Segregate ON_CHANGE and SAMPLE nAppInfos from targetInfos
	for _, nAppInfo := range tpInfo.response.ntfAppInfoTrgt {
		if !nAppInfo.isOnChangeSupported || nAppInfo.pType == Sample {
			sampleInfos = append(sampleInfos, nAppInfo)
		} else {
			onchgInfos = append(onchgInfos, nAppInfo)
		}
	}

	if glog.V(2) {
		glog.Infof("[%v] found %d on_change and %d sample targetInfos",
			sid, len(onchgInfos), len(sampleInfos))
	}

	// Path does not support ON_CHANGE. No extra re-grouping required.
	if len(onchgInfos) == 0 {
		glog.Infof("[%v] on_change={}; sample={%s}", sid, tpInfo.path)
		return nil
	}

	// Some targetInfos prefer SAMPLE mode.. Remove them from this translatedPathInfo
	// and create separate translatedPathInfo for them.
	if len(sampleInfos) != 0 {
		tpInfo.response.ntfAppInfoTrgt = onchgInfos
		for _, nAppInfo := range sampleInfos {
			matchingPInfo := findTranslatedParentInfo(samplePInfos, nAppInfo.path)
			if matchingPInfo != nil {
				// Matching entry exists.. append to its targetInfo.
				// Required for paths that map to multiple tables.
				matchingPInfo.response.ntfAppInfoTrgt = append(
					matchingPInfo.response.ntfAppInfoTrgt, nAppInfo)
			} else {
				newPInfo := tpInfo.cloneForSubPath(nAppInfo)
				samplePInfos = append(samplePInfos, newPInfo)
				samplePaths = append(samplePaths, newPInfo.path)
			}
		}
	}

	// Segregate ON_CHANGE and SAMPLE childInfos.
	onchgInfos = nil
	for _, nAppInfo := range tpInfo.response.ntfAppInfoTrgtChlds {
		if nAppInfo.isOnChangeSupported && nAppInfo.pType != Sample {
			onchgInfos = append(onchgInfos, nAppInfo)
			continue
		}

		parentPInfo := findTranslatedParentInfo(samplePInfos, nAppInfo.path)
		if parentPInfo != nil {
			// Parent path of nAppInfo is already recorded. Add as childInfo.
			parentPInfo.response.ntfAppInfoTrgtChlds = append(
				parentPInfo.response.ntfAppInfoTrgtChlds, nAppInfo)
		} else {
			// nAppInfo does not belong to any of the already recorded paths.
			// Treat it as a new targetInfo.
			newPInfo := tpInfo.cloneForSubPath(nAppInfo)
			samplePInfos = append(samplePInfos, newPInfo)
			samplePaths = append(samplePaths, newPInfo.path)
		}
	}

	if glog.V(2) {
		nSample := len(tpInfo.response.ntfAppInfoTrgtChlds) - len(onchgInfos)
		glog.Infof("[%v] found %d on_change and %d sample childInfos", sid, len(onchgInfos), nSample)
	}

	// Retain only ON_CHANGE childInfos in this translatedPathInfo.
	// SAMPLE mode childInfos would have been already moved to samplePInfos.
	if len(onchgInfos) != len(tpInfo.response.ntfAppInfoTrgtChlds) {
		tpInfo.response.ntfAppInfoTrgtChlds = onchgInfos
	}

	if samplePaths != nil {
		glog.Infof("[%v] on_change={%s}; sample={%s}", sid, tpInfo.path, strings.Join(samplePaths, ", "))
	} else {
		glog.Infof("[%v] on_change={%s}; sample={}", sid, tpInfo.path)
	}

	return samplePInfos
}

// cloneForSubPath creates a clone of this *translatedPathInfo with a
// fake response created from given *notificationAppInfo.
func (tpInfo *translatedPathInfo) cloneForSubPath(nAppInfo *notificationAppInfo) *translatedPathInfo {
	return &translatedPathInfo{
		path:     path.String(nAppInfo.path),
		appInfo:  tpInfo.appInfo,
		sContext: tpInfo.sContext,
		response: &translateSubResponse{
			ntfAppInfoTrgt: []*notificationAppInfo{nAppInfo},
		},
	}
}

func findTranslatedParentInfo(tpInfos []*translatedPathInfo, p *gnmi.Path) *translatedPathInfo {
	for _, tpInfo := range tpInfos {
		for _, targetInfo := range tpInfo.response.ntfAppInfoTrgt {
			if path.Matches(p, targetInfo.path) {
				return tpInfo
			}
		}
	}
	return nil
}

// translatedPathCache is the per-path cache in SubscribeSession.
type translatedPathCache struct {
	pathData map[string]*translatedSubData
}

func (tpCache *translatedPathCache) put(path string, trData *translatedSubData) {
	if tpCache.pathData == nil {
		tpCache.pathData = make(map[string]*translatedSubData)
	}
	tpCache.pathData[path] = trData
}

func (tpCache *translatedPathCache) get(path string) *translatedSubData {
	return tpCache.pathData[path]
}

func (tpCache *translatedPathCache) reset() {
	tpCache.pathData = nil
}
