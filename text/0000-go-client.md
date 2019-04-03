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

The example code below assumes the following datamodel:

```groovy
model Post {
  id: ID
  title: String
  body: String
  comments: [Comment]
  author: User
}

model Comment {
  id: ID
  text: String
  post: Post
  author: User
}

model User {
  id: ID
  firstName: String
  lastName: String
  email: String
  posts: [Post]
  comments: [Comment]
  friends: [User]
  profile: Profile
}

embed Profile {
  imageUrl: String
  imageSize: String
}
```

### Find By Primary Key

Find a single resource by its `primary key`.

#### Current Typescript API:

```ts
const bob: User = await prisma.users.findOne('bobs-id')
```

#### Go

```go
bob, err := users.FindByID(db, "bobs-id")
```

#### GraphQL

```graphql
query($where: UserWhereInput) {
  user(where: $where) {
    id
    firstName
    lastName
    email
    profile {
      # Profile is an embedded type and therefore will also be fetched by default
      imageUrl
      imageSize
    }
  }
}
```

Variables

```json
{
  "where": {
    "id": "bobs-id"
  }
}
```

#### Additional Notes:

- todo

### Find By Unique Constraint

Find a single resource with a `unique(email)` constraint.

#### Current Typescript API:

```ts
const alice: User = await prisma.users.findOne({
  where: { email: 'alice@prisma.io' }
})
```

#### Go

```go
alice, err := users.FindByEmail(db, "alice@prisma.io")
```

#### GraphQL

```graphql
query($where: UserWhereInput) {
  user(where: $where) {
    id
    firstName
    lastName
    email
    profile {
      # Profile is an embedded type and therefore will also be fetched by default
      imageUrl
      imageSize
    }
  }
}
```

Variables

```json
{
  "where": {
    "email": "alice@prisma.io"
  }
}
```

#### Additional Notes:

- todo

### Find By a Composite Unique Constraint

Find a single resource with a `unique(first_name, last_name)` constraint.

#### Current Typescript API:

```ts
const john: User = await prisma.users.findOne({
  name: { firstName: 'John', lastName: 'Doe' }
})
```

#### Go

```go
john, err := users.FindByFirstNameAndLastName(db, "John", "Doe")
```

#### GraphQL

```graphql
query($where: UserWhereInput) {
  user(where: $where) {
    id
    firstName
    lastName
    email
    profile {
      # Profile is an embedded type and therefore will also be fetched by default
      imageUrl
      imageSize
    }
  }
}
```

Variables

```json
{
  "where": {
    "firstName": "John",
    "lastName": "Doe"
  }
}
```

#### Additional Notes:

- todo

### Find By non-unique conditions

Find a single resource by non-unique fields.

#### Current Typescript API:

```ts
const john: User = await prisma.users.findOne({
  name: { firstName: 'John', lastName: 'Doe' }
})
```

#### Go

```go
john, err := users.Find(db, users.Where().FirstName("John").LastName("Doe"))
```

#### GraphQL

```graphql
query($where: UserWhereInput) {
  users(where: $where) {
    id
    firstName
    lastName
    email
    profile {
      # Profile is an embedded type and therefore will also be fetched by default
      imageUrl
      imageSize
    }
  }
}
```

Variables

```json
{
  "where": {
    "firstName": "John",
    "lastName": "Doe"
  }
}
```

#### Additional Notes:

- todo

### Find with a condition

Find the all items that match a condition.

#### Current Typescript API:

```ts
const allUsers: User[] = await prisma.users.findAll({ firstName: 'John', lastName: 'Doe' })
// or
const allUsersShortcut: User[] = await prisma.users({ firstName: 'John', lastName: 'Doe' })
```

#### Go

```go
allUsers, err := users.FindMany(db, users.Where().FirstName("John").LastName("Doe"))
// or
// no equivalent. users.Users(...) does not exist to avoid stutter
```

#### GraphQL

```graphql
query($where: UserWhereInput) {
  users(where: $where) {
    id
    firstName
    lastName
    email
    profile {
      # Profile is an embedded type and therefore will also be fetched by default
      imageUrl
      imageSize
    }
  }
}
```

Variables

```json
{
  "where": {
    "firstName": "John",
    "lastName": "Doe"
  }
}
```

#### Additional Notes:

- todo

### Find with First limit

Find the first N items of a resource.

#### Current Typescript API:

```ts
const allUsers: User[] = await prisma.users.findAll({ first: 100 })
```

#### Go

```go
allUsers, err := users.FindMany(db, users.First(100))
```

