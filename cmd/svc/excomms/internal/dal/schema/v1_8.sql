CREATE TABLE deleted_resource (
	resource VARCHAR(255) NOT NULL,
	resource_id VARCHAR(34) NOT NULL,
	deleted_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);