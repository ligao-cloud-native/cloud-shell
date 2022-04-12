package servers

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ligao-cloud-native/cloud-shell/pkg/client"
	"github.com/ligao-cloud-native/cloud-shell/pkg/webshell"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"time"
)

// StartHTTPServer starts the http service
func StartHttpServer() {
	// StrictSlash(true)，定义uri尾部斜线的行为。
	// true表示如果定义的路由为/path/，访问/path时会重定向到/path/
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/cluster/{clusterName}/cloudshell", getHandleFunc())

	server := &http.Server{
		Addr:         ":8088",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	klog.Fatal(server.ListenAndServe())
}

func getHandleFunc() func(http.ResponseWriter, *http.Request) {
	switch os.Getenv("PROXY_ENABLED") {
	case "true":
		return spdyExecutorFromreverseProxy
	default:
		return spdyExecutor
	}
}

//spdyExecutor 基于SPDY协议实现进入kubernetes pod 的终端操作。
func spdyExecutor(w http.ResponseWriter, r *http.Request) {
	clusterName := mux.Vars(r)["clusterName"]
	masterID := r.URL.Query().Get("masterid")
	action := r.URL.Query().Get("action")
	path := r.URL.Query().Get("path")
	user := r.URL.Query().Get("user")
	namespace := fmt.Sprintf("%s-%s", clusterName, masterID)

	// 获取需要进入的pod, 通过标签筛选
	_, kubeClient := client.NewRestClient("", "")
	pods, err := kubeClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", "webshell")})
	if err != nil {
		klog.Errorf("list supported shell pods error: %s", err.Error())
		return
	}
	podName := pods.Items[0].Name
	containerName := pods.Items[0].Spec.Containers[0].Name

	switch action {
	case "shell":
		webshell.ShellTerminal(w, r, namespace, podName, containerName, user)
	case "upload":
		webshell.CopyTo(w, r, namespace, podName, clusterName, path)
	case "download":
		webshell.CopyFrom(w, r, namespace, podName, clusterName, path)
	}

}

// spdyExecutorFromreverseProxy 基于SPDY协议实现进入kubernetes pod 的终端操作.
// 与spdyExecutor方法不同的是使用了反向代理
func spdyExecutorFromreverseProxy(w http.ResponseWriter, r *http.Request) {

}
