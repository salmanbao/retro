# Research: Authorization Module (RBAC)

**Feature**: 002-rbac-authorization | **Date**: 2026-05-15

## Research Questions

1. Permission resolution with role inheritance (single parent, max 3 levels deep)
2. Wildcard permission matching (e.g., "campaign.*" grants all campaign.* permissions)
3. Middleware pattern for requirePermission checks
4. Efficient data structures for permission lookup
5. Circular role inheritance prevention

---

## Decision 1: Permission Resolution Strategy

**Choice**: Graph traversal with memoization

**Rationale**:
- Role hierarchy is a directed acyclic graph (DAG) with single parent (tree)
- Max depth of 3 levels means traversal is never expensive
- Memoize effective permissions per profile to meet <50ms performance target
- BFS or DFS both work; BFS is simpler for collecting all inherited permissions

**Alternatives considered**:
- Recursive resolution: Simple but stack overflow risk on deep hierarchies
- SQL query with recursive CTE: Database-dependent, harder to test
- Precomputed flattened permissions: Fastest but requires cache invalidation on role changes

---

## Decision 2: Wildcard Permission Matching

**Choice**: Prefix-based glob matching

**Rationale**:
- "campaign.*" matches "campaign.create", "campaign.update", "campaign.delete"
- Simple string split on "." and prefix comparison
- Can be implemented with `strings.HasPrefix(permission, strings.TrimSuffix(wildcard, ".*") + ".")`

**Alternatives considered**:
- Full glob/regex: Overkill for dot-notation
- Explicit expansion at role creation: Harder to maintain when adding new permissions
- Separate wildcard permission type: Adds complexity to data model

---

## Decision 3: Middleware Pattern

**Choice**: Middleware chain with context propagation

**Rationale**:
- `AuthMiddleware.Authenticate` runs first, populates user/session/profile in context
- `AuthorizationMiddleware.RequirePermission(permission)` runs next, checks context
- Profile ID retrieved from context via `GetActiveProfileID(ctx)`
- Uses chi router's `Use` and `.Group` patterns for clean route grouping

**Alternatives considered**:
- Decorator pattern: Couples authz to specific handlers
- Service injection: Requires passing authz service to all handlers
- AOP/annotation: Not idiomatic in Go

---

## Decision 4: Data Structure for Permission Lookup

**Choice**: `map[string]struct{}` (set) for permissions per role; cached effective permissions per profile

**Rationale**:
- O(1) lookup for `hasPermission` checks
- Memory: 50 permissions × ~50 bytes = 2.5KB per role, acceptable
- Profile effective permissions cached with TTL or invalidated on role assignment change
- Use slices for permission lists (ordering matters for API responses)

**Alternatives considered**:
- BitSet: Memory efficient but hard to map permission strings to bits
- Trie: Overkill for flat dot-notation permissions
- HashSet from external library: Adds dependency; stdlib map is sufficient

---

## Decision 5: Circular Inheritance Prevention

**Choice**: Validation on role create/update by traversing parent chain

**Rationale**:
- Before assigning parent to role, traverse up to root checking for cycles
- Maximum 3 levels means max 3 lookups = O(1) operation
- Reject with error if proposed parent is self or in ancestor chain

**Alternatives considered**:
- Database constraint (foreign key cycle): Not supported by most DBs
- Periodic cleanup job: Risk of inconsistent state between checks
- Eventual consistency with compensation: Overkill for this use case

---

## Technical Decisions Summary

| Aspect | Decision |
|--------|----------|
| Permission resolution | BFS traversal + memoization cache |
| Wildcard matching | Prefix-based glob |
| Middleware pattern | Context-based chain |
| Permission storage | Set (map[string]struct{}) |
| Cache invalidation | On role assignment change |
| Cycle prevention | Pre-assignment validation |

---

## Implementation Notes

- Platform admin role (`platform_admin`) with permission `*` (all) should bypass wildcard matching and always return true
- Non-existent permission strings should return `false` (not error) per spec FR-AUTH-013
- Role hierarchy depth validation at role creation time (not runtime traversal)