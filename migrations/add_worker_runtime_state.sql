-- Migration: Replace pod_* columns with runtime_state JSON field in workers table
-- Date: 2026-01-07

-- Add runtime_state column
ALTER TABLE workers ADD COLUMN runtime_state JSON AFTER total_execution_time_ms;

-- Migrate existing data (if any pod_* columns exist)
-- UPDATE workers 
-- SET runtime_state = JSON_OBJECT(
--     'phase', COALESCE(pod_phase, ''),
--     'status', COALESCE(pod_status, ''),
--     'reason', COALESCE(pod_reason, ''),
--     'message', COALESCE(pod_message, ''),
--     'ip', COALESCE(pod_ip, ''),
--     'nodeName', COALESCE(pod_node_name, ''),
--     'createdAt', IF(pod_created_at IS NOT NULL, DATE_FORMAT(pod_created_at, '%Y-%m-%dT%H:%i:%sZ'), NULL),
--     'startedAt', IF(pod_started_at IS NOT NULL, DATE_FORMAT(pod_started_at, '%Y-%m-%dT%H:%i:%sZ'), NULL)
-- )
-- WHERE pod_phase IS NOT NULL OR pod_status IS NOT NULL;

-- Drop old pod_* columns (if they exist)
-- ALTER TABLE workers 
--     DROP COLUMN IF EXISTS pod_phase,
--     DROP COLUMN IF EXISTS pod_status,
--     DROP COLUMN IF EXISTS pod_reason,
--     DROP COLUMN IF EXISTS pod_message,
--     DROP COLUMN IF EXISTS pod_ip,
--     DROP COLUMN IF EXISTS pod_node_name;
