package httpapi

import "strings"

// Permission is a "<path>:<action>" grant as defined in architecture.md.
type Permission string

const (
	PermCatalogView   Permission = "catalog:view"
	PermPlaybackWatch Permission = "playback:watch"
	PermAccountView   Permission = "account:view"
	PermSessionCreate Permission = "session:create"
	PermSessionDelete Permission = "session:delete"
)

// Role is the set of grants a request holds.
type Role string

const (
	RoleVisitor Role = "visitor"
	RoleMember  Role = "member"
)

// roleGrants is the single place roles are defined. Handlers never compare
// permission strings themselves.
var roleGrants = map[Role][]string{
	RoleVisitor: {string(PermCatalogView), string(PermSessionCreate)},
	RoleMember: {
		string(PermCatalogView), string(PermPlaybackWatch),
		string(PermAccountView), string(PermSessionCreate), string(PermSessionDelete),
	},
}

// Allows reports whether role grants required.
func (role Role) Allows(required Permission) bool {
	return AllowsGrants(roleGrants[role], string(required))
}

// AllowsGrants implements the permission rule: a grant allows a required
// permission when its path equals the required path or is a dot-prefix of it,
// and its action is the same action or "all". There are no deny rules.
func AllowsGrants(grants []string, required string) bool {
	reqPath, reqAction, ok := splitPermission(required)
	if !ok {
		return false
	}
	for _, grant := range grants {
		grantPath, grantAction, ok := splitPermission(grant)
		if !ok {
			continue
		}
		if !pathCovers(grantPath, reqPath) {
			continue
		}
		if grantAction == reqAction || grantAction == "all" {
			return true
		}
	}
	return false
}

func splitPermission(value string) (path, action string, ok bool) {
	idx := strings.LastIndex(value, ":")
	if idx <= 0 || idx == len(value)-1 {
		return "", "", false
	}
	return value[:idx], value[idx+1:], true
}

// pathCovers reports whether grantPath is reqPath or a dot-prefix of it.
// "x.a" covers "x.a" and "x.a.b" but never "x.ab".
func pathCovers(grantPath, reqPath string) bool {
	return grantPath == reqPath || strings.HasPrefix(reqPath, grantPath+".")
}
