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

//go:build subscribe_annot
// +build subscribe_annot

package translib

import (
	"testing"

	"github.com/Azure/sonic-mgmt-common/translib/db"
)

// tamSwitchConfigNInfo returns mappings for /tam/switch/config
func tamSwitchConfigNInfo(subpath string) *notificationAppInfo {
	nInfo := &notificationAppInfo{
		dbno:                db.ConfigDB,
		table:               &db.TableSpec{Name: "TAM_SWITCH_TABLE"},
		key:                 db.NewKey("global"),
		isOnChangeSupported: true,
		mInterval:           0,
		pType:               OnChange,
	}
	switch subpath {
	case "*":
		nInfo.setFields(`{"":{"switch-id": "switch-id", "enterprise-id": "enterprise-id"}}`)
	default:
		nInfo.setFields(`{"":{"` + subpath + `": ""}}`)
	}
	return nInfo
}

// tamSamplerNInfo returns mappings for /tam/samplers/sampler/config
func tamSamplerConfigNInfo(subpath string) *notificationAppInfo {
	nInfo := &notificationAppInfo{
		dbno:                db.ConfigDB,
		table:               &db.TableSpec{Name: "TAM_SAMPLINGRATE_TABLE"},
		key:                 db.NewKey("*"),
		isOnChangeSupported: true,
		mInterval:           0,
		pType:               OnChange,
	}
	switch subpath {
	case "*":
		nInfo.setFields(`{"":{"name":"name","sampling-rate":"sampling-rate"}}`)
	default:
		nInfo.setFields(`{"":{"` + subpath + `": ""}}`)
	}
	return nInfo
}

// tamFlowgroupsConfigNInfo returns mappings for /tam/flowgroups/config
func tamFlowgroupsConfigNInfo(subpath string) *notificationAppInfo {
	nInfo := &notificationAppInfo{
		dbno:                db.ConfigDB,
		table:               &db.TableSpec{Name: "ACL_RULE"},
		key:                 db.NewKey("TAM", "*"),
		isOnChangeSupported: true,
		mInterval:           0,
		pType:               OnChange,
	}
	switch subpath {
	case "*":
		nInfo.setFields(`{"":{"IN_PORTS":"interfaces","PRIORITY":"priority","name":"name"}}`)
	default:
		nInfo.setFields(`{"":{"` + subpath + `": ""}}`)
	}
	return nInfo
}

// tamFlowgroupsConfigIdNInfo returns mappings for /tam/flowgroups/config/id
func tamFlowgroupsConfigIdNInfo(subpath string) *notificationAppInfo {
	nInfo := &notificationAppInfo{
		dbno:                db.ConfigDB,
		table:               &db.TableSpec{Name: "TAM_FLOWGROUP_TABLE"},
		key:                 db.NewKey("*"),
		isOnChangeSupported: true,
		mInterval:           0,
		pType:               OnChange,
	}
	nInfo.setFields(`{"":{"id":""}}`)
	return nInfo
}

// tamFlowgroupsNInfo returns mappings for /tam/flowgroups
func tamFlowgroupsNInfo(subpath string) *notificationAppInfo {
	nInfo := &notificationAppInfo{
		dbno:                db.ConfigDB,
		table:               &db.TableSpec{Name: "ACL_RULE"},
		key:                 db.NewKey("TAM", "*"),
		isOnChangeSupported: true,
		mInterval:           0,
		pType:               OnChange,
	}
	nInfo.setFields(`{"":{"name":"name"},"/config":{"IN_PORTS":"interfaces","PRIORITY":"priority","name":"name"},"/ipv4/config":
			{"DSCP":"dscp","DST_IP":"destination-address","IP_PROTOCOL":"protocol","SRC_IP":"source-address","hop-limit":"hop-limit","dscp-set":"dscp-set"},
			"/ipv4/state":{"DSCP":"dscp","DST_IP":"destination-address","IP_PROTOCOL":"protocol","SRC_IP":"source-address","hop-limit":"hop-limit","dscp-set":"dscp-set"},
			"/ipv6/config":{"DSCP":"dscp","DST_IPV6":"destination-address","IP_PROTOCOL":"protocol","SRC_IPV6":"source-address",
			"destination-flow-label":"destination-flow-label","hop-limit":"hop-limit","source-flow-label":"source-flow-label","dscp-set":"dscp-set"},
			"/ipv6/state":{"DSCP":"dscp","DST_IPV6":"destination-address","IP_PROTOCOL":"protocol","SRC_IPV6":"source-address",
			"destination-flow-label":"destination-flow-label","hop-limit":"hop-limit","source-flow-label":"source-flow-label","dscp-set":"dscp-set"},"/l2/config":
			{"DST_MAC":"destination-mac","ETHER_TYPE":"ethertype","SRC_MAC":"source-mac","VLAN":"vlan","destination-mac-mask":"destination-mac-mask","source-mac-mask":"source-mac-mask"},
			"/l2/state":{"DST_MAC":"destination-mac","ETHER_TYPE":"ethertype","SRC_MAC":"source-mac","VLAN":"vlan","destination-mac-mask":"destination-mac-mask",
			"source-mac-mask":"source-mac-mask"},"/transport/config":{"L4_DST_PORT":"destination-port","L4_SRC_PORT":"source-port","TCP_FLAGS":"tcp-flags"},
			"/transport/state":{"L4_DST_PORT":"destination-port","L4_SRC_PORT":"source-port","TCP_FLAGS":"tcp-flags"},
			"/state":{"IN_PORTS":"interfaces","PRIORITY":"priority","name":"name"}}`)
	return nInfo
}

