package v1

import (
	"flame/pkg/utils/k8s"
	"flame/pkg/watcher"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
)

type Flame struct {
	Web            *gin.Engine
	K8sClient      *kubernetes.Clientset
	PromController *watcher.PromController
	// RuleController
}

func NewAndRunFlame() {
	k8sClientSet := k8s.NewK8sClient()
	f := &Flame{
		Web:            gin.Default(),
		K8sClient:      k8sClientSet,
		PromController: watcher.NewPromController(k8sClientSet),
	}
	go f.PromController.RunPromController()
	Group(f)
	if err := f.Web.Run(); err != nil {
		panic(err)
	}
}