#### GraphQL

```graphql
query($first: Int) {
  users(first: $first) {
    id
    firstName
    lastName
    email
    profile {
      # Profile is an embedded type and therefore will also be fetched by default
      imageUrl
      imageSize
    }
  }
}
```

Variables

```json
{
  "first": 100
}
```

#### Additional Notes:

- todo

### Find all with Order

Find an ordered list of items matching the condition.

#### Current Typescript API:

```ts
const allUsers = await prisma.users({
  where: { firstName: 'John', lastName: 'Doe' },
  orderBy: { email: 'ASC' }
})
```

#### Go

```go
allUsers, err := users.FindMany(db,
  users.Where().FirstName("John").LastName("Doe"),
  users.Order().Email("ASC"),
)
```

#### GraphQL

```graphql
query($where: UserWhereInput, $orderBy: UserOrderByInput) {
  users(where: $where, orderBy: $orderBy) {
    id
    firstName
    lastName
    email
    profile {
      # Profile is an embedded type and therefore will also be fetched by default
      imageUrl
      imageSize
    }
  }
}
```

Variables

```json
{
  "where": {
    "firstName": "John",
    "lastName": "Doe"
  },
  "orderBy": "email_DESC"
}
```

#### Additional Notes:

- May also want to define constants here, e.g. `prisma.DESC` & `prisma.ASC`.

### Find all with Composite Order

Find an ordered list of items matching the condition: `order by email ASC name DESC`

#### Current Typescript API:

```ts
const allUsers = await prisma.users({
  where: { firstName: 'John', lastName: 'Doe' },
  orderBy: [{ email: 'ASC' }, { name: 'DESC' }]
})
```

#### Go

```go
allUsers, err := users.FindMany(db,
  users.Where().FirstName("John").LastName("Doe"),
  users.Order().Email("ASC").Name("DESC"),
)
```

#### GraphQL

```graphql
query($where: UserWhereInput, $orderBy: UserOrderByInput) {
  users(where: $where, orderBy: $orderBy) {
    id
    firstName
    lastName
    email
    profile {
      # Profile is an embedded type and therefore will also be fetched by default
      imageUrl
      imageSize
    }
  }
}
```

Variables

```json
{
  "where": {
    "firstName": "John",
    "lastName": "Doe"
  },
  // this is not yet possible with Prisma. You can only order by one field at a time
  "orderBy": "name_ASC email_DESC"
}
```

#### Additional Notes:

- todo

### Find all with Nested Order

Find an ordered list of items matching the condition.

#### Current Typescript API:

```ts
const usersByProfile = await prisma.users({
  orderBy: { profile: { imageSize: 'ASC' } }
})
```

#### Go

```go
// todo
```

#### GraphQL

This is not yet possible with Prisma

```graphql
# todo
```

#### Additional Notes:

- todo

### Find all where contains

Find all items where resource contains submatch: `where email like %@gmail.com%`.

#### Current Typescript API:

```ts
const users = await prisma.users({ where: { email_contains: '@gmail.com' } })
```

#### Go

```go
users, err := users.Find(db, users.Where().EmailContains("@gmail.com"))
```

#### GraphQL

```graphql
query($where: UserWhereInput) {
  users(where: $where) {
    id
    firstName
    lastName
    email
    profile {
      # Profile is an embedded type and therefore will also be fetched by default
      imageUrl
      imageSize
    }
  }
}
```

Variables

```json
{
  "where": {
    "email_contains": "@gmail.com"
  }
}
```

#### Additional Notes:

- todo

### Partially Typed Raw SQL

Escape hatch to handle cases where the generated API isn't expressive enough.

```sql
select * from users where email like '%@gmail.com%' order by age + postsViewCount desc
```

#### Current Typescript API:

```ts
const users = await prisma.users({
  where: { email_contains: '@gmail.com' },
  raw: { orderBy: 'age + postsViewCount DESC' }
})
```

#### Go

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

#### Underlying graphql query for JS Client:

```graphql
query($where: UserWhereInput, $raw: UserRawInput) {
  # this is not yet possible with Prisma
  users(where: $where, raw: $raw) {
    id
    firstName
    lastName
    email
    profile {
      # Profile is an embedded type and therefore will also be fetched by default
      imageUrl
      imageSize
    }
  }
}
```

Variables

```json
{
  "where": {
    "email_contains": "@gmail.com"
  },
  "raw": {
    "orderBy": "age + postsViewCount DESC"
  }
}
```

#### Underlying graphql query for Go Client:

