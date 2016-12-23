package controller

import (
	"fmt"
	"net/url"
	"time"

	"github.com/appscode/errors"
	"github.com/appscode/log"
	"k8s.io/kubernetes/pkg/api"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	rest "k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned/remotecommand"
	remotecommandserver "k8s.io/kubernetes/pkg/kubelet/server/remotecommand"
)

type RemoteExecutor interface {
	Execute(*rest.Config, string, *url.URL, []string) error
}

type RemoteBashExecutor struct{}

func (e *RemoteBashExecutor) Execute(config *rest.Config, method string, url *url.URL, cmds []string) error {
	exec, err := remotecommand.NewExecutor(config, method, url)
	if err != nil {
		return errors.New().WithCause(err).WithMessage("failed to create executor").Internal()
	}
	stdIn := newStringReader(cmds)
	err = exec.Stream(remotecommand.StreamOptions{
		SupportedProtocols: remotecommandserver.SupportedStreamingProtocols,
		Stdin:              stdIn,
		Stdout:             DefaultWriter,
		Stderr:             DefaultWriter,
		Tty:                false,
	})
	if err != nil {
		log.Errorln("Error in exec", err)
		return errors.New().WithCause(err).WithMessage("failed to exec").Internal()
	}
	return nil
}

type ExecOptions struct {
	Namespace     string
	PodName       string
	ContainerName string
	Command       []string

	Executor RemoteExecutor
	Client   clientset.Interface
	Config   *rest.Config
}

func (p *ExecOptions) Run(retry int) error {
	err := p.Validate()
	if err != nil {
		return errors.New().WithCause(err).WithMessage("failed to validate").Internal()
	}
	var pod *api.Pod
	for i := 0; i < retry; i++ {
		pod, err = p.Client.Core().Pods(p.Namespace).Get(p.PodName)
		if err != nil || pod.Status.Phase != api.PodRunning {
			log.Debugln("pod not running waiting, tries", i+1)
			time.Sleep(time.Second * 30)
			continue
		}
		if pod.Status.Phase == api.PodRunning {
			log.Debugln("pod running quiting loop, tries", i+1)
			break
		}
	}
	if pod.Status.Phase != api.PodRunning || err != nil {
		return errors.New().WithMessage(fmt.Sprintf("pod %s is not running and cannot execute commands; current phase is %s", p.PodName, pod.Status.Phase)).Failed()
	}

	req := p.Client.Core().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		Param("container", p.ContainerName)

	req.VersionedParams(&api.PodExecOptions{
		Container: p.ContainerName,
		Command:   []string{"/bin/bash"},
		Stdin:     true,
		Stdout:    true,
		Stderr:    false,
		TTY:       false,
	}, api.ParameterCodec)

	return p.Executor.Execute(p.Config, "POST", req.URL(), p.Command)
}

func (p *ExecOptions) Validate() error {
	if len(p.PodName) == 0 {
		return errors.New().WithMessage("pod name must be specified").Failed()
	}
	if len(p.Command) == 0 {
		return errors.New().WithMessage("you must specify at least one command for the container").Failed()
	}
	if p.Executor == nil || p.Client == nil || p.Config == nil {
		return errors.New().WithMessage("client, client config, and executor must be provided").Failed()
	}
	return nil
}

var DefaultWriter = &StringWriter{
	data: make([]byte, 0),
}

type StringWriter struct {
	data []byte
}

func (s *StringWriter) Flush() {
	s.data = make([]byte, 0)
}

func (s *StringWriter) Output() string {
	return string(s.data)
}

func (s *StringWriter) Write(b []byte) (int, error) {
	log.Infoln("$ ", string(b))
	s.data = append(s.data, b...)
	return len(b), nil
}