// tamSwitchStateNInfo returns mappings for /tam/switch/state
func tamSwitchStateNInfo(subpath string) *notificationAppInfo {
	nInfo := &notificationAppInfo{
		dbno:                db.ApplDB,
		table:               &db.TableSpec{Name: "TAM_APPL_SWITCH_TABLE"},
		key:                 db.NewKey("global"),
		isOnChangeSupported: false,
		mInterval:           20,
		pType:               Sample,
	}
	switch subpath {
	case "*":
		nInfo.setFields(`{"":{"switch-id": "switch-id", "enterprise-id": "enterprise-id"}}`)
	default:
		nInfo.setFields(`{"":{"` + subpath + `": ""}}`)
	}
	return nInfo
}

// tamFlowgroupStateStatsNInfo returns mappings for /tam/flowgroups/flowgroup/state/statistics
func tamFlowgroupStateStatsNInfo(subpath string) *notificationAppInfo {
	nInfo := &notificationAppInfo{
		dbno:                db.CountersDB,
		table:               &db.TableSpec{Name: "COUNTERS"},
		key:                 db.NewKey("TAM", "*"),
		isOnChangeSupported: false,
		mInterval:           20,
		pType:               Sample,
	}
	switch subpath {
	case "*":
		nInfo.setFields(`{"":{"Bytes":"bytes","Packets":"packets"}}`)
	default:
		nInfo.setFields(`{"":{"` + subpath + `": ""}}`)
	}
	return nInfo
}

// tamFeatureNInfo returns mappings for /tam/features/feature - for both config and state
func tamFeatureNInfo(subpath string) *notificationAppInfo {
	nInfo := &notificationAppInfo{
		dbno:                db.ConfigDB,
		table:               &db.TableSpec{Name: "TAM_FEATURES_TABLE"},
		key:                 db.NewKey("*"),
		isOnChangeSupported: false,
		mInterval:           20,
		pType:               Sample,
	}

	switch subpath {
	case "*":
		nInfo.setFields(`{"/config":{"feature-ref":"feature-ref","status":"status"},"/state":{"feature-ref":"feature-ref","status":"status"}}`)
	default:
		nInfo.setFields(`{"":{"` + subpath + `": ""}}`)
	}
	return nInfo
}

// tamFeatureCfgNInfo returns mappings for /tam/features/feature/config
func tamFeatureCfgNInfo(subpath string) *notificationAppInfo {
	nInfo := &notificationAppInfo{
		dbno:                db.ConfigDB,
		table:               &db.TableSpec{Name: "TAM_FEATURES_TABLE"},
		key:                 db.NewKey("*"),
		isOnChangeSupported: false,
		mInterval:           20,
		pType:               Sample,
	}

	switch subpath {
	case "*":
		nInfo.setFields(`{"":{"feature-ref":"feature-ref","status":"status"}}`)
	default:
		nInfo.setFields(`{"":{"` + subpath + `": ""}}`)
	}
	return nInfo
}

