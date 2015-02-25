Backend Monorepo
================
[![Build Status](https://magnum.travis-ci.com/SpruceHealth/backend.svg?token=NtmZSFxujHkPCqsPtfXC&branch=master)](https://magnum.travis-ci.com/SpruceHealth/backend)

Setting up your environment & running the `gotour`
---------------------------------

	$ brew update
	$ brew doctor # ensure there are no issues with Brew or your system
	$ brew install go
	$ brew install mercurial
	$ export GOPATH=$HOME/go
	$ go get code.google.com/p/go-tour/gotour
	$ $HOME/go/bin/gotour # runs the gotour executable and opens it in a browser window

Building the app
---------------------------------

	$ go get github.com/sprucehealth/backend # checks out to $GOPATH/src/github.com/sprucehealth/backend/

One of the great things about Go is getting external packages is as simple as the above command. You're encouraged to create your package under the path that you'd upload it if you wanted to open source your project. So any personal package you wrote you could create under github.com/GITHUB_USERNAME. Someone else may have the exact same package name but its under their own unique path.

Next, `cd` to the backend app's directory and `go build` it:

	$ cd /Users/YOU/go/src/github.com/sprucehealth/backend/apps/restapi
	$ go build # binary should be at /Users/YOU/go/src/github.com/sprucehealth/backend/apps/restapi/restapi

_Having issues? See the [troubleshooting](#troubleshooting) section._

Getting environment setup
---------------------------------

Set up the AWS keys as environment variables by adding the following to `~/.bashrc` or `~/.zshrc`:

	export GOPATH=$HOME/go
	export AWS_ACCESS_KEY='ASK_KUNAL_OR_SAM_FOR_ME'
	export AWS_SECRET_KEY='ASK_KUNAL_OR_SAM_FOR_ME'

Then:

	$ source ~/.bashrc # or source ~/.zshrc

Add the following lines to `/etc/hosts`. Reason you need this is because we currently use the same binary for the restapi
as well as the website, and requests are routed based on the incoming URI

	127.0.0.1       www.spruce.loc
	127.0.0.1       api.spruce.loc

Local database setup (automatic method)
---------------------------------

1. `cd mysql`
2. Find number of the latest migration file (ex: if `migration-395.sql` is the highest-numbered file, then `395` is your number)
3. ./bootstrap_mysql.sh carefront_db carefront changethis <latest-migration-id>


Local database setup (manual method)
---------------------------------

Before running the backend server locally, we want to get a local instance of mysql running, and setup with the database schema and boostrapped data.

Install MySQL and get it running:

	$ brew install mysql
	$ mysql.server start

Setup the expected user for the restapi (user="carefront" password="changethis"):

	$ mysql -u root
	mysql> CREATE USER 'carefront'@'localhost' IDENTIFIED BY 'changethis';
	mysql> CREATE DATABASE carefront_db;
	mysql> GRANT ALL on *.* to 'carefront'@localhost;

Ensure that you have access to your local mysql instance and the `carefront_db` as "carefront". *First ensure to log out of your session by typing exit*

	$ mysql -u carefront -pchangethis
	$ use carefront_db;

Now that you have mysql up and running, lets populate the database just created with the schema and boostrapped data. Anytime we have to update the schema we create a migration filed under the mysql directory in the form migration-X.sql where X represents the migration number. A validation script loads a database with the current schema, runs the migration on this database, and then spits out snapshot-X.sql and data-snapshot-X.sql files that represent the database schema and boostrapped data respectively.

Open a new terminal tab and `cd $GOPATH/src/github.com/sprucehealth/backend/mysql`:

	$ echo "use carefront_db;" | cat - snapshot-<latest_migration_id>.sql > temp.sql
	$ echo "use carefront_db;" | cat - data-snapshot-<latest_migration_id>.sql > data_temp.sql
	
Now seed the database:

	mysql -u carefront -pchangethis < temp.sql
	mysql -u carefront -pchangethis < data_temp.sql

Go back to your mysql session tab. Log the latest migration id in the migrations table to indicate to the application the last migration that was completed:

	mysql> insert into migrations (migration_id, migration_user) values (<latest_migration_id>, "carefront");


Running the server locally
---------------------------------

Let's try running the server locally.

`cd` to the restapi folder under apps:
```
	cd $GOPATH/src/github.com/sprucehealth/backend/apps/restapi
```

Build the app and execute the run_server.bash script which tells the application where to get the config file for the local config from:
```
	go build
	./run_server.bash
```

_Having issues? See the [troubleshooting](#troubleshooting) section._


Setting up an admin user (for `http://www.spruce.loc:8443/admin/`)
---------------------------------

Creating an admin account. The reason we need to create an admin account is because there are operational tasks we have to carry out to upload the patient visit intake and doctor review layouts, and only an admin user can do that. Currently, the easiest way to create an admin account is to create a _patient account_ and then modify its role type to be that of an admin user.

But first make sure to build and start running the app:

	$ go build
	$ ./run_server.bash

> Open the [PAW file](https://github.com/SpruceHealth/api-response-examples/tree/master/v1) in [PAW (Mac App Store)](https://itunes.apple.com/us/app/paw-http-client/id584653203?mt=12) and create a new patient (ex: `jon@sprucehealth.com`):
<img src="http://f.cl.ly/items/221c0k392Z3n2R3O3Z0z/Screen%20Shot%202014-11-26%20at%201.17.28%20PM.png" />

Log back in to mysql as `carefront` and change the account's role type to:
```
	$ mysql -u carefront -pchangethis;
	mysql> USE carefront_db;
	mysql> UPDATE account SET role_type_id=(select id from role_type where role_type_tag='ADMIN') WHERE email=<admin_email>;
```

Open the PAW file again:

* In the `Layout upload (Initial Visit)`, upload the latest versions of `intake-*`, `review-*`, and `diagnose-*` json files. You'll have to locate them on disk in `./info_intake/`.

> <img src="http://f.cl.ly/items/2E3X1k0X1X2l0y1I3O2g/Screen%20Shot%202014-11-26%20at%203.14.09%20PM.png" />

* In the `Layout upload (Initial Visit)`, upload the `followup-intake-*` and `followup-review-*` json files. You'll have to locate them on disk in `./info_intake/`.

> <img src="http://f.cl.ly/items/021Z1q3h1u3Z3i0O3I1A/Screen%20Shot%202014-11-26%20at%203.14.11%20PM.png" />

Create a cost entry for the initial and the followup visits so that there exists a cost type to query against for the patient app when attempting to determine the cost of each of the visits:
```
	insert into item_cost (sku_id, status) values ((select id from sku where type='acne_visit'), 'ACTIVE');
	insert into line_item (currency, description, amount, item_cost_id) values ('USD', 'Acne visit', 4000, 1);
	insert into item_cost (sku_id, status) values ((select id from sku where type='acne_followup'), 'ACTIVE');
	insert into line_item (currency, description, amount, item_cost_id) values ('USD', 'Acne Followup', 2000, 2);
```

Make yourself a boss:

	INSERT INTO carefront_db.account_group_member (group_id, account_id) VALUES ((SELECT * FROM carefront_db.account_group WHERE name='superuser'), (select id from carefront_db.account where email='<account_email>'));

Building to run the website(s)
---------------------------------

1. `cd` to `resources`
2. `$ ./build.sh`
3. `cd` to `apps/restapi`
4. `./run_server.bash`

If it fails, you'll need to install the following dependencies:

	npm install -g react-tools browserify
	npm install reactify uglifyify
	
If you find that you need more (perhaps more have been added since this writing), look for the `.travis.yml` file for dependencies that our CI server needs, and try installing those.

> The public-facing website will be at: https://www.spruce.loc:8443/

> The admin website will be at: https://www.spruce.loc:8443/admin/
	

Running integration tests locally
---------------------------------

Add the following to `~/.bashrc` or `~/.zshrc`:

	export CF_LOCAL_DB_USERNAME='carefront'
	export CF_LOCAL_DB_PASSWORD='changethis'
	export CF_LOCAL_DB_INSTANCE='127.0.0.1'
	export DOSESPOT_USER_ID=407

Then:

	$ source ~/.bashrc # or source ~/.zshrc

To run the tests serially:

	$ cd ./test/test_integration
	$ go test -v ./...

To run the tests in parallel:

	$ cd ./test/test_integration
	$ go test -v -parallel 4 ./...


Troubleshooting
---------------------------------

### Issues during `go build`:

Error:

	github.com/sprucehealth/backend/app_url
../../app_url/action.go:9: import /Users/jonsibley/go/pkg/darwin_amd64/github.com/sprucehealth/backend/libs/golog.a: object is [darwin amd64 go1.3.3 X:precisestack] expected [darwin amd64 go1.4.1 X:precisestack]

Solution:

	go clean -r -i
	go install -a
	go build

### Issues while attempting to run the app:

Error:

	dial tcp 127.0.0.1:3306: connection refused

Solution:

	mysql.server start # note: there are ways to automatically start mysql when your machine starts, too

Error:

	Error 1045: Access denied for user 'carefront'@'localhost' (using password: YES)

Solution:

You need to set up the `carefront` user with access to the database `carefront_db` (as described above).
