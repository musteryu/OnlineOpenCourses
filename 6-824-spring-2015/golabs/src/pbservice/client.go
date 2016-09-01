package pbservice

import "../viewservice"
import "net/rpc"
import "fmt"

import "crypto/rand"
import "math/big"
import "time"

type Clerk struct {
	vs *viewservice.Clerk
	// Your declarations here
	currentView viewservice.View
}

// this may come in handy.
func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(vshost string, me string) *Clerk {
	ck := new(Clerk)
	ck.vs = viewservice.MakeClerk(me, vshost)
	// Your ck.* initializations here
	ck.currentView = viewservice.View{}
	return ck
}

//
// call() sends an RPC to the rpcname handler on server srv
// with arguments args, waits for the reply, and leaves the
// reply in reply. the reply argument should be a pointer
// to a reply structure.
//
// the return value is true if the server responded, and false
// if call() was not able to contact the server. in particular,
// the reply's contents are only valid if call() returned true.
//
// you should assume that call() will return an
// error after a while if the server is dead.
// don't provide your own time-out mechanism.
//
// please use call() to send all RPCs, in client.go and server.go.
// please don't change this function.
//
func call(srv string, rpcname string,
	args interface{}, reply interface{}) bool {
	c, errx := rpc.Dial("unix", srv)
	if errx != nil {
		return false
	}
	defer c.Close()

	err := c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

//
// fetch a key's value from the current primary;
// if they key has never been set, return "".
// Get() must keep trying until it either the
// primary replies with the value or the primary
// says the key doesn't exist (has never been Put().
//
func (ck *Clerk) Get(key string) string {
	// Your code here.
	args := &GetArgs{}
	args.Key = key
	var value string
	for true {
		var reply GetReply
		ok := ck.currentView.Primary != "" && call(ck.currentView.Primary, "PBServer.Get", args, &reply)
		if !ok || reply.Err == ErrWrongServer {
			ck.currentView, _ = ck.vs.Get()
		} else {
			if reply.Err == OK {
				value = reply.Value
			} else {
				value = ""
			}
			break
		}
		time.Sleep(viewservice.PingInterval)
	}
	return value
}

//
// send a Put or Append RPC
//
func (ck *Clerk) PutAppend(key string, value string, op string) {
	// Your code here.
	args := &PutAppendArgs{}
	args.Key = key
	args.Value = value
	args.Op = op
	args.Id = nrand()
	for true {
		var reply PutAppendReply
		ok := ck.currentView.Primary != "" && call(ck.currentView.Primary, "PBServer.PutAppend", args, &reply)
		if !ok || reply.Err != OK {
			ck.currentView, _ = ck.vs.Get()
		} else {
			break
		}
		time.Sleep(viewservice.PingInterval)
	}
}

//
// tell the primary to update key's value.
// must keep trying until it succeeds.
//
func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}

//
// tell the primary to append to key's value.
// must keep trying until it succeeds.
//
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "Append")
}
