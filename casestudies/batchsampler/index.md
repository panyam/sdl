
# Design A GPU Inference Platform

We have a system that can route prompts to GPUs for completion.

For example the user might submit:

```
/inference/?completion="The three"
```

With the service returning "muskeeteers".

The basic way of doing this is for each of these API calls to going to a GPU that say already has models loaded.  The
GPU would then perform inference and return a response (we can assuming all vector embeding to/fro conversation is taken
care of).

---

### Functional Requirements (~2-3 minutes)

Our API is simple:

```
service Completion {
  Complete(prompt string) string {
  }
}
```

We will skip NFR and OOS requirements for now.

## High Level Design (~ 5-10 minutes)

We have the following components:

```
APIServer {
  g = GPU()
  Complete(string) string {
    return g.Infer(input)
  }
}

GPU {
  Infer(embeding) embeding {
    Delay(100ms)
  }
}
```

Pretty simple.

1. API Server gets a request (say through a load balancer)
2. It has access to a GPU.
3. Calls the GPU to infer and sends the response back.

The GPU can perform an inference in 100ms.

## Scale it

Now for the next part of the problem.   We are given 2 more facts:

1. The API Server gets 20k QPS
2. The GPU can take a batch request of atmost 100 and perform parallel inference/sampling of those (atmost) 100
   completions.
   
Eg it can do:

```
GPU {
  Infer(inputs [100]string) (completions [100]string) {
    // still take atmost 100ms
    Delay(100ms)
  }
}
```

Now our API server to serve 20k qps needs to have access to 2000 GPUs (since each GPU can handle 10 inferences a second)
since each request is being handled right away.  Instead, what we need is batching:

```
APIServer {
  g = GPU()
  
  Loop {
    batch = []
    for {
      select {
        case prompt, respChan := <- reqChan:
          batch.append((prompt, respChan))
          break
        case timer400ms:
          prompts = [b[0] for b in batch]
          resp = g.Infer(prompts)
          for i, (prompt, respChan) in enumerate(batch):
            respChan <- resp[i]
          break
      }
    }
  }
  
  Complete(reqId, prompt) string {
    respChan := make chan for response
    reqChan <- reqId, prompt, respChan
    resp := <- respChan
    close(respChan)
    return resp
  }
}
```

Here with a single GPU - we are sending 100 at a time - and we just wait for 400ms before sending the batch (the user is
ok to wait 500ms for their inference).

This means suddenly intead of an API server just hadning 10 requests - it can server 1000 requests per second.  Now all
we need is 20 GPUs.  A single API server is ok to do this batching - but can break down to 2 or 4 servers for reducing #
open connections.

How can we model this in SDL?  Synchronous processes are easy but async loops and comms between them is tricky.

Couple of options:

1. Is this even necessary? - This is important because we want to be able to describe a system where requests are
   "parked" with possible error modes that can arise and be modelled.
2. Can Loop and Complete be treated as just APIs that call each other? - This may be a bit too hacky?
3. Do we need another construct for hand-off between threads?

Some constructs are:

1. Parked Requests
2. Threads
3. Signals (between threads)
4. Request Identifiers
5. Streams
6. Spawn, Yield and Waits

* In this case we got a request.  Instead of serving it straight away, we parked it in the bather loop (via a channel)
* So perhaps just use a "Park" keyword?  Remember we are not modelling correctness - just SLOs

```
// Loopers are the run loops doing things in the background
Looper {
  OnRequest:
    batch.add(request)
  Every 400 MS:
    g.Infer(batch)
    batch = []
}

Complete(prompt) string {   // note no mention of request
  parkRequest(#request, Loop)
}
```

This seems more hacky.  We know that a batch of 100 can happen at a time over a window of 400ms.  Instead can we have
the concept of queues:

```
APIServer {
  @queue(maxTime = 400ms)     // <--- This queues each request and lets Complete handle a "batch"
                              // when the queue is full - the full trigger could be a max time or
                              // max requests
                              // request should be of type "batch"
  Complete(batch<prompt>) response {
  }
}
```
