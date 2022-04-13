package servers

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ligao-cloud-native/cloud-shell/pkg/client"
	"github.com/ligao-cloud-native/cloud-shell/webshell"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net/http"
	"net/url"
	"os"
	"strings"
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
		return spdyExecutorFromReverseProxy
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
	podName, containerName, _, err := getPodContainerd(namespace)
	if err != nil {
		klog.Errorf(err.Error())
		return
	}

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
func spdyExecutorFromReverseProxy(w http.ResponseWriter, r *http.Request) {
	clusterName := mux.Vars(r)["clusterName"]
	masterID := r.URL.Query().Get("masterid")
	action := r.URL.Query().Get("action")
	path := r.URL.Query().Get("path")
	user := r.URL.Query().Get("user")
	namespace := fmt.Sprintf("%s-%s", clusterName, masterID)

	// 获取需要进入的pod, 通过标签筛选
	podName, containerName, host, err := getPodContainerd(namespace)
	if err != nil {
		klog.Errorf(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	target := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/exec?", namespace, podName)
	switch action {
	case "shell":
		target = target + "command=/bin/bash" + fmt.Sprintf("&container=%s&stdin=true&stdout=true&stderr=true&tty=true", containerName)
		uri, _ := url.Parse(target)
		r.URL = uri
	case "upload":
		if validatedUser(path, user) {
			target = fmt.Sprintf(target+"command=tar&commant=-xmf&command=-&command=-C&command=%s&container=%s&stdin=true&stdout=true&stderr=true&tty=false", path, containerName)
			uri, _ := url.Parse(target)
			r.URL = uri
		} else {
			w.WriteHeader(http.StatusBadRequest)
			klog.Errorf("path %s and username %s is not matched", path, user)
			return
		}
	case "download":
		if validatedUser(path, user) {
			target = fmt.Sprintf(target+"command=tar&commant=cf&command=-&command=%s&container=%s&stdin=true&stdout=true&stderr=true&tty=false", path, containerName)
			uri, _ := url.Parse(target)
			r.URL = uri
		} else {
			w.WriteHeader(http.StatusBadRequest)
			klog.Errorf("path %s and username %s is not matched", path, user)
			return
		}
	}

	klog.Error(webshell.ServeReverseProxy(host, w, r))
}

func getPodContainerd(namespace string) (pod, container, host string, err error) {
	kubeConfig, kubeClient := client.NewRestClient("", "")
	pods, err := kubeClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", "webshell")})
	if err != nil {
		klog.Errorf("list supported shell pods error: %s", err.Error())
		err = fmt.Errorf("no shell pods")
		return
	}

	if len(pods.Items) > 0 {
		pod = pods.Items[0].Name
		container = pods.Items[0].Spec.Containers[0].Name
		host = kubeConfig.Host
		return
	}

	err = fmt.Errorf("shell pods is empty")
	return
}

func validatedUser(path, username string) bool {
	subPath := strings.TrimPrefix(path, "/home/")
	user := strings.Split(subPath, "/")
	if len(user) > 0 {
		if strings.Compare(user[0], username) == 0 {
			return true
		}
		return false
	}
	return false
}
