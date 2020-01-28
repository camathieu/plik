rs.initiate({
    _id: 'rs0',
    members: [
        {_id: 0, host: "mongo1:27017" },
        {_id: 1, host: "mongo2:27017" },
        {_id: 2, host: "mongo3:27017" }
    ]
});

for (i = 0; i < 60; i++) {
    if (db.isMaster().primary) {
        print("master is " + db.isMaster().primary);
        quit(0);
    } else {
        print("waiting for master election");
    }
    sleep(1000);
}
quit(1);