package services

import (
	"context"
	"fmt"
	"log/slog"

	"cloud.google.com/go/datastore"
)

type DataStore[T any] struct {
	DSClient      *datastore.Client
	kind          string
	autoCreateKey bool
	IndexExcludes []string
	IDToKey       func(key string) *datastore.Key
	GetEntityKey  func(entity *T) *datastore.Key
	SetEntityKey  func(entity *T, key *datastore.Key)
}

func NewDataStore[T any](kind string, autoCreateKey bool) *DataStore[T] {
	return &DataStore[T]{
		kind:          kind,
		autoCreateKey: autoCreateKey,
		IDToKey: func(id string) *datastore.Key {
			return datastore.NameKey(kind, id, nil)
		},
	}
}

func (ds *DataStore[T]) Kind() string {
	return ds.kind
}

func (ds *DataStore[T]) GetByID(id string, out *T) error {
	key := ds.IDToKey(id)
	err := ds.DSClient.Get(context.Background(), key, out)
	if err != nil {
		slog.Error("Error getting by ID: ", "kind", ds.kind, "id", id, "err", err)
		if err == datastore.ErrNoSuchEntity {
			return ErrNoSuchEntity
		}
	}
	return err
}

func (ds *DataStore[T]) DeleteByKey(key string) error {
	dbkey := datastore.NameKey(ds.kind, key, nil)
	return ds.DSClient.Delete(context.Background(), dbkey)
}

func (ds *DataStore[T]) SaveEntity(entity *T) (*datastore.Key, error) {
	ctx := context.Background()
	newKey := datastore.IncompleteKey(ds.kind, nil)
	entityKey := ds.GetEntityKey(entity)
	if entityKey == nil {
		if !ds.autoCreateKey {
			return nil, fmt.Errorf("Key cannot be autocreated for %s", ds.kind)
		} else {
			key, err := ds.DSClient.Put(ctx, newKey, entity)
			if err != nil {
				return nil, err
			}
			slog.Debug("Got Key: ", "key", key, "newKey", newKey, "err", err)
			if key.ID == 0 {
				return nil, fmt.Errorf("Key (%s) is invalid.  Save failed.", ds.kind)
			}
			newKey = key
			ds.SetEntityKey(entity, newKey)
		}
	} else if ds.autoCreateKey {
		newKey = entityKey
	} else {
		// key is already set
		newKey = datastore.NameKey(ds.kind, entityKey.Name, nil)
	}

	// Now update with the
	return ds.DSClient.Put(ctx, newKey, entity)
}

func (ds *DataStore[T]) NewQuery() *datastore.Query {
	return datastore.NewQuery(ds.kind)
}

func (ds *DataStore[T]) Select(query *datastore.Query) (out []*T, err error) {
	_, err = ds.DSClient.GetAll(context.Background(), query, &out)
	if err != nil {
		slog.Error("error selecting with query: ", "err", err)
		return nil, err
	}
	return
}

func (ds *DataStore[T]) ListEntities(offset int, count int) (out []*T, err error) {
	query := ds.NewQuery().Offset(offset).Limit(count)
	slog.Debug("Trying List Query: ", "kind", ds.kind, "query", query, "client", ds.DSClient)
	_, err = ds.DSClient.GetAll(context.Background(), query, &out)
	if err != nil {
		slog.Error("error listing entities: ", "kind", ds.kind, "err", err)
	}
	return
}

type StringMapField struct {
	Properties map[string]any
}

func (m *StringMapField) Load(ps []datastore.Property) error {
	if len(ps) > 0 {
		if m.Properties == nil {
			m.Properties = make(map[string]any)
		}
		for _, prop := range ps {
			m.Properties[prop.Name] = prop.Value
		}
	}
	return nil
}

func (m *StringMapField) Save() (out []datastore.Property, err error) {
	for key, value := range m.Properties {
		out = append(out, datastore.Property{
			Name:  key,
			Value: value,
		})
	}
	return
}
