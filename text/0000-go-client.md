- Start Date: 2019-03-28
- RFC PR:
- Prisma Issue:

# Summary

This spec details how to generate Prisma's client for the Go language
in a idiomatic way that looks handwritten.

# Motivation

The current Go client too verbose and doesn't look like how a normal Go
developer would write Go code.

The goal of this RFC is to refactor the Go client to look like handwritten Go code as if a developer had unlimited time to dream up the best API for their models interacting with their database of choice.

We should also be able to extend this client as we come up with new ways
of expressing relations and incorporate more databases.

# Detailed design

## Background

The basic design of this Go client is informed by 6 iterations I did last year while developing a type-safe PostgreSQL client for [Standup Jack](https://standupjack.com).

I built this client while simultaneously addressing the data needs of the app. In other words, if the app needed a feature, I added it. If it felt wrong in the app, I rethought the API.

The Prisma client has a much more ambitious API that spans multiple databases and languages. I'm breaking a lot of new ground in this RFC, particularly around nested operations. Everything in this RFC ought to be questioned and debated.

The goal is to have a memorable, simple API that is a delight to use for Go developers. Let's find a design together that charms!

## Generated API

The layout of this detailed design is:

- method name
- method description
- currently proposed typescript approach
- proposed Go approach
- underlying graphql query
- additional notes _(if needed)_

### Find One

#### Find By Primary Key

Find a single resource by its `primary key`.

##### Current Typescript API:

```ts
const bob: User = await prisma.users.findOne('bobs-id')
```

##### Proposed Go API:

```go
bob, err := users.FindByID(db, "bobs-id")
```

##### Underlying graphql query:

```gql
# todo
```

##### Additional Notes:

- todo

#### Find By Unique Constraint

Find a single resource with a `unique(email)` constraint.

##### Current Typescript API:

```ts
const alice: User = await prisma.users.findOne({
  where: { email: 'alice@prisma.io' }
})
```

##### Proposed Go API:

```go
alice, err := users.FindByEmail(db, "alice@prisma.io")
```

##### Underlying graphql query:

```gql
# todo
```

##### Additional Notes:

- todo

#### Find By a Composite Unique Constraint

Find a single resource with a `unique(first_name, last_name)` constraint.

##### Current Typescript API:

```ts
const john: User = await prisma.users.findOne({
  name: { firstName: 'John', lastName: 'Doe' }
})
```

##### Proposed Go API:

```go
john, err := users.FindByFirstNameAndLastName(db, "John", "Doe")
```

##### Underlying graphql query:

```gql
# todo
```

##### Additional Notes:

- todo

#### Find By non-unique conditions

Find a single resource by non-unique fields.

##### Current Typescript API:

```ts
const john: User = await prisma.users.findOne({
  name: { firstName: 'John', lastName: 'Doe' }
})
```

##### Proposed Go API:

```go
john, err := users.Find(db, users.Where().FirstName("John").LastName("Doe"))
```

##### Underlying graphql query:

```gql
# todo
```

##### Additional Notes:

- todo

### Find Many

#### Find with a condition

Find the all items that match a condition.

##### Current Typescript API:

```ts
const allUsers: User[] = await prisma.users.findAll({ firstName: 'John', lastName: 'Doe' })
// or
const allUsersShortcut: User[] = await prisma.users({ firstName: 'John', lastName: 'Doe' })
```

##### Proposed Go API:

```go
allUsers, err := users.FindMany(db, users.Where().FirstName("John").LastName("Doe"))
// or
// no equivalent. users.Users(...) does not exist to avoid stutter
```

##### Underlying graphql query:

```gql
# todo
```

##### Additional Notes:

- todo

#### Find with First limit

Find the first N items of a resource.

##### Current Typescript API:

```ts
const allUsers: User[] = await prisma.users.findAll({ first: 100 })
```

##### Proposed Go API:

```go
allUsers, err := users.FindMany(db, users.First(100))
```

##### Underlying graphql query:

```gql
# todo
```

##### Additional Notes:

- todo

#### Find all with Order

Find an ordered list of items matching the condition.

##### Current Typescript API:

```ts
const allUsers = await prisma.users({
  where: { firstName: 'John', lastName: 'Doe' },
  orderBy: { email: 'ASC' }
})
```

##### Proposed Go API:

```go
allUsers, err := users.FindMany(db,
  users.Where().FirstName("John").LastName("Doe"),
  users.Order().Email("ASC"),
)
```

##### Underlying graphql query:

```gql
# todo
```

##### Additional Notes:

- May also want to define constants here, e.g. `prisma.DESC` & `prisma.ASC`.

#### Find all with Composite Order

Find an ordered list of items matching the condition: `order by email ASC name DESC`

##### Current Typescript API:

```ts
const allUsers = await prisma.users({
  where: { firstName: 'John', lastName: 'Doe' },
  orderBy: [{ email: 'ASC' }, { name: 'DESC' }]
})
```

##### Proposed Go API:

```go
allUsers, err := users.FindMany(db,
  users.Where().FirstName("John").LastName("Doe"),
  users.Order().Email("ASC").Name("DESC"),
)
```

##### Underlying graphql query:

```gql
# todo
```

##### Additional Notes:

- todo

#### Find all with Nested Order

Find an ordered list of items matching the condition.

##### Current Typescript API:

```ts
const usersByProfile = await prisma.users({
  orderBy: { profile: { imageSize: 'ASC' } }
})
```

##### Proposed Go API:

```go
// todo
```

##### Underlying graphql query:

```gql
# todo
```

##### Additional Notes:

- todo

#### Find all where contains

Find all items where resource contains submatch: `where email like %@gmail.com%`.

##### Current Typescript API:

```ts
const users = await prisma.users({ where: { email_contains: '@gmail.com' } })
```

##### Proposed Go API:

```go
usrs, err := users.Find(db, users.Where().EmailContains("@gmail.com"))
```

##### Underlying graphql query:

```gql
# todo
```

##### Additional Notes:

- todo

#### Partially Typed Raw SQL

Escape hatch to handle cases where the generated API isn't expressive enough.

```sql
select * from users where email like '%@gmail.com%' order by age + postsViewCount desc
```

##### Current Typescript API:

```ts
const users = await prisma.users({
  where: { email_contains: '@gmail.com' },
  raw: { orderBy: 'age + postsViewCount DESC' }
})
```

##### Proposed Go API:

Use the database client directly, with type-safe constants generated by Prisma.

```go
results, err := db.Query(fmt.Sprintf(
  `select * from %s where %s like '%@gmail.com%' order by %s + %s desc`,
  users.Table.Name,
  users.Column.Email,
  users.Column.Age,
  users.Column.PostViewCount,
))
```

##### Underlying graphql query:

```gql
# todo
```

##### Additional Notes:

- todo

##### Fluent API

```sql
select * from posts where user_id = 'bobs-id' limit 50
```

##### Current Typescript API:

```ts
const bobsPosts: Post[] = await prisma.users.findOne('bobs-id').posts({ first: 50 })
```

##### Proposed Go API:

Use the database client directly, with type-safe constants generated by Prisma.

`FromX` would essentially enter a different Builder mode. The reason we can't use `users.FindByID(db, "bobs-id").Posts.FindMany(db, ...)` is because Go doesn't support lazy evaluation. We could have an API like this, but it would make 2 round-trip calls.

```go
// backwards
bobsPosts, err := posts.FromUserID("bobs-id").FindMany(db, posts.First(50))
// or, forwards
bobsPosts, err := users.FromID("bobs-id").Posts.FindMany(db, posts.First(50))
// or
bobsPosts, err := users.FromID("bobs-id").FindManyPosts(db, posts.First(50))
```

##### Underlying graphql query:

```gql
# todo
```

##### Fluent API (Chained)

##### Current Typescript API:

```ts
const bobsLastPostComments: Comment[] = await prisma.users
  .findOne('bobs-id')
  .post({ last: 1 })
  .comments()
```

##### Proposed Go API:

```go
// backwards
//   no API. backwards starts to get really confusing after you
//   go more than one level deep
// or, forwards
bobsLastPostComments, err := users.FromID('bobs-id').FromPost(posts.Last(1)).Comments.FindMany(db)
// or
bobsLastPostComments, err := users.FromID('bobs-id').FromPost(posts.Last(1)).FindManyComments(db)
```

##### Underlying graphql query:

```gql
# todo
```

##### Additional Notes:

- We may be able to ditch the 2nd `FromX`, e.g. `FromPost()`, though it may make things more clear regardless of if it's possible to remove.

#### Select API

```graphql
type User {
  id: String!
  name: String!
  posts: [Posts!]!
  friends: [User!]!
}

type Post {
  title: String
  comments: [Comment!]!
}

type Comment {
  id: String!
  comment: String!
}
```

##### Current Typescript API:

```ts
type DynamicResult1 = (User & {
  posts: (Post & { comments: Comment[] })[]
  friends: User[]
  best_friend: User
})[]

// Query Object API
const dynamicResult1: DynamicResult1 = await prisma.users.findOne({
  where: 'bobs-id',
  select: {
    posts: { select: { comments: true } },
    friends: true,
    best_friend: true
  }
})
```

##### Proposed Go API:

```go
var user struct {
  ID string
  Name string
  Posts []struct {
    ID string
    Title string
    Comments []struct {
      ID string
      Comment string
    }
  }
  Friends []struct {
    ID string
    Name string
  }
}

// use reflection (like json.Unmarshal)
err := users.SelectByID(db, "bobs-id", &user)
```

##### Underlying graphql query:

```graphql
query user(id: $id) {
  id
  name
  posts {
    id
    title
    comments {
      id
      comment
    }
  }
  friends {
    id
    name
  }
}
```

#### Page Info / Streaming Iterator

##### Typescript API

```ts
// PageInfo
const bobsPostsWithPageInfo: PageInfo<Post> = await prisma.users
  .findOne('bobs-id')
  .posts({ first: 50 })
  .$withPageInfo()

// Streaming data
for await (const post of prisma.posts().$stream()) {
  console.log(post)
}
```

##### Proposed Go API:

```go
// reader with defaults
bobsPostReader, err := users.FromID("bobs-id").FindManyPosts(posts.First(50)).Reader()

// or reader with config
bobsPostReader, err := users.FromID("bobs-id").FindManyPosts(posts.First(50)).Reader(&prisma.ReaderConfig{})

// Streaming row API
dec := json.NewDecoder(bobsPostReader)
for dec.More() {
  var post posts.Post
  if err := dec.Decode(&post); err != nil {
    return err
  }
}
```

##### Underlying graphql query:

```graphql
# todo
```

#### Aggregations

##### Typescript API

```ts
type DynamicResult2 = (User & { aggregate: { age: { avg: number } } })[]

const dynamicResult2: DynamicResult2 = await prisma.users({
  select: { aggregate: { age: { avg: true } } }
})

type DynamicResult3 = User & {
  posts: (Post & { aggregate: { count: number } })[]
}

const dynamicResult3: DynamicResult3 = await prisma.users.findOne({
  where: 'bobs-id',
  select: { posts: { select: { aggregate: { count: true } } } }
})
```

##### Proposed Go API:

```go
// avg(age)
var u1 []struct {
  ID string
  Name string
  Aggregate struct {
    Age struct {
      Avg float32
    }
  }
}
// reflection, like json.Unmarshal(...)
err := users.SelectMany(db, &u1)
err := users.SelectMany(db, &u1, users.Where().NameStartsWith("p"))

// nested count(*)
var u2 []struct {
  ID string
  Name string
  Posts []struct {
    Aggregate struct {
      Count int
    }
  }
}
// reflection, like json.Unmarshal(...)
err := users.SelectMany(db, &u2)
err := users.SelectMany(db, &u2, users.Where().NameStartsWith("p"))
```

#### GroupBy with Select

##### Typescript API

```ts
type DynamicResult3 = User & {
  posts: (Post & { aggregate: { count: number } })[]
}

// TODO wrong type
type DynamicResult4 = {
  lastName: string
  records: User[]
  aggregate: { age: { avg: number } }
}

const groupByResult: DynamicResult4 = await prisma.users.groupBy({
  key: 'lastName',
  having: { age: { avgGt: 10 } },
  where: { isActive: true },
  first: 100,
  orderBy: { lastName: 'ASC' },
  select: {
    records: { first: 100 },
    aggregate: { age: { avg: true } }
  }
})

const groupByResult2: DynamicResult5 = await prisma.users.groupBy({
  raw: { key: 'firstName || lastName', having: 'AVG(age) > 50' },
  select: {
    records: { $first: 100 },
    aggregate: { age: { avg: true } }
  }
})
```

##### Proposed Go API:

```go

err :=
```

## What is `db` and why pass it in each time?

`db` in the Go examples above is an instance of the `prisma.DB` interface. That interface would look like this:

```go
// DB interface for Prisma's client APIs.
//
// This will work with database/sql.DB and database/sql.Tx
type DB interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
}
```

**Advantages:**

- Generally speaking, I think this is a better design. You're layering on complexity by composition rather than by encapsulation.
- However the main reason is that some people will want to pass in a request-level `context.Context`, some will not want to bother. `db` makes `ctx` optional.
  - For example, you can pass in a context-wrapped database: `db := &DBContext{ctx, db}` if you're Google. If you're me, I prefer skipping this step.
  - Generated code becomes stateless, `goimports` (automatic imports) works better
- Easier to layer in logging: `db := &DBLogger{logger, DBContext{ctx, db}}`
  - _counterpoint_: this could be alleviated by exposing a logging hook `prisma.Log = logger` or as an option.
- Easier to mock, you just need to mock `prisma.DB`, you don't need to mock every generated function
  - _counterpoint_: this could be alleviated by also generating mockable functions)
