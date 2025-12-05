
## Modelling Parallel Execution

A big of part of system modelling is async/parallel calls to services.  For example we may write to 5 services in
parallel and either forget about it.   A few ways to model this are:

### `go` keyword to start a green thread or a future

```
x = db.read()
y = log.write()

async {
  // creates a new binding of vars
  x = db.write()
  c = cache.put()
  stmt2
  ...
}

```

In the above example the result of future are lost because they were fire and forget

We need to be able to "wait", perhaps we can do:

```
x = db.read()
y = log.write()

go z = {    // started at T = 5
  // creates a new binding of vars
  x = db.write()
  c = cache.put()
  stmt2
  ...
}                 // Assume this takes 20 seconds

// do other work taking 8 seconds

// wait for the parallel to finish - at this point since z took 20 seconds and 8 seconds already elapsed,
// here T = 5 + 20 (NOT 5 + 8 + 20)
// Also Z will a Outcome[T] where T will be inferred based on the types in the future's body
wait z      

Note all variables inside the parallel block are not visible here
```

You could wait for multiple parallels in a single wait

In this does the "any"/"all" succeeded even matter?  May be not.

The parallel returns an Outcome just like a normal Expr where the underlying type is just another union of sets of
outcomes.  So z can be used as is - if there is no "return" then just the delay is used otherwise types can be used.

Using futures in assignents.   Instead of creating futures statements you could also do:

```
x = go Expr1
y = go Expr2
z = go Expr3
```

and then wait for them later on:

```
wait x, y, z
```

What would happen if we did not wait but used it?   Couple of ways to handle this:

1. Firstly of all an identifier have to be used "somewhere" so the first time it is invokved we can see if it was a
   future and check if it was waited on (have a IsFuture and WaitDone flags)
2. Throw an error if is future and not waited on
3. Option2 - force a wait when it is first used.  This would ensure that until an identifer is actually referenced, it
   wont be waited on - makes the decl much easier for the user too - less verbose.

Both options need us to have to implement the underlying wait - just that whether expose as a keyword or not is the
issue.   Let us go with Option1 now so it is more explicit for the user.  User decides when/where they want to wait.

