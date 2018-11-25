# wow
WoW API tools

Tools that use the World of Warcraft developer API.
https://develop.battle.net/

SQL statements
 # Create WoW environment
 sudo mysql -u root -p
 create database wow;
 GRANT ALL PRIVILEGES ON wow.* TO 'wow'@'localhost' IDENTIFIED BY 'wowpassword';

 CREATE TABLE items (
    id int,
    name varchar(128),
    sellPrice int,
    json varchar(5000)
 );

 CREATE UNIQUE INDEX id ON items ( id );

 # Change root password
 service mysql stop
 mkdir -p /var/run/mysqld
 chown mysql:mysql /var/run/mysqld
 mysqld_safe --skip-grant-tables &
 mysql
 UPDATE mysql.user SET authentication_string=PASSWORD('new_password') WHERE User='root';
 FLUSH PRIVILEGES;
 (kill any mysql procs running)
 service mysql stop
 service mysql start
 sudo mysql -u root -p

Backup / Restore

mysqldump -u wow -p db_name t1 > dump.sql
mysql -u wow -p db_name < dump.sql


// Auction Keys:
// rand
// petSpeciesId
// owner
// buyout
// context
// bid
// timeLeft
// ownerRealm
// petQualityId
// modifiers
// petBreedId
// petLevel
// quantity
// item
// seed
// bonusLists
// auc

