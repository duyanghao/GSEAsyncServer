package handle

import (
	"Configuration"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	PROCESS_MAX_TIMEOUT time.Duration = time.Second * 30
	WORK_CHANNEL_LEN    int           = 1000
)

type Message struct {
	Msg string `json:"msg"`
}

type TaskChan struct {
	MessageTask Message
	ErrChan     chan error
}

type Task struct {
	client *http.Client
	db     *sql.DB
	config *Configuration.MysqlConfig
}

type TaskWork struct {
	taskClient *Task
	taskChan   chan TaskChan
	queueChan  chan int
	sync.RWMutex
}

func NewTaskWork(c *Configuration.MysqlConfig) (*TaskWork, error) {
	//create sql.DB
	addrs := c.Addrs
	port := c.Port
	username := c.Username
	password := c.Password
	database := c.Database
	connMaxLifetime := c.ConnMaxLifetime
	maxIdleConns := c.MaxIdleConns
	maxOpenConns := c.MaxOpenConns
	//connect to the database
	par := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&timeout=30s", username, password, addrs, port, database)
	db, err := sql.Open("mysql", par) //第一个参数为驱动名
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to mysql: %s", err)
	}
	//ping the mysql
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Failed to ping mysql: %s", err)
	}
	//set db

	//reuse the connection forever(Expired connections may be closed lazily before reuse)
	//If d <= 0, connections are reused forever.
	db.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Hour)

	//SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	//If n <= 0, no idle connections are retained.
	db.SetMaxIdleConns(maxIdleConns)

	//SetMaxOpenConns sets the maximum number of open connections to the database.
	//If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than MaxIdleConns, then MaxIdleConns will be reduced to match the new MaxOpenConns limit
	//If n <= 0, then there is no limit on the number of open connections. The default is 0 (unlimited).
	db.SetMaxOpenConns(maxOpenConns)

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 200,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	task := &Task{
		client: client,
		db:     db,
		config: c,
	}

	return &TaskWork{
		taskClient: task,
		taskChan:   make(chan TaskChan),
		queueChan:  make(chan int, WORK_CHANNEL_LEN),
	}, nil
}

func (tw *TaskWork) Run() error {
	defer func() {
		close(tw.taskChan)
		close(tw.queueChan)
		tw.taskClient.db.Close()
	}()
	for task := range tw.taskChan {
		tw.queueChan <- 1
		go func(task TaskChan) {
			// handle task
			CustomizeLog(2, fmt.Sprintf("Handle task: %+v", task))
			err := tw.work(task)
			if err != nil {
				CustomizeLog(0, fmt.Sprintf("Async task: %+v error: %s", task, err))
			} else {
				CustomizeLog(2, fmt.Sprintf("Successfully async task: %+v", task))
			}
			glog.V(5).Infof("\n=======================分割线======================\n")

			<-tw.queueChan

		}(task)
	}

	return nil
}

func (tw *TaskWork) work(task TaskChan) error {
	CustomizeLog(2, fmt.Sprintf("Message: %s received ...", task.MessageTask.Msg))
	// Do whatever you want to do(eg: send http request using tw.taskClient.client or interact with mysql using tw.taskClient.db) ...
	time.Sleep(1 * time.Minute)
	return nil
}

func (tw *TaskWork) AsyncTask(task TaskChan) error {
	select {
	case tw.taskChan <- task:
		return nil
	case <-time.After(PROCESS_MAX_TIMEOUT):
		return fmt.Errorf("Async task %+v timeout", task)
	}
}
