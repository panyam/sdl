package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"cloud.google.com/go/datastore"
)

type IDGen struct {
	ds         *DataStore[GenID]
	NextIDFunc func() string
}

func NewIDGen(kind string) *IDGen {
	client, err := datastore.NewClient(context.Background(), "leetcoach")
	if err != nil {
		log.Fatal(err)
	}
	ds := NewDataStore[GenID](fmt.Sprintf("IDGen_%s", kind), true)
	ds.DSClient = client
	ds.GetEntityKey = func(gid *GenID) *datastore.Key {
		if gid.Id == "" {
			return nil
		}
		return datastore.NameKey(ds.kind, gid.Id, nil)
	}
	return &IDGen{ds: ds}
}

func (i *IDGen) NextID(owner string, expiresAt time.Time) (*GenID, error) {
	id := i.NextIDFunc()
	var gid GenID
	for {
		err := i.ds.GetByID(id, &gid)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				// all good
				break
			} else {
				log.Println("Error getting id: ", err)
				return nil, err
			}
		} else {
			id = i.NextIDFunc()
		}
	}
	gid.Id = id
	gid.OwnerId = owner
	gid.ExpiresAt = expiresAt
	if _, err := i.ds.SaveEntity(&gid); err != nil {
		return nil, err
	}
	return &gid, nil
}

func SimpleIDFunc(letters []rune, maxdigits int) func() string {
	if letters == nil || len(letters) == 0 {
		letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	}
	if maxdigits <= 0 {
		maxdigits = 4
	}
	return func() string {
		return randSeq(maxdigits, letters)
	}
}

func randSeq(n int, chars []rune) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
