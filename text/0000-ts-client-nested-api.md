- Start Date: 2019-01-12
- RFC PR: (leave this empty)
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

The idea is to mimic a GraphQL query in JS (or languages):
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

## Drawback 1: "Ugly" query value token

Queries in this API proposal are expressed using JS objects which always require a key **and a value** opposing to a GraphQL query where just providing a key (i.e. the field you're querying for) is enough. This means we need some kind of *token* for the value. Ideally we could come up with some kind of better syntax or avoid having to provide a token value altogether.

# Alternatives

## Drawback 1: "Ugly" query value token

- Use `{}` instead of `true`/`false`. The downside of this approach is that it wouldn't allow for explicitly excluding certain fields from being queried unless we'd opt for allowing `{} | false`.
- Another approach could be to use a builder API (e.g. `.$nested(posts(comments(), tags()).friends())`). The downside of this approach would probably be the following:
  - Lost automatic type-safety of the return types
  - "Free floating functions" tend to be less intuitive/pragmatic to use as they don't provide contextual auto-completion like a typed object would do.

# Adoption strategy

This is an additional feature and would replace the need for almost all uses of the `$fragment` API. We need to make sure to update all examples and recommendations in the docs.

# How we teach this

By documenting it properly across our docs and examples.

# Unresolved questions

- [ ] Is the proposed API even possible to implement in TS? The best approach seems to be a combination of conditional types and code generation.
- [ ] Should we use `true` as a token in the query API or can we find a better/different approach?
- [ ] Should we call the API `$nested` or can we find a better name?
- [ ] Should we introduce a special `$depth` feature for recursive relations?
