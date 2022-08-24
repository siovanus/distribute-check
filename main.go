/**
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/polynetwork/distribute-check/http/restful"
	"github.com/polynetwork/distribute-check/listener"
	"github.com/polynetwork/distribute-check/log"
	"github.com/polynetwork/distribute-check/store"
	"os/signal"
	"syscall"
	"time"
)

var zionRpc string
var port uint64
var caseNum uint64

func init() {
	flag.StringVar(&zionRpc, "zion", "", "zion rpc endpoint")
	flag.Uint64Var(&port, "port", 0, "server rest port")
	flag.Uint64Var(&caseNum, "case", 0, "case number")
	flag.Parse()
}

func main() {
	log.InitLog(log.InfoLog, "./Log/", log.Stdout)

	db, err := store.ConnectToDb(fmt.Sprintf("postgresql://postgres:123456@localhost:5432/zion_%d?sslmode=disable", caseNum))
	if err != nil {
		log.Errorf("store.ConnectToDb error: %s", err)
		return
	}

	l := listener.New(zionRpc, db)
	err = l.Init()
	if err != nil {
		log.Errorf("listener.Init error: %s", err)
		return
	}
	restServer := restful.InitRestServer(l, port)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go restServer.Start()
	go checkLogFile()
	l.Listen(ctx)
}

func checkLogFile() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.CheckRotateLogFile()
		}
	}
}
