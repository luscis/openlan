package database

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/ovn-org/libovsdb/cache"
	"github.com/ovn-org/libovsdb/client"
	"github.com/ovn-org/libovsdb/model"
	"github.com/ovn-org/libovsdb/ovsdb"
	"log"
	"os"
)

type OvSDB struct {
	client.Client
	Ops []ovsdb.Operation
}

func (o *OvSDB) Context() context.Context {
	return context.Background()
}

func (o *OvSDB) Switch() (*Switch, error) {
	var listSw []Switch
	if err := o.List(&listSw); err != nil {
		return nil, err
	}
	if len(listSw) == 0 {
		return nil, libol.NewErr("hasn't switch")
	}
	return &listSw[0], nil
}

func (o *OvSDB) Execute(ops []ovsdb.Operation) {
	o.Ops = append(o.Ops, ops...)
}

func (o *OvSDB) Commit() ([]ovsdb.OperationResult, error) {
	ops := o.Ops
	o.Ops = nil
	return o.Client.Transact(o.Context(), ops...)
}

func (o *OvSDB) Get(m model.Model) error {
	return o.Client.Get(o.Context(), m)
}

func (o *OvSDB) Transact(ops ...ovsdb.Operation) ([]ovsdb.OperationResult, error) {
	return o.Client.Transact(o.Context(), ops...)
}

func (o *OvSDB) List(result interface{}) error {
	return o.Client.List(o.Context(), result)
}

func (o *OvSDB) WhereList(predict interface{}, result interface{}) error {
	return o.Client.WhereCache(predict).List(o.Context(), result)
}

var Client *OvSDB

type DBClient struct {
	Server   string
	Database string
	Verbose  bool
	Client   *OvSDB
}

func (c *DBClient) Context() context.Context {
	return context.Background()
}

func (c *DBClient) NilLog() *logr.Logger {
	// create a new logger to log to /dev/null
	writer, err := libol.OpenWrite(os.DevNull)
	if err != nil {
		writer = os.Stderr
	}
	l := stdr.NewWithOptions(log.New(writer, "", log.LstdFlags), stdr.Options{LogCaller: stdr.All})
	return &l
}

func (c *DBClient) Open(handler *cache.EventHandlerFuncs) error {
	server := c.Server
	database := c.Database
	dbModel, err := model.NewClientDBModel(database, models)
	if err != nil {
		return err
	}
	ops := []client.Option{
		client.WithEndpoint(server),
	}
	if !c.Verbose {
		ops = append(ops, client.WithLogger(c.NilLog()))
	}
	ovs, err := client.NewOVSDBClient(dbModel, ops...)
	if err != nil {
		return err
	}
	if err := ovs.Connect(c.Context()); err != nil {
		return err
	}
	Client = &OvSDB{Client: ovs}
	c.Client = Client
	if handler != nil {
		processor := ovs.Cache()
		if processor == nil {
			return libol.NewErr("can't get cache.")
		}
		processor.AddEventHandler(handler)
	}
	if _, err := ovs.MonitorAll(c.Context()); err != nil {
		return err
	}
	return nil
}

var Conf *DBClient

func NewDBClient(handler *cache.EventHandlerFuncs) (*DBClient, error) {
	var err error
	if Conf == nil {
		Conf = &DBClient{
			Server:   api.Server,
			Database: api.Database,
			Verbose:  api.Verbose,
		}
		err = Conf.Open(handler)
	}
	return Conf, err
}
