- Start Date: 2019-01-12
- RFC PR: [rfcs#1](https://github.com/prisma/rfcs/pull/1)
- Prisma Issue: (leave this empty)

# Summary

A new API for the JS/TS Prisma client to allow for more complex, type-safe, relational queries using the given programming language instead of having to fallback to GraphQL.

# Basic example

```ts
const dynamicResult: DynamicResult = await prisma
  .user('bobs-id')
  .$nested({ posts: { comments: true }, friends: true })

// assuming the following type definitions
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
  name: string
}

// Note: This is just here for the sake of demonstration and would be automatically
// derived based on conditional types in TS
type DynamicResult = {
  posts: (Post & { comments: Comment[] })[]
  friends: User[]
}
```

# Motivation

As described in e.g. [prisma#3104](https://github.com/prisma/prisma/issues/3104), the current Prisma client JS/TS API is quite limited when it comes to more complex relational queries (i.e. you can just query exactly one level of nodes). As a fallback it's possible to use the $fragment API which however can feel a bit cumbersome and lacks better tooling to make it nice to use (e.g. auto-completion for the GraphQL query/fragment + auto-generated interfaces for the response type generics).

# Detailed design

## Goal

- Intuitive API (high discoverablity and self-documenting)
- Full type-safety (incl. arguments and return type) without the need for explicitly provided generics
- Also works for plain JavaScript/Node.js

## Approach

The idea is to mimic a GraphQL query in JS (or other languages):
- To express whether you fetch a field/relation you set it to `true` or `false`. Empty/nested objects are equivalent to specifying `true`.
- **Scalars `true` by default** Similar to how the client automatically fetches all scalar fields, the queried relations inside the `$nested` query also contain all scalar values by default.
- **Relations `false` by default** Also same as for the typically client behavior, relations are not fetched by default but can be fetched by setting the corresponding field to `true` or `{ ... }`.
- The response value is strongly typed given on which fields you're querying. In TS this can be derived automatically using conditional types. In other programming languages this needs to be handled via generics (+ optionally code generation).

The `$nested` API works both on the root level (e.g. `prisma.$nested({ ... })`) as well on a node level (e.g. `prisma.users().$nested({ ... })`. See full example:


```ts
const nestedResult = await prisma.$nested({
  users: {
    $args: { first: 100 },
    firstName: false,
    posts: {
      comments: true,
    },
    friends: true,
  },
})
```

# Drawbacks

## Drawback 1: "Ugly" object value token

Queries in this API proposal are expressed using JS objects which always require a key **and a value** opposing to a GraphQL query where just providing a key (i.e. the field you're querying for) is enough. This means we need some kind of *token* for the value. Ideally we could come up with some kind of better syntax or avoid having to provide a token value altogether.

# Alternatives

Before diving into the details of the alternative suggestions, note that most proposals below are not mutually exclusive with the initial proposal. That means at the expense of additional implementation cost we could allow for a combination of all different syntaxes.

### Alternative 1.1 (addresses Drawback 1)

Use `{}` instead of `true`/`false`. The downside of this approach is that it wouldn't allow for explicitly excluding certain fields from being queried unless we'd opt for allowing `{} | false`.

### Alternative 1.2 (addresses Drawback 1)

Another approach could be to use a builder API.

```ts
const nestedResult = await prisma.$nested(
  posts(
   comments(), 
   tags()
  ),
  friends()
)
```

The downside of this approach would probably be the following:

- Lost automatic type-safety of the return types
- "Free floating functions" tend to be less intuitive/pragmatic to use as they don't provide contextual auto-completion like a typed object would do.

### Alternative 1.3  (addresses Drawback 1)

Yet another approach would be to allow a combination of the presented nested object notation and string (literal) arrays in the case of leaf query fields. Here are a few examples:

```ts
// 1.3.1 Simple example to query a user their related friends and posts
const dynamicResult: DynamicResult = await prisma
  .user('bobs-id')
  .$nested('friends', 'posts')

// 1.3.2 Slightly more complex query that also includes the comments for each post
const dynamicResult: DynamicResult = await prisma
  .user('bobs-id')
  .$nested('friends', { posts: 'comments' })

// 1.3.3 Even more complex query that uses pagination for the queried post comments
const dynamicResult: DynamicResult = await prisma
  .user('bobs-id')
  .$nested(
    'friends',
    { posts: { comments: { $args: { first: 100 } } } }
  )
```

Note that for the sake of convenience the proposal suggests to support both an explicit and an optional array notation. That means that the following example both work and are equivalent `$nested(['field1', 'field2'])` and `$nested('field1', 'field2')`. The same also applies for deeper fields in the object hierarchy.

This string array based approach successfully avoids the need for a object value token altogether and is intuitive and straightforward to read/write in simple cases. However, as seen in the 1.3.2 example above a combination of string (here `'friends'`) and object notation (here `{ posts: ... }`) can be unintuitive to read since the difference in notion suggests for the two fields to be not on the same hierarchy level.

### Alternative 1.4

To make it easier for a developer to structure their code in a flexible way, we could also support chained `.$nested` calls following the Fluent API style. This would allow for the following:

```ts
const dynamicResult: DynamicResult = await prisma
  .user('bobs-id')
  .$nested('friends')
  .$nested({ posts: 'comments' })
```

However, we have to carefuly evaluate whether this additional API functionality is worth it.

# Adoption strategy

This is an additional feature and would replace the need for almost all uses of the `$fragment` API. We need to make sure to update all examples and recommendations in the docs.

# How we teach this

By documenting it properly across our docs and examples.

# Unresolved questions

- [ ] Is the proposed API even possible to implement in TS? The best approach seems to be a combination of conditional types and code generation.
- [ ] Should we use `true` as a token in the query API, the string array syntax or can we find a better/different approach?
- [ ] Should we call the API `$nested` or can we find a better name?
- [ ] Should we introduce a special `$depth` feature for recursive relations?
- [ ] TS implementation: How would generic variables be handled?
