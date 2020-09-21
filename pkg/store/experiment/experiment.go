// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package experiment

import (
	"context"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/store/dbstore"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("experimentStore")

// NewStore returns a new ExperimentStore.
func NewStore(db *dbstore.DB) core.ExperimentStore {
	db.AutoMigrate(&core.ArchiveExperiment{})

	es := &experimentStore{db}

	if err := es.DeleteIncompleteExperiments(context.Background()); err != nil && !gorm.IsRecordNotFoundError(err) {
		log.Error(err, "failed to delete all incomplete experiments")
	}

	return es
}

type experimentStore struct {
	db *dbstore.DB
}

// DeleteIncompleteExperiments implement core.ExperimentStore interface.
func (e *experimentStore) DeleteIncompleteExperiments(_ context.Context) error {
	return e.db.Where("finish_time IS NULL").Unscoped().
		Delete(core.Event{}).Error
}

func (e *experimentStore) List(_ context.Context, kind, ns, name string) ([]*core.ArchiveExperiment, error) {
	archives := make([]*core.ArchiveExperiment, 0)
	query, args := constructQueryArgs(kind, ns, name, "")

	db := e.db.Model(core.ArchiveExperiment{})
	if len(args) > 0 {
		db = db.Where(query, args)
	}

	if err := db.Where("archived = ?", true).Find(&archives).Error; err != nil &&
		!gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return archives, nil
}

func (e *experimentStore) ListMeta(_ context.Context, kind, ns, name string) ([]*core.ArchiveExperimentMeta, error) {
	archives := make([]*core.ArchiveExperimentMeta, 0)
	query, args := constructQueryArgs(kind, ns, name, "")

	db := e.db.Table("archive_experiments")
	if len(args) > 0 {
		db = db.Where(query, args)
	}

	if err := db.Where("archived = ?", true).
		Find(&archives).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return archives, nil
}

func (e *experimentStore) Archive(_ context.Context, ns, name string) error {
	if err := e.db.Model(core.ArchiveExperiment{}).
		Where("namespace = ? AND name = ? AND archived = ?", ns, name, false).
		Updates(map[string]interface{}{"archived": true, "finish_time": time.Now()}).Error; err != nil &&
		!gorm.IsRecordNotFoundError(err) {
		return err
	}
	return nil
}

func (e *experimentStore) Set(_ context.Context, archive *core.ArchiveExperiment) error {
	return e.db.Model(core.ArchiveExperiment{}).Save(archive).Error
}

func (e *experimentStore) getUID(_ context.Context, kind, ns, name string) (string, error) {
	archives := make([]*core.ArchiveExperimentMeta, 0)

	if err := e.db.Table("archive_experiments").Where(
		"namespace = ? and name = ? and kind = ?", ns, name, kind).
		Find(&archives).Error; err != nil {
		return "", err
	}

	if len(archives) == 0 {
		return "", fmt.Errorf("get UID failure")
	}

	UID := archives[0].UID
	st := archives[0].StartTime

	for _, archive := range archives {
		if st.Before(archive.StartTime) {
			st = archive.StartTime
			UID = archive.UID
		}
	}
	return UID, nil
}

// DetailList returns a list of archive experiments from the datastore.
func (e *experimentStore) DetailList(ctx context.Context, kind, namespace, name, uid string) ([]*core.ArchiveExperiment, error) {
	if kind != "" && namespace != "" && name != "" && uid == "" {
		var err error
		uid, err = e.getUID(context.TODO(), kind, namespace, name)
		if err != nil {
			return nil, err
		}
	}

	archives := make([]*core.ArchiveExperiment, 0)
	query, args := constructQueryArgs(kind, namespace, name, uid)

	if len(args) == 0 {
		if err := e.db.Table("archive_experiments").Find(&archives).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
	} else {
		if err := e.db.Table("archive_experiments").Where(query, args).
			Find(&archives).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
	}

	return archives, nil
}

// TODO: implement the left core.EventStore interfaces
func (e *experimentStore) Find(context.Context, int64) (*core.ArchiveExperiment, error) {
	return nil, nil
}

func (e *experimentStore) Delete(context.Context, *core.ArchiveExperiment) error { return nil }

// DeleteByFinishTime deletes experiments whose time difference is greater than the given time from FinishTime.
func (e *experimentStore) DeleteByFinishTime(_ context.Context, ttl time.Duration) error {
	expList, err := e.List(context.Background(), "", "", "")
	if err != nil {
		return err
	}
	nowTime := time.Now()
	for _, exp := range expList {
		if exp.Archived == false {
			continue
		}
		if exp.FinishTime.Add(ttl).Before(nowTime) {
			if err := e.db.Table("archive_experiments").Unscoped().Delete(*exp).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *experimentStore) FindByUID(_ context.Context, uid string) (*core.ArchiveExperiment, error) {
	archive := new(core.ArchiveExperiment)

	if err := e.db.Where(
		"uid = ?", uid).
		First(archive).Error; err != nil {
		return nil, err
	}

	return archive, nil
}

// FindMetaByUID returns an archive experiment by UID.
func (e *experimentStore) FindMetaByUID(_ context.Context, uid string) (*core.ArchiveExperimentMeta, error) {
	archive := new(core.ArchiveExperimentMeta)

	if err := e.db.Table("archive_experiments").Where(
		"uid = ?", uid).
		First(archive).Error; err != nil {
		return nil, err
	}

	return archive, nil
}

func constructQueryArgs(kind, ns, name, uid string) (string, []string) {
	args := make([]string, 0)
	query := ""
	if kind != "" {
		query += "kind = ?"
		args = append(args, kind)
	}
	if ns != "" {
		if len(args) > 0 {
			query += " AND namespace = ?"
		} else {
			query += "namespace = ?"
		}
		args = append(args, ns)
	}

	if name != "" {
		if len(args) > 0 {
			query += " AND name = ?"
		} else {
			query += "name = ?"
		}
		args = append(args, name)
	}

	if uid != "" {
		if len(args) > 0 {
			query += " AND uid = ?"
		} else {
			query += "uid = ?"
		}
		args = append(args, uid)
	}
	return query, args
}
