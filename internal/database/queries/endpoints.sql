-- name: LookupEndpointByAgentAndAddress :one
SELECT id FROM endpoints WHERE agent_id = ? AND address = ? LIMIT 1;
