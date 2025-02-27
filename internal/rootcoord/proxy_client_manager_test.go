// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rootcoord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/milvus-io/milvus/internal/proto/proxypb"
	"github.com/milvus-io/milvus/internal/types"
	"github.com/milvus-io/milvus/internal/util/etcd"
	"github.com/milvus-io/milvus/internal/util/sessionutil"
)

func TestProxyClientManager_GetProxyClients(t *testing.T) {
	Params.Init()

	core, err := NewCore(context.Background(), nil)
	assert.Nil(t, err)
	cli, err := etcd.GetEtcdClient(&Params.EtcdCfg)
	defer cli.Close()
	assert.Nil(t, err)
	core.etcdCli = cli

	core.SetNewProxyClient(
		func(se *sessionutil.Session) (types.Proxy, error) {
			return nil, errors.New("failed")
		},
	)

	pcm := newProxyClientManager(core)

	session := &sessionutil.Session{
		ServerID: 100,
		Address:  "localhost",
	}

	sessions := []*sessionutil.Session{session}
	pcm.GetProxyClients(sessions)
}

func TestProxyClientManager_AddProxyClient(t *testing.T) {
	Params.Init()

	core, err := NewCore(context.Background(), nil)
	assert.Nil(t, err)
	cli, err := etcd.GetEtcdClient(&Params.EtcdCfg)
	assert.Nil(t, err)
	defer cli.Close()
	core.etcdCli = cli

	core.SetNewProxyClient(
		func(se *sessionutil.Session) (types.Proxy, error) {
			return nil, errors.New("failed")
		},
	)

	pcm := newProxyClientManager(core)

	session := &sessionutil.Session{
		ServerID: 100,
		Address:  "localhost",
	}

	pcm.AddProxyClient(session)
}

func TestProxyClientManager_InvalidateCollectionMetaCache(t *testing.T) {
	Params.Init()
	ctx := context.Background()

	core, err := NewCore(ctx, nil)
	assert.Nil(t, err)
	cli, err := etcd.GetEtcdClient(&Params.EtcdCfg)
	assert.Nil(t, err)
	defer cli.Close()
	core.etcdCli = cli

	pcm := newProxyClientManager(core)

	ch := make(chan struct{})
	pcm.helper = proxyClientManagerHelper{
		afterConnect: func() { ch <- struct{}{} },
	}

	err = pcm.InvalidateCollectionMetaCache(ctx, nil)
	assert.NoError(t, err)

	core.SetNewProxyClient(
		func(se *sessionutil.Session) (types.Proxy, error) {
			return &proxyMock{}, nil
		},
	)

	session := &sessionutil.Session{
		ServerID: 100,
		Address:  "localhost",
	}

	pcm.AddProxyClient(session)
	<-ch

	err = pcm.InvalidateCollectionMetaCache(ctx, &proxypb.InvalidateCollMetaCacheRequest{
		CollectionName: "collection0",
	})
	assert.NoError(t, err)

	// test releaseDQLMessageStream failed
	for _, v := range pcm.proxyClient {
		v.(*proxyMock).returnError = true
	}
	err = pcm.InvalidateCollectionMetaCache(ctx, nil)
	assert.Error(t, err)

	for _, v := range pcm.proxyClient {
		v.(*proxyMock).returnGrpcError = true
	}
	err = pcm.InvalidateCollectionMetaCache(ctx, nil)
	assert.Error(t, err)
}

func TestProxyClientManager_ReleaseDQLMessageStream(t *testing.T) {
	Params.Init()
	ctx := context.Background()

	core, err := NewCore(ctx, nil)
	assert.Nil(t, err)
	cli, err := etcd.GetEtcdClient(&Params.EtcdCfg)
	assert.Nil(t, err)
	defer cli.Close()
	core.etcdCli = cli

	pcm := newProxyClientManager(core)

	ch := make(chan struct{})
	pcm.helper = proxyClientManagerHelper{
		afterConnect: func() { ch <- struct{}{} },
	}

	err = pcm.ReleaseDQLMessageStream(ctx, nil)
	assert.NoError(t, err)

	core.SetNewProxyClient(
		func(se *sessionutil.Session) (types.Proxy, error) {
			return &proxyMock{}, nil
		},
	)

	session := &sessionutil.Session{
		ServerID: 100,
		Address:  "localhost",
	}

	pcm.AddProxyClient(session)
	<-ch

	err = pcm.ReleaseDQLMessageStream(ctx, nil)
	assert.NoError(t, err)

	// test releaseDQLMessageStream failed
	for _, v := range pcm.proxyClient {
		v.(*proxyMock).returnError = true
	}
	err = pcm.ReleaseDQLMessageStream(ctx, nil)
	assert.Error(t, err)

	for _, v := range pcm.proxyClient {
		v.(*proxyMock).returnGrpcError = true
	}
	err = pcm.ReleaseDQLMessageStream(ctx, nil)
	assert.Error(t, err)
}

func TestProxyClientManager_InvalidateCredentialCache(t *testing.T) {
	Params.Init()
	ctx := context.Background()

	core, err := NewCore(ctx, nil)
	assert.Nil(t, err)
	cli, err := etcd.GetEtcdClient(&Params.EtcdCfg)
	assert.Nil(t, err)
	defer cli.Close()
	core.etcdCli = cli

	pcm := newProxyClientManager(core)

	ch := make(chan struct{})
	pcm.helper = proxyClientManagerHelper{
		afterConnect: func() { ch <- struct{}{} },
	}

	err = pcm.InvalidateCredentialCache(ctx, nil)
	assert.NoError(t, err)

	core.SetNewProxyClient(
		func(se *sessionutil.Session) (types.Proxy, error) {
			return &proxyMock{}, nil
		},
	)

	session := &sessionutil.Session{
		ServerID: 100,
		Address:  "localhost",
	}

	pcm.AddProxyClient(session)
	<-ch

	err = pcm.InvalidateCredentialCache(ctx, nil)
	assert.NoError(t, err)

	// test releaseDQLMessageStream failed
	for _, v := range pcm.proxyClient {
		v.(*proxyMock).returnError = true
	}
	err = pcm.InvalidateCredentialCache(ctx, nil)
	assert.Error(t, err)

	for _, v := range pcm.proxyClient {
		v.(*proxyMock).returnGrpcError = true
	}
	err = pcm.InvalidateCredentialCache(ctx, nil)
	assert.Error(t, err)
}