```graphql
query($raw: UserRawInput) {
  # this is not yet possible with Prisma
  users(raw: $raw) {
    id
    firstName
    lastName
    email
    profile {
      # Profile is an embedded type and therefore will also be fetched by default
      imageUrl
      imageSize
    }
  }
}
```

Variables

```json
{
  "raw": "select * from %s where %s like '%@gmail.com%' order by %s + %s desc"
}
```

#### Additional Notes:

- todo

#### Fluent API

```sql
select * from posts where user_id = 'bobs-id' limit 50
```

#### Current Typescript API:

```ts
const bobsPosts: Post[] = await prisma.users.findOne('bobs-id').posts({ first: 50 })
```

#### Go

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

#### Underlying graphql query as of now:

```graphql
query($where: UserWhereInput, $first: Int) {
  user(where: $where) {
    id
    firstName
    lastName
    email
    profile {
      # Profile is an embedded type and therefore will also be fetched by default
      imageUrl
      imageSize
    }
    posts(first: $first) {
      id
      title
      body
    }
  }
}
```

Variables

```json
{
  "where": {
    "id": "bobs-id"
  },
  "first": 50
}
```

#### Optimized underlying graphql query

```graphql
query($where: UserWhereInput, $first: Int) {
  user(where: $where) {
    posts(where: $where, first: $first) {
      id
      title
      body
    }
  }
}
```

Variables

```json
{
  "where": {
    "id": "bobs-id"
  },
  "first": 50
}
```

#### Advanced optimized underlying graphql query

```graphql
query($where: PostWhereInput, $first: Int) {
  posts(where: $where, first: $first) {
    id
    title
    body
    comments
  }
}
```

Variables

```json
{
  "where": {
    "author": {
      "id": "bobs-id"
    }
  },
  "first": 50
}
```

Here we would need to understand back relations.

#### Fluent API (Chained)

#### Current Typescript API:

```ts
const bobsLastPostComments: Comment[] = await prisma.users
  .findOne('bobs-id')
  .posts({ last: 1 })
  .comments()
```

#### Go

```go
// backwards
//   no API. backwards starts to get really confusing after you
//   go more than one level deep
// or, forwards
bobsLastPostComments, err := users.FromID('bobs-id').FromPost(posts.Last(1)).Comments.FindMany(db)
// or
bobsLastPostComments, err := users.FromID('bobs-id').FromPost(posts.Last(1)).FindManyComments(db)
```

#### Underlying graphql query as of now:

```graphql
query($where: UserWhereInput, $last: Int) {
  users(where: $where) {
    id
    firstName
    lastName
    email
    profile {
      # Profile is an embedded type and therefore will also be fetched by default
      imageUrl
      imageSize
    }
    posts(last: $last) {
      id
      title
      body
      comments {
        id
        text
      }
    }
  }
}
```

Variables

```json
{
  "where": {
    "id": "bobs-id"
  },
  "last": 1
}
```

### Optimized graphql query:

```graphql
query($where: UserWhereInput, $last: Int) {
  users(where: $where) {
    posts(last: $last) {
      comments {
        id
        text
      }
    }
  }
}
```

Variables

```json
{
  "where": {
    "id": "bobs-id"
  },
  "last": 1
}
```

#### Additional Notes:

- We may be able to ditch the 2nd `FromX`, e.g. `FromPost()`, though it may make things more clear regardless of if it's possible to remove.

### Select API

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

#### Current Typescript API:

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

#### Go

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

#### GraphQL

```graphql
query($where: UserWhereInput) {
  user(where: $where) {
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
}
```

Variables

```json
{
  "where": {
    "id": "bobs-id"
  }
}
```

### Page Info / Streaming Iterator

#### Typescript API

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

#### Go

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

#### GraphQL

The first API using `findOne.posts.$withPageInfo` is not yet possible with Prisma,
as you would need to use the `postsConnection` field on User, which only exists as a toplevel query right now.

```graphql
query($first: Int, $after: String) {
  postsConnection(first: $first, after: $after) {
    pageInfo {
      hasNextPage
      hasPreviousPage
      startCursor
      endCursor
    }
    edges {
      node {
        id
        title
        body
      }
    }
  }
}
```

Variables

```json
{
  "first": 100,
  "after": null
}
```

### Aggregations

#### Typescript API

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

#### Go

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

#### GraphQL

```graphql
# todo
```

### GroupBy with Select

#### Typescript API

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

#### Go

