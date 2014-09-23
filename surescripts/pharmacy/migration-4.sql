create table updated_pharmacy (
	id serial primary key,
	ncpdpid text not null unique,
	store_number text not null,
	store_name text not null,
	address_line_1 text not null,
	address_line_2 text,
	city text not null,
	state text not null,
	zip text not null,
	phone_primary text not null,
	fax text not null,
	active_start_time timestamp not null,
	active_end_time timestamp not null,
	service_level integer,
	specialty integer,
	last_modified_date timestamp not null,
	twenty_four_hour_flag text,
	version text not null,
	cross_street text,
	is_from_surescripts boolean not null
);
\copy updated_pharmacy FROM '/Users/kunaljham/Dropbox/personal/workspace/backend/carefront/src/github.com/sprucehealth/backend/apps/pharmacydb/2014-07-11-08-32-42-PharmacyData_20140711.csv' WITH DELIMITER  ',' QUOTE '"' CSV;


create table pharmacy_migration (
	id serial primary key,
	file_name text not null,
	tstamp timestamp default current_timestamp,
	rows_inserted int,
	rows_updated int,
	status text not null,
	error text
);

	create table pharmacy_test_data_mapping (
		pharmacy_id integer references pharmacy(id),
		dosespot_test_id integer,
		ncpdpid text not null
	);