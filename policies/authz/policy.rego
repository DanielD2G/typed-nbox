# METADATA
# title: NBOX Authorization Policy
# description: Role-based access control with security validations
package authz

import rego.v1

default allow := false

default action_allowed := false

default roles := {"anonymous"}

roles := input.payload.roles if {
	count(input.payload.roles) > 0
}

# ==============================================================================
# SECURITY HELPERS
# Prevent common attack vectors
# ==============================================================================

# Check for path traversal attempts
has_path_traversal if {
	contains(input.action, "..")
}

# Check for dangerous characters that could indicate SQL injection or other attacks
has_dangerous_chars if {
	contains(input.action, "'")
}

has_dangerous_chars if {
	contains(input.action, "\"")
}

has_dangerous_chars if {
	contains(input.action, ";")
}

has_dangerous_chars if {
	contains(input.action, "--")
}

has_dangerous_chars if {
	contains(input.action, "/*")
}

has_dangerous_chars if {
	contains(input.action, "*/")
}

# Security check: reject if dangerous patterns detected
is_safe_request if {
	not has_path_traversal
	not has_dangerous_chars
}

# ==============================================================================
# AUTHORIZATION RULES
# ==============================================================================

# whitelist - public endpoints that don't require authentication
allow if {
	some action in data.whitelist
	regex.match(action, input.action)
}

# Main authorization rule - requires security checks
allow if {
	action_allowed
	is_safe_request
}

action_allowed if {
	some role in roles
	some permission in data.roles[role].permissions
	some path in data.permissions[permission].patterns
	regex.match(path, input.action)
}
