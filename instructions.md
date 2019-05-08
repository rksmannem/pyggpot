# Table of Contents

1.  [Introduction](#org2356a34)
    1.  [First, A Spec](#org8b3d19b)
2.  [Step 1: Build Pyggpot](#orgbcc504f)
3.  [Step 2: Understand what you've built](#org1758c95)
    1.  [Walkthrough](#org59f912a)
        1.  [Schema](#org0b3cccd)
        2.  [Service contract, aka Protobuf schema](#org6ccd965)
        3.  [Service providers](#org82642fd)
        4.  [Main](#orgade0616)
4.  [The Assignment](#org8c11443)
    1.  [Fix a validation error message](#org86adf10)
    2.  [Code the RemoveCoins service](#org4146bca)
    3.  [Document a better design to support the RemoveCoins service](#org7ec16c5)
5.  [I'm Done!](#org8297943)

<a id="org2356a34"></a>

# Introduction

The point of this example project assignment is to challenge you with
the kind of code and design challenges we developers face daily. A
common first step is figuring out some else's code, getting it to
work, and fixing or adding features.

While there's no hard time limit, we think this project will take most
developers with some Golang experience 45 to 90 minutes. We certainly
don't want to tie you up for more than 90 minutes, unless it's for
your own extracurricular purposes. Please let us know a rough estimate
on your time spent to get through this work and answer the questions
at the end.

<a id="org8b3d19b"></a>

## First, A Spec

The Pyggpot service at <https://github.com/aspiration-labs/pyggpot>
emulates a piggy bank. The service supports the following actions

- Create a Pot with a name and a capacity of some number of coins
- Add gold, silver, bronze, or lead Coins to the Pot. All coins are
  the same size. Gold are worth 100, silver 10, bronze 1, and lead 0.
- Remove some number of Coins from the Pot. You shake the Pot upside
  down and Coins come out in a uniform random distribution.

A pot is opaque, so it's not possible to see inside how many or what
kind of coins are in a pot. Plus, we have a very bad memory, and
always forget what we put in the pot.

<a id="orgbcc504f"></a>

# Step 1: Build Pyggpot

The top level README.md in <https://github.com/aspiration-labs/pyggpot>
should cover building and running the code. OSX and Ubuntu Linux are
currently supported. Most other Linuxes should work, but haven't been
tested. Windows is not yet supported.

If you get to the point of a working Swagger site at
<http://localhost:8080/swaggerui/> you're ready for the next step. If
you believe there's a bug in the build, feel free to email with repro
details.

<a id="org1758c95"></a>

# Step 2: Understand what you've built

You should be able to work through this assignment without significant
knowledge of the tooling details. If you have worked with [Google
Protocol Buffers](https://developers.google.com/protocol-buffers/), [Twirp](https://github.com/twitchtv/twirp), and/or [xo](https://github.com/xo/xo/) then you're way ahead of the
game. However, the makefile should take care of that machinery, and
you just need to understand some sql, regular expressions, and go code
to push-on through.

<a id="org59f912a"></a>

## Walkthrough

Usually when you get thrown into a new code, you get some kind of
walkthrough from another developer. Sometimes it's even
helpful. That's the intent of this section.

<a id="org0b3cccd"></a>

### Schema

Peruse `sql/schema.sqlite3.sql`. It's a simple two table schema, many
coins to each pot. A Coin record is some number of one kind of coin.

<a id="org6ccd965"></a>

### Service contract, aka Protobuf schema

Look at `proto/coin/service.proto` and
`proto/pot/service.proto`. These two specs define our service API,
request and response formats. The `message` schema will be compiled
into native Golang data types, which is all you really need to focus
on.

After building you'll find the generated code in `rpc/go/coin` and
`rpc/go/pot`. You don't need to fully grok all the generated code, but
worth knowing that the Go data types for protobuf messages are in the
`service.pb.go` files. The `service.twirp.go` holds the interface spec
that needs to be implemented for the service.

<a id="org82642fd"></a>

### Service providers

For this project we refer to our service implementations as
"providers", which you'll find in internal/providers/&#x2026; You can
review the pot and coin provider in the `provider.go` files. You
should note, for example, a `type Pot interface` from
rpc/go/pot/service.twirp.go is implemented here for a
`potServer`. This interface implementation, `potServer` in this case,
also holds useful state data for the service, like database
connections.

<a id="orgade0616"></a>

### Main

Everything is initialized, wired up, and the service listener started
in `cmd/server/main.go`

<a id="org8c11443"></a>

# The Assignment

At this point you should have Pyggpot built and running, and perhaps
have thrown a few curls at it directly or via the swaggerui. Here's
your todo list.

<a id="org86adf10"></a>

## TODO Fix a validation error message

If I

    curl -X POST "http://localhost:8080/twirp/pyggpot.pot.Pot/CreatePot" \
      -H "accept: application/json" -H "Content-Type: application/json" \
      -d "{ \"pot_name\": \"PP\", \"max_coins\": 10}"

I get a 400 status back with response

    {
      "code": "invalid_argument",
      "msg": "invalid field PotName: Can contain only alphanumeric characters, dot and underscore. ",
      "meta": {
        "argument": "invalid field PotName: Can contain only alphanumeric characters, dot and underscore."
      }
    }

Similarly, `pot_name` of "PP." and "PP-", while "PPP" return 200 status
and a new Pot.

Your task: find the validation message and the rule behind it; write
an accurate error message that will avoid end user rage clicking.

<a id="org4146bca"></a>

## TODO Code the RemoveCoins service

In `internal/providers/coin/provider.go` the `RemoveCoins` function is
unimplemented. The spec is like shaking a piggy bank upside down until
coins come out. The type (denomination) of the coin is random, based
on the proportion of that coin type in the pot.

<a id="org7ec16c5"></a>

## TODO Document a better design to support the RemoveCoins service

We (royal) have decided that how coins are tracked in pots is not
ideal. In particular, the `RemoveCoins` service has to do more work
than we'd like to retrieve coins randomly.

Write a short redesign proposal which may include schema, service
spec, or provider implementation changes. You don't have to write the
code for this, but justify your redesign based on simplicity,
performance, or any other positive you find compelling.

<a id="org8297943"></a>

# I'm Done!

The easist way to respond is

    git add . && git diff --cached >MyPyggpotMods.diff

and email the diff file back to whoever sent you these instructions.

Tia!
