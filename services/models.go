package services

import (
	"fmt"
	"strings"
	"time"
)

type BaseModel struct {
	CreatedAt time.Time `datastore:"createdAt"`
	UpdatedAt time.Time `datastore:"updatedAt"`
	Version   int       `datastore:"version"` // used for optimistic locking
}

type Acme struct {
	Id       string `datastore:"id"`
	Response string `datastore:"response"`
}

type GenID struct {
	BaseModel
	Id        string `datastore:"id"`
	OwnerId   string
	ExpiresAt time.Time
}

type Tag struct {
	BaseModel
	Name           string
	NormalizedName string
	Description    string
	ImageUrl       string
	FirstUserId    string `datastore:"firstUser"`
	NumDesigns     int64
}

type Design struct {
	BaseModel
	Id         string   `datastore:"id"`
	OwnerId    string   `datastore:"userId"`
	Visibility string   `datastore:"visibility"`
	VisibleTo  []string `datastore:"visibleTo"`

	/**
	 * Name of the design
	 */
	Name string `datastore:"name"`

	/**
	 * Description of this design.
	 */
	Description string `datastore:"description"`

	/**
	 * IDs of section in this design.
	 */
	SectionIds []string `datastore:"sectionIds"`

	// Metadata about the content itself that user may want to
	// highlight (or the system extracts)
	ContentMetadata StringMapField `datastore:"contentMetadata,noindex"`
}

/**
 * An identify is a unique global "address" corresponding to a user.
 * For example the identify abc@example.com is a unique identify regardless
 * of which Channel is verifying it.  Multiple channels can verify the same
 * entity, eg open auth by github, FB or Google can verify the same email
 * address.
 */
type Identity struct {
	BaseModel
	IsActive bool `datastore:"isActive"`

	// Type of identity being verified (eg email, phone etc).
	IdentityType string `datastore:"identityType"`

	// The key specific to the identity (eg an email address or a phone number etc).
	//
	// type + key should be unique through out the system.
	IdentityKey string `datastore:"identityKey"`

	// The primary user that this identity can be associated with.
	// Identities do not need to be explicitly associted with a user especially
	// in systems where a single Identity can be used to front several users
	PrimaryUser string `datastore:"primaryUser"`
}

func (i *Identity) HasUser() bool {
	return strings.TrimSpace(i.PrimaryUser) != ""
}

func (i *Identity) HasKey() bool {
	return strings.TrimSpace(i.IdentityType) != "" && strings.TrimSpace(i.IdentityKey) != ""
}

func (i *Identity) Key() string {
	out := fmt.Sprintf("%s:%s", i.IdentityType, i.IdentityKey)
	if out == ":" {
		out = ""
	}
	return out
}

/**
 * Once a channel has verified an Identity, the end result is a mapping to
 * a local user object that is the entry for authenticated actions within
 * the system.  The User can also mean a user profile and can be extended
 * to be customized by the user of this library in their own specific app.
 */
type User struct {
	BaseModel

	Id string `datastore:"id"`

	IsActive bool `datastore:"isActive"`
	// A globally unique user ID.  This User ID cannot be used as a login key.
	// Login's need to happen via the Identiites above and a username could be
	// one of the identities (which can be verified say via login/password mechanism)
	// Alternatively an email can be used as an identity that can also map to
	// a particular user.
	// Name    string `datastore:"name"`
	// Email   string `datastore:"email"`
	// Picture string `datastore:"picture"`
	Profile StringMapField `datastore:"profile"`
}

type AuthFlow struct {
	BaseModel

	Id string `datastore:"id"`

	// Kind of login being done
	Provider string `datastore:"provider"`

	// When this Auth session expires;
	ExpiresIn time.Time `datastore:"expiresIn"`

	// Call back URL for where the session needs to endup on success
	// callback: CallbackRequest;

	// Handler that will continue the flow after a successful AuthFlow.
	HandlerName string `datastore:"handlerName"`

	// Parameters for the handler to continue with.
	HandlerParams StringMapField `datastore:"handlerParams"`
}

/**
 * Channel's represented federated verification objects.  For example a Google
 * Signin would ensure that the user that goes through this flow will end up with
 * a Google signin Channel - which would verify a particular identity type.
 */
type Channel struct {
	BaseModel
	Provider string `datastore:"provider"`
	LoginId  string `datastore:"loginId"`

	/**
	 * Credentials for this channel (like access tokens, passwords etc).
	 */
	Credentials StringMapField `datastore:"credentials"`

	/**
	 * Profile as passed by the provider of the channel.
	 */
	Profile StringMapField `datastore:"profile"`

	/**
	 * When does this channel expire and needs another login/auth.
	 */
	ExpiresIn time.Time `datastore:"expiresIn"`

	// The identity that this channel is verifying.
	IdentityKey string `datastore:"identityKey"`
}

func (c *Channel) HasKey() bool {
	return strings.TrimSpace(c.Provider) != "" && strings.TrimSpace(c.LoginId) != ""
}

func (c *Channel) Key() string {
	return c.Provider + ":" + c.LoginId
}

func (c *Channel) HasIdentity() bool {
	return strings.TrimSpace(c.IdentityKey) != ""
}
