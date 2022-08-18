/*
----------------------------------------
@Create 2022-08-17
@Author 冷朴承<lengpucheng@qq.com>
@Program KubeOperator
@Describe addon_plugins_helm
----------------------------------------
@Version 1.0 2022/8/17-22:18
@Memo create this file
*/

package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KubeOperator/KubeOperator/pkg/constant"
	"github.com/KubeOperator/KubeOperator/pkg/model"
	"strings"
)

type AddonPluginHelm struct {
	Tool                *model.ClusterTool
	Cluster             *Cluster
	LocalHostName       string
	LocalRepositoryPort int
	chartName           string
	ingressPort         int
	serviceName         string
	workloadType        string
	workloadName        string
	workloadMinR        int
}

func NewAddonPluginHelm(cluster *Cluster, tool *model.ClusterTool, chartName string) (*AddonPluginHelm, error) {
	return &AddonPluginHelm{
		Cluster:             cluster,
		Tool:                tool,
		LocalHostName:       constant.LocalRepositoryDomainName,
		LocalRepositoryPort: cluster.helmRepoPort,
		chartName:           chartName,
	}, nil
}

func (a AddonPluginHelm) Install(toolDetail model.ClusterToolDetail) error {
	if err := a.setDefaultVars(toolDetail); err != nil {
		return err
	}
	if err := installChart(a.Cluster.HelmClient, a.Tool, a.chartName, toolDetail.ChartVersion); err != nil {
		return err
	}
	// 创建ingress route
	ingressItem := &Ingress{
		name:    fmt.Sprintf("%s-ingress", a.Tool.Name),
		url:     fmt.Sprintf("%s.%s", a.Tool.Name, constant.DefaultIngress),
		service: a.serviceName,
		port:    a.ingressPort,
		version: a.Cluster.Version,
	}
	if err := createRoute(a.Cluster.Namespace, ingressItem, a.Cluster.KubeClient); err != nil {
		return err
	}

	// 等待执行
	switch strings.ToLower(a.workloadType) {
	case "sts":
		fallthrough
	case "statefulset":
		return waitForStatefulSetsRunning(a.Cluster.Namespace, a.workloadName, int32(a.workloadMinR), a.Cluster.KubeClient)
	case "deploy":
		fallthrough
	case "deployment":
		return waitForRunning(a.Cluster.Namespace, a.workloadName, int32(a.workloadMinR), a.Cluster.KubeClient)
	default:
		return errors.New("The addon plugin helm workload type not support, please check manifest file ")
	}
}

func (a AddonPluginHelm) Upgrade(toolDetail model.ClusterToolDetail) error {
	if err := a.setDefaultVars(toolDetail); err != nil {
		return err
	}
	return upgradeChart(a.Cluster.HelmClient, a.Tool, a.chartName, toolDetail.ChartVersion)
}

func (a AddonPluginHelm) Uninstall() error {
	return uninstall(a.Cluster.Namespace, a.Tool, fmt.Sprintf("%s-ingress", a.Tool.Name), a.Cluster.Version, a.Cluster.HelmClient, a.Cluster.KubeClient)

}

func (a AddonPluginHelm) setDefaultVars(toolDetail model.ClusterToolDetail) error {
	// 取出默认值
	defVar := map[string]interface{}{}
	if err := json.Unmarshal([]byte(toolDetail.Vars), &defVar); err != nil {
		return err
	}
	// 取出参数
	vars := map[string]interface{}{}
	if err := json.Unmarshal([]byte(a.Tool.Vars), &vars); err != nil {
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

	// 插值替换
	strings.ReplaceAll(a.Tool.Vars, "${AddonRepo}", fmt.Sprintf("%s:%d", a.LocalHostName, a.LocalRepositoryPort))
	a.serviceName, _ = vars[model.AddonPluginHelmServiceName].(string)
	a.ingressPort, _ = vars[model.AddonPluginHelmIngressPort].(int)
	a.workloadType, _ = vars[model.AddonPluginHelmWorkloadType].(string)
	a.workloadName, _ = vars[model.AddonPluginHelmWorkloadName].(string)
	a.workloadMinR, _ = vars[model.AddonPluginHelmMinReplicas].(int)

	if a.workloadName == "" || a.workloadType == "" || a.ingressPort == 0 {
		return errors.New("The addon plugin helm workloadType or workloadName or ingressPort is not exist , please check manifest file ")
	}
	return nil
}