// tamFeatureNInfo returns mappings for /tam/features/feature/state
func tamFeatureStateNInfo(subpath string) *notificationAppInfo {
	nInfo := &notificationAppInfo{
		dbno:                db.ConfigDB,
		table:               &db.TableSpec{Name: "TAM_FEATURES_TABLE"},
		key:                 db.NewKey("*"),
		isOnChangeSupported: false,
		mInterval:           20,
		pType:               Sample,
	}

	switch subpath {
	case "*":
		nInfo.setFields(`{"":{"feature-ref":"feature-ref","status":"status"}}`)
	default:
		nInfo.setFields(`{"":{"` + subpath + `": ""}}`)
	}
	return nInfo
}

// tamDropMonitorNInfo returns mappings /tam/dropmonitor/global
func tamDropMonitorGlobalNInfo(subpath string) *notificationAppInfo {
	nInfo := &notificationAppInfo{
		dbno:                db.ConfigDB,
		table:               &db.TableSpec{Name: "TAM_DROPMONITOR_TABLE"},
		key:                 db.NewKey("global"),
		isOnChangeSupported: false,
		mInterval:           20,
		pType:               Sample,
	}
	switch subpath {
	case "*":
		nInfo.setFields(`{"/config":{"aging-interval":"aging-interval"},"/state":{"aging-interval":"aging-interval"}}`)
	default:
		nInfo.setFields(`{"":{"` + subpath + `": ""}}`)
	}
	return nInfo
}

// InterfaceStateNInfo returns mappings for /interfaces/interface/state/oper-status
func InterfaceEthernetStateOperStatusNInfo(ifName string) *notificationAppInfo {
	nInfo := &notificationAppInfo{
		dbno:                db.ApplDB,
		table:               &db.TableSpec{Name: "PORT_TABLE"},
		key:                 db.NewKey(ifName),
		isOnChangeSupported: true,
		deleteAction:        1,
		mInterval:           0,
		pType:               OnChange,
	}
	nInfo.setFields(`{"":{"oper_status":"", "under_test":""}}`)
	return nInfo
}

// ON_CHANGE for /tam/switch

func TestSubscribeOnChange_tam_switch(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/switch", OnChange)
	tv.VerifyCount(translErr, 0)
}

func TestSubscribeOnChange_tam_switch_config(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/switch/config", OnChange)
	tv.VerifyCount(1, 0)
	tv.VerifyTarget("/openconfig-tam:tam/switch/config", tamSwitchConfigNInfo("*"))
}

func TestSubscribeOnChange_tam_switch_config_id(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/switch/config/switch-id", OnChange)
	tv.VerifyCount(1, 0)
	tv.VerifyTarget("/openconfig-tam:tam/switch/config/switch-id", tamSwitchConfigNInfo("switch-id"))
}

func TestSubscribeOnChange_tam_switch_state(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/switch/state", OnChange)
	tv.VerifyCount(translErr, 0)
}

func TestSubscribeOnChange_tam_switch_state_id(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/switch/state/switch-id", OnChange)
	tv.VerifyCount(translErr, 0)
}

// SAMPLE for /tam/switch

func TestSubscribeSample_tam_switch(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/switch", Sample)
	tv.VerifyCount(2, 0)
	tamCfgNInfo := tamSwitchConfigNInfo("*")
	tamCfgNInfo.pType = Sample
	tamCfgNInfo.isOnChangeSupported = false
	tamCfgNInfo.mInterval = 25
	tv.VerifyTarget("/openconfig-tam:tam/switch/config", tamCfgNInfo)
	tamStateNInfo := tamSwitchStateNInfo("*")
	tamStateNInfo.mInterval = 20
	tv.VerifyTarget("/openconfig-tam:tam/switch/state", tamStateNInfo)
}

func TestSubscribeSample_tam_switch_config(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/switch/config", Sample)
	tv.VerifyCount(1, 0)
	tamCfgNInfo := tamSwitchConfigNInfo("*")
	tamCfgNInfo.pType = Sample
	tamCfgNInfo.isOnChangeSupported = false
	tamCfgNInfo.mInterval = 25 // got overriden by annot.subscribe
	tv.VerifyTarget("/openconfig-tam:tam/switch/config", tamCfgNInfo)
}

