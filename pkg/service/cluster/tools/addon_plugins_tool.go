/*
----------------------------------------
@Create 2022-08-18
@Author 冷朴承<lengpucheng@qq.com>
@Program KubeOperator
@Describe addon_plugins_tool
----------------------------------------
@Version 1.0 2022/8/18-12:55
@Memo create this file
*/

package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KubeOperator/KubeOperator/pkg/model"
	"strings"
)

func NewAddonPluginsTool(cluster *Cluster, tool *model.ClusterTool) (Interface, error) {
	// 取出来vars
	defVars := map[string]interface{}{}
	if err := json.Unmarshal([]byte(tool.Vars), &defVars); err != nil {
		return nil, err
	}
	addonType, is := defVars[model.AddonPluginType]
	if !is {
		return nil, errors.New("The addon plugin type is not found ,please check plugin json/yaml file is legal ")
	}
	addonResource, is := defVars[model.AddonPluginResource]
	if !is {
		return nil, errors.New("The addon plugin resource is not found ,please check plugin json/yaml file is legal ")
	}
	switch strings.ToLower(fmt.Sprintf("%v", addonType)) {
	case "helm":
		return NewAddonPluginHelm(cluster, tool, fmt.Sprintf("%v", addonResource))
	case "ansible":
		return NewAddonPluginAnsible(cluster, tool, fmt.Sprintf("%v", addonResource))
	}

	return nil, nil
}
