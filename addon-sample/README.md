# 说明

`Addon-Plugins`对Kobeoperator做了如下改动：

1. 在初始化时将加载`addon-plugins.json/yaml`文件,其中的`plugins`将会被转换为`ClusterToolDetail`写入数据库(其中的参数会被转换为vars一并写入)
   ,并将其中的内容缓存到内存中供后续使用，这个改动使得tools可以被动态修改
2. tools的操作除了`helm`外还支持`ansible`,只需要将helmChart替换为playbook脚本名称即可（需要放在`/opt/kubeoperator/data/kobe/project/ko`目录下)
3. 在创建集群部署时，会将`addon-plugins`下的`globals`加入到ansible的全局变量中，从而支持对playbook自定义修改，同时加载的tools将会被写入到`ClusterTool`中
4. 在安装集群时就自动部署部分tools，实现自定义的集群部署（todo）

# 设计理练
`addon-plugins.yaml`如下：
```yaml
kind: addonPluginsManifest
apiVersion: v1
spec:
  # 全局参数，将会写入到ansible hosts执行playbook的全局变量中
  globals:
    isString: '222'
    isNumber: 30
    isBoolean: true
  # 插件 该部分会在集群创建时候被写入到应用库中
  plugins:
    - name: add-nfs
        metadata:
          # 对应CTD
          architecture: all
          version: v1.05
          # 对应CT
          describe: "简介"
          logo: nfs.jpg
          # 将对应default_vars
          addonType: ansible
          # ansible 为ansible playbookName helm 为 chartName
          resource: 05-addon-version
          isInit: true
        # 当类型为helm时候的字段
        helm:
          workloadType: stsOrDeploy
          workloadName: name
          serviceName: serviceName
          IngressPort: 80
        # 索引 在前端展示出来的选项卡
        schema:
          addonType: number
          resource: string
        # 将合并vars 到 ctd的vars中 使用${AddonRepo}可以动态替换为当前ko仓库
        vars: { }
```
