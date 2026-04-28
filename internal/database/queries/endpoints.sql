-- name: GetEndpointByIDAndAgentID :one
-- Validates that an agent is permitted to submit check results for a given endpoint.
-- Permission is granted when the endpoint and the agent share a common active topology
-- where the endpoint carries the 'endpoint' role and the agent carries the 'agent' role.
SELECT e.id FROM endpoints e
WHERE e.id = ?
  AND e.enabled = 1
  AND EXISTS (
    SELECT 1
    FROM topology_members tm_e
    JOIN endpoint_tags et     ON et.tag_id = tm_e.tag_id AND et.endpoint_id = e.id
    JOIN topologies t         ON t.id = tm_e.topology_id AND t.enabled = 1
    JOIN topology_members tm_a ON tm_a.topology_id = t.id AND tm_a.role = 'agent'
    JOIN agent_tags at        ON at.tag_id = tm_a.tag_id AND at.agent_id = ?
    WHERE tm_e.role = 'endpoint'
  )
LIMIT 1;

-- name: CreateEndpoint :one
-- Creates a standalone (non-agent) endpoint in a section.
INSERT INTO endpoints (id, section_id, address, port, enabled)
VALUES (?, ?, ?, ?, ?)
RETURNING id;

-- name: CreateAgentEndpoint :one
-- Creates an endpoint that represents one of our own agents (auto-registered on claim).
INSERT INTO endpoints (id, section_id, address, is_agent, linked_agent_id)
VALUES (?, ?, ?, 1, ?)
RETURNING id;

-- name: GetEndpointsForAgent :many
-- Resolves the full set of endpoints an agent should monitor based on active topology memberships.
-- An endpoint is included when:
--   1. It is in the same section as the agent (section guard).
--   2. One of its tags appears in topology_members with role='endpoint'.
--   3. The same topology has a member with role='agent' whose tag is assigned to the agent.
--   4. That topology is enabled.
--   5. The endpoint is not the agent's own self-registered endpoint (prevents self-monitoring).
SELECT DISTINCT e.id, e.address, e.port, e.enabled
FROM endpoints e
JOIN endpoint_tags   et    ON et.endpoint_id = e.id
JOIN topology_members tm_e ON tm_e.tag_id = et.tag_id AND tm_e.role = 'endpoint'
JOIN topologies      t     ON t.id = tm_e.topology_id  AND t.enabled = 1
JOIN topology_members tm_a ON tm_a.topology_id = t.id  AND tm_a.role = 'agent'
JOIN agent_tags      at    ON at.tag_id = tm_a.tag_id   AND at.agent_id = ?1
WHERE e.enabled = 1
  AND e.section_id = (SELECT section_id FROM agents WHERE id = ?1)
  AND NOT (e.is_agent = 1 AND e.linked_agent_id = ?1);

-- name: ListEndpointsBySection :many
SELECT id, section_id, address, port, enabled, is_agent, linked_agent_id, updated_at, created_at
FROM endpoints
WHERE section_id = ?
ORDER BY created_at DESC;

-- name: UpdateEndpoint :exec
UPDATE endpoints
SET address = ?, port = ?, enabled = ?
WHERE id = ?;

-- name: DeleteEndpoint :exec
DELETE FROM endpoints WHERE id = ?;
