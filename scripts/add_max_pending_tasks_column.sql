-- Migration: Add max_pending_tasks column to endpoints table
-- Date: 2025-11-18
-- Description: Adds max_pending_tasks field to control task submission eligibility checks

-- Add max_pending_tasks column if it doesn't exist
ALTER TABLE `endpoints`
ADD COLUMN IF NOT EXISTS `max_pending_tasks` int NOT NULL DEFAULT '1'
COMMENT 'Maximum allowed pending tasks before warning clients';
