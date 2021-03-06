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

package psmdbcluster

import (
	"os"
	"testing"

	controllerv1beta1 "github.com/percona-platform/dbaas-api/gen/controller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/percona-platform/dbaas-controller/tests"
)

func TestPSMDBClusterAPI(t *testing.T) {
	kubeconfig := os.Getenv("PERCONA_TEST_DBAAS_KUBECONFIG")
	if kubeconfig == "" {
		t.Skip("PERCONA_TEST_DBAAS_KUBECONFIG env variable is not provided")
	}
	name := "api-psmdb-test-cluster"

	clusters, err := tests.PSMDBClusterAPIClient.ListPSMDBClusters(tests.Context, &controllerv1beta1.ListPSMDBClustersRequest{
		KubeAuth: &controllerv1beta1.KubeAuth{
			Kubeconfig: kubeconfig,
		},
	})
	require.NoError(t, err)

	var clusterFound bool
	for _, cluster := range clusters.Clusters {
		if cluster.Name == name {
			clusterFound = true
		}
	}
	require.Falsef(t, clusterFound, "There should not be cluster with name %s", name)

	createPSMDBClusterResponse, err := tests.PSMDBClusterAPIClient.CreatePSMDBCluster(tests.Context, &controllerv1beta1.CreatePSMDBClusterRequest{
		KubeAuth: &controllerv1beta1.KubeAuth{
			Kubeconfig: kubeconfig,
		},
		Name: name,
		Params: &controllerv1beta1.PSMDBClusterParams{
			ClusterSize: 3,
			Replicaset: &controllerv1beta1.PSMDBClusterParams_ReplicaSet{
				ComputeResources: &controllerv1beta1.ComputeResources{
					CpuM:        1000,
					MemoryBytes: 1024 * 1024 * 1024,
				},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, createPSMDBClusterResponse)

	clusters, err = tests.PSMDBClusterAPIClient.ListPSMDBClusters(tests.Context, &controllerv1beta1.ListPSMDBClustersRequest{
		KubeAuth: &controllerv1beta1.KubeAuth{
			Kubeconfig: kubeconfig,
		},
	})
	assert.NoError(t, err)

	for _, cluster := range clusters.Clusters {
		if cluster.Name == name {
			assert.Equal(t, int32(3), cluster.Params.ClusterSize)
			assert.Equal(t, int64(1024*1024*1024), cluster.Params.Replicaset.ComputeResources.MemoryBytes)
			assert.Equal(t, int32(1000), cluster.Params.Replicaset.ComputeResources.CpuM)
			clusterFound = true
		}
	}
	assert.True(t, clusterFound)

	deletePSMDBClusterResponse, err := tests.PSMDBClusterAPIClient.DeletePSMDBCluster(tests.Context, &controllerv1beta1.DeletePSMDBClusterRequest{
		KubeAuth: &controllerv1beta1.KubeAuth{
			Kubeconfig: kubeconfig,
		},
		Name: name,
	})
	require.NoError(t, err)
	require.NotNil(t, deletePSMDBClusterResponse)
}
