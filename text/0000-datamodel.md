- Start Date: 2019-03-23
- RFC PR: (leave this empty)
- Prisma Issue: (leave this empty)

# Summary

This RFC proposes a new syntax for the Prisma Datamodel. Main focus areas:

- Break from the existing GraphQL SDL syntax where it makes sense
- Clearly separate responsibilities into two categories: Core Prisma primitives and Connector specific primitives

# Basic example

This example illustrate many aspects of the proposed syntax:

```groovy
@db(name: "user")
model User {
  id: ID! @id
  createdAt: DateTime @createdAt
  email: String @unique
  name: String?
  role: Role = USER
  posts: [Post]
  profile: Profile? @relation(link: INLINE)
}

enum Role {
  USER
  ADMIN
}

@db(name: "profile")
model Profile {
  id: ID @id
  user: User
  bio: String
}

@db(name: "post")
model Post {
  id: ID @id
  createdAt: DateTime @createdAt
  updatedAt: DateTime @updatedAt
  title: String
  author: User
  published: Boolean = false
  categories: [Category]? @relation(link: TABLE, name: "PostToCategory")
}

@db(name: "category")
model Category {
  id: ID @id
  name: String
  posts: [Post] @relation(name: "PostToCategory")
}

@db(name: "post_to_category")
@linkTable
model PostToCategory {
  post: Post
  category: Category
}
```

# Motivation



# Detailed design



# Drawbacks



# Alternatives



# Adoption strategy



# How we teach this



# Unresolved questions


