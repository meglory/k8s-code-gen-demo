#### k8s提供了一系列code-generator工具，可以很方便的为k8s中的资源生成代码

- deepcopy-gen: 生成深度拷贝方法,避免性能开销
- client-gen: 为资源生成标准的操作方法(get,list,create,update,patch,delete,deleteCollection,watch)
- informer-gen: 生成informer,提供事件机制来相应kubernetes的event
- lister-gen: 为get和list方法提供只读缓存层

其中informer和listers是构建controller的基础,kubebuilder也是基于informer的机制生成的代码.code-generator还专门整合了这些gen,形成了generate-groups.sh和generate-internal-groups.sh这两个脚本.

#### 下面以一个简单的crd demo工程介绍下脚本的使用方法

##### 1.首先，创建一个demo工程

```shell
# 创建工程目录
$ mkdir k8s-code-gen-demo && cd k8s-code-gen-demo
# go mod初始化
$ go mod init k8s-code-gen-demo
# 下载code-generator依赖
$ go get k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
$ go get k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
$ go get k8s.io/code-generator v0.0.0-20191003035328-700b1226c0bd
```

##### 2.编写doc.go(包描述)和types.go(crd定义)

- doc.go代码如下

```go
// +k8s:deepcopy-gen=package
// +groupName=samplecontroller.k8s.io

// v1alpha1版本的api包
package v1alpha1
```
> 上面这两行注释是code-generator的global tag，需要和其他注释空行隔开

+k8s:deepcopy-gen=: 它告诉deepcopy默认为该包中的每一个类型创建deepcopy方法，如果不需要深度复制,可以选择关闭此功能// +k8s:deepcopy-gen=false如果不启用包级别的深度复制,那么就需要在每个类型上加入深度复制// +k8s:deepcopy-gen=true

+groupName: 定义group的名称,注意别弄错了.注意 这里是 +k8s:deepcopy-gen=,最后是 = ,和local中的区别开来.

- types.go代码

```go
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Demo is a specification for a Foo resource
type Demo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DemoSpec   `json:"spec"`
	Status DemoStatus `json:"status"`
}

// DemoSpec is the spec for a Foo resource
type DemoSpec struct {
	DeploymentName string `json:"deploymentName"`
	Replicas       *int32 `json:"replicas"`
}

// DemoStatus is the status for a Foo resource
type DemoStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DemoList is a list of Foo resources
type DemoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Demo `json:"items"`
}
```
> // +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
上面这行注释，是code-generator的local tag

- +genclient: 此标签是告诉client-gen,为此类型创建clientset,但也有以下几种用法.

1. 对于集群范围内的资源(没有namespace限制的),需要使用// +genclient:nonNamespaced,生成的clientset中的namespace()方法就不再需要传入参数
2. 使用子资源分离的,例如/status分离的,则需要使用+genclient:noStatus,来避免更新到status资源(当然代码的struct中也没有status)
3. 另外还支持下面这些值

```go
// +genclient:noVerbs
// +genclient:onlyVerbs=create,delete
// +genclient:skipVerbs=get,list,create,update,patch,delete,deleteCollection,watch
// +genclient:method=Create,verb=create,result=k8s.io/apimachinery/pkg/apis/meta/v1.Status
```

- +k8s:deepcopy-gen:interfaces=: 为struct生成实现 tag值的DeepCopyXxx方法,例如:// +k8s:deepcopy-gen:interfaces=example.com/pkg/apis/example.SomeInterface将生成 DeepCopySomeInterface()方法

##### 3. 编写代码生成脚本
- 在工程下新建一个hack目录，然后往里面添加hack/tools.go代码文件，引入code-generator依赖，然后添加api文件头boilerplate.go.txt文件，最后编写生成脚本update-codegen.sh，代码如下：

```shell
set -o errexit
set -o nounset
set -o pipefail

../vendor/k8s.io/code-generator/generate-groups.sh \
  "deepcopy,client,informer,lister" \
  k8s-code-gen-demo/generated \
  k8s-code-gen-demo/pkg/apis \
  democontroller:v1alpha1 \
  --go-header-file $(pwd)/boilerplate.go.txt \
  --output-base $(pwd)/../../
```

##### 4. 生成代码

按以下步骤操作：

```shell
# 生成vendor文件夹
$ go mod vendor
# 赋予脚本权限
$ chmod -R 777 vendor
# 执行脚本生成代码，需要预先配置$GOPATH env
$ cd hack && ./update-codegen.sh
Generating deepcopy funcs
Generating clientset for democontroller:v1alpha1 at k8s-code-gen-demo/generated/clientset
Generating listers for democontroller:v1alpha1 at k8s-code-gen-demo/generated/listers
Generating informers for democontroller:v1alpha1 at k8s-code-gen-demo/generated/informers

#此时目录变为如下情况
$cd ../ && tree -L 4
.
├── cmd
│   └── demo
├── generated
│   ├── clientset
│   │   └── versioned
│   │       ├── clientset.go
│   │       ├── doc.go
│   │       ├── fake
│   │       ├── scheme
│   │       └── typed
│   ├── informers
│   │   └── externalversions
│   │       ├── democontroller
│   │       ├── factory.go
│   │       ├── generic.go
│   │       └── internalinterfaces
│   └── listers
│       └── democontroller
│           └── v1alpha1
├── go.mod
├── go.sum
├── hack
│   ├── boilerplate.go.txt
│   ├── tools.go
│   └── update-codegen.sh
├── pkg
│   └── apis
│       └── democontroller
│           └── v1alpha1

```
##### 5. 编写register代码
在pkg/apis/democontroller/v1alpha1下添加register.go，代码如下：

```go
/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects
// 注册自己的自定义资源
var SchemeGroupVersion = schema.GroupVersion{Group: "samplecontroller.k8s.io", Version: "v1alpha1"}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	//注意,添加了foo/foolist 两个资源到scheme
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Demo{},
		&DemoList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
```
##### 6. 编写程序入口main.go，执行编译
```go
package main

import (
	"k8s-code-gen-demo/generated/clientset/versioned"
	"k8s-code-gen-demo/generated/clientset/versioned/typed/democontroller/v1alpha1"
	"k8s-code-gen-demo/generated/informers/externalversions"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

var log = logf.Log.WithName("cmd")

func main() {
	flag.Parse()
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	client, err := v1alpha1.NewForConfig(cfg); if err != nil {
		panic(err.Error())
	}
	demoList, err := client.Demos("test").List(metav1.ListOptions{}); if err != nil {
		panic(err.Error())
	}
	log.Info(fmt.Sprintf("demoList: [%s]", demoList))
	clientset, err := versioned.NewForConfig(cfg); if err != nil {
		panic(err.Error())
	}
	factory := externalversions.NewSharedInformerFactory(clientset, 30*time.Second)
	demo, err := factory.Democontroller().V1alpha1().Demos().Lister().Demos("test").Get("test"); if err != nil {
		panic(err.Error())
	}
	log.Info(fmt.Sprintf("demo: [%s]", demo))
}
```
然后，编译成功就大功告成了
```shell
$ cd .. && go build cmd/demo/main.go
$ 
```