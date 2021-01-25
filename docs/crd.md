# CRD 定义说明

主要定义三层 `CRD type`，从上到下分别为：`ClusterTypes`、`AppSetTypes`、`AdvDeploymentTypes`。依次介绍如下：

## 一、ClusterTypes

描述单元中集群的抽象类型。

### 1. Spec

- KubeConfig `string`
- SymNodeName `string`
- Additionals `map[string]string`
- DisplayName `string`
- Description `string`
- HelmSpec `*HelmSpec`
  - Namespace `string`
  - OverrideImageSpec `string`
  - MaxHistory `int`
- AlertSpec `*AlertSpec`
  - Enable `bool`
- Apps `[]*HelmChartSpec`
  - Name `string`
  - Namespace `string`
  - Repo `string`
  - ChartName `string`
  - ChartVersion `string`
  - OverrideValue `string`
  - Values `map[string]string`
- Pause `bool`

### 2. Status

- AppStatus `[]AppHelmStatuses`
  - Name `string`
  - ChartVersion `string`
  - RlsName `string`
  - RlsStatus `string`
  - RlsVersion `int32`
  - OverrideVa  `string`
- ClusterStatus `[]ClusterComponentStatus`
  - Name `string`
  - Conditions `[]v1.ComponentCondition`
- Version `*version.Info`
- MonitoringStatus `*MonitoringStatus`
  - GrafanaEndpoint `*string`
  - AlertManagerEndpoint `*string`
  - PrometheusEndpoint `*string`
- NodeDetail `*NodeDetail`
  - NodeStatus `[]*NodeStatus`
      - NodeName `string`
      - Etcd `bool`
      - ControlPlane `bool`
      - Worker `bool`
      - Capacity `v1.ResourceList`
      - Allocatable `v1.ResourceList`
      - Requested `v1.ResourceList`
      - Limits `v1.ResourceList`
      - Ready `string`
      - KernelDeadlock `string`
      - NetworkUnavailable `string`
      - OutOfDisk `string`
      - MemoryPressure `string`
      - DiskPressure `string`
      - PIDPressure `string`
      - CpuUsagePercent `int32`
      - MemoryUsagePercent `int32`
      - PodUsagePercent `int32`
      - StorageUsagePercent `int32`
  - Capacity `v1.ResourceList`
  - Allocatable `v1.ResourceList`
  - Requested `v1.ResourceList`
  - Limits `v1.ResourceList`
  - CpuUsagePercent `int32`
  - PodUsagePercent `int32`
  - StorageUsagePercent `int32`
  - MemoryUsagePercent `int32`


## 二、AppSetTypes

### 1. Spec

- Labels `map[string]string`
- Meta `map[string]string`
- Replicas `*int32`
- ServiceName `*string`
- DeployType `string`
- PodSpec `PodSpec`
  - Selector `*metav1.LabelSelector`
  - Template `*corev1.PodTemplateSpec`
  - Chart `ChartSpec`
- UpdateStrategy `AppSetUpdateStrategy`
  - UpgradeType `string`
  - MinReadySeconds `int32`
  - PriorityStrategy `*UpdatePriorityStrategy`
    - OrderPriority `[]UpdatePriorityOrderTerm`
      - OrderedKey `string`
    - WeightPriority `[]UpdatePriorityWeightTerm`
      - Weight `int32`
      - MatchSelector `metav1.LabelSelector`
  - CanaryClusters `[]string`
  - Paused `bool`
  - NeedWaitingForConfirm `bool`
- ClusterTopology `ClusterTopology`
  - Clusters `[]TargetCluster`
    - Name `string`
    - PodSets `[]PodSet`
      - Name `string`
      - NodeSelectorTerm `corev1.NodeSelectorTerm`
      - Replicas `intstr.IntOrString`
      - RawValues `string`

### 2. Status

- ObservedGeneration `int64`
- ReadyReplicas `int32`
- Replicas `int32`
- UpdatedReplicas `int32`
- UpdatedReadyReplicas `int32`
- Conditions `[]AppSetCondition`
  - Type `AppSetConditionType`
  - Status `corev1.ConditionStatus`
  - LastTransitionTime `metav1.Time`
  - Reason `string`
  - Message `string`
- Status `AppStatus`
- AppActual `AppActual`
  - Total `int32`
  - Items `[]*AppActualItem`
  - Pods `[]*Pod`
  - WarnEvents `[]*Event`
  - Service `*Service`

## 三、AdvDeploymentTypes

### 1. Spec

- DeployType `string`
- Replicas `*int32`
- PodSpec `PodSpec`
  - Selector `*metav1.LabelSelector`
  - Template `*corev1.PodTemplateSpec`
  - Chart `*ChartSpec`
    - RawChart `*[]byte`
    - Url `*ChartUrl`
      - Url `string`
      - Version `string`
- ServiceName `*string`
- UpdateStrategy `AdvDeploymentUpdateStrategy`
  - UpgradeType `string`
  - StatefulSetStrategy `*StatefulSetStrategy`
    - Partition `*int32`
    - MaxUnavailable `*intstr.IntOrString`
    - PodUpdatePolicy `PodUpdateStrategyType`
  - MinReadySeconds `int32`
  - Meta `map[string]string`
  - PriorityStrategy `*UpdatePriorityStrategy`
    - OrderPriority `[]UpdatePriorityOrderTerm`
      - OrderedKey `string`
    - WeightPriority `[]UpdatePriorityWeightTerm`
      - Weight `int32`
      - MatchSelector `metav1.LabelSelector`
  - Paused `bool`
  - NeedWaitingForConfirm `bool`
- Topology `Topology`
  - PodSets `[]PodSet`
    - Name `string`
    - NodeSelectorTerm `corev1.NodeSelectorTerm`
    - Replicas `*intstr.IntOrString`
    - RawValues `string`
- RevisionHistoryLimit `*int32`

### 2. Status

- Version `string`
- Message `string`
- Replicas `int32`
- ReadyReplicas `int32`
- CurrentReplicas `int32`
- UpdatedReplicas `int32`
- PodSets `[]PodSetStatus`
  - Name `string`
  - ObservedGeneration `int64`
  - Replicas `int32`
  - ReadyReplicas `int32`
  - CurrentReplicas `int32`
  - UpdatedReplicas `int32`
- Conditions `[]AdvDeploymentCondition`
  - Type `AdvDeploymentConditionType`
  - Status `corev1.ConditionStatus`
  - LastUpdateTime `metav1.Time`
  - LastTransitionTime `metav1.Time`
  - Reason `string`
  - Message `string`
- ObservedGeneration `int64`
- CurrentRevision `string`
- UpdateRevision `string`
- CollisionCount `*int32`

