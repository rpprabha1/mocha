package reply

import (
	"fmt"
	"github.com/vitorsalgado/mocha/mock"
	"math/rand"
	"net/http"
)

type RandomReply struct {
	replies []mock.Reply
}

func Random() *RandomReply { return &RandomReply{replies: make([]mock.Reply, 0)} }

func (mr *RandomReply) Add(reply ...mock.Reply) *RandomReply {
	mr.replies = append(mr.replies, reply...)
	return mr
}

func (mr *RandomReply) Then(reply mock.Reply) *RandomReply {
	mr.Add(reply)
	return mr
}

func (mr *RandomReply) Build(r *http.Request, m *mock.Mock) (*mock.Response, error) {
	size := len(mr.replies)
	if size == 0 {
		return nil,
			fmt.Errorf("you need to set at least one response when using multiple response builder")
	}

	index := rand.Intn(len(mr.replies)-1) + 0
	reply := mr.replies[index]

	return reply.Build(r, m)
}
