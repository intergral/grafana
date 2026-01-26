package clients

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/mail"
	"strconv"

	"go.opentelemetry.io/otel/trace"

	claims "github.com/grafana/authlib/types"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/authn"
	"github.com/grafana/grafana/pkg/services/login"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/services/user"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util"
)

var _ authn.ProxyClient = new(Grafana)
var _ authn.PasswordClient = new(Grafana)

func ProvideGrafana(cfg *setting.Cfg, userService user.Service, orgService org.Service, tracer trace.Tracer) *Grafana {
	return &Grafana{cfg, userService, orgService, tracer, log.New("authn.grafana")}
}

type Grafana struct {
	cfg         *setting.Cfg
	userService user.Service
	orgService  org.Service
	tracer      trace.Tracer
	log         log.Logger
}

func (c *Grafana) String() string {
	return "grafana"
}

func (c *Grafana) AuthenticateProxy(ctx context.Context, r *authn.Request, username string, additional map[string]string) (*authn.Identity, error) {
	ctx, span := c.tracer.Start(ctx, "authn.grafana.AuthenticateProxy") //nolint:ineffassign,staticcheck
	defer span.End()

	identity := &authn.Identity{
		AuthenticatedBy: login.AuthProxyAuthModule,
		AuthID:          username,
		ClientParams: authn.ClientParams{
			SyncUser:        true,
			SyncTeams:       true,
			FetchSyncedUser: true,
			SyncOrgRoles:    true,
			SyncPermissions: true,
			AllowSignUp:     c.cfg.AuthProxy.AutoSignUp,
		},
	}

	switch c.cfg.AuthProxy.HeaderProperty {
	case "username":
		identity.Login = username
		addr, err := mail.ParseAddress(username)
		if err == nil {
			identity.Email = addr.Address
		}
	case "email":
		identity.Login = username
		identity.Email = username
	default:
		return nil, errInvalidProxyHeader.Errorf("invalid auth proxy header property, expected username or email but got: %s", c.cfg.AuthProxy.HeaderProperty)
	}

	if v, ok := additional[proxyFieldName]; ok {
		identity.Name = v
	}

	if v, ok := additional[proxyFieldEmail]; ok {
		identity.Email = v
	}

	if v, ok := additional[proxyFieldLogin]; ok {
		identity.Login = v
	}

	if v, ok := additional[proxyFieldOrgName]; ok {
		identity.OrgName = v
		orgByName, err := c.orgService.GetByName(ctx, &org.GetOrgByNameQuery{Name: v})
		if err != nil {
			return nil, fmt.Errorf("failed to get org by name: %w", err)
		}
		identity.OrgID = orgByName.ID
	}

	if v, ok := additional[proxyFieldRole]; ok {
		orgRoles, isGrafanaAdmin, _ := getRoles(c.cfg, func() (org.RoleType, *bool, error) {
			return org.RoleType(v), nil, nil
		})
		identity.OrgRoles = orgRoles
		identity.IsGrafanaAdmin = isGrafanaAdmin
	}

	if v, ok := additional[proxyFieldGroups]; ok {
		identity.Groups = util.SplitString(v)
	}

	identity.ClientParams.LookUpParams.Email = &identity.Email
	identity.ClientParams.LookUpParams.Login = &identity.Login

	// Log complete auth proxy identity for debugging
	c.log.FromContext(ctx).Info("Auth proxy authentication completed",
		"username", username,
		"login", identity.Login,
		"email", identity.Email,
		"name", identity.Name,
		"orgName", identity.OrgName,
		"orgID", identity.OrgID,
		"authID", identity.AuthID,
		"orgRoles", identity.OrgRoles,
		"isGrafanaAdmin", identity.IsGrafanaAdmin,
		"groups", identity.Groups,
		"additionalHeaders", additional,
		"remoteAddr", r.HTTPRequest.RemoteAddr,
	)

	return identity, nil
}

func (c *Grafana) AuthenticatePassword(ctx context.Context, r *authn.Request, username, password string) (*authn.Identity, error) {
	ctx, span := c.tracer.Start(ctx, "authn.grafana.AuthenticatePassword")
	defer span.End()

	usr, err := c.userService.GetByLogin(ctx, &user.GetUserByLoginQuery{LoginOrEmail: username})
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, errIdentityNotFound.Errorf("no user found: %w", err)
		}
		return nil, err
	}

	// user was found so set auth module in req metadata
	r.SetMeta(authn.MetaKeyAuthModule, "grafana")

	if ok := comparePassword(password, usr.Salt, string(usr.Password)); !ok {
		return nil, errInvalidPassword.Errorf("invalid password")
	}

	return &authn.Identity{
		ID:              strconv.FormatInt(usr.ID, 10),
		Type:            claims.TypeUser,
		OrgID:           r.OrgID,
		ClientParams:    authn.ClientParams{FetchSyncedUser: true, SyncPermissions: true},
		AuthenticatedBy: login.PasswordAuthModule,
	}, nil
}

func comparePassword(password, salt, hash string) bool {
	// It is ok to ignore the error here because util.EncodePassword can never return a error
	hashedPassword, _ := util.EncodePassword(password, salt)
	return subtle.ConstantTimeCompare([]byte(hashedPassword), []byte(hash)) == 1
}
