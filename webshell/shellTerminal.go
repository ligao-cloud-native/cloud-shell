package webshell

import (
	"fmt"
	"github.com/ligao-cloud-native/cloud-shell/pkg/client"
	"k8s.io/client-go/tools/remotecommand"
	"net/http"
	"os"
	"path"
)

func ShellTerminal(w http.ResponseWriter, r *http.Request, namespace, pod, container string, user string) {
	wsIO := NewTerminalSession(client.NewWsClient(w, r))
	wsIO.sizeChan <- remotecommand.TerminalSize{Width: 200, Height: 40}
	defer wsIO.Close()

	//command := []string{"/bin/sh"}
	subcmd := fmt.Sprintf("id %s > /dev/null 2>&1;"+
		"if [ $? -eq 1 ]; then "+
		"useradd -U -m %s; chmod 700 /home/%s > /dev/null 2>&1;"+
		"mkdir -p /home/%s/.kube;"+
		"cp /.kube/kubeconfig /home/%s/.kube/config;"+
		"chown -R %s:%s /home/%s/.kube;"+
		"su --sesion-command bash - %s;"+
		"else su --sesion-command bash - %s;"+
		"fi", user, user, user, user, user, user, user, user, user, user)
	command := []string{"/bin/bash", "-c", "--", subcmd}

	client.SPDYExecuteStream(namespace, pod, container, command, true, remotecommand.StreamOptions{
		Stdin:             wsIO,
		Stdout:            wsIO,
		Stderr:            wsIO,
		TerminalSizeQueue: wsIO,
		Tty:               true,
	})
}

// CopyTo copy local file to pod dir
func CopyTo(w http.ResponseWriter, r *http.Request, namespace, pod, container string, podDir string) {
	//reader, writer, _ := os.Pipe()
	//go func() {
	//	file, _, _ := r.FormFile("uploadfile")
	//	data, _ := ioutil.ReadAll(file)
	//	klog.Infof("Copy to pod data from file: %s", string(data))
	//	defer file.Close()
	//	io.Copy(writer, bytes.NewReader(data))
	//}()
	//go func() {
	//	 data, _ := ioutil.ReadAll(reader)
	//	klog.Infof("Copy to pod data from reader: %s", string(data))
	//}()

	wsIO := NewTerminalSession(client.NewWsClient(w, r))
	defer wsIO.Close()

	command := []string{"tar", "-xf", "-", "-C", path.Dir(podDir + "/")}

	client.SPDYExecuteStream(namespace, pod, container, command, false, remotecommand.StreamOptions{
		Stdin:  wsIO,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    false,
	})
}

// CopyFrom copy pod dir to local
func CopyFrom(w http.ResponseWriter, r *http.Request, namespace, pod, container string, podDir string) {
	wsIO := NewTerminalSession(client.NewWsClient(w, r))
	defer wsIO.Close()

	command := []string{"tar", "cf", "-", podDir}

	client.SPDYExecuteStream(namespace, pod, container, command, false, remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: wsIO,
		Stderr: os.Stderr,
		Tty:    false,
	})
}
