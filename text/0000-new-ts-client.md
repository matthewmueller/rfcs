- Start Date: 2019-01-16
- RFC PR: (leave this empty)
- Prisma Issue: (leave this empty)

# Summary


# Basic example


# Motivation



# Detailed design

The example code below assumes the following datamodel:

```
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

```ts
// NOTE: Explicit type annotations aren't required and just here for illustration
async function main() {
  // Get single node
  const bob: User = await prisma.users.findOne('bobs-id')
  const alice: User = await prisma.users.findOne({
    where: { email: 'alice@prisma.io' },
  })

  // Lookup-by Multi-field indexes
  const john: User = await prisma.users.findOne({
    name: { firstName: 'John', lastName: 'Doe' },
  })

  // Get many nodes
  const allUsers: User[] = await prisma.users.findAll({ first: 100 })
  const allUsersShortcut: User[] = await prisma.users({ first: 100 })

  // Ordering
  const usersByEmail = await prisma.users({ orderBy: { email: 'ASC' } })
  const usersByEmailAndName = await prisma.users({
    orderBy: [{ email: 'ASC' }, { name: 'DESC' }],
  })
  const usersByProfile = await prisma.users({
    orderBy: { profile: { imageSize: 'ASC' } },
  })

  // Where / filtering
  await prisma.users({ where: { email_contains: '@gmail.com' } })

  // Raw
  await prisma.users({
    where: { email_contains: '@gmail.com' },
    raw: { orderBy: 'age + postsViewCount DESC' },
  })

  const someEmail = 'bob@prisma.io'
  await prisma.users({
    // where is overwritten when provided in raw
    // where: { email_contains: '@gmail.com' },
    raw: {
      orderBy: 'age + postsViewCount DESC',
      where: ['email = $1', someEmail],
    },
  })

  // Fluent API
  const bobsPosts: Post[] = await prisma.users
    .findOne('bobs-id')
    .posts({ first: 50 })

  // Query Object API
  const dynamicResult1: DynamicResult1 = await prisma.users.findOne({
    where: 'bobs-id',
    select: {
      posts: { select: { comments: true } },
      friends: true,
    },
  })

  const dynamicResult4: DynamicResult1 = await prisma.users.findOne({
    where: 'bobs-id',
    select: {
      posts: { select: { comments: true } },
      friends: true,
      // $raw: {
      //   name: {
      //     query: `firstName || ' ' || lastName`,
      //     type: 'String'
      //   }
      // },
      // expr: {
      //   name2: user => user.firstName.add(' ').add(user.lastName)
      // }
    },
  })

  // PageInfo
  const bobsPostsWithPageInfo: PageInfo<Post> = await prisma.users
    .findOne('bobs-id')
    .posts({ first: 50 })
    .$withPageInfo()

  // Streaming data
  for await (const post of prisma.posts().$stream()) {
    console.log(post)
  }

  const postStreamWithPageInfo = await prisma
    .posts()
    .$stream()
    .$withPageInfo()

  prisma
    .posts({ first: 10000 })
    .$stream({ chunkSize: 100, fetchThreshold: 0.5 /*, tailable: true*/ })

  // Aggregations
  const dynamicResult2: DynamicResult2 = await prisma.users({
    select: { aggregate: { age: { avg: true } } },
  })

  const dynamicResult3: DynamicResult3 = await prisma.users.findOne({
    where: 'bobs-id',
    select: { posts: { select: { aggregate: { count: true } } } },
  })

  // GroupBy
  const groupByResult: DynamicResult4 = await prisma.users.groupBy({
    key: 'lastName',
    having: { age_avg_gt: 10 },
    where: { isActive: true },
    first: 100,
    orderBy: { lastName: 'ASC' },
    select: {
      records: { first: 100 },
      aggregate: { age: { avg: true } },
    },
  })

  const groupByResult2: DynamicResult5 = await prisma.users.groupBy({
    raw: { key: 'firstName || lastName', having: 'AVG(age) > 50' },
    select: {
      records: { $first: 100 },
      aggregate: { age: { avg: true } },
    },
  })

  // Writing data
  const newUser: User = await prisma.users.create({ firstName: 'Alice' })

  // Updates
  const updatedUser: User = await prisma.users.update({
    where: 'bobs-id',
    data: { firstName: 'Alice' },
  })

  const updatedUserByEmail: User = await prisma.users.update({
    where: { email: 'bob@prisma.io' },
    data: { firstName: 'Alice' },
  })

  const upsertedUser: User = await prisma.users.upsert({
    where: 'bobs-id',
    update: { firstName: 'Alice' },
    create: { id: '...', firstName: 'Alice' },
  })

  // NOTE has Fluent API disabled (incl. nested queries)
  const deletedUser: User = await prisma.users.delete('bobs-id')

  // OCC
  const updatedUserOCC: User = await prisma.users.update({
    where: 'bobs-id',
    if: { version: 12 },
    data: { firstName: 'Alice' },
  })

  const upsertedUserOCC: User = await prisma.users.upsert({
    where: 'bobs-id',
    if: { version: 12 },
    update: { firstName: 'Alice' },
    create: { id: '...', firstName: 'Alice' },
  })

  const deletedUserOCC: User = await prisma.users.delete({
    if: { version: 12 },
    where: 'bobs-id',
  })

  const deletedCount: number = await prisma.users.deleteMany()

  // Batching
  const m1 = prisma.users.create({ firstName: 'Alice' })
  const m2 = prisma.posts.create({ title: 'Hello world' })
  const [u1, p1]: [User, Post] = await prisma.batch([m1, m2])

  // Batching with transaction
  await prisma.batch([m1, m2], { transaction: true })

  // Explicit $exec terminator
  const usersQueryWithTimeout = await prisma.users.$exec({ timeout: 1000 })

  // Top level $query API
  const nestedResult = await prisma.query({
    users: {
      first: 100,
      select: {
        posts: { select: { comments: true } },
        friends: true,
      },
    },
  })
}

