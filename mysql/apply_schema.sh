#!/bin/bash -e

# This script makes it easy to apply changes to the development and production database once
# the schema has been validated.

LOCAL_DB_USERNAME="carefront"
LOCAL_DB_NAME="carefront_db"
DEV_USERNAME="spruce"
DEV_DB_NAME="spruce"
DEV_HOST="dev-mysql-1.node.staging-us-east-1.spruce"
PROD_DB_NAME="carefront"
PROD_DB_INSTANCE="master.mysql.service.prod-us-east-1.spruce"
STAGING_DB_NAME="carefront"
STAGING_DB_USER_NAME="restapi"
STAGING_DB_INSTANCE="staging-mysql-1.node.staging-us-east-1.spruce"
DEMO_DB_USER_NAME="carefront"
DEMO_DB_NAME="carefront_db"

argsArray=($@)
len=${#argsArray[@]}

if [ $len -lt 2 ];
then
	echo "ERROR: Usage ./apply_schema.sh [local|dev|prod|demo|staging] migration1 migration2 .... migrationN"
	exit 1;
fi

env=${argsArray[0]}
for migrationNumber in ${argsArray[@]:1:$len}
do
	echo "Applying migration-$migrationNumber.sql"

	# ensure that the file exists
	if [ ! -f snapshot-$migrationNumber.sql ] || [ ! -f data-snapshot-$migrationNumber.sql ] || [ ! -f migration-$migrationNumber.sql ]; then
		echo "ERROR: Looks like migration $migrationNumber has not yet been validated using validate_schema.sql and so they will not be applied to database"
		exit 1
	fi


	case "$env" in

		"local" )
			echo "use $LOCAL_DB_NAME; insert into migrations (migration_id, migration_user) values ($migrationNumber, '$USER');" > temp-migration.sql
			echo "use $LOCAL_DB_NAME;" | cat - migration-$migrationNumber.sql > temp.sql
			mysql -u $LOCAL_DB_USERNAME -p$LOCAL_DB_PASSWORD < temp.sql
			mysql -u $LOCAL_DB_USERNAME -p$LOCAL_DB_PASSWORD < temp-migration.sql
		;;

		"staging" )
			if [ "$STAGING_DB_PASSWORD" == "" ]; then
				echo "STAGING_DB_PASSWORD not set" > /dev/stderr
				exit 1
			fi
			echo "use $STAGING_DB_NAME; insert into migrations (migration_id, migration_user) values ($migrationNumber, '$USER');" > temp-migration.sql
			LOGMSG="{\"env\":\"$env\",\"user\":\"$USER\",\"migration_id\":\"$migrationNumber\"}"
			echo "use $STAGING_DB_NAME;" | cat - migration-$migrationNumber.sql > temp.sql
			scp temp.sql $STAGING_DB_INSTANCE:~
			scp temp-migration.sql $STAGING_DB_INSTANCE:~
			ssh -t $USER@$STAGING_DB_INSTANCE "sudo ec2-consistent-snapshot -mysql.config /mysql-data/mysql/backup.cnf -tag migrationId=migration-$migrationNumber"
			ssh -t $USER@$STAGING_DB_INSTANCE "mysql -h 127.0.0.1 -u $STAGING_DB_USER_NAME -p$STAGING_DB_PASSWORD < temp.sql ; mysql -h $STAGING_DB_INSTANCE -u $STAGING_DB_USER_NAME -p$STAGING_DB_PASSWORD < temp-migration.sql; logger -p user.info -t schema '$LOGMSG'"
		;;

		"dev" )
			echo "use $DEV_DB_NAME; insert into migrations (migration_id, migration_user) values ($migrationNumber, '$USER');" > temp-migration.sql
			echo "use $DEV_DB_NAME;" | cat - migration-$migrationNumber.sql > temp.sql
			mysql -h $DEV_HOST -u $DEV_USERNAME -p$DEV_RDS_PASSWORD < temp.sql
			mysql -h $DEV_HOST -u $DEV_USERNAME -p$DEV_RDS_PASSWORD < temp-migration.sql
		;;

		"demo" )
			echo "use $DEMO_DB_NAME; insert into migrations (migration_id, migration_user) values ($migrationNumber, '$USER');" > temp-migration.sql
			LOGMSG="{\"env\":\"$env\",\"user\":\"$USER\",\"migration_id\":\"$migrationNumber\"}"
			echo "use $DEMO_DB_NAME;" | cat - migration-$migrationNumber.sql > temp.sql
			scp temp.sql 54.210.97.69:~
			scp temp-migration.sql 54.210.97.69:~
			ssh 54.210.97.69 "mysql -u $DEMO_DB_USER_NAME -p$DEMO_DB_PASSWORD < temp.sql ; mysql -u $DEMO_DB_USER_NAME -p$DEMO_DB_PASSWORD < temp-migration.sql; logger -p user.info -t schema '$LOGMSG'"
		;;

		"prod" )
			if [ "$PROD_DB_USER_NAME" == "" ]; then
				echo "PROD_DB_USER_NAME not set"
				exit 1
			fi
			if [ "$PROD_DB_PASSWORD" == "" ]; then
				echo "PROD_DB_PASSWORD not set"
				exit 1
			fi
			echo "use $PROD_DB_NAME; insert into migrations (migration_id, migration_user) values ($migrationNumber, '$USER');" > temp-migration.sql
			LOGMSG="{\"env\":\"$env\",\"user\":\"$USER\",\"migration_id\":\"$migrationNumber\"}"
			echo "use $PROD_DB_NAME;" | cat - migration-$migrationNumber.sql > temp.sql
			scp temp.sql 54.209.10.66:~
			scp temp-migration.sql 54.209.10.66:~
			ssh -t $PROD_DB_INSTANCE "sudo ec2-consistent-snapshot -mysql.config /mysql-data/mysql/backup.cnf -tag migrationId=migration-$migrationNumber"
			ssh 54.209.10.66 "mysql -h $PROD_DB_INSTANCE -u $PROD_DB_USER_NAME -p$PROD_DB_PASSWORD < temp.sql ; mysql -h $PROD_DB_INSTANCE -u $PROD_DB_USER_NAME -p$PROD_DB_PASSWORD < temp-migration.sql; logger -p user.info -t schema '$LOGMSG'"
		;;
	esac

	rm temp.sql temp-migration.sql

done
