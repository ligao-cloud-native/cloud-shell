package client

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
)

func SPDYExecuteStream(namespace, pod, container string, command []string, tty bool, streamOpts remotecommand.StreamOptions) {
	kubeConfig, kubeClient := NewRestClient("", "")
	req := kubeClient.CoreV1().RESTClient().Post().
		Namespace(namespace).Resource("pods").Name(pod).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       tty,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(kubeConfig, "POST", req.URL())
	if err != nil {
		klog.Errorf("NewSPDYExecutor error: %s", err.Error())
		return
	}

	err = exec.Stream(streamOpts)
	if err != nil {
		klog.Errorf("SPDY Executor Stream error: %s", err.Error())
	}
}
