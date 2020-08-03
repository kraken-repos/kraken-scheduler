# Mock Up Backing Services and Run Application

## Summary
1. Create a hana schema for your own. (Refer: [create HANA schema](#create-HANA-schema))
1. Start these docker container(s). (Refer: [Start and Stop Services](#Start-and-Stop-Services))
1. Create all these kafka topics. (Refer [create kafka topics](#create-kafka-topics))
1. Build and run application. (Refer [Build and run Application](#Build-and-run-Application))

## Start and Stop Services
```bash
# start up
docker-compose up -d

# stop without wiping data
docker-compose down

# stop and wipe data
docker-compose down -v
```

## create HANA schema
1. Connect to a HANA instance. (Refer: [team wiki](https://wiki.wdf.sap.corp/wiki/x/21-ffw))
2. Create a schema for your own and update the name of created schema into [hana.properties](./hana.properties).
```sql
CREATE SCHEMA KRAKEN_SCHEDULER
```

## Build and run Application
1. Build the jar and start application. (Refer: [README.md](../README.md))

## create kafka topics
```bash
docker-compose exec kafka bash
kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic DocumentExtracted
kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic ClaimRequestCreatedByUpload
kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic ClaimRequestFinalized
kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic TenantLcmServiceRegister
kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic TenantFailTask
kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic TenantCompleteTask
kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic TenantEvent
```

## list kafka topics
```bash
kafka-topics --list --zookeeper localhost:2181
```

## send messages
```bash
kafka-console-producer --broker-list localhost:9092 \
 --topic claims.notification \
 --property "parse.key=true" \
 --property "key.separator=:"

1:{"message":"hello1","code":"ERR001","employeeID":"1"}
2:{"message":"hello2","employeeID":"1"}
3:{"message":"hello3","employeeID":"2"}
```

## Reset (clear) a topic
```bash
kafka-configs --zookeeper localhost:2181 --entity-type topics --alter --entity-name claims.notification --add-config retention.ms=1000
kafka-configs --zookeeper localhost:2181 --entity-type topics --alter --entity-name claims.notification --add-config retention.ms=604800000   #7days
```

## Delete a topic
```bash
kafka-topics --delete --zookeeper localhost:2181 --topic claims.notification
```


