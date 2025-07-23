## develop policy
```shell
opa eval --data policies/  --data data.json --input test-input.json "data.policies.rbac.allow" --explain=full


opa eval --data authz/policy.rego --data roles.json --data permissions.json --input test-input.json "data.authz.allow"

# format
opa fmt --write authz

# test
opa test authz --verbose --coverage
opa test authz --verbose
```

### compilar politicas a wasm
```shell
 opa build -t wasm -e authz/allow authz/policy.rego permissions.json roles.json
```

### test expression regular

https://go.dev/play/p/Zza13T47R4u

### validate json 
```shell
jq . permissions.json
jq . roles.json
```

## Docs

- [course OPA Policy Authoring](https://academy.styra.com/courses/opa-rego)
- [Building With Open Policy Agent (OPA) for Better Policy as Code](https://dzone.com/articles/building-with-open-policy-agent-opa-for-better-pol)

- [playground](https://play.openpolicyagent.org/p/CJIq9dnzfC)
- [*rego* cheat sheet](https://docs.styra.com/opa/rego-cheat-sheet)
- [*rego* style guide](https://docs.styra.com/opa/rego-style-guide#use-opa-fmt)

- [SDK Go](https://www.openpolicyagent.org/integrations/opa-golang/)