func TestSubscribeSample_tam_sampler_config(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/samplers/sampler/config", Sample)
	tv.VerifyCount(1, 0)
	tamSamplerCfgNInfo := tamSamplerConfigNInfo("*")
	tamSamplerCfgNInfo.pType = Sample
	tamSamplerCfgNInfo.isOnChangeSupported = false
	tamSamplerCfgNInfo.mInterval = 27 // got overriden by annot.subscribe
	tv.VerifyTarget("/openconfig-tam:tam/samplers/sampler[name=*]/config", tamSamplerCfgNInfo)
}

func TestSubscribeSample_tam_switch_config_leaf(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/switch/config/enterprise-id", Sample)
	tv.VerifyCount(1, 0)
	tamCfgNInfo := tamSwitchConfigNInfo("enterprise-id")
	tamCfgNInfo.pType = Sample
	tamCfgNInfo.isOnChangeSupported = false
	tamCfgNInfo.mInterval = 25
	tv.VerifyTarget("/openconfig-tam:tam/switch/config/enterprise-id", tamCfgNInfo)
}

func TestSubscribeSample_tam_xxswitch_state(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/switch/state", Sample)
	tv.VerifyCount(1, 0)
	tv.VerifyTarget("/openconfig-tam:tam/switch/state", tamSwitchStateNInfo("*"))
}

func TestSubscribeSample_tam_flowgroups(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/flowgroups", Sample)
	tv.VerifyCount(1, 3)
	tamFlowGrpNInfo := tamFlowgroupsNInfo("*")
	tamFlowGrpNInfo.pType = Sample
	tamFlowGrpNInfo.isOnChangeSupported = false
	tamFlowGrpNInfo.mInterval = 22
	tv.VerifyTarget("/openconfig-tam:tam/flowgroups/flowgroup[name=*]", tamFlowGrpNInfo)
}

func TestSubscribeSample_tam_flowgroups_config(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/flowgroups/flowgroup/config", Sample)
	tv.VerifyCount(1, 1)
	tamFlowGrpCfgNInfo := tamFlowgroupsConfigNInfo("*")
	tamFlowGrpCfgNInfo.pType = Sample
	tamFlowGrpCfgNInfo.isOnChangeSupported = false
	tamFlowGrpCfgNInfo.mInterval = 22 // since min interval "22" configured at it parent node flowgroup
	tv.VerifyTarget("/openconfig-tam:tam/flowgroups/flowgroup[name=*]/config", tamFlowGrpCfgNInfo)
	tamFlowGrpCfgIdNInfo := tamFlowgroupsConfigIdNInfo("*")
	tamFlowGrpCfgIdNInfo.pType = Sample
	tamFlowGrpCfgIdNInfo.isOnChangeSupported = false
	tamFlowGrpCfgIdNInfo.mInterval = 22 // since min interval "22" configured at it parent node flowgroup
	tv.VerifyChild("/openconfig-tam:tam/flowgroups/flowgroup[name=*]/config/id", tamFlowGrpCfgIdNInfo)
}

// non subtree
func TestSubscribeSample_tam_feature(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/features/feature", Sample)
	tv.VerifyCount(1, 0)
	tamFeatureInfo := tamFeatureNInfo("*")
	tamFeatureInfo.pType = Sample
	tamFeatureInfo.isOnChangeSupported = false
	tamFeatureInfo.mInterval = 21
	tv.VerifyTarget("/openconfig-tam:tam/features/feature[feature-ref=*]", tamFeatureInfo)
}

// non subtree config
func TestSubscribeSample_tam_feature_config(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/features/feature/config", Sample)
	tv.VerifyCount(1, 0)
	tamFeatureCfgInfo := tamFeatureCfgNInfo("*")
	tamFeatureCfgInfo.pType = Sample
	tamFeatureCfgInfo.isOnChangeSupported = false
	tamFeatureCfgInfo.mInterval = 20
	tv.VerifyTarget("/openconfig-tam:tam/features/feature[feature-ref=*]/config", tamFeatureCfgInfo)
}

// non subtree state
func TestSubscribeSample_tam_feature_state(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/features/feature/state", Sample)
	tv.VerifyCount(1, 0)
	tamFeatureStateInfo := tamFeatureStateNInfo("*")
	tamFeatureStateInfo.pType = Sample
	tamFeatureStateInfo.isOnChangeSupported = false
	tamFeatureStateInfo.mInterval = 21
	tv.VerifyTarget("/openconfig-tam:tam/features/feature[feature-ref=*]/state", tamFeatureStateInfo)
}

