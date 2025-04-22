package services

import (
	"log"
	"math/rand"
	"time"
)

type IDGen struct {
	MaxRetries int
	SaveID     func(*GenID) error
	GetID      func(kind, id string) (*GenID, error)
	NextIDFunc func(kind string) string
}

func (i *IDGen) NextID(kind string, owner string, expiresAt time.Time) (genid *GenID, err error) {
	id := i.NextIDFunc(kind)
	// Only check for ID collissions if a getter is provided.
	// if a getter is provided then the NextIDFunc is guaranteeing that IDs will be collision free (eg based on
	// monotonically increasing timestamp etc)
	if i.GetID != nil {
		for {
			genid, err = i.GetID(kind, id)
			if err != nil {
				log.Println("Error getting id: ", err)
				return nil, err
			} else if genid == nil {
				// ID does not exist so we are good
				break
			} else {
				// try again
				id = i.NextIDFunc(kind)
			}
		}
	}
	genid = &GenID{
		Id:        id,
		OwnerId:   owner,
		ExpiresAt: expiresAt,
	}

	// Only save ID if these are not persistent
	if i.SaveID != nil {
		err = i.SaveID(genid)
	}
	return
}

type SimpleIDGen struct {
	Letters    []rune
	MaxDigits  int
	RandSource *rand.Rand
}

func (s *SimpleIDGen) NextID(kind string) string {
	if s.Letters == nil || len(s.Letters) == 0 {
		s.Letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	}
	if s.MaxDigits <= 0 {
		s.MaxDigits = 4
	}
	if s.RandSource == nil {
		s.RandSource = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return s.randSeq()
}

func (s *SimpleIDGen) randSeq() string {
	b := make([]rune, s.MaxDigits)
	for i := range b {
		b[i] = s.Letters[s.RandSource.Intn(len(s.Letters))] // Use s.idChars and s.randSource
	}
	return string(b)
}