// NOTE the following types are auto-generated
type Post = {
  id: string
  title: string
  body: string
}

type Comment = {
  id: string
  text: string
}

type User = {
  id: string
  firstName: string
  lastName: string
  email: string
  profile: Profile
}

type Profile = {
  imageUrl: string
  imageSize: number
}

type PageInfo<Data> = {
  data: Data[]
  hasNext: boolean
  hasPrev: boolean
}

type DynamicResult1 = (User & {
  posts: (Post & { comments: Comment[] })[]
  friends: User[]
})[]

type DynamicResult2 = (User & { aggregate: { age: { avg: number } } })[]

type DynamicResult3 = User & {
  posts: (Post & { aggregate: { count: number } })[]
}

// TODO wrong type
type DynamicResult4 = {
  lastName: string
  records: User[]
  aggregate: { age: { avg: number } }
}

// TODO wrong type
type DynamicResult5 = {
  raw: any
  records: User[]
  aggregate: { age: { avg: number } }
}
```

## `$withPageInfo`

- Can be applied to every paginable list and stream

## Nested "thenable" API

```ts
await user.create({
  role,
  email,
  posts: post.upsertByEmail({
    title
  })
})
```

## Aggregations

TODO:

- Aggregations API
- Extend `where` API to support aggregations

## Group By

TODO

## `raw` fallbacks

## Batching

TODO:

- Add option to not return data

## API shortcuts

### Where

```ts
prisma.users.deleteMany('id')
prisma.users.deleteMany(['id1', 'id2'])

prisma.users({
  where: {
    id: ['id1', 'id2'], // instead of `_in` or `OR`
    email: { endsWith: '@gmail.com' },
  }
})

prisma.users({
  where: {
    name: { contains: 'Bob' },
    email: { contains: ['prisma.io', 'gmail.com'] }, // instead of `_in` or `OR`
  }
})
```

# Drawbacks


# Alternatives

- `$nested` API


# Adoption strategy


# How we teach this


# Unresolved questions

- [ ] Type mapping and static field preselection (see [comment in #4](https://github.com/prisma/rfcs/pull/4#issuecomment-471202364))]
- [ ] (Type-safe) raw field selection
- [ ] API shortcuts (`where`, `aggregate`, `select`, ...)
- [ ] Rethink generated type name scheme (incl. pluralization)
- [ ] Real-time API (subscriptions/live queries)
- [ ] Unit of work API (see #5)
- [ ] Life-cycle hooks
- [ ] `exists` API
- [ ] API for atomic operations
- [ ] Explore idea of nested "thenable` API
- [ ] Consolidate with Go client API
  - [ ] `findAll` vs `findMany`
- [ ] Double check cursor, streaming and batching API
- [ ] Update(many) API to use existing values
- [ ] Better `raw` integration with query builders like Knex
- [ ] Do we need `$` prefix (e.g. `$withPageInfo`)

# Future topics

- [ ] Rails-like scopes (see [Sequelize](http://docs.sequelizejs.com/manual/tutorial/scopes.html))