// non subtree with field transformer case
func TestSubscribeSample_tam_feature_state_leaf(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/features/feature/state/feature-ref", Sample)
	tv.VerifyCount(1, 0)
	tamFeatureStateInfo := tamFeatureStateNInfo("feature-ref")
	tamFeatureStateInfo.pType = Sample
	tamFeatureStateInfo.isOnChangeSupported = false
	tamFeatureStateInfo.mInterval = 21
	tv.VerifyTarget("/openconfig-tam:tam/features/feature[feature-ref=*]/state/feature-ref", tamFeatureStateInfo)
}

// non subtree state
func TestSubscribeSample_tam_dropmonitor_global(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/dropmonitor/global", Sample)
	tv.VerifyCount(1, 0)
	tmGlobalInfo := tamDropMonitorGlobalNInfo("*")
	tmGlobalInfo.pType = Sample
	tmGlobalInfo.isOnChangeSupported = false
	tmGlobalInfo.mInterval = 20
	tv.VerifyTarget("/openconfig-tam:tam/dropmonitor/global", tmGlobalInfo)
}

// TARGET_DEFINED for /tam/switch

func TestSubscribeTrgtDef_tam_switch(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/switch", TargetDefined)
	tv.VerifyCount(2, 0)
	tv.VerifyTarget("/openconfig-tam:tam/switch/config", tamSwitchConfigNInfo("*"))
	tv.VerifyTarget("/openconfig-tam:tam/switch/state", tamSwitchStateNInfo("*"))
}

func TestSubscribeTrgtDef_tam_switch_config(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/switch/config", TargetDefined)
	tv.VerifyCount(1, 0)
	tv.VerifyTarget("/openconfig-tam:tam/switch/config", tamSwitchConfigNInfo("*"))
}

func TestSubscribeTrgtDef_tam_switch_state(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/switch/state", TargetDefined)
	tv.VerifyCount(1, 0)
	tv.VerifyTarget("/openconfig-tam:tam/switch/state", tamSwitchStateNInfo("*"))
}

func TestSubscribeTrgtDef_tam_flowgroup(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/flowgroups/flowgroup", TargetDefined)
	tv.VerifyCount(1, 3)
	tamFlowgrpStats := tamFlowgroupStateStatsNInfo("*")
	tamFlowgrpStats.mInterval = 22
	tv.VerifyChild("/openconfig-tam:tam/flowgroups/flowgroup[name=*]/state/statistics", tamFlowgrpStats)
	tamFlowgrpId := tamFlowgroupsConfigIdNInfo("*")
	tv.VerifyChild("/openconfig-tam:tam/flowgroups/flowgroup[name=*]/config/id", tamFlowgrpId)
	tv.VerifyChild("/openconfig-tam:tam/flowgroups/flowgroup[name=*]/state/id", tamFlowgrpId)
}

func TestSubscribeTrgtDef_tam_flowgroup_state_statistics(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/flowgroups/flowgroup/state/statistics", TargetDefined)
	tv.VerifyCount(1, 0)
	tamFlowgrpStats := tamFlowgroupStateStatsNInfo("*")
	tamFlowgrpStats.mInterval = 22
	tv.VerifyTarget("/openconfig-tam:tam/flowgroups/flowgroup[name=*]/state/statistics", tamFlowgrpStats)
}

// TestSubscribeTrgtDef_tam_flowgroup_test - before running this test - make /state as "on change disable"
// and /state/statistics as "on change enable" for the flowgroup path in the xmfr_tam.go
//func TestSubscribeTrgtDef_tam_flowgroup_test(t *testing.T) {
//	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/flowgroups/flowgroup", TargetDefined)
//	tv.VerifyCount(1, 4)
//	tamFlowgrpStats := tamFlowgroupStateStatsNInfo("*")
//	tamFlowgrpStats.mInterval = 22
//	tamFlowgrpStats.pType = Sample
//	tv.VerifyChild("/openconfig-tam:tam/flowgroups/flowgroup[name=*]/state/statistics", tamFlowgrpStats)
//	tamFlowgrpId := tamFlowgroupsConfigIdNInfo("*")
//	tv.VerifyChild("/openconfig-tam:tam/flowgroups/flowgroup[name=*]/config/id", tamFlowgrpId)
//	tamFlowgrpId.mInterval = 22
//	tamFlowgrpId.pType = Sample
//	tamFlowgrpId.isOnChangeSupported = false
//	tv.VerifyChild("/openconfig-tam:tam/flowgroups/flowgroup[name=*]/state/id", tamFlowgrpId)
//}

