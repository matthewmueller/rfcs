- Start Date: 2019-01-22
- RFC PR: (leave this empty)
- Prisma Issue: (leave this empty)

# Summary

_This RFC is still an early draft_

# Basic example

```ts
const prisma: any = {}

// TODO figure out better name
const uow = prisma.unitOfWork()

const bob = await prisma.users.findOne('bobs-id')

// TODO figure out better name
uow.ensureConsistent({ model: 'User', data: bob })

const m1 = prisma.posts.update({
  where: 'post-id',
  data: { title: 'Hello Universe' },
})

// TODO figure out better name
uow.add(m1)

await uow.commit()
```

# Motivation


# Detailed design


# Drawbacks


# Alternatives


# Adoption strategy


# How we teach this


# Unresolved questions
