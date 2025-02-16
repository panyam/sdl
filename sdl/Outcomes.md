# Modelling Outcome Distributions

We need a way to model distribution of Outcomes
Idea is for any "method" on a component we are not going to be returning actual results.
Instead we will only returning distribution of outcomes.   This is because we are only
modelling SLOs for various operations and not functional correctness.  We want the following
features from our modelling:

1. Ease of adding/removing/re-distributing weights across outcomes
2. Make it easy to model code paths via union/intersection/negation of decision subtrees
3. We want our component "behavior" code to look as natural to writing "normal" code in achieving (2)

## Option 1 - List of (Outcome, Weight) pairs

For the  disk read operations that has a success of 99.8% and .2% failure rates, with further latency
distribution on each of these as 90%, 9% and .8% for 1, 10 and 1000 ms (for success) and 0.1% and 0.1%
with 10 and 50ms on failures as:

```
Outcomes : [
  (900, ("Success", "1ms"))
  (90, ("Success", "10ms"))
  (8, ("Success", "1000ms"))
  (1, ("Failure", "10ms"))
  (1, ("Failure", "50ms"))
]
```

This is pretty simple and the advantage is for any "leaf" outcome we can see what its final probability is, eg Failure of 10ms = 1/1000. 

## Option 2 - As map[Outcome][[]WeightedValue]

For the same scenario above, we would have:

```
 Outcomes : {
  "Success": {
		"1ms": 900,
		"10ms": 90,
		"1000ms": 8,
	},
  "Failure": {
		"10ms": 1,
		"50ms": 1,
	}
}
```

This is a bit more complicated but easier to model as each node is just a map of "sub" outcomes.  But it is a bit confusing if the leaf weights are dependent on the parent or not.  For example here it is required that the probabilities of the Failure cases (10 and 50ms) are relative to "overall" probabilities.   It would be easier for modelling if the weights are specified "locally" but somehow calculated globally implicitly.

Also this option(like the previous one) makes it harder to compute outcomes in a more generic way.  For example we may be interested in all calls whose latencies are <= 10 instead of just filtering based on Success or Failure.   This would put the burden of this slicing on the user.

Option 3: Weights and Attribs

Instead of giving Outcomes a special meaning, we could perhaps treat the value as a map of attributes which would make filtering very easy. 

eg:

```
Outcome : [
  (900, {Outcome: "Success", Latency: "1ms"})
  (90, {Outcome: "Success", Latency: "10ms"})
  (8, {Outcome: "Success", Latency: "1000ms"})
  (1, {Outcome: "Failure", Latency: "10ms"})
  (1, {Outcome: "Failure", Latency: "50ms"})
]
```

This let us partition our outcomes based on filtering by attribtues instead of just labels so our caller can decide how to proceed.
