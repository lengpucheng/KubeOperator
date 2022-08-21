/*
----------------------------------------
@Create 2022-08-18
@Author 冷朴承<lengpucheng@qq.com>
@Program KubeOperator
@Describe addon_plugins_ansible
----------------------------------------
@Version 1.0 2022/8/18-12:54
@Memo create this file
*/

package tools

import (
	"encoding/json"
	"fmt"
	"github.com/KubeOperator/KubeOperator/pkg/constant"
	"github.com/KubeOperator/KubeOperator/pkg/model"
	"github.com/KubeOperator/KubeOperator/pkg/service/cluster/adm/phases"
	"github.com/KubeOperator/KubeOperator/pkg/util/ansible"
	"github.com/KubeOperator/KubeOperator/pkg/util/kobe"
	"io"
	"strings"
	"unsafe"
)

type AddonPluginAnsible struct {
	Tool                *model.ClusterTool
	Cluster             *Cluster
	LocalHostName       string
	LocalRepositoryPort int
	playbook            string
	vars                map[string]interface{}
	writerPoint         int
}

func NewAddonPluginAnsible(cluster *Cluster, tool *model.ClusterTool, playbook string) (*AddonPluginAnsible, error) {
	return &AddonPluginAnsible{
		Cluster:             cluster,
		Tool:                tool,
		LocalHostName:       constant.LocalRepositoryDomainName,
		LocalRepositoryPort: cluster.helmRepoPort,
		playbook:            playbook,
	}, nil
}

func (a *AddonPluginAnsible) Install(toolDetail model.ClusterToolDetail) error {
	if err := a.setDefaultVars(toolDetail); err != nil {
		return err
	}
	adm := a.initKobe()
	adm.SetVar(model.AddonPluginAnsibleAction, "install")
	writer, err := a.initWrite()
	if err != nil {
		return err
	}
	return phases.RunPlaybookAndGetResult(adm, a.playbook, "", writer)
}

func (a *AddonPluginAnsible) Upgrade(toolDetail model.ClusterToolDetail) error {
	if err := a.setDefaultVars(toolDetail); err != nil {
		return err
	}
	adm := a.initKobe()
	adm.SetVar(model.AddonPluginAnsibleAction, "upgrade")
	writer, err := a.initWrite()
	if err != nil {
		return err
	}
	return phases.RunPlaybookAndGetResult(adm, a.playbook, "", writer)
}

func (a *AddonPluginAnsible) Uninstall() error {
	if err := a.setDefaultVars(); err != nil {
		return err
	}
	adm := a.initKobe()
	adm.SetVar(model.AddonPluginAnsibleAction, "uninstall")
	writer, err := a.initWrite()
	if err != nil {
		return err
	}
	return phases.RunPlaybookAndGetResult(adm, a.playbook, "", writer)
}

func (a *AddonPluginAnsible) setDefaultVars(toolDetails ...model.ClusterToolDetail) error {
	// 取出参数
	vars := map[string]interface{}{}
	if err := json.Unmarshal([]byte(a.Tool.Vars), &vars); err != nil {
		return err
	}
	if toolDetails != nil && len(toolDetails) > 0 {
		toolDetail := toolDetails[0]
		// 取出默认值
		defVar := map[string]interface{}{}
		if err := json.Unmarshal([]byte(toolDetail.Vars), &defVar); err != nil {
			return err
		}
		// 当该默认值不存在时设置默认值
		for k, v := range defVar {
			if _, p := vars[k]; !p {
				vars[k] = v
			}
		}
		marshal, err := json.Marshal(vars)
		if err != nil {
			return err
		}
		a.Tool.Vars = string(marshal)
	}
	// 插值替换
	a.Tool.Vars = strings.ReplaceAll(a.Tool.Vars, "${AddonRepo}", fmt.Sprintf("%s:%d", a.LocalHostName, a.LocalRepositoryPort))
	if err := json.Unmarshal([]byte(a.Tool.Vars), &a.vars); err != nil {
		a.vars = vars
	}
	return nil
}

func (a *AddonPluginAnsible) initKobe() *kobe.Kobe {
	adm := kobe.NewAnsible(&kobe.Config{
		Inventory: a.Cluster.ParseInventory(),
	})
	for k, v := range a.vars {
		adm.SetVar(k, fmt.Sprintf("%v", v))
	}
	return adm
}

func (a *AddonPluginAnsible) initWrite() (writer io.Writer, err error) {
	defer func() {
		if r := recover(); r != nil {
			writer, err = ansible.CreateAnsibleLogWriterWithId(a.Cluster.Name, a.Tool.Name)
		}
	}()

	// 初始化writer
	if a.writerPoint == 0 {
		if writer, err = ansible.CreateAnsibleLogWriterWithId(a.Cluster.Name, a.Tool.Name); err != nil {
			return
		}
	} else {
		writer = *(*io.Writer)(unsafe.Pointer(uintptr(a.writerPoint)))
	}

	return
}
