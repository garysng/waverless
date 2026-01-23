-- Add gpu_count column to endpoints table
ALTER TABLE `endpoints` ADD COLUMN `gpu_count` int NOT NULL DEFAULT 1 COMMENT 'GPU count per replica (resources = per-gpu-config * gpuCount)' AFTER `replicas`;