```go
var user struct {
  ID string
  Name string
  Posts []struct {
    ID string
    Title string
    Aggregate struct {
      Count int
    }
  }
}

err := users.Select(db, &user,
  users.Where().IsActive(true),
  users.First(100),
  users.Order().LastName("ASC"),
  users.Group().LastName(),
  users.Having().AgeGt(10),
)


err := users.Select(db, &user,
  users.Where().IsActive(true),
  users.First(100),
  users.Order().LastName("ASC"),
  users.Group().Raw(fmt.Sprintf("%s || %s", users.Field.FirstName, users.Field.LastName)),
  users.Having().Raw(fmt.Sprintf("AVG(%s) > 50", users.Field.Age)),
)
```

#### GraphQL

```graphql
# not possible yet
```

### Create

Insert data into the database.

#### Typescript

```ts
const newUser: User = await prisma.users.create({ firstName: 'Alice' })
```

#### Go

```go
newUser, err := users.Create(db, users.New().FirstName("Alice"))
```

#### GraphQL

```graphql
# todo
```

### Update by ID

#### Typescript

```ts
// Updates
const updatedUser: User = await prisma.users.update({
  where: 'bobs-id',
  data: { firstName: 'Alice' }
})
```

#### Go

```go
updatedUser, err := users.UpdateByID(db, "bobs-id", users.New().FirstName("Alice"))
```

#### GraphQL

```graphql
# todo
```

### Update by Unique Field

#### Typescript

```ts
const u: User = await prisma.users.update({
  where: { email: 'bob@prisma.io' },
  data: { firstName: 'Alice' }
})
```

#### Go

```go
u, err := users.UpdateByEmail(db, "bob@prisma.io", users.New().FirstName("Alice"))
```

#### GraphQL

```graphql
# todo
```

### Update by Composite Unique Fields

#### Typescript

```ts
// todo
```

#### Go

```go
u, err := users.UpdateByFirstNameAndLastName(db, "alice", "baggins",
  users.New().FirstName("Martha"),
)
```

#### GraphQL

```graphql
# todo
```

### Update by a Condition

Update by a condition, returning the first user.

> WARNING: your condition is expected to match one user, this could update more if you're not specific enough. This call will return the first updated user

#### Typescript

```ts
// todo
```

#### Go

```go
u, err := users.Update(db, user.New().FirstName("Martha"), user.Where().FirstName("Alice"))
```

#### GraphQL

```graphql
# todo
```

### Update Many by a Condition

Update many rows by a condition, returning all updated users

#### Typescript

```ts
// todo
```

#### Go

```go
uu, err := users.UpdateMany(db,
  user.New().FirstName("Martha"),
  user.Where().FirstName("Alice"),
)
```

#### GraphQL

```graphql
## todo
```

### Upsert By ID

Upsert a resource by its ID. If the ID matches, it's an update, otherwise it's a create

#### Typescript

```ts
const upsertedUser: User = await prisma.users.upsert({
  where: 'bobs-id',
  update: { firstName: 'Alice' },
  create: { id: '...', firstName: 'Alice' }
})
```

#### Go

```go
upsertedUser, err := users.UpsertByID(db, "bobs-id", users.New().ID("...").FirstName("Alice"))
```

#### GraphQL

```graphql
## todo
```

### Upsert By Unique Constraint

Upsert a resource by its unique constraint. If the unique constraint matches, it's an update, otherwise it's a create

#### Typescript

```ts
// todo
```

#### Go

```go
upsertedUser, err := users.UpsertByEmail(db, "bob@bob.com",
  users.New().Email("...").FirstName("Alice"),
)
```

#### GraphQL

```graphql
## todo
```

### Upsert By Composite Unique Constraint

Upsert a resource by its composite unique constraint. If the composite unique constraint matches, it's an update, otherwise it's a create

#### Typescript

```ts
// todo
```

#### Go

```go
upsertedUser, err := users.UpsertByFirstNameAndLastName(db, "Alice", "Bobbins",
  users.New().FirstName("Mark").LastName("Anthony"),
)
```

#### GraphQL

```graphql
## todo
```

### Delete by ID

Delete by an ID

#### Typescript

```ts
// NOTE has Fluent API disabled (incl. nested queries)
const deletedUser: User = await prisma.users.delete('bobs-id')
```

#### Go

```go
deletedUser, err := users.DeleteByID(db, "bobs-id")
```

#### GraphQL

```graphql
## todo
```

### Delete by Unique Constraint

Delete by a unique constraint.

#### Typescript

```ts
// todo
```

#### Go

```go
deletedUser, err := users.DeleteByEmail(db, "bob@bob.com")
```

