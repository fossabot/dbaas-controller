// dbaas-controller
// Copyright (C) 2020 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Package kubectl provides kubectl CLI wrapper.
package kubectl

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/percona-platform/dbaas-controller/utils/app"
)

func TestNewKubeCtl(t *testing.T) {
	ctx := app.Context()

	cmd, err := getKubectlCmd(ctx)
	require.NoError(t, err)

	validKubeconfig, err := run(ctx, cmd, []string{"config", "view", "-o", "json"}, nil)
	require.NoError(t, err)

	t.Run("BasicNewKubeCtl", func(t *testing.T) {
		kubeCtl, err := NewKubeCtl(ctx, string(validKubeconfig))
		require.NoError(t, err)
		// lookup for kubeconfig path
		var kubeconfigFlag string
		for _, option := range kubeCtl.cmd {
			if strings.HasPrefix(option, "--kubeconfig") {
				kubeconfigFlag = option
				break
			}
		}

		assert.True(t, strings.HasSuffix(kubeconfigFlag, kubeconfigFileName))

		kubeconfigFilePath := strings.Split(kubeconfigFlag, "=")[1]
		kubeconfigActual, err := ioutil.ReadFile(kubeconfigFilePath) //nolint:gosec
		require.NoError(t, err)
		assert.Equal(t, validKubeconfig, kubeconfigActual)

		tmpDir := strings.TrimSuffix(kubeconfigFilePath, "/"+kubeconfigFileName)
		assert.Equal(t, kubeCtl.tmpDir, tmpDir)

		err = kubeCtl.Cleanup()
		require.NoError(t, err)

		_, err = os.Stat(tmpDir)
		require.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})
}

const kubernetsVersions = `
{
	"clientVersion": {
	  "major": "1",
	  "minor": "16",
	  "gitVersion": "v1.16.8",
	  "gitCommit": "ec6eb119b81be488b030e849b9e64fda4caaf33c",
	  "gitTreeState": "clean",
	  "buildDate": "2020-03-12T21:00:06Z",
	  "goVersion": "go1.13.8",
	  "compiler": "gc",
	  "platform": "darwin/amd64"
	},
	"serverVersion": {
	  "major": "1",
	  "minor": "16",
	  "gitVersion": "v1.16.8",
	  "gitCommit": "ec6eb119b81be488b030e849b9e64fda4caaf33c",
	  "gitTreeState": "clean",
	  "buildDate": "2020-03-12T20:52:22Z",
	  "goVersion": "go1.13.8",
	  "compiler": "gc",
	  "platform": "linux/amd64"
	}
}
`

func Test_selectCorrectKubectlVersions(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		got, err := selectCorrectKubectlVersions([]byte(kubernetsVersions))
		require.NoError(t, err)
		expected := []string{
			"kubectl-1.17",
			"kubectl-1.16",
			"kubectl-1.15",
		}
		assert.Equal(t, got, expected)
	})

	t.Run("empty", func(t *testing.T) {
		got, err := selectCorrectKubectlVersions([]byte(""))
		assert.Errorf(t, err, "unexpected end of JSON input")
		assert.Nil(t, got)
	})
}

func Test_getKubectlCmd(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		ctx := context.TODO()
		got, err := getKubectlCmd(ctx)
		require.NoError(t, err)
		assert.Equal(t, got, []string{"minikube", "kubectl", "--"})
	})
}

func Test_lookupCorrectKubectlCmd(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		args := []string{
			"kubectl-1.17",
			"kubectl-1.16",
			"kubectl-1.15",
		}
		got, err := lookupCorrectKubectlCmd(args)
		require.NoError(t, err)
		assert.Equal(t, got, []string{"minikube", "kubectl", "--"})
	})

	t.Run("empty_kubectl_list_of_correct_version", func(t *testing.T) {
		got, err := lookupCorrectKubectlCmd(nil)
		require.NoError(t, err)
		assert.Equal(t, got, []string{"minikube", "kubectl", "--"})
	})
}