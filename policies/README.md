# OPA Authorization Policy - Test Results and Documentation

## Executive Summary

**Status**: ✅ All tests passing
**Total Tests**: 58/58 (100%)
**Code Coverage**: 96.64%
**Security Status**: Enhanced with path traversal and injection protection

---

## Role Permissions Matrix

| Role | Read Non-Prod | Read Prod | Write | Delete | Secrets | Templates |
|------|--------------|-----------|-------|---------|---------|-----------|
| anonymous | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| viewer | ✅ | ❌ | ❌ | ❌ | ❌ | Non-Prod Only |
| viewer_prod | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| editor | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ |
| secrets_reader | - | - | ❌ | ❌ | ✅ | ❌ |
| maintainer | - | - | ❌ | ✅ | ❌ | ❌ |
| cicd | ✅ | ✅ | ❌ | ❌ | ❌ | Read/Build Only |
| admin | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

**Note**: Roles can be combined to grant multiple permissions.

---

## Example Use Cases

### Use Case 1: Developer in Development Environment
**Roles**: `["viewer", "secrets_reader"]`

**Allowed Actions**:
- Read entries from development, QA, global
- Read plain secret values
- View tracking history
- Read non-production templates

**Blocked Actions**:
- Read production entries
- Write or delete entries
- Write templates

### Use Case 2: DevOps Engineer
**Roles**: `["viewer_prod", "cicd"]`

**Allowed Actions**:
- Read all entries (including production)
- Build templates for deployment
- Export configurations
- Check template existence (HEAD)

**Blocked Actions**:
- Write entries or templates
- Delete entries
- Read plain secret values (needs secrets_reader)

### Use Case 3: Infrastructure Manager
**Roles**: `["editor", "maintainer", "secrets_reader"]`

**Allowed Actions**:
- Full read access (all environments)
- Write entries and templates
- Delete entries
- Read plain secret values
- Build templates

**Blocked Actions**:
- None (has comprehensive access except admin-only endpoints)

### Use Case 4: CI/CD Pipeline
**Roles**: `["cicd"]`

**Allowed Actions**:
- Read templates and entries
- Build templates with variables
- Check if templates exist (HEAD)
- Read entry history

**Blocked Actions**:
- Write or delete operations
- Read plain secret values

### Use Case 5: Security Auditor
**Roles**: `["viewer_prod", "secrets_reader"]`

**Allowed Actions**:
- Read all entries including production
- Export configurations for audit
- Read plain secret values
- View change history

**Blocked Actions**:
- Any write operations
- Delete operations
- Template management

---

## Security Best Practices

### 1. Principle of Least Privilege
- Start with minimal permissions (`viewer`)
- Add roles incrementally as needed
- Regularly audit role assignments

### 2. Separate Concerns
- Use `secrets_reader` only for authorized personnel
- Combine `editor` + `maintainer` carefully (deletion is permanent)
- Reserve `admin` for infrastructure team leads

### 3. Environment Separation
- Use `viewer` for developers in non-production
- Require `viewer_prod` for production access
- Audit all production access regularly

### 4. CI/CD Security
- Grant `cicd` role to automation only
- Never grant `editor` or `maintainer` to CI/CD without approval
- Use read-only access for build/deploy verification

---

## Develop OPA

### Evaluate Policies

```shell
# Evaluate with full explanation
opa eval --data policies/ --data data.json --input test-input.json "data.policies.rbac.allow" --explain=full

# Test authorization policy
opa eval --data authz/policy.rego --data roles.json --data permissions.json --input test-input.json "data.authz.allow"
```

### Expression regular

https://go.dev/play/p/Zza13T47R4u

### validate json
```shell
jq . permissions.json
jq . roles.json
```

---

## Running Tests

### Basic Test Execution
```bash
# Run all tests
opa test policies/authz/ -v

# Run with coverage
opa test policies/authz/ --coverage

# Check if files need formatting (list files)
opa fmt --list policies/authz/

# Check format and fail if changes needed
opa fmt --fail policies/authz/

# Format policies (write changes)
opa fmt --write policies/authz/

# Show diff without applying changes
opa fmt --diff policies/authz/
```

### Expected Output
```
PASS: 58/58
Coverage: 96.64%
```

---

## Policy Files Structure

```
policies/authz/
├── policy.rego              # Main authorization policy with security checks
├── policy_test.rego         # Comprehensive test suite (58 tests)
├── roles.json              # Role definitions and permissions mapping
├── permissions.json        # Permission patterns (regex)
└── whitelist.json         # Public endpoints (no auth required)
```

---

## Continuous Integration

### Recommended CI Pipeline Steps

1. **Format Check**

   ```bash
   # validate json
   jq . permissions.json
   jq . roles.json
   
   # List files that need formatting
   opa fmt --list policies/authz/
   
   # Or fail if any files need formatting (recommended for CI)
   opa fmt --fail policies/authz/
   ```

2. **Validation**
   ```bash
   opa check policies/authz/
   ```

3. **Run Tests**
   ```bash
   opa test policies/authz/ -v
   ```

4. **Coverage Check**
   ```bash
   opa test policies/authz/ --coverage --threshold=95
   ```

5. **Build for Production**
   ```bash
   opa build -t wasm -e authz/allow policies/authz/
   ```

---

## Future Enhancements

### Potential Improvements
1. **Rate Limiting**: Add rate limit checks for sensitive operations
2. **Audit Logging**: Integrate with audit logging for denied requests
3. **Dynamic Permissions**: Support for user-specific permissions
4. **Temporal Access**: Time-based access grants
5. **Resource-Level Controls**: Fine-grained permissions per resource

### Monitoring Recommendations
1. Alert on repeated authorization failures
2. Monitor production secret access patterns
3. Track role assignment changes
4. Audit `admin` role usage

---

## Support and Documentation

### Resources
- [OPA Documentation](https://www.openpolicyagent.org/docs/)
- [Rego Language Guide](https://www.openpolicyagent.org/docs/latest/policy-language/)
- [OPA Playground](https://play.openpolicyagent.org/)
- [Rego Style Guide](https://docs.styra.com/opa/rego-style-guide)

### Testing New Policies
Use the OPA playground to test policy changes before deployment:
1. Copy `policy.rego` content
2. Add test input/data
3. Validate expected behavior
4. Add corresponding test cases