// TestSubscribeTrgtDef_tam_flowgroup_state_test - before running this test - make /state as "on change disable"
// and /state/statistics as "on change enable" for the flowgroup path in the xmfr_tam.go
//func TestSubscribeTrgtDef_tam_flowgroup_state_test(t *testing.T) {
//	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/flowgroups/flowgroup/state", TargetDefined)
//	tv.VerifyCount(1, 2)
//	tamFlowgrpStats := tamFlowgroupStateStatsNInfo("*")
//	tamFlowgrpStats.mInterval = 22
//	tamFlowgrpStats.pType = Sample
//	tv.VerifyChild("/openconfig-tam:tam/flowgroups/flowgroup[name=*]/state/statistics", tamFlowgrpStats)
//	tamFlowgrpId := tamFlowgroupsConfigIdNInfo("*")
//	tamFlowgrpId.mInterval = 22
//	tamFlowgrpId.pType = Sample
//	tamFlowgrpId.isOnChangeSupported = false
//	tv.VerifyChild("/openconfig-tam:tam/flowgroups/flowgroup[name=*]/state/id", tamFlowgrpId)
//}

func TestSubscribeTrgtDef_tam_switch_config_leaf(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/switch/config/enterprise-id", TargetDefined)
	tv.VerifyCount(1, 0)
	tamCfgNInfo := tamSwitchConfigNInfo("enterprise-id")
	tv.VerifyTarget("/openconfig-tam:tam/switch/config/enterprise-id", tamCfgNInfo)
}

func TestSubscribeTrgtDef_tam_flowgroups_config(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/flowgroups/flowgroup/config", TargetDefined)
	tv.VerifyCount(1, 1)
	tamFlowGrpCfgNInfo := tamFlowgroupsConfigNInfo("*")
	tv.VerifyTarget("/openconfig-tam:tam/flowgroups/flowgroup[name=*]/config", tamFlowGrpCfgNInfo)
	tamFlowGrpCfgIdNInfo := tamFlowgroupsConfigIdNInfo("*")
	tv.VerifyChild("/openconfig-tam:tam/flowgroups/flowgroup[name=*]/config/id", tamFlowGrpCfgIdNInfo)
}

// non subtree
func TestSubscribeTrgtDef_tam_feature(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/features/feature", TargetDefined)
	tv.VerifyCount(1, 0)
	tamFeature := tamFeatureNInfo("*")
	tamFeature.pType = OnChange
	tamFeature.mInterval = 0
	tamFeature.isOnChangeSupported = true
	tv.VerifyTarget("/openconfig-tam:tam/features/feature[feature-ref=*]", tamFeature)
}

// non subtree config
func TestSubscribeTrgtDef_tam_feature_config(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/features/feature/config", TargetDefined)
	tv.VerifyCount(1, 0)
	tamFeatureCfgInfo := tamFeatureCfgNInfo("*")
	tamFeatureCfgInfo.pType = OnChange
	tamFeatureCfgInfo.mInterval = 0
	tamFeatureCfgInfo.isOnChangeSupported = true
	tv.VerifyTarget("/openconfig-tam:tam/features/feature[feature-ref=*]/config", tamFeatureCfgInfo)
}

// non subtree state
func TestSubscribeTrgtDef_tam_feature_state(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/features/feature/state", TargetDefined)
	tv.VerifyCount(1, 0)
	tamFeatureStateInfo := tamFeatureStateNInfo("*")
	tamFeatureStateInfo.pType = OnChange
	tamFeatureStateInfo.isOnChangeSupported = true
	tamFeatureStateInfo.mInterval = 0
	tv.VerifyTarget("/openconfig-tam:tam/features/feature[feature-ref=*]/state", tamFeatureStateInfo)
}

