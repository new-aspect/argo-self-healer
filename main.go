package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	argoclientset "github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// 极简自愈脚本，这个脚本会连接到k8s, 像雷达扫描所有ArgoCD Application, 一旦发现状态不对，就执行自愈

// 思路
// (1)创建连接K8s的ArgoCD客户端
// (2)循环检查所有命名空间的 ArgoCD Application，10秒1次
// (3)模拟救火逻辑，如果状态是Unknow或我自定义的异常，模拟自愈
// (4)给该Argo Application打标签，让ArgoCD重新评估它的状态
func main() {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	argocdClient, err := argoclientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	fmt.Println("Askning 的 ArgoCD 自愈雷达已经启动，正在监控集群...")

	for {
		//列出来命名空间下的 Application

		apps, err := argocdClient.ArgoprojV1alpha1().Applications("").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("❌ 获取列表失败: %v\n", err)
			continue
		}

		for _, app := range apps.Items {
			status := string(app.Status.Sync.Status)

			// 模拟救火
			if status == "Unknown" || status == "" {
				fmt.Printf("发现异常应用：[%s]，当前状态：%v\n", app.Name, app.Status)

				fmt.Printf("🛠 正在为 [%s] 执行自愈动作：强制刷新缓存并同步...\n", app.Name)

				// 这里模拟自愈
				app.Annotations = map[string]string{"self-healed-by": "zhaoning-bot", "time": time.Now().String()}
				_, err = argocdClient.ArgoprojV1alpha1().Applications(app.Namespace).Update(context.Background(), &app, metav1.UpdateOptions{})
				if err != nil {
					fmt.Printf("❌ 自愈失败: %v\n", err)
				} else {
					fmt.Printf("✅ 应用 [%s] 自愈指令已下发！\n", app.Name)
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}
