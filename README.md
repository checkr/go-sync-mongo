# Go-Sync-Mongo

Simple tool that polls operations from the replication oplog of a remote server, and applies them to the local/remote server. This capability supports certain classes of real-time migrations that require that the source server remain online and in operation throughout the migration process.

To see list of commands available use:
```
$ go run main.go --help
```

Typically this command will take the following form:
```
$ go run main.go sync --src "mongodb://localhost:27018" --src-username mongooplog --src-password some-password --src-ssl=false --dst "mongodb://localhost:27019" --dst-username mongooplog --dst-password some-password --dst-ssl=false
```

This command copies oplog entries from the mongod instance running on the host mongodb0.example.net and duplicates operations to the host mongodb1.example.net. If you do not need to keep the --src host running during the migration, consider using mongodump and mongorestore or another backup operation, which may be better suited to your operation.

### Running and testing with docker

start two mongo services:
```
$ docker run --name mongo1 -p 27018:27017 -d mongo mongod --logpath ./tmp/1.log --port 27017 --replSet rs
$ docker run --name mongo2 -p 27019:27017 -d mongo mongod --logpath ./tmp/1.log --port 27017 --replSet rs
```

set up replication, user and permissions for **mongo1** server
```
$ docker exec -it mongo1 mongo admin
> rs.initiate({_id: 'rs', members: [ {_id: 1, host:'localhost:27017'}]})
> db.createUser({ user: 'mongooplog', pwd: 'some-password', roles: [ { role: "userAdminAnyDatabase", db: "admin" } ] });
> db.createRole( 
{ 
    role: "anyAction", 
    privileges: [ { 
        resource: { anyResource: true }, 
        actions: [ "anyAction" ] } ], 
    roles: []
})
> db.grantRolesToUser("mongooplog",["anyAction"])
```

set up replication, user and permissions for **mongo2** server
```
$ docker exec -it mongo2 mongo admin
> rs.initiate({_id: 'rs', members: [ {_id: 1, host:'localhost:27017'}]})
> db.createUser({ user: 'mongooplog', pwd: 'some-password', roles: [ { role: "userAdminAnyDatabase", db: "admin" } ] });
> db.createRole( 
{ 
    role: "anyAction", 
    privileges: [ { 
        resource: { anyResource: true }, 
        actions: [ "anyAction" ] } ], 
    roles: []
})
> db.grantRolesToUser("mongooplog",["anyAction"])
```

start mongo oplog sync:
```
$ go run main.go sync --src "mongodb://localhost:27018" --src-username mongooplog --src-password some-password --src-ssl=false --dst "mongodb://localhost:27019" --dst-username mongooplog --dst-password some-password --dst-ssl=false
```

Try adding, deleting and modify some records in **mongo1** server and check if they persist in **mongo2** server. You can also use the `status` command to check the record count in each cluster.
```
$ go run main.go status --src "mongodb://localhost:27018" --src-username mongooplog --src-password some-password --src-ssl=false --dst "mongodb://localhost:27019" --dst-username mongooplog --dst-password some-password --dst-ssl=false
+--------+--------+-------------+------+
|   DB   | SOURCE | DESTINATION | DIFF |
+--------+--------+-------------+------+
| checkr |      1 |           1 |    0 |
+--------+--------+-------------+------+
```

### Releases
You can download binary releases for linux, macos and windows [here](https://github.com/checkr/go-sync-mongo/releases)
