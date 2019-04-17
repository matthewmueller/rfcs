# Datamodel 2 (Experimental Syntax)

This RFC attempts to improve the syntax laid out in [current RFC on Data Model 2](https://github.com/prisma/rfcs/blob/datamodel/text/0000-datamodel.md#1-1), spec out a few additional features and unify some concepts.

This RFC is a potential answer to an open question posed in the previous RFC:

> "If we were a little more radical with the syntax, could we create something much better?"

- **Warning:** There is a lot missing from this RFC that is properly speced out in the previous RFC. If we like the direction of this syntax, I can start bringing concepts over or moving concepts to that spec.

## Requirements

- Break from the existing GraphQL SDL syntax where it makes sense
- Clearly separate responsibilities into two categories: Core Prisma primitives and Connector specific primitives
- High-level relationships without ambiguities
- Easily parsable ([avoid symbol tables, ideally](https://golang.org/doc/faq#different_syntax))
- Abstraction over raw column names via field aliasing

## Nice to Have

- One configuration file for prisma (WIP)
- Can be rendered into raw JSON
- Strict Machine formatting (␀ bikeshedding)
- Multi-line support and optional single-line via commas
- Unicode (emoji) support

## Summary of adjustments from previous DM2 RFC

- Explicit join tables. CLI/VSCode makes suggestions when you don't have the join table
- Removed colon between name and type
- Functions (or attributes) instead of directives, removing @.
- Moved model directives from the top of the model into the block
- Lowercase primitives, capitalized higher-level types
- Removed `ID` as a primitive type
- Merged prisma.yml configuration into the datamodel (WIP)
- Replaced relation metadata with `Model@field`
- Introduced `source` block for connectors
- Renamed `embedded` to `embed`
- Replaced `=` in favor of `default(...)`
- Replaced `text` with `text`
- Added "embedded embeds"
- Added metadata support to any block
- Added top-level configuration support
- Added connector constraints
- Revised connector-specific types
- Added model embedding (fragments in GraphQL)
- Adjust many-to-many convention `BlogToWriter` to `BlogsWriters`

## Syntax (WIP)

This following syntax is primarily inspired by [HCL2's blocks and attributes](https://github.com/hashicorp/hcl2/#information-model-and-syntax)
concepts for the configuration and block formatting.

```hcl
ConfigFile   = Body;
Body         = (Attribute | Block | OneLineBlock)*;
Attribute    = Identifier "=" Expression Newline;
Block        = Identifier (StringLit|Identifier)* "{" Newline Body "}" Newline;
OneLineBlock = Identifier (StringLit|Identifier)* "{" (Identifier "=" Expression)? "}" Newline;
```

> https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#structural-elements

Additionally, this syntax has a concept of alias definitions, field definitions & block expressions.

```
SelectorExpression = Identifier "." Identifier
AliasDefinition    = Identifier Identifier "=" (SelectorExpression|Identifier|Expression)
// TODO: include multi-line expressions
FieldDefinition    = Identifier (SelectorExpression|Identifier|Expression) (Expression)*;
// Added Expressions from Body above
Body               = (Attribute | Block | OneLineBlock | Expression)*;
```

## Basic Example

This example illustrate many aspects of the proposed syntax:

```groovy
postgres_url = env("POSTGRES_URL")
mongo_url    = env("MONGO_URL")

// postgres datasource
source postgres {
  url     = postgres_url
  default = true
}

// mongo datasource
source mongo {
  url = mongo_url
}

// connect to our secondary Mongo DB
source mongo {
  alias    = mgo2
  host     = "localhost"
  port     = 27017
  database = "neato2"

  query = {
    sslMode = "disable"
  }
}

type Numeric = postgres.Numeric(5, 2)

model User {
  meta = {
    // adjust name from "users" convention
    name = "people"
  }

  // model fields
  id             int              primary() serial
  email          postgres.Citext  unique() postgres.Like(".%.com")
  name           text?            check(name > 2)
  role           Role
  profile        Profile?         alias("my_profile")
  createdAt      datetime         default(now())
  updatedAt      datetime         onChange(now())

  weight         Numeric          alias("my_weight")
  posts          Post[]

  // composite indexes
  unique(email, name)             alias("email_name_index")
}

enum Role {
  USER  text // unless explicit, defaults to "USER"
  ADMIN text default("A")
}

model Profile {
  meta = {
    from = mongo
    name = "people_profiles"
  }

  // model fields
  id       int            primary() serial()
  author   User@id
  bio      text

  // nullable array items and nullable photos field
  photos   Photo?[]?
}

// named embed (reusable)
embed Photo {
  url  text

  // anonymous embed (optional)
  size {
    height int
    width int
  }?

  // anonymous embed
  alternatives {
    height int
    width int
  }[]
}

model Post {
  // intentionally messy 😅
  // multi-line support, separated with commas
  id                     int          primary(),
                                      serial(),
                                      // this is okay too...
                                      default("some default") // default value
  // model attribute right after also fine
  unique(title, author)

  title                  text
  author                 User@id
  reviewer               User@id
  published              boolean      default(false)

  createdAt              datetime     createdAt() default(now())
  updatedAt              datetime     updatedAt() default(now())

  categories:            CategoriesPosts[]
}

model MoviePost {
  // mixin Post's fields into MoviePost
  // Based on Go's struct embedding syntax
  Post

  stars   text
  review  text

  // duplicate title would replace the included model field
  title   text
}

// Comments in this datamodel have meaning. Look to godoc for inspiration:
// https://blog.golang.org/godoc-documenting-go-code

// Comments directly above a model are attached to the model
model Category {
  id     int    primary(),
  name   text
  posts  CategoriesPosts[]
}

// Many-to-Many naming based on Rails conventions
model CategoriesPosts {
  post      Post@id
  category  Category@id
  unique(post, category)
}
```

## Configuration

We're essentially a superset of the HCL2, so the configuration of Terraform would apply here:

```groovy
postgres_url = env("POSTGRES_URL")
mongo_url    = env("MONGO_URL")

generate javascript {
  target = "es3"
  output = "generated/js"
}

generate typescript {
  // ...
}

generate flow {
  // ...
}
```

## Expression generators

> Note: this section is experimental and might never be implemented in Prisma.
> This might be a better fit in the modifiers layer.

Prisma could support expressions in generators:

```groovy
model User {
  name             text
  age              int
  someRandomField  text  default(`${this.name} is ${this.age} years old`)
  ageInDays        int     default(`${this.age * 365}`)
}
```

## Relations

### 1-1

#### Specifying Relation id side

```groovy
model User {
  id        int           primary() serial()
  customer  Customer@id?
  name      text
}

model Customer {
  id       int     primary() serial()
  user     User?
  address  text
}
```

The relationship can be made on either side, but the `@id` indicates where the data is stored. You can think of this as a pointer to the Customer id field.

### 1-M

```groovy
model Writer {
  id      int        primary() serial()
  blogs   Blog[]
}

model Blog {
  id      int        primary() serial()
  author  Writer@id
}
```

- `author Writer@id` points to the `id int` on `Writer` model, establishing the
  has-many relationship.
- `blogs Blog[]` names the back-relation, but is entirely optional

### M-N

Blogs can have multiple writers

```groovy
model Blog {
  id       int        primary() serial()
  authors  Writer[]
}

model Writer {
  id      int      primary() serial()
  blogs   Blog[]
}

// many to many
model BlogsWriters {
  blog      Blog@id
  author    Writer@id
  is_owner  bool

  // enforce a composite unique
  unique(author, blog)
}
```

- Many-to-Many relationships always require an explicit join table.
- Many-to-many's are always awkward to name. I find ActiveRecord's convention is the least awkward of the bunch, until you decide to provide a custom name. Alphabetical order and plural: https://guides.rubyonrails.org/association_basics.html#the-has-and-belongs-to-many-association
- The `BlogsWriters` holds data about the relationship and points to the data types
  in the `Blog` and `Writer` models

### Ambiguous Relations

With explicit join tables, we have less ambiguities, but we may still have issues like this:

```groovy
model User {
  id        int                  primary() serial()
  asked     Question@asker[]
  answered  Question@answerer[]
}

// @id could probably be implied here
model Question {
  id        int      primary() serial()
  asker     User@id
  answerer  User@id
}
```

### Self-Referential Models

```groovy
// @id could probably be implied here
model Employee {
  id         int          primary() serial()
  reportsTo  Employee@id
}
```

### Embedded Models

```groovy
model Human {
  id     int   primary() serial()
  name   text
  height int
}

model Employee {
  Human
  employer  text
  height    float
}
```

Models can be embedded inside of other models, resulting in an Employee that looks like this:

```groovy
model Employee {
  id      int   primary() serial()
  name    text
  height  int
  height  float
}
```

# Open Questions

**1. Lowercase primitives, uppercase indentifiers?**

- If primitives are considered special and cannot be overridden
  then I think we should have special syntax for them. If they are
  simply types that are booted up at start (GraphQL), they should
  be treated like every other type.

**2. Do we want back-relations to be optional?**

- My typical stance is to enforce good practices (e.g. prettier),
  and provide one way to do it. We have some options with back-relations though:

  1. Implied back-relation when not provided (affects client API)
  2. No back-relation when not provided (affects client API)
  3. Build-time (`prisma generate`) error when no back-relation provided

**3. Should we enforce link tables?**

- Link tables are not usually needed right away, but are often good practice
  since you often want to attach metadata to that relation later on
  (e.g. `can_edit boolean`). Some options:

  1. We could make them optional at first, but create a table in the background (we'd need to do this anyway), but then when they specify the table and migrate, we'll be aware that this implicit join table became explicit in the datamodel

  2. Enforce the link table at build-time when we run `prisma generate`. A bit simpler to implement and less magic, at the expense of cluttering up your datamodel file and forcing you to think more about your data layout earlier on (might not be a bad thing).

**4. Can back-relations have a different nullability than the forward relations?**

e.g. Is this possible?

```groovy
model User {
  id        int          primary() serial()
  customer  Customer     relates(id)
  name      text
}

model Customer {
  id       int    primary() serial()
  user     User
  address  text
}
```

Or is it always:

```groovy
model User {
  id        int         primary() serial()
  customer  Customer.id?
  name      text
}

model Customer {
  id       int    primary() serial()
  user     User?
  address  text
}
```

**5. `text` or `text`?**

🙃

- text is more familiar to programmers.
- Text is more familiar to English speakers
- Text is shorter.

**6. Should we support Mongo's many-to-many relationship in relational databases?**

This is what facebook needed to do to scale MySQL. It looks like the [original
video from 2011](https://www.facebook.com/Engineering?sk=app_260691170608423) was taken down,
but the gist is that they successfully scaled MySQL by add the foreign keys as an array in
your table aand writing "integrity checkers" to ensure that if a row gets deleted, the foreign
keys that link to it will eventually get deleted. In short, Facebook wrote a NoSQL
implementation on a MySQL database.

How this architecture plays out in your application is that instead of looking up
posts by a user's ID, you can lookup each post by ID within the user table. This
can tie into their memcache layer too so it's easily parallelizable.

In researching this question a bit better, it seems like they got the benefits of NoSQL,
without migrating all their data. Additionally, I came across Google Spanner (considered [NewSQL](https://en.wikipedia.org/wiki/NewSQL) 😑) which does
seem to give you the benefits of this approach with better guarentees so you don't need
to write your own "integrity checkers".

The question is can we enforce relations at the application layer without enforcement from
the database layer? Furthermore, is there value in doing this when something that's not using
Prisma client can muck up the data?

If there's value we need to be able to declare this:

```groovy
model User {
  id     int
  posts  Post[]
}

model Post {
  id     int
  users  User[]
}
```

How do we indicate that these are logical relationships without actually enforcing
those guarentees.

**7. Should we have named arguments?**

I generally like named arguments more, but with this new syntax it feels
like named arguments could generally be different attributes:

```
db(name: "users")
```

to

```
db("users") or name("users") or db.name("users")
```

In the cases where it does matter, we could maybe get assistance from
VSCode on this one. Here's what Android Studio does for Kotlin:

![named params](https://cldup.com/bXwmV_70KW.png)

**8. Should `id`, `created_at`, `updated_at` be special types that the database adds?**

The DM2 proposal proposes:

```groovy
model User {
  id: ID @id
}
```

The reason I avoided this is because it's not always clear (at least in SQL databases)
whether you want your ID to be an `int`, a `text`, or a `postgres.UUID`. I could
definitely see value in higher-level types where Prisma chooses the datatype based
on the most common cases. We'd just need to have a way to override them.

Same could go for `created_at` or `updated_at`. Where they bring a type and
some default functionality.

```groovy
model User {
  id        id
  createdAt created_at
  updatedAt updated_at
}
```

More questions: https://github.com/prisma/rfcs/blob/datamodel/text/0000-datamodel.md#open-questions
