package descope

import (
	"strings"

	"github.com/descope/go-sdk/descope/logger"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/exp/maps"
)

// TOTPResponse - returns all relevant data to complete a TOTP registration
// One can select which method of registration to use for handshaking with an Authenticator app
type TOTPResponse struct {
	ProvisioningURL string `json:"provisioningURL,omitempty"`
	Image           string `json:"image,omitempty"`
	Key             string `json:"key,omitempty"`
}

type AuthenticationInfo struct {
	SessionToken *Token        `json:"token,omitempty"`
	RefreshToken *Token        `json:"refreshToken,omitempty"`
	User         *UserResponse `json:"user,omitempty"`
	FirstSeen    bool          `json:"firstSeen,omitempty"`
}

type WebAuthnTransactionResponse struct {
	TransactionID string `json:"transactionId,omitempty"`
	Options       string `json:"options,omitempty"`
	Create        bool   `json:"create,omitempty"`
}

type WebAuthnFinishRequest struct {
	TransactionID string `json:"transactionID,omitempty"`
	Response      string `json:"response,omitempty"`
}

type AuthFactor string

const AuthFactorUnknown AuthFactor = ""
const AuthFactorEmail AuthFactor = "email"
const AuthFactorPhone AuthFactor = "sms"
const AuthFactorSaml AuthFactor = "fed"
const AuthFactorOAuth AuthFactor = "oauth"
const AuthFactorWebauthn AuthFactor = "webauthn"
const AuthFactorTOTP AuthFactor = "totp"
const AuthFactorMFA AuthFactor = "mfa"

type Token struct {
	RefreshExpiration int64                  `json:"refreshExpiration,omitempty"`
	Expiration        int64                  `json:"expiration,omitempty"`
	JWT               string                 `json:"jwt,omitempty"`
	ID                string                 `json:"id,omitempty"`
	ProjectID         string                 `json:"projectId,omitempty"`
	Claims            map[string]interface{} `json:"claims,omitempty"`
}

func (to *Token) GetTenants() []string {
	tenants := to.getTenants()
	return maps.Keys(tenants)
}

func (to *Token) GetTenantValue(tenant, key string) any {
	tenants := to.getTenants()
	if info, ok := tenants[tenant].(map[string]any); ok {
		return info[key]
	}
	return nil
}

func (to *Token) getTenants() map[string]any {
	if to.Claims != nil {
		if tenants, ok := to.Claims[ClaimAuthorizedTenants].(map[string]any); ok {
			return tenants
		}
	}
	return make(map[string]any)
}

func (to *Token) CustomClaim(value string) interface{} {
	if to.Claims != nil {
		return to.Claims[value]
	}
	return nil
}

func (to *Token) AuthFactors() []AuthFactor {
	if to.Claims == nil {
		return nil
	}
	var afs []AuthFactor
	factors, ok := to.Claims["amr"]
	if ok {
		factorsArr, ok := factors.([]interface{})
		if ok {
			for i := range factorsArr {
				af, ok := factorsArr[i].(string)
				if ok {
					afs = append(afs, AuthFactor(af))
				} else {
					logger.LogInfo("Unknown auth-factor type [%T]", factorsArr[i]) //notest
				}
			}
		} else {
			logger.LogInfo("Unknown amr value type [%T]", factors) //notest
		}
	}
	// cases of no factors are not interesting, so not going to log them
	return afs
}

func (to *Token) IsMFA() bool {
	return len(to.AuthFactors()) > 1
}

type LoginOptions struct {
	Stepup       bool                   `json:"stepup,omitempty"`
	MFA          bool                   `json:"mfa,omitempty"`
	CustomClaims map[string]interface{} `json:"customClaims,omitempty"`
}

func (lo *LoginOptions) IsJWTRequired() bool {
	return lo != nil && (lo.Stepup || lo.MFA)
}

type JWTResponse struct {
	SessionJwt       string        `json:"sessionJwt,omitempty"`
	RefreshJwt       string        `json:"refreshJwt,omitempty"`
	CookieDomain     string        `json:"cookieDomain,omitempty"`
	CookiePath       string        `json:"cookiePath,omitempty"`
	CookieMaxAge     int32         `json:"cookieMaxAge,omitempty"`
	CookieExpiration int32         `json:"cookieExpiration,omitempty"`
	User             *UserResponse `json:"user,omitempty"`
	FirstSeen        bool          `json:"firstSeen,omitempty"`
}

type EnchantedLinkResponse struct {
	PendingRef string `json:"pendingRef,omitempty"` // Pending referral code used to poll enchanted link authentication status
	LinkID     string `json:"linkId,omitempty"`     // Link id, on which link the user should click
}

