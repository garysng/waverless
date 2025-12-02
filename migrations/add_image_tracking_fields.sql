-- Add image tracking fields to endpoints table
-- This migration adds fields to track Docker image updates

ALTER TABLE endpoints
ADD COLUMN image_prefix VARCHAR(500) NOT NULL DEFAULT '' COMMENT 'Image prefix for matching updates (e.g., "wavespeed/model-deploy:wan_i2v-default-")' AFTER image;

ALTER TABLE endpoints
ADD COLUMN  image_digest VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Current image digest from DockerHub' AFTER image_prefix;

ALTER TABLE endpoints
ADD COLUMN latest_image VARCHAR(500) NOT NULL DEFAULT '' COMMENT 'Latest available image if update is detected' AFTER image_digest;

ALTER TABLE endpoints
ADD COLUMN image_last_checked DATETIME(3) NULL COMMENT 'Last time image was checked for updates' AFTER latest_image;

ALTER TABLE endpoints
ADD COLUMN  description VARCHAR(500) NOT NULL DEFAULT '';
