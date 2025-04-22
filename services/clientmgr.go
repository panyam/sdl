package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strconv"

	"cloud.google.com/go/datastore"
	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const APP_ID = "leetcoach"

var ErrNoSuchEntity = errors.New("entity not found")

type ClientMgr struct {
	svcAddr         string
	designSvcClient protos.DesignServiceClient
	tagSvcClient    protos.TagServiceClient
	authSvc         *AuthService
	tagDS           *DataStore[Tag]
	acmeDS          *DataStore[Acme]
	designDS        *DataStore[Design]
	userDS          *DataStore[User]
	identityDS      *DataStore[Identity]
	channelDS       *DataStore[Channel]
	authFlowDS      *DataStore[AuthFlow]
}

func NewClientMgr(svc_addr string) *ClientMgr {
	log.Println("Client Mgr Svc Addr: ", svc_addr)
	if svc_addr == "" {
		panic("Service Address is nil")
	}
	return &ClientMgr{svcAddr: svc_addr}
}

func (c *ClientMgr) ClientContext(ctx context.Context, loggedInUserId string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return metadata.AppendToOutgoingContext(context.Background(), "LoggedInUserId", loggedInUserId)
}

func (c *ClientMgr) GetAuthService() *AuthService {
	if c.authSvc == nil {
		c.authSvc = &AuthService{clients: c}
	}
	return c.authSvc
}

func (c *ClientMgr) GetDesignSvcClient() (out protos.DesignServiceClient, err error) {
	if c.designSvcClient == nil {
		log.Println("Addr: ", c.svcAddr)
		designSvcConn, err := grpc.NewClient(c.svcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("cannot connect with server %v", err)
			return nil, err
		}

		c.designSvcClient = protos.NewDesignServiceClient(designSvcConn)
	}
	return c.designSvcClient, nil
}

/*
func (c *ClientMgr) getDSClient() *datastore.Client {
	client, err := datastore.NewClient(context.Background(), APP_ID)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return client
}
*/

func (c *ClientMgr) GetTagDSClient() *DataStore[Tag] {
	if c.tagDS == nil {
		c.tagDS = NewDataStore[Tag]("Score", false)
		// c.tagDS.DSClient = c.getDSClient()
		c.tagDS.GetEntityKey = func(tag *Tag) *datastore.Key {
			if tag.Name == "" {
				return nil
			}
			return datastore.NameKey(c.tagDS.kind, tag.Name, nil)
		}
	}
	return c.tagDS
}

func (c *ClientMgr) GetDesignDSClient() *DataStore[Design] {
	if c.designDS == nil {
		c.designDS = NewDataStore[Design]("Score", false)
		// c.designDS.DSClient = c.getDSClient()
		c.designDS.GetEntityKey = func(design *Design) *datastore.Key {
			if design.Id == "" {
				return nil
			}
			return datastore.NameKey(c.designDS.kind, design.Id, nil)
		}
	}
	return c.designDS
}

func (c *ClientMgr) GetAuthFlowDSClient() *DataStore[AuthFlow] {
	if c.authFlowDS == nil {
		c.authFlowDS = NewDataStore[AuthFlow]("AuthFlow", true)
		// c.authFlowDS.DSClient = c.getDSClient()
		c.authFlowDS.GetEntityKey = func(authFlow *AuthFlow) *datastore.Key {
			if authFlow.Id == "" {
				return nil
			}
			return datastore.NameKey(c.authFlowDS.kind, authFlow.Id, nil)
		}
	}
	return c.authFlowDS
}

func (c *ClientMgr) GetIdentityDSClient() *DataStore[Identity] {
	if c.identityDS == nil {
		c.identityDS = NewDataStore[Identity]("Identity", false)
		// c.identityDS.DSClient = c.getDSClient()
		c.identityDS.GetEntityKey = func(identity *Identity) *datastore.Key {
			if identity.Key() == "" {
				return nil
			}
			return datastore.NameKey(c.identityDS.kind, identity.Key(), nil)
		}
	}
	return c.identityDS
}

func (c *ClientMgr) GetChannelDSClient() *DataStore[Channel] {
	if c.channelDS == nil {
		c.channelDS = NewDataStore[Channel]("Channel", false)
		// c.channelDS.DSClient = c.getDSClient()
		c.channelDS.GetEntityKey = func(channel *Channel) *datastore.Key {
			if channel.Key() == "" {
				return nil
			}
			return datastore.NameKey(c.channelDS.kind, channel.Key(), nil)
		}
	}
	return c.channelDS
}

func (c *ClientMgr) GetUserDSClient() *DataStore[User] {
	if c.userDS == nil {
		c.userDS = NewDataStore[User]("User", true)
		// c.userDS.DSClient = c.getDSClient()
		c.userDS.SetEntityKey = func(user *User, key *datastore.Key) {
			user.Id = fmt.Sprintf("%d", key.ID)
		}
		c.userDS.IDToKey = func(id string) *datastore.Key {
			if id == "" {
				return nil
			}
			uid, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				slog.Warn("Invalid user ID: ", id, nil)
			}
			return datastore.IDKey(c.userDS.kind, uid, nil)
		}
		c.userDS.GetEntityKey = func(user *User) *datastore.Key {
			return c.userDS.IDToKey(user.Id)
		}
	}
	return c.userDS
}

func (c *ClientMgr) GetAcmeDSClient() *DataStore[Acme] {
	if c.acmeDS == nil {
		c.acmeDS = NewDataStore[Acme]("Acme", false)
		// c.acmeDS.DSClient = c.getDSClient()
		c.acmeDS.GetEntityKey = func(acme *Acme) *datastore.Key {
			if acme.Id == "" {
				return nil
			}
			return datastore.NameKey(c.acmeDS.kind, acme.Id, nil)
		}
	}
	return c.acmeDS
}