func NewAuthenticationInfo(jRes *JWTResponse, sessionToken, refreshToken *Token) *AuthenticationInfo {
	if jRes == nil {
		jRes = &JWTResponse{}
	}

	if sessionToken == nil || refreshToken == nil {
		logger.LogDebug("Building new authentication info object with empty sessionToken(%t)/refreshToken(%t)", sessionToken == nil, refreshToken == nil)
	}

	return &AuthenticationInfo{
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		User:         jRes.User,
		FirstSeen:    jRes.FirstSeen,
	}
}

func NewToken(JWT string, token jwt.Token) *Token {
	if token == nil {
		return nil
	}

	parts := strings.Split(token.Issuer(), "/")
	projectID := parts[len(parts)-1]

	return &Token{
		JWT:        JWT,
		ID:         token.Subject(),
		ProjectID:  projectID,
		Expiration: token.Expiration().Unix(),
		Claims:     token.PrivateClaims(),
	}
}

type User struct {
	Name  string `json:"name,omitempty"`
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
}

type WebauthnUserRequest struct {
	LoginID string `json:"loginId,omitempty"`
	Name    string `json:"name,omitempty"`
	Icon    string `json:"icon,omitempty"`
}

type UserResponse struct {
	User          `json:",inline"`
	UserID        string              `json:"userId,omitempty"`
	LoginIDs      []string            `json:"loginIds,omitempty"`
	VerifiedEmail bool                `json:"verifiedEmail,omitempty"`
	VerifiedPhone bool                `json:"verifiedPhone,omitempty"`
	RoleNames     []string            `json:"roleNames,omitempty"`
	UserTenants   []*AssociatedTenant `json:"userTenants,omitempty"`
	Status        string              `json:"status,omitempty"`
	Picture       string              `json:"picture,omitempty"`
}

type AccessKeyResponse struct {
	ID          string              `json:"id,omitempty"`
	Name        string              `json:"name,omitempty"`
	RoleNames   []string            `json:"roleNames,omitempty"`
	KeyTenants  []*AssociatedTenant `json:"keyTenants,omitempty"`
	Status      string              `json:"status,omitempty"`
	CreatedTime int32               `json:"createdTime,omitempty"`
	ExpireTime  int32               `json:"expireTime,omitempty"`
	CreatedBy   string              `json:"createdBy,omitempty"`
}

// Represents a tenant association for a User or an Access Key. The tenant ID is required
// to denote which tenant the user / access key belongs to. Roles is an optional list of
// roles for the user / access key in this specific tenant.
type AssociatedTenant struct {
	TenantID string   `json:"tenantId"`
	Roles    []string `json:"roleNames,omitempty"`
}

// Represents a mapping between a set of groups of users and a role that will be assigned to them.
type RoleMapping struct {
	Groups []string
	Role   string
}

// Represents a mapping between Descope and IDP user attributes
type AttributeMapping struct {
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Group       string `json:"group,omitempty"`
}

type Tenant struct {
	ID                      string   `json:"id"`
	Name                    string   `json:"name"`
	SelfProvisioningDomains []string `json:"selfProvisioningDomains"`
}

type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type Role struct {
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	PermissionNames []string `json:"permissionNames,omitempty"`
}

type GroupMember struct {
	LoginID string `json:"loginID,omitempty"`
	UserID  string `json:"userId,omitempty"`
	Display string `json:"display,omitempty"`
}

type Group struct {
	ID      string        `json:"id"`
	Display string        `json:"display,omitempty"`
	Members []GroupMember `json:"members,omitempty"`
}

type DeliveryMethod string

type OAuthProvider string

type ContextKey string

const (
	MethodWhatsApp DeliveryMethod = "whatsapp"
	MethodSMS      DeliveryMethod = "sms"
	MethodEmail    DeliveryMethod = "email"

	OAuthFacebook  OAuthProvider = "facebook"
	OAuthGithub    OAuthProvider = "github"
	OAuthGoogle    OAuthProvider = "google"
	OAuthMicrosoft OAuthProvider = "microsoft"
	OAuthGitlab    OAuthProvider = "gitlab"
	OAuthApple     OAuthProvider = "apple"

	SessionCookieName = "DS"
	RefreshCookieName = "DSR"

	RedirectLocationCookieName = "Location"

	ContextUserIDProperty               = "DESCOPE_USER_ID"
	ContextUserIDPropertyKey ContextKey = ContextUserIDProperty
	ClaimAuthorizedTenants              = "tenants"

	EnvironmentVariableProjectID     = "DESCOPE_PROJECT_ID"
	EnvironmentVariablePublicKey     = "DESCOPE_PUBLIC_KEY"
	EnvironmentVariableManagementKey = "DESCOPE_MANAGEMENT_KEY"
)