- Transactions are now the developer's choice, no limits on how they're used.
  - _counterpoint_: may lead to more opportunities to shoot yourself in the foot, but this would be the developer's concern.
- You can query SQL with your familiar DB client directly as you need `db.Query(...)`
  - _counterpoint:_ We could expose the raw query method via `prisma.Query(...)`

# Drawbacks

- Generated API isolated from `prisma.DB`, might not be expected at first.
- It's a departure from the existing Typescript client
- It's a major breaking change to users of the existing Go client

# Alternatives

The builder pattern above could be replaced with structs. While I don't think this is quite as nice, it's what a lot of other client libraries do. Here's Stripe's client:

```go
client.Customers.New(&stripe.CustomerParams{
  Email: stripe.String("tony@jam.com"),
})
```

# Adoption strategy

> If we implement this proposal, how will existing Prisma developers adopt it? Is this a breaking change? Can we write a codemod? Should we coordinate with other projects or libraries?

**TODO**

# How we teach this

> What names and terminology work best for these concepts and why? How is this idea best presented? As a continuation of existing Prisma patterns?

> Would the acceptance of this proposal mean the Prisma documentation must be re-organized or altered? Does it change how Prisma is taught to new developers at any level?

> How should this feature be taught to existing Prisma developers?

**TODO**

# Unresolved questions

> Optional, but suggested for first drafts. What parts of the design are still TBD?

- [ ] Consider for `db.Query` (and everything "raw" access related) that models can live in different datasources
- [ ] Enable chaining syntax to traverse relations (Idea: `_.chain`)

**TODO**
