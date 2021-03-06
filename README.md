# Signavio Workflow Accelerator Connector

The Workflow Accelerator Connector is a RESTful web service which acts as a proxy between Signavio's Workflow Accelerator and an external SQL database.

![Overview](docs/images/workflow-connector-overview.png?raw=true "Overview")

In order to use the connector with Signavio's Workflow Accelerator, the connector must first be running on a server that is accessible to the public internet. After the connector is running and the databas has been provisioned, a Workflow Accelerator administrator can [add](https://docs.signavio.com/userguide/workflow/en/integration/connectors.html#configuring-a-connector) the connector to a workspace under _Services & Connector_ menu entry. A process owner can then use the connector in a process to populate a drop down field dynamically using data from the database. More information about this can be found [here](https://docs.signavio.com/userguide/workflow/en/integration/connectors.html)

## Features

- Insert and update table rows using a standard RESTful API
- Supports multiple SQL databases (MicrosoftSQL Server, Sqlite3, MySQL, PostgresSQL)

## Deployment

The following examples will demonstrate how to deploy the workflow connector web service in the cloud, or on premise.

### In the cloud (using Heroku) ###

The following [screencast](https://drive.google.com/file/d/1V8Kizoka-5L-56SpqTRxshCBBDyerB7v/view?usp=sharing) will show how to use heroku to install, configure and deploy the workflow connector.

### On premise (bare metal) ###

Deploying the workflow connector on premise is as simple as copying the executable file, and related configuration files, to a directory on the server where it should be run. The following section demonstrate how this can be done in an ubuntu linux environment.

#### Installation ####

The connector can be downloaded from the [release page](https://github.com/signavio/workflow-connector/releases) for linux and windows platforms. Alternatively, the executable can be generated by compiling the source code as shown below.

##### Install from source #####

1. Download and install go from your distrubtion's package manager (for ubuntu `apt-get install go`) and make sure you are using version >= 1.9
2. Set the `$GOPATH` environment variable and add `$GOPATH/bin` to your `$PATH`. Here we assume that `$GOPATH` points to `~/go`. 

```sh
mkdir ~/go
echo "export GOPATH=$HOME/go" >> ~/.bashrc
echo "export PATH=\$PATH:\$GOPATH/bin" >> ~/.bashrc
source ~/.bashrjc
```
3. Download and install the workflow-connector using the `go get` command on the command line.

```sh
go get -v github.com/signavio/workflow-connector
```
4. Download the `dep` utility in order to install the projects dependencies locally.

```sh
# install dep utility
go get github.com/golang/dep/cmd/dep
# install project's dependencies
dep ensure -vendor-only
```
5. compile the source code into an architecture specific executable. Adjust your `$GOARCH` and `$GOOS` variables as needed.

```sh
# export the required environment variables
export CGO_ENABLED=1 GOARCH=amd64 GOOS=linux
# compile the code and generate the `workflow-connector` binary
go build -o workflow-connector cmd/wfadb/main.go
```

6. The executable is now located in `$GOPATH/github.com/signavio/workflow-connector/workflow-connector`. At this point you can copy the executable somewhere in your `$GOPATH`

```sh
# Here we assume ~/bin/ is in your $PATH
copy $GOPATH/github.com/signavio/workflow-connector/workflow-connector ~/binj/
```

7. If you have TLS enabled (highly recommended) and want to listen on port 443 without running the executable as root, you can set the proper permissions using the `setcap` command

```sh
setcap 'cap_net_bind_service=+ep' ~/bin/wfc
```
8. You can then run the executable with `systemd-run` as a non-root user and track the status of the workflow-connector process using the regular systemd tooling.

```sh
export WFC_PORT=443 && systemd-run --user ~/bin/wfc
```

#### Configuration ####

##### config.yaml #####
All program and environment specific configuration settings (like database connection information, username, password, etc.) should be saved in a `config.yaml` file that is located in the same directory as the executable, or in one of the following directories:

| **Linux**                           |
|-------------------------------------|
| ./config/                           |
| /etc/                               |
| ~/.config/workflow-connector        |

All configuration settings in `config.yaml` can also be specified as environment variables. For example, you can specify the database connection url by exporting the environment variable `DATABASE_URL=sqlserver://john:84mj29rSgHz@172.17.8.2?database=test`. This means that nested fields in the yaml file are delimited with a '_' (underscore) character when used in an environment variable. All configuration settings declared via environment variables will take precedence over the settings in your `config.yaml` file.

##### descriptor.json #####

The workflow connector also needs to know the schema of the data it will receive from the database. This is stored in the connector descriptor file `descriptor.json` and an example is provided in this repository. You can also refer to the [workflow documentation](https://docs.signavio.com/userguide/workflow/en/integration/connectors.html#connector-descriptor) for more information. 

##### HTTP basic auth #####

The webservice will only respond to clients using HTTP basic auth. This can be enabled by setting `tls.enabled = true` and providing valid TLS certificates in the `config.yaml` file. The username for HTTP basic auth is stored in as plain text in `config.yaml` but the password is stored salted and hashed using [argon2](https://passlib.readthedocs.io/en/stable/lib/passlib.hash.argon2.html). You can use the following commands to generate a argon2 password hash using python.

1. Install passlib using python `pip`

```sh
pip install passlib
```

2. Use the python shell in the command line to generate an argon2 password hash with a digest size of 32 bytes

```python
from passlib.hash import argon2
argon2.using(digest_size=32).hash("password")
```

#### Testing ####

You can test the deployment with a local sqlite database to make sure that the REST API is behaving properly. The following sections demonstrate how this can be done.

##### Prerequisites #####

1. Download and install sqlite

```sh
apt-get install sqlite
```

2. Create the database file in the config directory we used earlier.
```sh
touch ~/.config/workflow-connector/test.db
```

##### Populate the database #####

For testing purposes, we will create a table called `equipment` and populate it with data. The table will end up looking like this: 

**Equipment**

| id | name                           | acquisition_cost | purchase_date       |
|----|--------------------------------|------------------|---------------------|
|  1 | Stainless Steel Cooling Spiral |            119.0 | 2017-09-07 12:00:00 |
|  2 | Fermentation Tank (50L)        |            250.0 | 2014-09-07 11:00:00 |
|  3 | Temperature Gauge              |            49.99 | 2017-09-04 11:00:00 |
|  4 | Masch Tun (50L)                |           199.99 | 2016-09-04 11:00:00 |


We will write the necessary sql statements to a temporary file and then import the file into a sqlite database.

```sh
echo "\
CREATE TABLE IF NOT EXISTS equipment ( \
id integer not null primary key, \
name text, \
acquisition_cost real, \
purchase_date datetime); \
INSERT INTO equipment(id, name, acquisition_cost, purchase_date) \
VALUES \
(1,'Stainless Steel Cooling Spiral',119.0,'2017-09-07 12:00:00'), \
(2,'Fermentation Tank (50L)',250.0,'2014-09-07 11:00:00'), \
(3,'Temperature Gauge',49.99,'2017-09-04 11:00:00'), \
(4,'Masch Tun (50L)',199.99,'2016-09-04 11:00:00'); \
" > /tmp/test.db
```

```sh
cd ~/.config/workflow-connector
sqlite3 test.db < /tmp/test.db
```

The table should now look like this: 

```sqlite
sqlite3 test.db
SQLite version 3.20.1 2017-08-24 16:21:36
Enter ".help" for usage hints.
sqlite> SELECT * FROM equipment;
1|Stainless Steel Cooling Spiral|119.0|2017-09-07 12:00:00
2|Fermentation Tank (50L)|250.0|2014-09-07 11:00:00
3|Temperature Gauge|49.99|2017-09-04 11:00:00
4|Masch Tun (50L)|199.99|2016-09-04 11:00:00
sqlite> .quit

```
Now exit out of the sqlite command line interface (using the command `.quit`)

##### Run the workflow connector #####

Before running the `workflow-connector` command, either edit the `config.yaml` file to include the database connection parameters and other settings, or export these settings as environment variables.

```sh
# Export environment variables
#
export PORT=:8080 DATABASE_URL=test.db DATABASE_DRIVER=sqlite3
#
# Run the connector
~/bin/workflow-connector
Listening on :8080

```

##### Exercise the REST API #####

Now we can test the functionality of the connector's REST API either in a new terminal, or using the following the [postman collection](TODO). All HTTP requests are sent using HTTP basic auth with the default username (`wfauser`) password (`Foobar`) combination here.

Go ahead and fetch the product with id 1 by sending a `HTTP GET` request to the connector using the `curl` command (you can `apt-get install curl` if `curl` is not yet installed):

```sh
curl --verbose --header "Authorization: Basic $(echo -n "wfauser:Foobar" | base64)" --request GET http://localhost:8080/equipment/1
# Response:
## Headers
> GET /equipment/1 HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.55.1
> Accept: */*
> Authorization: Basic d2ZhdXNlcjpGb29iYXI=
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Fri, 23 Mar 2018 21:33:47 GMT
< Content-Length: 595
<
## Data
{
  "cost": {
    "amount": 119,
    "currency": "EUR"
  },
  "equipmentMaintenance": [],
  "equipmentWarranty": [],
  "id": "1",
  "name": "Stainless Steel Cooling Spiral",
  "purchaseDate": "2017-09-07T12:00:00Z"
}
```

###### Insert a new product in the database ######

You can create a new product by sending a `HTTP POST` to the webservice

```sh
curl --verbose --header "Authorization: Basic $(echo -n "wfauser:Foobar" | base64)" --request POST --data 'name=Malt+mill+550&acquisitionCost=1270&purchaseDate=2016-09-04+11:00:00' http://localhost:8080/equipment

# Response:
## Headers
> POST /equipment HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.56.1
> Accept: */*
> Authorization: Basic d2ZhdXNlcjpGb29iYXI=
> Content-Type: application/x-www-form-urlencoded
> Content-Length: 45
>
* upload completely sent off: 45 out of 45 bytes
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Fri, 23 Mar 2018 21:33:47 GMT
< Content-Length: 2
<
## Data
{
  "cost": {
    "amount": 1270,
    "currency": "EUR"
  },
  "equipmentMaintenance": [],
  "equipmentWarranty": [],
  "id": "5",
  "name": "Malt mill 550",
  "purchaseDate": "2017-09-04T11:00:00Z"
}
```

###### Updating an existing product ######

By sending a `HTTP PUT` to the web service you can change existing with entries. Let's go ahead an adjust the name of the malt mill we just added recently.


```sh
curl --verbose --header "Authorization: Basic $(echo -n "wfauser:Foobar" | base64)" --request PUT --data 'name=Malt+mill+400' http://localhost:8080/equipment/5

# Response:
## Headers
> PUT /equipment HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.56.1
> Accept: */*
> Authorization: Basic d2ZhdXNlcjpGb29iYXI=
> Content-Type: application/x-www-form-urlencoded
> Content-Length: 45
>
* upload completely sent off: 45 out of 45 bytes
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Fri, 23 Mar 2018 21:33:47 GMT
< Content-Length: 2
<
## Data
{
  "cost": {
    "amount": 1270,
    "currency": "EUR"
  },
  "equipmentMaintenance": [],
  "equipmentWarranty": [],
  "id": "5",
  "name": "Malt mill 400",
  "purchaseDate": "2017-09-04T11:00:00Z"
}
```

###### Deleting an existing product ######

TODO: deletion is not supported at the moment.

## Technical Overview

TODO

## Support

Any inquiries for support can be sent to [support](mailto:support@signavio.com). 

## Authors

The development team at Signavio with input from Stefano Da Ros and Peter Hilton 

## Licence

Apache License 2.0
