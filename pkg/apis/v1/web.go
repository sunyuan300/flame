package v1

import (
	"flame/pkg/utils/k8s"
	"flame/pkg/watcher"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
)

type Flame struct {
	Web             *gin.Engine
	K8sClient       *kubernetes.Clientset
	PromController  *watcher.PromController
	RulesController *watcher.RulesController
}

func NewAndRunFlame() {
	k8sClientSet := k8s.NewK8sClient()
	f := &Flame{
		Web:             gin.Default(),
		K8sClient:       k8sClientSet,
		PromController:  watcher.NewPromController(k8sClientSet),
		RulesController: watcher.NewRulesController(k8sClientSet),
	}
	f.PromController.Instance.Lock.Lock()
	go f.PromController.RunPromController()
	f.RulesController.Instance.Lock.Lock()
	go f.RulesController.RunRulesController()
	Group(f)
	if err := f.Web.Run(); err != nil {
		panic(err)
	}
}
