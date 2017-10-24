/*
Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package etcdtopo

import (
	"encoding/json"
	"fmt"

	"github.com/coreos/go-etcd/etcd"
	"github.com/youtube/vitess/go/vt/topo"
	"golang.org/x/net/context"

	topodatapb "github.com/youtube/vitess/go/vt/proto/topodata"
)

// CreateKeyspace implements topo.Server.
func (s *Server) CreateKeyspace(ctx context.Context, keyspace string, value *topodatapb.Keyspace) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	global := s.getGlobal()

	if _, err = global.Create(keyspaceFilePath(keyspace), string(data), 0 /* ttl */); err != nil {
		return convertError(err)
	}
	return nil
}

// UpdateKeyspace implements topo.Server.
func (s *Server) UpdateKeyspace(ctx context.Context, keyspace string, value *topodatapb.Keyspace, existingVersion int64) (int64, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return -1, err
	}

	var resp *etcd.Response
	if existingVersion == -1 {
		// Set unconditionally.
		resp, err = s.getGlobal().Set(keyspaceFilePath(keyspace), string(data), 0 /* ttl */)
	} else {
		resp, err = s.getGlobal().CompareAndSwap(keyspaceFilePath(keyspace),
			string(data), 0 /* ttl */, "" /* prevValue */, uint64(existingVersion))
	}
	if err != nil {
		return -1, convertError(err)
	}
	if resp.Node == nil {
		return -1, ErrBadResponse
	}
	return int64(resp.Node.ModifiedIndex), nil
}

// GetKeyspace implements topo.Server.
func (s *Server) GetKeyspace(ctx context.Context, keyspace string) (*topodatapb.Keyspace, int64, error) {
	resp, err := s.getGlobal().Get(keyspaceFilePath(keyspace), false /* sort */, false /* recursive */)
	if err != nil {
		return nil, 0, convertError(err)
	}
	if resp.Node == nil {
		return nil, 0, ErrBadResponse
	}

	value := &topodatapb.Keyspace{}
	if err := json.Unmarshal([]byte(resp.Node.Value), value); err != nil {
		return nil, 0, fmt.Errorf("bad keyspace data (%v): %q", err, resp.Node.Value)
	}

	return value, int64(resp.Node.ModifiedIndex), nil
}

// GetKeyspaces implements topo.Server.
func (s *Server) GetKeyspaces(ctx context.Context) ([]string, error) {
	resp, err := s.getGlobal().Get(keyspacesDirPath, true /* sort */, false /* recursive */)
	if err != nil {
		err = convertError(err)
		if err == topo.ErrNoNode {
			return nil, nil
		}
		return nil, err
	}
	return getNodeNames(resp)
}

// DeleteKeyspace implements topo.Server.
func (s *Server) DeleteKeyspace(ctx context.Context, keyspace string) error {
	_, err := s.getGlobal().Delete(keyspaceDirPath(keyspace), true /* recursive */)
	if err != nil {
		return convertError(err)
	}
	return nil
}