// non subtree with field transformer case
func TestSubscribeTrgtDef_tam_feature_state_leaf(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/features/feature/state/feature-ref", TargetDefined)
	tv.VerifyCount(1, 0)
	tamFeatureStateInfo := tamFeatureStateNInfo("feature-ref")
	tamFeatureStateInfo.pType = OnChange
	tamFeatureStateInfo.isOnChangeSupported = true
	tamFeatureStateInfo.mInterval = 0
	tv.VerifyTarget("/openconfig-tam:tam/features/feature[feature-ref=*]/state/feature-ref", tamFeatureStateInfo)
}

// non subtree state
func TestSubscribeTrgtDef_tam_dropmonitor_global(t *testing.T) {
	tv := testTranslateSubscribe(t, "/openconfig-tam:tam/dropmonitor/global", TargetDefined)
	tv.VerifyCount(1, 0)
	tmGlobalInfo := tamDropMonitorGlobalNInfo("*")
	tmGlobalInfo.pType = OnChange
	tmGlobalInfo.isOnChangeSupported = true
	tmGlobalInfo.mInterval = 0
	tv.VerifyTarget("/openconfig-tam:tam/dropmonitor/global", tmGlobalInfo)
}

// Target defined - composite Db field mappings on the leaf node with delete_as_update option, and the
// target node of the subscribe path is all the interfaces
func TestSubscribeTrgtDefDelAsUpdateLeafWithCompositeDbFieldsCount(t *testing.T) {
	path := "/openconfig-interfaces:interfaces/interface[name=*]/state"
	tv := testTranslateSubscribe(t, path, TargetDefined)
	tv.VerifyCount(5, 6)
}

// Target defined - composite Db field mappings on the leaf node with delete_as_update option; and the
// target node of the subscribe path is Ethernet interface, verifies the child "oper-status"
func TestSubscribeTrgtDefDelAsUpdateLeafWithCompositeDbFieldsMapping(t *testing.T) {
	path := "/openconfig-interfaces:interfaces/interface[name=Ethernet10]/state"
	ethState := InterfaceEthernetStateOperStatusNInfo("Ethernet10")
	tv := testTranslateSubscribe(t, path, TargetDefined)
	tv.VerifyCount(1, 2)
	tv.VerifyChild("/openconfig-interfaces:interfaces/interface[name=Ethernet10]/state/oper-status", ethState)
}

// Sample - composite Db field mappings on the leaf node with delete_as_update option; and the
// target node of the subscribe path is all the interfaces
func TestSubscribeSampleDelAsUpdateLeafWithCompositeDbFieldsCount(t *testing.T) {
	path := "/openconfig-interfaces:interfaces/interface[name=*]/state"
	tv := testTranslateSubscribe(t, path, Sample)
	tv.VerifyCount(5, 6)
}

// Sample - composite Db field mappings on the leaf node with delete_as_update option; and the
// target node of the subscribe path is Ethernet interface, verifies the child "oper-status"
func TestSubscribeSampleDelAsUpdateLeafWithCompositeDbFieldsMapping(t *testing.T) {
	path := "/openconfig-interfaces:interfaces/interface[name=Ethernet10]/state"
	ethState := InterfaceEthernetStateOperStatusNInfo("Ethernet10")
	tv := testTranslateSubscribe(t, path, Sample)
	tv.VerifyCount(1, 2)
	ethState.mInterval = 20
	tv.VerifyChild("/openconfig-interfaces:interfaces/interface[name=Ethernet10]/state/oper-status", ethState)
}

// Onchange - omposite Db field mappings on the leaf node with delete_as_update option; and the
// target node of the subscribe path is oper-status
func TestSubscribeOnchangeDelAsUpdateLeafWithCompositeDbFieldsMapping(t *testing.T) {
	path := "/openconfig-interfaces:interfaces/interface[name=Ethernet10]/state/oper-status"
	ethStateOperStatus := InterfaceEthernetStateOperStatusNInfo("Ethernet10")
	tv := testTranslateSubscribe(t, path, OnChange)
	tv.VerifyCount(1, 0)
	tv.VerifyTarget("/openconfig-interfaces:interfaces/interface[name=Ethernet10]/state/oper-status", ethStateOperStatus)
}

// Target Defined - To test the target node of the path with the OnchangeDisable
func TestSubscribeTrgtDefInterfaceSatteCounters(t *testing.T) {
	path := "/openconfig-interfaces:interfaces/interface[name=*]/state/counters"
	tv := testTranslateSubscribe(t, path, TargetDefined)
	tv.VerifyCount(1, 0)
}
