package ngalert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/grafana/grafana/pkg/services/ngalert/models"
	"github.com/grafana/grafana/pkg/services/org"
)

// Intergral: guards the fork's grant of protected-field write to Editors.
//
// Upstream gives alert.notifications.receivers.protected:write to fixed:alerting:admin
// only, which on OSS means nobody — we never make a customer an org Admin, so an Editor
// could not change the destination URL of a contact point it did not create. We add the
// action to receiversWriterRole, which reaches Editors by composition:
//
//	receiversWriterRole -> notificationsWriterRole -> alertingWriterRole -> Grants: Editor
//
// That chain is three ConcatPermissions hops away from the grant, so a rebase that
// restructures any of them drops the permission with nothing failing. This test is the
// tripwire: it asserts the flattened, composed result rather than the literal we edited.
func TestIntergralEditorsCanWriteProtectedReceiverFields(t *testing.T) {
	protectedWrite := accesscontrol.Permission{
		Action: accesscontrol.ActionAlertingReceiversUpdateProtected,
		Scope:  models.ScopeReceiversAll,
	}

	t.Run("receiversWriterRole carries the action", func(t *testing.T) {
		assert.Contains(t, receiversWriterRole.Role.Permissions, protectedWrite,
			"the fork's grant was dropped from receiversWriterRole")
	})

	t.Run("composition carries it through to alertingWriterRole", func(t *testing.T) {
		assert.Contains(t, alertingWriterRole.Role.Permissions, protectedWrite,
			"receiversWriterRole no longer composes into alertingWriterRole")
	})

	t.Run("alertingWriterRole is granted to Editor", func(t *testing.T) {
		require.Contains(t, alertingWriterRole.Grants, string(org.RoleEditor),
			"alertingWriterRole is no longer the role Editors receive")
	})

	// The counterpart to the above: upstream's separation of Editor from Admin is
	// otherwise intact. If one of these starts failing, the fork has widened beyond
	// the one action it meant to add.
	t.Run("Editors gain nothing else that belongs to Admin", func(t *testing.T) {
		for _, action := range []string{
			accesscontrol.ActionAlertingReceiversReadSecrets,
			accesscontrol.ActionAlertingReceiversPermissionsRead,
			accesscontrol.ActionAlertingReceiversPermissionsWrite,
		} {
			for _, p := range alertingWriterRole.Role.Permissions {
				assert.NotEqual(t, action, p.Action,
					"alertingWriterRole should not carry %s — that stays with Admin", action)
			}
		}
	})

	// Viewers must not pick the action up by any route: alertingReaderRole is what
	// org.RoleViewer receives, and viewers_can_edit (set in our deployment) only
	// touches the data sources explorer role, never alerting.
	t.Run("Viewers do not gain it", func(t *testing.T) {
		require.Contains(t, alertingReaderRole.Grants, string(org.RoleViewer))
		assert.NotContains(t, alertingReaderRole.Role.Permissions, protectedWrite)
	})
}
