// Copyright 2013-3014 Adam Presley. All rights reserved
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package mailslurper

import (
	"log"
	"net"
	"time"

	"github.com/adampresley/webframework/sanitizer"
)

/*
ServerPool represents a pool of SMTP workers. This will
manage how many workers may respond to SMTP client requests
and allocation of those workers.
*/
type ServerPool chan *SMTPWorker

/*
JoinQueue adds a worker to the queue.
*/
func (pool ServerPool) JoinQueue(worker *SMTPWorker) {
	pool <- worker
}

/*
Create a new server pool with a maximum number of SMTP
workers. An array of workers is initialized with an ID
and an initial state of SMTP_WORKER_IDLE.
*/
func NewServerPool(maxWorkers int) ServerPool {
	xssService := sanitizer.NewXSSService()
	emailValidationService := NewEmailValidationService()

	pool := make(ServerPool, maxWorkers)

	for index := 0; index < maxWorkers; index++ {
		pool.JoinQueue(NewSMTPWorker(
			index+1,
			pool,
			emailValidationService,
			xssService,
		))
	}

	log.Println("libmailslurper: INFO - Worker pool configured for", maxWorkers, "worker(s)")
	return pool
}

/*
NextWorker retrieves the next available worker from
the queue.
*/
func (pool ServerPool) NextWorker(connection net.Conn, receiver chan MailItem) (*SMTPWorker, error) {
	select {
	case worker := <-pool:
		worker.Prepare(
			connection,
			receiver,
			SMTPReader{Connection: connection},
			SMTPWriter{Connection: connection},
		)

		log.Println("libmailslurper: INFO - Worker", worker.WorkerID, "queued to handle connection from", connection.RemoteAddr().String())
		return worker, nil

	case <-time.After(time.Second * 2):
		return &SMTPWorker{}, NoWorkerAvailable()
	}
}