#### GraphQL

```graphql
## todo
```

### Delete by Composite Unique Constraints

Delete by composite unique constraints

#### Typescript

```ts
// todo
```

#### Go

```go
deletedUser, err := users.DeleteByFirstNameAndLastName(db, "Alice", "Baggins")
```

#### GraphQL

```graphql
## todo
```

### Delete by a condition

Delete by a condition

> WARNING: your condition is expected to match one user, it could delete more if you're not specific enough. This call will return the first user

#### Typescript

```ts
// todo
```

#### Go

```go
// delete by a condition returning the first deleted
deletedUser, err := users.Delete(db, users.Where().FirstName("alice"))
```

#### GraphQL

```graphql
## todo
```

### Delete Many by a condition

Delete many by conditions returning all deleted users

> Note: this might not be portable, but it sure is nice in Postgres.
> It'd stink to cater to the lowest common database features.

#### Typescript

```ts
// todo
```

#### Go

```go
deletedUsers, err := users.DeleteMany(db, users.Where().FirstName("alice"))
```

### Delete count

Get the number of deleted users

#### Typescript

```ts
const deletedCount: number = await prisma.users.deleteMany()
```

#### Go

```go
deletedUsers, err := users.DeleteMany(db, users.Where().FirstName("alice"))
deletedCount := len(deletedUsers)
```

#### GraphQL

```graphql
## todo
```

### Update OCC

Update if the version matches

#### Typescript

```ts
const updatedUserOCC: User = await prisma.users.update({
  where: 'bobs-id',
  if: { version: 12 },
  data: { firstName: 'Alice' }
})
```

#### Go

```go
updatedUserOCC, err := users.Update(db,
  users.New().FirstName("Alice"),
  users.Where().ID('bobs-id'),
  users.If().Version(12),
)
```

#### GraphQL

```graphql
## todo
```

### Upsert OCC

Upsert if the version matches

> TODO: what's this query do? Would it reject both update and create if the version doesn't match?

#### Typescript

```ts
const upsertedUserOCC: User = await prisma.users.upsert({
  where: 'bobs-id',
  if: { version: 12 },
  update: { firstName: 'Alice' },
  create: { id: '...', firstName: 'Alice' }
})
```

#### Go

```go
// todo
```

#### GraphQL

```graphql
## todo
```

### Delete OCC

Delete if the version matches

#### Typescript

```ts
const deletedUserOCC: User = await prisma.users.delete({
  if: { version: 12 },
  where: 'bobs-id'
})
```

#### Go

```go
deletedUser, err := users.Delete(db,
  users.Where().FirstName("alice"),
  users.If().Version(12),
)
```

#### GraphQL

```graphql
## todo
```

### Batching

Batch multiple statements into one request.

#### Typescript

```ts
const m1 = prisma.users.create({ firstName: 'Alice' })
const m2 = prisma.posts.create({ title: 'Hello world' })
const [u1, p1]: [User, Post] = await prisma.batch([m1, m2])
```

#### Go

```go
b := prisma.Batch()
b.Users.Create(users.New().FirstName("Alice"))
b.Posts.Create(posts.New().Title("Hello World"))
if err := b.Run(db); err != nil {
  return err
}
```

#### GraphQL

```graphql
## todo
```

### Batching (with Transaction)

Batch multiple statements into one request as a transaction.

#### Typescript

```ts
const m1 = prisma.users.create({ firstName: 'Alice' })
const m2 = prisma.posts.create({ title: 'Hello world' })
const [u1, p1]: [User, Post] = await prisma.batch([m1, m2], { transaction: true })
```

#### Go

```go
b := prisma.Batch()
b.Users.Create(users.New().FirstName("Alice"))
b.Posts.Create(posts.New().Title("Hello World"))

tx, err := db.Begin()
if err != nil {
  return err
}
defer db.Rollback()

if err := b.Run(tx); err != nil {
  return err
}

if err := tx.Commit(); err != nil {
  return err
}
```

#### GraphQL

```graphql
## todo
```

### Explicit \$exec terminator

#### Typescript

```ts
// todo: Not clear what the command actually does
const usersQueryWithTimeout = await prisma.users.$exec({ timeout: 1000 })
```

#### Go

```go
// create a context with a timeout
ctx, cancel := context.WithDeadline(request.Context(), 10*time.Second))
defer cancel()
db := &DBContext{db, context}
u, err := users.Create(db, users.New().FirstName("Alice"))
```

#### GraphQL

```
# todo
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
