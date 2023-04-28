## Promscraper

Queries Prometheus light levels of a record and sutffs them into a mysqldb. 


Example config.yml

```yml

sql:
	name: db
	address: localhost:3306
	user: root
	password: root

prometheus: 
	endpoint: "http://prometheus:9090"

```

