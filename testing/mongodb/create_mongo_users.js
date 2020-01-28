if (db.isMaster().ismaster) {
    print("");
    print(db.isMaster().me + " is master. creating users");
    print("");
} else {
    print("");
    print(db.isMaster().me + " is not master. exiting");
    print("");
    quit(1);
}

use admin;
db.createUser({user:"admin", pwd:"secret", roles:["root"]});

db.auth("admin","secret");

use plik;
db.createUser({user:"plik", pwd:"password", roles:["readWrite"]});

print("");
print("USERS SUCCESSFULLY CREATED");
print("");

quit(0);