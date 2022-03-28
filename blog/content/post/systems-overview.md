---
author: Sriram Panyam
title: Systems Overview
date: 2020-01-15
description: A brief of overview of systems and how to think about composition.
math: true
draft: true
---

It is hard to imagine a world where we do not interact with complex systems and platforms.   Twitter, Facebook, Tiktok, Whatsapp - they are all around us.  While the complexity is fascinating, viewing these systems in a more system lens would reveal a lot about identifying, managing and evolving their complexity.


Systems perform work.   A blander explanation is hard to find.   Let us add some life to this.   Systems perform work by taking inputs and generating outputs.   Here is a system:

```
TODO - A block diagram showing inputs, a calculator block and an output
```

This system takes a bunch of numbers as input and returns a number as an output.   This system is not very useful.  It is unclear in its purpose and hard to reason about what it actually does.  Is it a random number generator?  Is it a weather monitor?  Is it a calculator?


Let us flesh out this system by making it a basic calculator.  Even for a calculator a stream of numbers as input is less useful.   We could give the calculator methods so give its operations meaning.  As a basic calculator, we would expect it to perform some arithmetic - perhaps addition and subtraction.

```
TODO - Drawing a calc showing with a block showing two "methods" - add and subtract
```

This system is becoming more purposeful now.  Dedicated calculators (ie not your smartphone) originally were built from [Adder/Subtractor](https://en.wikipedia.org/wiki/Adder%E2%80%93subtractor) circuits.  But for our purposes we can avoid the hardware waters and assume that the adder and subtractor are two distinct logical systems "hosted" by the calculator.  ie the calculator receives commands and (perhaps a "router" component) forwards the command its arguments/operands to the right subsystem (adder/subtractor).


```
TODO - Show an image of a calculator with 2 API bits connected to a router wich forwards requests to internal adder/subtractor
```

Basic arithmetic is more than simply addition and subtraction.  How about multiplication and division.   Let us add multiplier and divider components/sub-systems to the calculator and update the router accordingly.  Since we are on a time-crunch, we implement multiplication and addition as repeated addition and subtraction respectively:

```
TODO - System showing calculator, router, and mult/div that call the adder/subtractor respectively.
```

We can continue down this path and add more methods/functions as functions of the existing functions (addition, subtraction, multiplication and division).   Some examples are powers, exponents, logarithms, roots or even trignometric functions (eg sin, cos, tan etc).

```
TODO - show a diagram of calc with extra functiosn calling previously built functions
```

At some point in time we realize that performing higher level functions interms of lower level functions is not very efficient.  We can replace this with specialized components.  This could take the form of specialized assembly instructions or hardware for acceleration (eg custom [multiplier](https://en.wikipedia.org/wiki/Binary_multiplier)/[trignometric](https://en.wikipedia.org/wiki/CORDIC) circuits).


```
TODO - Show the system with specialized modules
```

By now, our calculator has evolved from a basic to a scientific one.   Though the mathematical functions provided by a calculator may have peaked our use cases have evolved.   It is quite tedious to repeat calculations (or resort to good old fashioned pen/paper) in larger equations.   Some form of (limited) memory to store intermediate results would be useful.  Additionally we storing basic formulas (or user programmable equations with variables) seems highly desirable.   Here we can add memory and equation components to our system.  

```
TODO - Show the calc with RAM + Eqn modules too
```

Our calculator is getting more complex.   An item of interest here is that the calculator's mathematical functions (sub-systems) are quite isolated.  The equation module calls upon math functions to evalute values.  The user saves results of the math functions into memory or passes values from memory into the functions.   Intuitively the math function can be grouped and packaged as a seperate sub-system much like the calculator had humbly began! 

```
TODO - Show calc with math processor as a sub-system
```

Let us take a pause here.   More and more features can be added to our calculator (a screen/display for graphing equations, a hand-writing recognition system for easier input of math, connectivity to share your math with your friends).  

What is of note here is that smaller and simpler subsystems are crucial to building and evolving systems.  This approach has several benefits:

* Modularity:  Systems that are modular can be built and tested in isolation for specific guarantees and be plugged into larger systems that need the sub-system's functionality.  Additionally modular systems (or modules) can be interchangeable as long as the meet the guarantees of their specifications.
* Ability to reason: Systems can be reasoned about on their behaviours, performance characteristics, reliablity and costs in a uniform and consistent way.  This in turn influences our ability to reason about the behaviours up the system's hierarchy.
* Simplicity from abstractions: Systems can be understood because we can reason at a higher level of abstractions.  A calculator is easier to understand as a system enabling "math functions with memory and display" rather than a system with "adders, multipliers, dividers, sin, cos with graphing abilities and memory".

This modularity and composability will be a key theme through out the blog as we look at some of the most popular systems we interact with daily and their evolution from simple systems (hosted from a single server) to massively distribute systems spanning several regions globally serving billions of users each day.
