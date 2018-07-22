package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func openDB(dbuser, dbpass, dbhost, dbname, socket string, port int) (*sql.DB, error) {
	userpass := fmt.Sprintf("%s:%s", dbuser, dbpass)
	var conn string
	if socket != "" {
		conn = fmt.Sprintf("unix(%s)", socket)
	} else {
		conn = fmt.Sprintf("tcp(%s:%d)", dbhost, port)
	}

	return sql.Open("mysql", fmt.Sprintf("%s@%s/%s", userpass, conn, dbname))
}

func openRedis(host, socket string, db, port int, password string) (redis.Conn, error) {
	options := []redis.DialOption{}

	if db != 0 {
		options = append(options, redis.DialDatabase(db))
	}

	if password != "" {
		options = append(options, redis.DialPassword(password))
	}

	options = append(options, redis.DialConnectTimeout(3600*time.Second))
	options = append(options, redis.DialReadTimeout(3600*time.Second))
	options = append(options, redis.DialWriteTimeout(3600*time.Second))
	options = append(options, redis.DialKeepAlive(3600*time.Second))

	if socket != "" {
		return redis.Dial("unix", socket, options...)
	}

	return redis.Dial("tcp", fmt.Sprintf("%s:%d", host, port), options...)
}

func redisArgs(tpl *template.Template, args interface{}, sep string) ([]interface{}, string, error) {
	var txt bytes.Buffer
	err := tpl.Execute(&txt, args)
	if err != nil {
		return []interface{}{}, "", err
	}

	sargs := strings.Split(txt.String(), sep)
	iargs := make([]interface{}, len(sargs))

	for i, arg := range sargs {
		iargs[i] = arg
	}

	return iargs, strings.Join(sargs, " "), nil
}

func main() {
	var app = kingpin.New("mysql2redis", "MySQL to Redis")

	var dbuser = app.Flag("dbuser", "Database user").Default("root").String()
	var dbpass = app.Flag("dbpass", "Database password").String()
	var dbhost = app.Flag("dbhost", "Database host").Default("localhost").String()
	var dbport = app.Flag("dbport", "Database port").Default("3306").Int()
	var dbsock = app.Flag("dbsock", "Database socket").String()
	var dbname = app.Flag("dbname", "Database name").Required().String()
	var query = app.Flag("query", "SQL").Required().String()
	var redisPass = app.Flag("redis-pass", "Redis password").String()
	var redisHost = app.Flag("redis-host", "Redis host").Default("localhost").String()
	var redisPort = app.Flag("redis-port", "Redis port").Default("6379").Int()
	var redisSock = app.Flag("redis-sock", "Redis socket").String()
	var redisDB = app.Flag("redis-db", "Redis Database").Default("0").Int()
	var redisCmd = app.Flag("redis-cmd", "Redis command").String()
	var redisCmdArgs = app.Flag("redis-cmd-args", "Redis command args (Go text/template syntax)").String()
	var separator = app.Flag("separator", "Separator").Short('F').Default(" ").String()
	var noLogs = app.Flag("no-logs", "No output redis command").Bool()

	app.Version("0.1.0")

	kingpin.MustParse(app.Parse(os.Args[1:]))

	db, err := openDB(*dbuser, *dbpass, *dbhost, *dbname, *dbsock, *dbport)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	redisClient, err := openRedis(*redisHost, *redisSock, *redisDB, *redisPort, *redisPass)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	rows, err := db.Query(*query)
	if err != nil {
		log.Fatal(err)
	}
	cols, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}

	colNames := make(map[string]struct{})
	for _, col := range cols {
		colNames[col] = struct{}{}
	}

	b := make([][]byte, len(cols))

	row := make([]interface{}, len(cols))
	for i, _ := range b {
		row[i] = &b[i]
	}

	tpl, err := template.New("").Option("missingkey=error").Parse(*redisCmdArgs)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		if err := rows.Scan(row...); err != nil {
			log.Fatal(err)
		}

		values := make(map[string]string)
		for i, val := range b {
			values[cols[i]] = string(val)
		}

		iargs, args, err := redisArgs(tpl, values, *separator)
		if err != nil {
			log.Fatal(err)
		}

		if !*noLogs {
			log.Println(fmt.Sprintf("%s %s", *redisCmd, args))
		}

		_, err = redisClient.Do(*redisCmd, iargs...)
		if err != nil {
			log.Fatal(err)
		}
	}
}
