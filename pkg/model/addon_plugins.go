/*
----------------------------------------
@Create 2022-08-18
@Author 冷朴承<lengpucheng@qq.com>
@Program KubeOperator
@Describe addon_plugins
----------------------------------------
@Version 1.0 2022/8/18-21:07
@Memo create this file
*/

package model

import (
	"encoding/json"
	"github.com/KubeOperator/KubeOperator/pkg/constant"
	"github.com/KubeOperator/KubeOperator/pkg/db"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"reflect"
	"strings"
)

const (
	AddonPluginsFileJson = "/etc/ko/addon-plugins.json"
	AddonPluginsFileYaml = "/etc/ko/addon-plugins.yaml"

	AddonPluginSchema   = "addon.plugin.schema"
	AddonPluginResource = "addon.plugin.resource"
	AddonPluginType     = "addon.plugin.type"
	AddonPluginIsInit   = "addon.plugin.isinit"
	AddonPluginVersion  = "addon.plugin.version"

	AddonPluginAnsibleAction = "addon_plugin_action"

	AddonPluginHelmWorkloadName = "addon.plugin.helm.workload.name"
	AddonPluginHelmWorkloadType = "addon.plugin.helm.workload.type"
	AddonPluginHelmIngressPort  = "addon.plugin.helm.ingress.port"
	AddonPluginHelmServiceName  = "addon.plugin.helm.service.name"
	AddonPluginHelmMinReplicas  = "addon.plugin.helm.minreplicas"
)

type AddonPluginsManifest struct {
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion" yaml:"apiVersion"`
	Spec       struct {
		Globals map[string]interface{} `json:"globals"`
		Plugins []AddonPlugin          `json:"plugins"`
	} `json:"spec"`
}
type AddonPlugin struct {
	Name     string                 `json:"name"`
	Metadata Metadata               `json:"metadata"`
	Schema   map[string]interface{} `json:"schema"`
	Helm     AddonHelm              `json:"helm"`
	Vars     map[string]interface{} `json:"vars"`
}
type Metadata struct {
	Describe      string `json:"describe"`
	Logo          string `json:"logo"`
	Architecture  string `json:"architecture"`
	Version       string `json:"version"`
	IsInit        bool   `json:"isInit" yaml:"isInit"`
	AddonType     string `json:"addonType" yaml:"addonType"`
	AddonResource string `json:"addonResource" yaml:"addonResource"`
}
type AddonHelm struct {
	WorkloadName string `json:"workloadName" yaml:"workloadName"`
	WorkloadType string `json:"workloadType" yaml:"workloadType"`
	MinReplicas  int    `json:"minReplicas" yaml:"minReplicas"`
	ServiceName  string `json:"serviceName" yaml:"serviceName"`
	IngressPort  int    `json:"ingressPort" yaml:"ingressPort"`
}

func (addon *AddonPlugin) ToClusterToolDetail() ClusterToolDetail {
	return ClusterToolDetail{
		Name:         addon.Name,
		Version:      addon.Metadata.Version,
		ChartVersion: addon.Metadata.Version,
		Architecture: addon.Metadata.Architecture,
		Vars:         addon.toAddonParameter(addon.Vars),
	}
}

func (addon *AddonPlugin) ToClusterTool(clusterId string) ClusterTool {
	if addon.Metadata.Architecture == "" {
		addon.Metadata.Architecture = "all"
	}
	if addon.Metadata.Logo == "" {
		addon.Metadata.Logo = "kubeapps.png"
	}
	return ClusterTool{
		Name:         addon.Name,
		ClusterID:    clusterId,
		Version:      addon.Metadata.Version,
		Status:       constant.StatusWaiting,
		Logo:         addon.Metadata.Logo,
		Describe:     addon.Metadata.Describe,
		Architecture: addon.Metadata.Architecture,
		Vars:         addon.toAddonParameter(nil),
		Frame:        false,
		Url:          "",
	}
}

func (addon *AddonPlugin) toAddonParameter(vars map[string]interface{}) string {
	if vars == nil {
		vars = make(map[string]interface{})
	}
	vars[AddonPluginSchema] = addon.Schema
	vars[AddonPluginResource] = addon.Metadata.AddonResource
	vars[AddonPluginType] = addon.Metadata.AddonType
	vars[AddonPluginIsInit] = addon.Metadata.IsInit
	vars[AddonPluginVersion] = addon.Metadata.Version
	if strings.ToLower(addon.Metadata.AddonType) == "helm" {
		vars[AddonPluginHelmWorkloadType] = addon.Helm.WorkloadType
		vars[AddonPluginHelmWorkloadName] = addon.Helm.WorkloadName
		vars[AddonPluginHelmServiceName] = addon.Helm.ServiceName
		vars[AddonPluginHelmIngressPort] = addon.Helm.IngressPort
		if addon.Helm.MinReplicas < 1 {
			addon.Helm.MinReplicas = 1
		}
		vars[AddonPluginHelmMinReplicas] = addon.Helm.MinReplicas
	}
	data, _ := json.Marshal(vars)
	return string(data)
}

var addonPlugins AddonPluginsManifest

func GetAddonPlugins() AddonPluginsManifest {
	if reflect.DeepEqual(addonPlugins, AddonPluginsManifest{}) {
		_, _ = LoadAddonPlugins()
	}
	return addonPlugins
}

func LoadAddonPlugins() (AddonPluginsManifest, error) {
	var manifest AddonPluginsManifest
	var err error
	var data []byte

	if data, err = ioutil.ReadFile(AddonPluginsFileJson); err != nil {
		if data, err = ioutil.ReadFile(AddonPluginsFileYaml); err != nil {
			return manifest, err
		}
		err = yaml.Unmarshal(data, &manifest)
	} else {
		err = json.Unmarshal(data, &manifest)
	}

	if err == nil {
		// 更新cache
		addonPlugins = manifest
		// 写入ctd
		if manifest.Spec.Plugins != nil && db.DB != nil {
			tx := db.DB.Begin()
			var toolDetails []ClusterToolDetail
			if err := tx.Find(&toolDetails).Error; err != nil {
				tx.Rollback()
				return manifest, err
			}
			for _, plugin := range manifest.Spec.Plugins {
				detail := plugin.ToClusterToolDetail()
				for i, td := range toolDetails {
					if td.Name == detail.Name {
						detail.ID = td.ID
						if err := tx.Model(&td).Update(detail).Error; err != nil {
							tx.Rollback()
							return manifest, err
						}
						break
					} else if i == len(toolDetails)-1 {
						if err := tx.Create(&detail).Error; err != nil {
							tx.Rollback()
							return manifest, err
						}
					}
				}
			}
			tx.Commit()

		}
	}

	return manifest, err
}

func init() {
	_, _ = LoadAddonPlugins()
}
