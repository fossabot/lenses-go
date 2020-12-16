package processor

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/lensesio/lenses-go/pkg/api"
	config "github.com/lensesio/lenses-go/pkg/configs"
	test "github.com/lensesio/lenses-go/test"
	"github.com/stretchr/testify/assert"
)

var aksk8s = api.KubernetesTarget{Cluster: "aks", Namespaces: []string{"prod"}, Version: "1.0.0"}
var aksk8Res = ListTargetsResult{Type: "Kubernetes", ClusterName: "aks", Namespace: "prod", Version: "1.0.0"}
var eksk8s = api.KubernetesTarget{Cluster: "eks", Namespaces: []string{"dev"}, Version: "1.0.0"}
var eksk8Res = ListTargetsResult{Type: "Kubernetes", ClusterName: "eks", Namespace: "dev", Version: "1.0.0"}
var connect = api.KafkaConnectTarget{Cluster: "my-kafka-connect", Version: "1.0.0"}
var connectRes = ListTargetsResult{Type: "Connect", ClusterName: "my-kafka-connect", Namespace: "", Version: "1.0.0"}
var targetList = &api.DeploymentTargets{
	Kubernetes: []api.KubernetesTarget{aksk8s, eksk8s},
	Connect:    []api.KafkaConnectTarget{connect},
}

var targetsAsJSON, _ = json.Marshal(targetList)

func TestListTargetDeploymentCommand(t *testing.T) {

	list := [3]ListTargetsResult{aksk8Res, eksk8Res, connectRes}
	e, _ := json.Marshal(list)

	//setup http client
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(string(targetsAsJSON)))
	})
	httpClient, teardown := test.TestingHTTPClient(h)
	defer teardown()

	client, err := api.OpenConnection(test.ClientConfig, api.UsingClient(httpClient))

	assert.Nil(t, err)

	config.Client = client

	cmd := NewGetProcessorsCommand()
	var outputValue string
	cmd.PersistentFlags().StringVar(&outputValue, "output", "json", "")
	output, err := test.ExecuteCommand(cmd, "targets")

	assert.Nil(t, err)
	assert.NotEmpty(t, output)
	assert.Equal(t, string(e), strings.TrimSuffix(output, "\n"))

	config.Client = nil
}

func TestListTargetK8sDeploymentCommand(t *testing.T) {

	//result := `[{"type":"Kubernetes","clusterName":"aks","namespace":"prod","version":"1.0.0"},{"type":"Kubernetes","clusterName":"eks","namespace":"dev","version":"1.0.0"}]`
	list := [2]ListTargetsResult{aksk8Res, eksk8Res}
	e, _ := json.Marshal(list)

	//setup http client
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(string(targetsAsJSON)))
	})
	httpClient, teardown := test.TestingHTTPClient(h)
	defer teardown()

	client, err := api.OpenConnection(test.ClientConfig, api.UsingClient(httpClient))

	assert.Nil(t, err)

	config.Client = client

	cmd := NewGetProcessorsCommand()
	var outputValue string
	cmd.PersistentFlags().StringVar(&outputValue, "output", "json", "")
	output, err := test.ExecuteCommand(cmd, "targets", "--target-type=kubernetes")

	assert.Nil(t, err)
	assert.NotEmpty(t, output)
	assert.Equal(t, string(e), strings.TrimSuffix(output, "\n"))

	config.Client = nil
}

func TestListTargetK8sClusterNameDeploymentCommand(t *testing.T) {

	list := [1]ListTargetsResult{aksk8Res}
	e, _ := json.Marshal(list)

	//setup http client
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(string(targetsAsJSON)))
	})
	httpClient, teardown := test.TestingHTTPClient(h)
	defer teardown()

	client, err := api.OpenConnection(test.ClientConfig, api.UsingClient(httpClient))

	assert.Nil(t, err)

	config.Client = client

	cmd := NewGetProcessorsCommand()
	var outputValue string
	cmd.PersistentFlags().StringVar(&outputValue, "output", "json", "")
	output, err := test.ExecuteCommand(cmd, "targets", "--target-type=kubernetes", "--cluster-name=aks")

	assert.Nil(t, err)
	assert.NotEmpty(t, output)
	assert.Equal(t, string(e), strings.TrimSuffix(output, "\n"))

	config.Client = nil
}

func TestListTargetConnectDeploymentCommand(t *testing.T) {

	list := [1]ListTargetsResult{connectRes}
	e, _ := json.Marshal(list)

	//setup http client
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(string(targetsAsJSON)))
	})
	httpClient, teardown := test.TestingHTTPClient(h)
	defer teardown()

	client, err := api.OpenConnection(test.ClientConfig, api.UsingClient(httpClient))

	assert.Nil(t, err)

	config.Client = client

	cmd := NewGetProcessorsCommand()
	var outputValue string
	cmd.PersistentFlags().StringVar(&outputValue, "output", "json", "")
	output, err := test.ExecuteCommand(cmd, "targets", "--target-type=connect")

	assert.Nil(t, err)
	assert.NotEmpty(t, output)
	assert.Equal(t, string(e), strings.TrimSuffix(output, "\n"))

	config.Client = nil
}
