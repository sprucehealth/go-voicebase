#!/bin/bash

# This script assumes that there is a live RDS instance
# that is publically accessible. The RDS instance is needed
# to validate the migration DDL and create an updated snapshot
RDS_INSTANCE="127.0.0.1"
if [ "$RDS_USERNAME" = "" ]; then
	RDS_USERNAME="carefront"
fi

DATABASE_NAME="database_$RANDOM"

# The RDS password for this instance is expected to be set as an environment variable
# named DEV_RDS_PASSWORD
PASSWORD_ARG="-p$DEV_RDS_PASSWORD"
if [ "$DEV_RDS_PASSWORD" = "" ]; then
	PASSWORD_ARG=""
fi

# trapping the TERM signal enables us to instruct the
# process executing the bash script to exit if the TERM
# signal is sent to it
trap "exit 1" TERM
export TOP_PID=$$

function cleanup {
	echo -e "--- Cleaning up temp files created and dropping database $DATABASE_NAME\n"
	rm temp-migration.sql
	rm temp.sql
	rm temp-data.sql
	echo "drop database $DATABASE_NAME;" > temp-drop-database.sql
	mysql -h $RDS_INSTANCE -u $RDS_USERNAME $PASSWORD_ARG < temp-drop-database.sql
	if [ $? -ne 0 ]; then
		echo "--- ERROR: Unable to drop database $DATABASE_NAME from rds instance"
	fi
	rm temp-drop-database.sql
}

# Identify the latest snapshot of the database that exists
# The latest snapshot is essentially the snapshot with the largest number in the snapshot-N.sql format
latestSnapshotNumber=`ls -r snapshot-*.sql | cut -d- -f 2  | cut -d. -f1 | sort -nr | head -1`
latestDataSnapshotNumber=`ls -r data-snapshot-*.sql | cut -d- -f 3  | cut -d. -f1 | sort -nr | head -1`
latestMigrationNumber=$((latestSnapshotNumber + 1))

if [ ! -f migration-$latestMigrationNumber.sql ]; then
	echo "FAILED: migration-$latestMigrationNumber.sql file does not exist"
	exit 1
fi

if [ ! $latestMigrationNumber -gt $latestSnapshotNumber ]; then
	echo "FAILED: Latest snapshot $latestSnapshotNumber >= migration $latestMigrationNumber" > /dev/stderr
	exit 1
fi

if [ ! $latestMigrationNumber -gt $latestDataSnapshotNumber ]; then
	echo "FAILED: Latest data snapshot $latestDataSnapshotNumber >= migration $latestMigrationNumber" > /dev/stderr
	exit 1
fi

## add the create database and use database statements before the rest of the sql statements
echo "create database $DATABASE_NAME; use $DATABASE_NAME;"  | cat - snapshot-$latestSnapshotNumber.sql > temp.sql
echo "use $DATABASE_NAME;" | cat - data-snapshot-$latestDataSnapshotNumber.sql > temp-data.sql

# Use this snapshot as the base to create a random database on a test mysql instance
echo -e "--- Creating database $DATABASE_NAME and restoring schema from snapshot-$latestSnapshotNumber.sql\n"
mysql -h $RDS_INSTANCE -u $RDS_USERNAME $PASSWORD_ARG < temp.sql
mysql -h $RDS_INSTANCE -u $RDS_USERNAME $PASSWORD_ARG < temp-data.sql

# Apply the latest migration file to the database
echo -e "--- Applying DDL in migrate-$latestMigrationNumber.sql to database\n"
echo "use $DATABASE_NAME;" | cat - migration-$latestMigrationNumber.sql > temp-migration.sql
mysql -h $RDS_INSTANCE -u $RDS_USERNAME $PASSWORD_ARG < temp-migration.sql

if [ $? -ne 0 ]; then
	cleanup
	kill -s TERM $TOP_PID
fi

dataSnapshotTables="app_text localized_text answer_type region health_condition languages_supported tips \
	tips_section section screen_type question_type question question_fields extra_question_fields potential_answer photo_tips \
	patient_layout_version layout_blob_storage layout_version dr_layout_version diagnosis_layout_version care_providing_state dispense_unit \
	drug_name drug_route drug_form drug_supplemental_instruction deny_refill_reason state photo_slot \
	photo_slot_type role_type account_available_permission account_group account_group_permission \
	email_sender sku_category sku"

# If migration successful, snapshotting database again to generate new schema
newSnapshotNumber=$((latestSnapshotNumber + 1))
newDataSnapshotNumber=$((latestDataSnapshotNumber + 1))
echo -e "--- Creating new snapshot from database into snapshot-$newSnapshotNumber.sql\n"
`mysqldump -h $RDS_INSTANCE -u $RDS_USERNAME --no-data $DATABASE_NAME $PASSWORD_ARG > snapshot-$newSnapshotNumber.sql`
`mysqldump -h $RDS_INSTANCE -u $RDS_USERNAME $PASSWORD_ARG $DATABASE_NAME $dataSnapshotTables > data-snapshot-$newDataSnapshotNumber.sql`
cleanup
