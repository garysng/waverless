-- Migration: Add runtime_state column to endpoints table
-- Date: 2026-01-07

ALTER TABLE `endpoints` ADD COLUMN IF NOT EXISTS `runtime_state` JSON DEFAULT NULL COMMENT 'K8s runtime state: namespace, readyReplicas, availableReplicas, shmSize, volumeMounts';
