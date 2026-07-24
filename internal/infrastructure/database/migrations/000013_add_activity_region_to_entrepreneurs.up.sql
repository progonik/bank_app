ALTER TABLE entrepreneurs
ADD COLUMN activity_region_id INT NOT NULL DEFAULT 0,
ADD COLUMN activity_region VARCHAR(255) NOT NULL DEFAULT '';
