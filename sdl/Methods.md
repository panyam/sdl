
# Modelling Probabilistic Executions

We currently have Outcomes that model probabilistic distributions of various outcomes of primitive operation.

However we want to start composing these.  For example a top level operation could be a tree of multiple method calls with loops, conditions nd so on.   For example consider the following method:

```
fund SaveTweet(tweet) Outcomes{
  o1 := database.ReadTweetById()
  if o1.Code == "404" {
    // tweet = generated id
    o2 := idgen.GenId()
    
    if o2.Error == false {
      o3 := data.SaveTweetById(tweet)
    }
  }
  return o4 <- ???
}
```

We have 3 "probablistic" calls (o1, o2 and o3) and finally our method needs to return o4 - its own distribution of errors and latencies etc.

When we model our componewnts we want code to reflect "functional" behavior atleast logically (we dont care about actual values beign passed) while also providing strong modelling of SLOs.

What is needed is some kind of "tracker" that tracks branches, loops, returns and calls. So in our lib we would have:

```
fund SaveTweet(tracker) Outcomes {
  tracker.Call(database.ReadTweetById())
  // if we return here then distribution of SaveTreet is just o1
  // but we want to go foward
  
  tracker.ReturnIf(func(ch => ch.Code == "404"))
  // its like if we return here, then our return is same as O1 we no other "branches" have been pushed - they were only started
  
  // Now somehow trakcer should have 2 branches - code == 404 and != 404 which should be this branch
  tracker.Call(idgen.GenId())
  
  tracker.With(func(ch Choice) => if "Latency" is an attribute, ch.Latench += ProcessingTime)
  
  tracker.Call(data.SaveTweetById(tweet))
  
    // tweet = generated id
    o2 := idgen.GenId()
    
    if o2.Error == false {
      o3 := data.SaveTweetById(tweet)
    }
  return o4 <- ???
}

// Exp := Literal | Dist
//    | ExpList
//    | "busy" Exp
//    | "if" Exp Exp
//    | "if" Exp Exp "else" Exp
//    | Exp "(" Exp1, Exp2 ... ExpN ")"
//    | "for" Literal Do Exp    <- Literal can become an Exp later?
//
// FunDecl := "fun" Ident(params) ExpList 

// ExpList := "{" Exp* "}"

// So above becomes
func SaveTweet(tweet) Outcomes{
  o1 := database.ReadTweetById()
  if o1.Code == "404" {
    // tweet = generated id
    o2 := idgen.GenId()
    
    if o2.Error == false {
      o3 := data.SaveTweetById(tweet)
    }
  }
  return o4 <- ???
}
```


