# mysql2redis

## Usage

```console
$ ./mysql2redis --help
usage: mysql2redis --dbname=DBNAME --query=QUERY [<flags>]

MySQL to Redis

Flags:
      --help                    Show context-sensitive help (also try --help-long and --help-man).
      --dbuser="root"           Database user
      --dbpass=DBPASS           Database password
      --dbhost="localhost"      Database host
      --dbport=3306             Database port
      --dbsock=DBSOCK           Database socket
      --dbname=DBNAME           Database name
      --query=QUERY             SQL
      --redis-pass=REDIS-PASS   Redis password
      --redis-host="localhost"  Redis host
      --redis-port=6379         Redis port
      --redis-sock=REDIS-SOCK   Redis socket
      --redis-db=0              Redis Database
      --redis-cmd=REDIS-CMD     Redis command
      --redis-cmd-args=REDIS-CMD-ARGS
                                Redis command args (Go text/template syntax)
  -F, --separator=" "           Separator
      --no-logs                 No output redis command
      --version                 Show application version.
```

```console
$ ./mysql2redis --dbname isubata --query "SELECT id, name FROM image WHERE id IN (1,2)" --redis-cmd SET --redis-cmd-args "image_id:{{ .id }} {{ .name }}"
2018/07/22 18:55:44 SET image_id:1 default.png
2018/07/22 18:55:44 SET image_id:2 1ce0c4ff504f19f267e877a9e244d60ac0bf1a41.png

$ redis-cli GET image_id:1
"default.png"
$ redis-cli GET image_id:2
"1ce0c4ff504f19f267e877a9e244d60ac0bf1a41.png"
```

```console
$ ./mysql2redis --dbname isubata --query "SELECT id, name FROM image WHERE id IN (1,2)" --redis-cmd ZADD --redis-cmd-args "myzset {{ .id }} {{ .name }}"
2018/07/22 18:58:14 ZADD myzset 1 default.png
2018/07/22 18:58:14 ZADD myzset 2 1ce0c4ff504f19f267e877a9e244d60ac0bf1a41.png

$ redis-cli ZRANGE myzset 0 -1
1) "default.png"
2) "1ce0c4ff504f19f267e877a9e244d60ac0bf1a41.png"
```
