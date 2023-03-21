////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2022 Broadcom. The term Broadcom refers to Broadcom Inc. and/or //
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

/* This file includes debug APIs for the subscription feature */

package translib

import (
	"context"
	"strings"

	"github.com/Azure/sonic-mgmt-common/translib/path"
	"github.com/Azure/sonic-mgmt-common/translib/tlerr"
	"github.com/Azure/sonic-mgmt-common/translib/transformer"
	"github.com/openconfig/gnmi/proto/gnmi"
)

// SubscribePrefsRequest holds the request parameters for the GetSubscribePreferences API.
// It extends the IsSubscribeRequest to add context and response channel info.
type SubscribePrefsRequest struct {
	IsSubscribeRequest
	Recurse bool                      // recurse into subpaths
	Context context.Context           // API context
	Channel chan *IsSubscribeResponse // write responses to this channel
}

// GetSubscribePreferences returns subscribe preferences for the requested
// paths and its subpaths. Preferences are written to the output channel
// provided by the caller (req.Channel).
func GetSubscribePreferences(req *SubscribePrefsRequest) error {
	reqID := subscribeCounter.Next()
	dbs, err := getAllDbs(withWriteDisable)
	if err != nil {
		return err
	}

	defer closeAllDbs(dbs[:])

	sc := subscribeContext{
		id:      reqID,
		dbs:     dbs,
		version: req.ClientVersion,
	}

	for _, p := range req.Paths {
		trInfo, err := sc.translateSubscribe(p.Path, TargetDefined)
		if err != nil {
			return err
		}
		spw := subscribePrefsWorker{req: req, reqPath: p}
		if err := spw.Process(trInfo.response); err != nil {
			return err
		}
	}

	req.Channel <- nil // end of data marker
	return nil
}

// notificationAppInfoGroup is an array of notificationAppInfo objects
// with same path.
type notificationAppInfoGroup []*notificationAppInfo

func (ng *notificationAppInfoGroup) Path() *gnmi.Path {
	if ng == nil || len(*ng) == 0 {
		return nil
	}
	return (*ng)[0].path
}

type subscribePrefsWorker struct {
	req     *SubscribePrefsRequest // for sending responses
	reqPath IsSubscribePath

	targetGroups []*notificationAppInfoGroup
	childGroups  []*notificationAppInfoGroup
}

// Process translateSubscribe response and stream out subscription preferences.
func (spw *subscribePrefsWorker) Process(data *translateSubResponse) error {
	spw.targetGroups = spw.group(data.ntfAppInfoTrgt)
	spw.childGroups = spw.group(data.ntfAppInfoTrgtChlds)

	// App sent multiple subpaths in the target info.
	if len(spw.targetGroups) != 1 {
		p, err := path.New(spw.reqPath.Path)
		if err == nil {
			_, err = spw.processPath(p, nil)
		}
		if err != nil || !spw.req.Recurse {
			return err
		}
	}

	for _, g := range spw.targetGroups {
		if err := spw.processGroup(g); err != nil {
			return err
		}
	}
	if !spw.req.Recurse {
		return nil
	}
	for _, g := range spw.childGroups {
		if err := spw.processGroup(g); err != nil {
			return err
		}
	}
	return nil
}

func (spw *subscribePrefsWorker) processGroup(g *notificationAppInfoGroup) error {
	p := g.Path()
	if skip, err := spw.processPath(p, g); err != nil {
		return err
	} else if skip || !spw.req.Recurse {
		return nil
	}

	// process child paths
	ddup := make(map[string]struct{})

	for _, nInfo := range *g {
		for _, dbMap := range nInfo.dbFldYgPathInfoList {
			for _, leaf := range spw.getSubpaths(dbMap) {
				if _, ok := ddup[leaf]; ok {
					continue // already processed
				}
				leafPath := path.Copy(p)
				err := path.AppendPathStr(leafPath, leaf)
				if err != nil {
					return err
				}
				ddup[leaf] = struct{}{}
				if _, err := spw.processPath(leafPath, g); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (spw *subscribePrefsWorker) processPath(p *gnmi.Path, g *notificationAppInfoGroup) (skip bool, err error) {
	modeFilter := spw.reqPath.Mode
	var pType NotificationType
	var hasOnChange bool

	resp := newIsSubscribeResponse(spw.reqPath.ID, path.String(p))
	if g != nil {
		collectNotificationPreferences(*g, resp)
		pType = resp.PreferredType
	}

	// Skip processing p and its subpaths if p does not support on_change
	// and requester wants on_change enabled only.
	if !resp.IsOnChangeSupported && modeFilter == OnChange {
		skip = true
		return
	}

	// Collect preferences of other target and child groups, whose path is a subpath
	// of p. This step can overwrite 'PreferredType', we'll restore it later.
	for _, tg := range spw.targetGroups {
		if tg != g && path.Matches(tg.Path(), p) {
			collectNotificationPreferences(*tg, resp)
			hasOnChange = hasOnChange || (resp.PreferredType == OnChange)
		}
	}
	for _, cg := range spw.childGroups {
		if cg != g && path.Matches(cg.Path(), p) {
			collectNotificationPreferences(*cg, resp)
		}
	}

	// Mode filter fails; ignore p.. Skip subpaths also if all of them support
	// on_change but requester wants on_change disabled only.
	if modeFilter != 0 && (modeFilter == OnChange) != resp.IsOnChangeSupported {
		skip = (modeFilter == Sample)
		return
	}

	// Restore preferred type of the group
	if pType != 0 {
		resp.PreferredType = pType
	} else if hasOnChange {
		resp.PreferredType = OnChange
	}

	err = spw.Send(resp)
	return
}

func (spw *subscribePrefsWorker) getSubpaths(fldMap *dbFldYgPathInfo) []string {
	var subpaths []string
	var prefix string
	for _, p := range transformer.SplitPath(fldMap.rltvPath) {
		if len(p) != 0 {
			prefix += ("/" + p)
			subpaths = append(subpaths, prefix)
		}
	}
	for _, leaf := range fldMap.dbFldYgPathMap {
		if len(leaf) != 0 && leaf[0] == '{' {
			leaf = strings.TrimSuffix(leaf[1:], "}")
		}
		for _, p := range strings.Split(leaf, ",") {
			p = strings.TrimSpace(p)
			if len(p) != 0 {
				subpaths = append(subpaths, prefix+"/"+p)
			}
		}
	}
	return subpaths
}

// group notificationAppInfo array based on path
func (spw *subscribePrefsWorker) group(nInfos []*notificationAppInfo) []*notificationAppInfoGroup {
	var groups []*notificationAppInfoGroup
outer:
	for _, nInfo := range nInfos {
		for _, g := range groups {
			if path.Equals(g.Path(), nInfo.path) {
				*g = append(*g, nInfo)
				continue outer
			}
		}
		g := &notificationAppInfoGroup{nInfo}
		groups = append(groups, g)
	}
	return groups
}

func (spw *subscribePrefsWorker) Send(resp *IsSubscribeResponse) error {
	select {
	case <-spw.req.Context.Done():
		return spw.req.Context.Err()
	case spw.req.Channel <- resp:
		return nil // success
	default:
		return tlerr.New("Too many entries")
	}
}
