# Datamodel 2.0

- [Env Block](#env-block)
- [Connector Block](#connector-block)
- [Generator Block](#generator-block)
- [Model Block](#model-block)
  - [Field Names](#field-names)
  - [Data Types](#data-types)
    - [Core Data Type to Connector](#core-data-type-to-connector)
    - [Core Data Type to Generator](#core-data-type-to-generator)
    - [Optional Types](#optional-types)
    - [List Types](#list-types)
    - [Relations](#relations)
      - [One-to-One (1:1) Relationships](#one-to-one-11-relationships)
      - [One-to-Many (1:N) Relationships](#one-to-many-1n-relationships)
      - [Implicit Many-to-Many (M:N) Relationships](#implicit-many-to-many-mn-relationships)
      - [Explicit Many-to-Many (M:N) Relationships](#explicit-many-to-many-mn-relationships)
      - [Self-Referential Relationships](#self-referential-relationships)
      - [Multiple-Reference Relationships](#multiple-reference-relationships)
      - [Referencing Primary Composite Keys](#referencing-primary-composite-keys)
  - [Attributes](#attributes)
    - [Field Attributes](#field-attributes)
    - [Core Field Attributes](#core-field-attributes)
    - [Block Attributes](#block-attributes)
    - [Core Block Attributes](#core-block-attributes)
    - [Connector Provided Attributes](#connector-provided-attributes)
  - [Extending Models](#extending-models)
- [Enum Block](#enum-block)
- [Embed Block](#embed-block)
  - [Inline Embeds](#inline-embeds)
- [Type Definition](#type-definition)
- [Attribute Definition](#attribute-definition)
- [Import Block](#import-block)
- [Function](#function)
- [Boolean Expressions](#boolean-expressions)
- [Configuration Layout](#configuration-layout)
  - [Soloists](#soloists)
  - [Team](#team)
    - [Multiple .prisma in the same directory get concatenated](#multiple-prisma-in-the-same-directory-get-concatenated)
    - [Multiple directories for different environments](#multiple-directories-for-different-environments)
    - [Organization](#organization)
- [Auto Formatting](#auto-formatting)
  - [Formatting Rules](#formatting-rules)
    - [Configuration blocks are align by their `=` sign.](#configuration-blocks-are-align-by-their--sign)
    - [Field definitions are aligned into columns separated by 2 or more spaces.](#field-definitions-are-aligned-into-columns-separated-by-2-or-more-spaces)
- [Grammar](#grammar)

## Env Block

The datamodel can require certain environment expectations to be met. The
purpose of the `env` block is to:

- Keeps secrets out of the datamodel
- Improve portability of the datamodel

```groovy
env POSTGRES_URL {
  type    = "string"
  default = "postgres://localhost:4321/db"
}

connector pg {
  type = "postgres"
  url  = env.POSTGRES_URL
}
```

In this case `env` represents the outside environment. Tools will be expected to
provide this environment variable when they perform operations based on the
datamodel. For example:

```sh
$ prisma deploy
! required POSTGRES_URL variable not provided

$ POSTGRES_URL="postgres://user:secret@rds.amazon.com:4321/db" prisma deploy
```

**TODO:** Does introspect write secrets to the DM or does it autogenerate these
env blocks?

## Connector Block

Connector blocks tell the datamodel where the models are backed. You can have
multiple connectors from different datasources.

```groovy
connector pg {
  type    = "postgres"
  url     = env.POSTGRES_URL
  default = true
}

connector mgo {
  type = "mongo"
  url  = env.MONGO_URL
}

connector mgo2 {
  type = "mongo"
  url  = env.MONGO2_URL
}
```

Connectors may bring their own attributes to allow users to tailor their
datamodels according to specific features of their connected datasources. We'll
talk about this more in **Connector Attributes**.

## Generator Block

Generator blocks configure what clients are generated and how they're generated.
Language preferences and configuration will go in here:

```groovy
generate js {
  target = "es3"
}

generate ts {
  target = "es5"
}

generate go {
  snakeCase = true
}
```

Generator blocks also generate a namespace. This namespace allows fine-grained
control over how a model generates it's types:

```groovy
generate go {
  snakeCase = true
}

type UUID String @go.type("uuid.UUID")

model User {
  id     UUID    @key
  email  String  @go.bytes(100)
}
```

This namespace is determined by the capabilities of the generator. The generator
will export a schema of capabilities we'll plug into.

## Model Block

Models are the high-level entities of our business. They are the nouns: the
User, the Comment, the Post and the Tweet.

Models may be backed by different datasources:

- In postgres, a model is a table
- In mongodb, a model is a collection
- In REST, a model is a resource

Here's an example of the Model block:

```groovy
model User {
  id         Int       @key
  email      String    @unique
  posts      Post[]
  createdAt  DateTime  @default(now())
  updatedAt  DateTime  @updatedAt
}

model Post {
  id          Int        @key
  title       String
  draft       Boolean
  categories  String[]
  slug        String
  author      User
  comments    Comment[]
  createdAt   DateTime   @default(now())
  updatedAt   DateTime   @updatedAt

  @@unique([ title, slug ])
}

model Comment {
  id         Int       @key
  email      String?
  comment    String
  createdAt  DateTime  @default(now())
  updatedAt  DateTime  @updatedAt
}
```

### Field Names

Field names are the first column of identifier inside the model block.

```
model _ {
  id         _
  email      _
  comment    _
  createdAt  _
  updatedAt  _
}
```

Field names are:

- Display name for the field
  - Affects the UI in studio, lift, etc.
- Not opinionated on casing
  - camel, snake, pascal are fine
- Name of the underlying field in the data source
  - Unless there's a handwritten `@field` override
  - If introspected, exactly always the same
- Basis for client generation
  - Generators may adjust casing depending on the language though

### Data Types

Prisma has a couple core primitive types. How these core types are defined may
vary across connectors. Every connector **must** implement these core types.
It's part of the connectors interface to Prisma. If a connector doesn't have a
core type, it should provide a **best-effort implementation**.

| Type     | Description           |
| -------- | --------------------- |
| String   | Variable length text  |
| Boolean  | True or false value   |
| Int      | Integer value         |
| Float    | Floating point number |
| Datetime | Timestamp             |

**TODO: Should we keep Datetime?**

**TODO: Should we add Binary?**

Here's how some of the databases we're tracking map to the core types:

#### Core Data Type to Connector

| Type     | Postgres  | MySQL     |
| -------- | --------- | --------- |
| String   | text      | TEXT      |
| Boolean  | boolean   | BOOLEAN   |
| Int      | integer   | INT       |
| Float    | real      | FLOAT     |
| Datetime | timestamp | TIMESTAMP |

| Type     | SQLite  | Mongo  | Raw JSON |
| -------- | ------- | ------ | -------- |
| String   | TEXT    | string | string   |
| Boolean  | _N/A_   | bool   | boolean  |
| Int      | INTEGER | int32  | number   |
| Float    | REAL    | double | number   |
| Datetime | _N/A_   | date   | _N/A_    |

_N/A:_ here means no perfect equivalent, but polyfill-able.

#### Core Data Type to Generator

| Type     | JS / TS | Go        |
| -------- | ------- | --------- |
| String   | string  | string    |
| Boolean  | boolean | bool      |
| Int      | number  | int       |
| Float    | number  | float64   |
| Datetime | Date    | time.Time |

#### Optional Types

All field types support optional fields. By default, fields are required, but if
you want to make them optional, you add a `?` at the end.

```groovy
model User {
  names    String[]?
  ages     Int?
  heights  Float?
  card     Card?
}

enum Card {
  Visa        = "VISA"
  Mastercard  = "MASTERCARD"
}
```

The default output for a nullable field is null.

#### List Types

All primitive `types`, `enums`, `relations` and `embeds` natively support lists.
Lists are denoted with `[]` at the end of the type.

```groovy
model User {
  names    String[]
  ages     Int[]
  heights  Float[]
}
```

Lists can also be optional and will give the list a 3rd state, null:

- `Blog[]`: empty list or non-empty list of blogs (default: [])
- `Blog[]?`: null, empty list or non-empty list of blogs (default: null)

The default value for a required list is an empty list. The default value for an
optional list is null.

#### Relations

Prisma provides a high-level syntax for defining relationships.

There are three kinds of relations: `1-1`, `1-m` and `m-n`. In relational
databases `1-1` and `1-m` is modeled the same way, and there is no built-in
support for `m-n` relations.

Prisma core provides explicit support for all 3 relation types and connectors
must ensure that their guarantees are upheld:

- `1-1` The return value on both sides is a nullable single value. Prisma
  prevents accidentally storing multiple records in the relation. This is an
  improvement over the standard implementation in relational databases that
  model 1-1 and 1-m relations the same, relying on application code to uphold
  this constraint.
- `1-m` The return value on one side is a optional single value, on the other
  side a list that might be empty.
- `m-n` The return value on both sides is a list that might be empty. This is an
  improvement over the standard implementation in relational databases that
  require the application developer to deal with implementation details such as
  an intermediate table / join table. In Prisma, each connector will implement
  this concept in the way that is most efficient on the given storage engine and
  expose an API that hides the implementation details.

##### One-to-One (1:1) Relationships

```groovy
model User {
  id        Int           @key
  customer  Customer?
  name      String
}

model Customer {
  id       Int     @key
  user     User?
  address  String
}
```

For 1:1 relationships, it doesn't matter which side you store the foreign key.
Therefore Prisma has a convention that the foreign key is added to the model who
appears first alphanumerically. In the example above, that's the `Customer`
model.

Under the hood, the models looks like this:

| **users** |         |
| --------- | ------- |
| id        | integer |
| name      | text    |

| **customers** |         |
| ------------- | ------- |
| id            | integer |
| user_id       | integer |
| address       | text    |

We require you to name both sides of the relationship and will error out if you
don't include both fields.

If you're introspecting an existing database and the foreign key does not follow
the alphanumeric convention, then we'll use the
`@relation(_ field: Identifier, foreignKey: Boolean?)` attribute to clarify.

```groovy
model User {
  id        Int        @key
  customer  Customer?  @relation(foreignKey: true)
  name      String
}

model Customer {
  id       Int     @key
  user     User?
  address  String
}
```

##### One-to-Many (1:N) Relationships

A writer can have multiple blogs.

```groovy
model Writer {
  id      Int     @key
  blogs   Blog[]
}

model Blog {
  id      Int     @key
  author  Writer
}
```

- `Blog.author`: points to the primary key on writer

Connectors for relational databases will implement this as two tables with a
foreign-key constraint on the blogs table:

| **writers** |         |
| ----------- | ------- |
| id          | integer |

| **blogs** |         |
| --------- | ------- |
| id        | integer |
| author_id | integer |

We require you to name both sides of the relationship and will error out if you
don't include both fields.

##### Implicit Many-to-Many (M:N) Relationships

Blogs can have multiple writers and a writer can write many blogs. Prisma
supports implicit join tables as a great for getting started.

```groovy
model Blog {
  id       Int       @key
  authors  Writer[]
}

model Writer {
  id      Int     @key
  blogs   Blog[]
}
```

Connectors for relational databases should implment this as two data tables and
a single join table.

For data sources that support composite primary keys, we'll use
`primary key(blog_id, writer_id)` to ensure that there can't be no more than one
unique association.

| **Blog** |         |
| -------- | ------- |
| id       | integer |

| **Writer** |         |
| ---------- | ------- |
| id         | integer |

| **BlogsWriters** |         |
| ---------------- | ------- |
| blog_id          | integer |
| writer_id        | integer |

##### Explicit Many-to-Many (M:N) Relationships

Many-to-many relationships are simply 2 one-to-many relationships.

```groovy
model Blog {
  id       Int       @key
  authors  Writer[]
}

model Writer {
  id      Int     @key
  blogs   Blog[]
}

// many to many
model BlogsWriters {
  blog      Blog
  author    Writer
  is_owner  Boolean
  @@unique(author, blog)
}
```

| **Blog** |         |
| -------- | ------- |
| id       | integer |

| **Writer** |         |
| ---------- | ------- |
| id         | integer |

| **BlogsWriters** |         |
| ---------------- | ------- |
| blog_id          | integer |
| author_id        | integer |
| is_owner         | boolean |

##### Self-Referential Relationships

Prisma supports self-referential relationships:

```groovy
model Employee {
  id         Int       @key
  reportsTo  Employee
}
```

| **Employee** |         |
| ------------ | ------- |
| id           | integer |
| reports_to   | integer |

##### Multiple-Reference Relationships

Models may have multiple references to the same model. To prevent ambiguities,
we explicitly link to the foreign key field using a `@relation` attribute:

```groovy
model User {
  id        Int         @key
  asked     Question[]
  answered  Question[]
}

model Question {
  id        Int   @key
  asker     User  @relation([ asked ])
  answerer  User  @relation([ answered ])
}
```

##### Referencing Primary Composite Keys

You can have also relationships to composite primary keys

```groovy
model Document {
  @@key([ projectID, revision ])

  projectID  String   @default('')
  revision   Int      @default(1)
  blocks     Block[]
}

model Block {
  id        Int       @key
  document  Document
}
```

### Attributes

Attributes modify the behavior of a field or block. Field attributes are
prefixed with a `@`, while block attributes are prefixed with `@@`.

Depending on their signature, attributes may be called in the following cases:

1. **No arguments** `@attribute`: parenthesis **must** be omitted. Examples:

   - `@key`
   - `@unique`
   - `@updatedAt`

2. **One positional argument** `@attribute(_ p0: T0, p1: T1, ...)`: There may be
   up to one positional argument that doesn't need to be named. This positional
   argument **must** appear first.

   - `@field("my_column")`
   - `@default(10)`

3. **Many named arguments** `@attribute(p1: T1, p2: T2, ...)`: There may be any
   number of named arguments. If there is a positional argument, all named
   arguments **must** follow it. Named arguments may appear in any order:

   - `@@pg.index([ email, first_name ], name: "my_index", partial: true)`
   - `@@pg.index([ first_name, last_name ], unique: true, name: "my_index")`
   - `@@check(a > b, name: "a_b_constraint")`
   - `@pg.numeric(precision: 5, scale: 2)`

**TODO** For (2) This means that we won't be able to retrofit a function with a
positional argument later without breaking existing Datamodels. This is probably
an acceptable tradeoff.

#### Field Attributes

Field attributes are marked by an `@` prefix placed at the end of the field
definition. You can have as many field attributes as you want and they may also
span multiple lines:

```
model _ {
  _ _ @attribute
}

embed _ {
  _ _ @attribute @attribute2
}

type _ _ @attribute("input")
         @attribute2("input", key: "value", key2: "value2")
         @attribute3
```

#### Core Field Attributes

Prisma supports the following core field attributes. Field attributes may be
used in `model` and `embed` blocks as well as `type` definitions. These
attributes **must** be implemented by every connector with a **best-effort
implementation**:

- `@key`: Defines the primary key
- `@unique`: Defines the unique constraint
- `@field(_ name: String)`: Defines the raw column name the field is mapped to
- `@check(_ expr: Expr, name: String?)`: Check creates a single field constraint
- `@default(_ expr: Expr)`: Specifies a default value if null is provided
- `@relation(_ field: Identifier[], foreignKey: Boolean?)`: Disambiguates
  relationships when needed
  - **field** explicitly links a field to another field
  - **foreign key** defines where the foreign key is placed
- `@updatedAt`: Updates the time to `now()` whenever the model is updated

#### Block Attributes

Field attributes are marked by an `@@` prefix placed anywhere inside the block.
You can have as many block attributes as you want and they may also span
multiple lines:

```
model _ {
  @@attribute0
  _ _ _
  @@attribute1("input")
  @attribute2("input", key: "value", key2: "value2")
  _ _ _
  @@attribute3
}

embed _ {
  @@attribute0
  _ _ _
  @@attribute1 @@attribute2("input")
}
```

#### Core Block Attributes

Prisma supports the following core block attributes. Block attributes may be
used in `model` and `embed` blocks. These attributes **must** be implemented by
every connector with a **best-effort implementation**:

- `@@key`: Defines a composite primary key across fields
- `@@unique(_ fields: Identifier[], name: String?)`: Defines a composite unique
  constraint across fields
- `@@check(_ expr: Expr, name: String?)`: Check creates a constraint across
  multiple model attributes

#### Connector Provided Attributes

In order to live up to our promise of not tailoring Prisma to the lowest-common
database feature-set, connectors may bring their own attributes to the
datamodel.

This will make your datamodel less universal, but more capable for the
datasource you're using. Connectors will export a schema of capabilities that
you can apply to your datamodel field and blocks

```groovy
connector pg {
  type = "postgres"
  url  = "postgres://localhost:5432/jack?sslmode=false"
}

connector ms {
  type = "mysql"
  url  = "mysql://localhost:5522/jack"
}

type PGCitext String @pg.Citext
type PGUUID String @pg.UUID

embed Point2D {
  X Int
  Y Int
  @@pg.Point
  @@ms.Point
}

embed Point3D {
  X Int
  Y Int
  Z Int
  @@pg.Point
  @@ms.Point
}

model User {
  id         UUID
  email      Citext
  location1  Point2D
  location2  Point3D
}
```

### Extending Models

You can break up large models into separate fragments by placing a model inside
another model:

```groovy
model Human {
  id      Int     @key
  name    String
  height  Int
}

model Employee {
  Human
  employer  String
  height    Float
}
```

In this example, Employee extends Human. Writing this manually it look like
this:

```groovy
model Employee {
  id        Int     @key
  name      String
  employer  String
  height    Float
}
```

- If there's a type conflict, the child's type will have precedence. In this
  case `Employee.height` overrides `Human.height`.

## Enum Block

Enums must include their corresponding value. This determines how an `enum` is
stored under the hood.

```groovy
enum Color {
  Red  = "RED"
  Teal = "TEAL"
}
```

```groovy
enum Status {
  STARTED = 0
  DOING   = 1
  DONE    = 2
}
```

An enum **cannot** have multiple different types:

```groovy
// not supported
enum Color {
  Red  = "RED"
  Teal = 10
}
```

For now, we'll only support `String` and `Int` enum value types.

## Embed Block

Embeds are supported natively by Prisma. There are 2 types of embeds: named
embeds (just called embeds) and inline embeds.

Unlike relations, embed tells the clients that this data _comes with the model_.
How the data is actually stored (co-located or not) is not a concern of the data
model.

```groovy
model User {
  id        String
  customer  StripeCustomer?
}

embed StripeCustomer {
  id     String
  cards  Source[]
}

enum Card {
  Visa        = "VISA"
  Mastercard  = "MASTERCARD"
}

embed Sources {
  type Card
}
```

### Inline Embeds

There's another way to use embeds.

When you don't need to reuse an embed, inline embeds are handy. Inlines embeds
are supported in `model` and `embed` blocks. They can be nested as deep as you
want. Please don't go too deep though.

```groovy
model User {
  id        String
  customer  {
    id     String
    cards  {
      type Card
    }[]
  }?
}

enum Card {
  Visa        = "VISA"
  Mastercard  = "MASTERCARD"
}
```

## Type Definition

Type definitions can be used to consolidate various type implementations into
one type.

```groovy
type Numeric Int @pg.numeric(precision: 5, scale: 2)
                 @ms.decimal(precision: 5, scale: 2)

model Customer {
  id       Int      @key
  weight   Numeric
}
```

**TODO:** Consider replacing Type Definition with Attribute Definition below

## Attribute Definition

Attribute definitions can be used to consolidate attributes so you're not
repeating yourself.

```groovy
attr numeric @pg.numeric(precision: 5, scale: 2)
             @ms.decimal(precision: 5, scale: 2)

model Customer {
  id       Int  @key
  weight   Int  @numeric
}
```

## Import Block

You can use the `import` block to stitch together a schema across an
organization.

The import will recursively resolve imported imports.It will include all blocks
and type definitions by default. In the future, we may want to lock this down
more with a `@@public` attribute.

```groovy
env profile_url {
  type = "string"
}

env feed_url {
  type = "string"
}

import profile {
  from = profile_url
}

import feed {
  from = feed_url
}

model User {
  profile.User
  feed.User
}
```

## Function

Prisma core provides a set of functions that **must** be implemented by every
connector with a **best-effort implementation**. Functions only work inside
field and block attributes that accept them.

- `uuid()` - generates a fresh UUID
- `cuid()` - generates a fresh cuid
- `between(min, max)` - generates a random int in the specified range
- `now()` - current date and time

Default values using a dynamic generator can be specified as follows:

```groovy
model User {
  age        Int       @default(between([ 1, 5 ]))
  height     Float     @default(between([ 1, 5 ]))
  createdAt  Datetime  @default(now())
}
```

These are backed in the database where it's supported, otherwise they're
provided at the Prisma level by the query engine.

The data types that these functions return will be defined by the connectors.
For example, `now()` in Postgres will return a `timestamp with time zone`, while
`now()` with a JSON connector would return an `ISOString`.

## Boolean Expressions

Boolean expressions are expressions that evaluate to either true or false. Some
examples include:

- a > b
- a == b
- a <= b
- a && b
- a && (b || c)

Like functions, boolean expressions are backed by the database where it's
supported, otherwise they're provided at the Prisma level by the query engine.

Prisma core provides a set of functions that **must** be implemented by every
connector with a **best-effort implementation**. Boolean expressions only work
inside field and block attributes that accept them.

- `==`: equal to
- `>`: greater than
- `>=`: greater than or equal
- `<`: less than
- `<=`: less than or equal
- `&&`: and
- `||`: or

**TODO:** Figure out how this would actually work as part of the exported
connector schema.

## Configuration Layout

Our prisma datamodel is designed to serve the requirements of a variety of
consumers.

### Soloists

Single developer just wants to shove everything in a single file. They don't
want to clutter their application up with your configuration. This is supported
by using the same language for configuration and your datamodel.

```sh
/app
  datamodel.prisma
```

### Team

A team may have a lot of configuration or many different models. They may also
have many environments they need to deploy to. We support a couple options to
break up the datamodel.

#### Multiple .prisma in the same directory get concatenated

```sh
/app
  /prisma
    user.prisma
    post.prisma
    comment.prisma
```

The above could also address the preference to separate configuration from the
datamodel. For example:

```sh
/app
  /prisma
    schema.prisma
    connectors.prisma
    generators.prisma
```

#### Multiple directories for different environments

If your environment widely vary, you can separate environment by directory. You
can use symlinks or the `import` block to share configuration:

```sh
/app
  /prisma
    /production
      user.prisma
      post.prisma
      comment.prisma
    /development
      user.prisma
      post.prisma
      comment.prisma
```

You should only reach for this capability if you absolutely need it.
[Twelve-Factor App](https://12factor.net/config?S_TACT=105AGX28) encourages you
to only manage environment differences inside environment variables. We supports
this use case very well with `env` blocks.

### Organization

Different parts of the datamodel are owned by different teams across the
organization. These teams may have different access control requirements. Prisma
configuration can be spread across a network.

You can use the `import` block to stitch together a schema across an
organization:

```groovy
env ACL {
  type = "string"
}

import profile {
  from = ACL.PROFILE_URL
}

import feed {
  from = ACL.FEED_URL
}

model User {
  profile.User
  feed.User
}
```

## Auto Formatting

Following the lead of [gofmt](https://golang.org/cmd/gofmt/) and
[prettier](https://github.com/prettier/prettier), our syntax ships with a
formatter for `.prisma` files.

Like `gofmt` and unlike `prettier`, we offer no options for configurability
here. **There is one way to format a prisma file**.

This strictness serves two benefits:

1. No bikeshedding. There's a saying in the Go community that, "Gofmt's style is
   nobody's favorite, but gofmt is everybody's favorite."
2. No pull requests with different spacing schemes.

### Formatting Rules

#### Configuration blocks are align by their `=` sign.

```
block _ {
  key      = "value"
  key2     = 1
  long_key = true
}
```

Formatting may be reset up by a newline.

```
block _ {
  key   = "value"
  key2  = 1
  key10 = true

  long_key   = true
  long_key_2 = true
}
```

Multiline objects follow their own nested formatting rules:

```
block _ {
  key   = "value"
  key2  = 1
  key10 = {
    a = "a"
    b = "b"
  }
  key10 = [
    1,
    2
  ]
}
```

#### Field definitions are aligned into columns separated by 2 or more spaces.

```
block _ {
  id          String       @key
  first_name  LongNumeric  @default
}
```

Multiline field attributes are properly aligned with the rest of the field
attributes:

```
block _ {
  id          String       @key
                           @default
  first_name  LongNumeric  @default
}
```

Formatting may be reset by a newline.

```
block _ {
  id  String  @key
              @default

  first_name  LongNumeric  @default
}
```

Inline embeds follow their own nested formatting rules:

```groovy
model User {
  id    String
  name  String
  customer {
    id         String
    full_name  String
    cards  {
      type  Card
    }[]
  }?
  age   Int
  email String
}
```

## Grammar

You can copy & paste the following syntax into https://pegjs.org/online to
experiment with the Prisma syntax.

```pegjs
DataModel   = body:Body {
  return body
}

Body         = body:(_ (Attribute / Block / Comment /* / OneLineBlock */) _)* {
  return body.map(inner => inner[1])
}

Attribute = ident:Identifier _ "=" _ expr:Expression Newline {
  return {
    type: 'Attribute',
    ident: ident,
    expr: expr,
  }
}

Field = ident:Identifier __ type:FieldType attrs:(__ (Comment / FieldAttribute))* {
  return {
    type: "field",
    ident: ident,
    fieldType: type,
    attrs: attrs.map(attr => attr[1])
  }
}

// TODO: FieldType should be narrowed to primitive or block name
FieldType = EmbeddedField / ListFieldType / FieldRelation / OptionalFieldType / RequiredFieldType
// TODO: we probably want attributes to have the same shape
FieldAttribute = '@' attr:(FunctionCall / SelectorExpression / Identifier) {
  return attr
}

ModelAttribute = '@' '@' attr:(FunctionCall / SelectorExpression / Identifier) {
  return attr
}

ListFieldType = RequiredFieldType '[]' '?'?
RequiredFieldType = Identifier
OptionalFieldType = Identifier '?'

FieldRelation = (SelectorExpression / Identifier) Arguments

Block = ident:Identifier idents:(_ (StringLit / Identifier))* _ "{" body:BlockBody "}" {
  return {
    type: 'Block',
    idents: [].concat(ident).concat(idents.map(ident => ident[1])),
    body: body,
  }
}

BlockBody = body:(_ (Attribute / EmbeddedField / Comment / Field /  EmbeddedBlock / ModelAttribute) _)* {
  return body.map(inner => inner[1])
}

EmbeddedField = ident:Identifier __ type:EmbeddedFieldMap
EmbeddedFieldMap = '{' BlockBody '}' ('[]' / '?')?

// OneLineBlock = Identifier (StringLit / Identifier)* "{" (Identifier "=" Expression)? "}" Newline

// TODO: the end of this is hacked on to support: @postgres.index(first_name + " " + last_name)
// I doubt order of operations is properly supported here.
Expression = expr:(Operation / /*Conditional /*/ ExprTerm) (_ binaryOperator _ ExprTerm)* {
  return {
    type: 'Expression',
    expr: expr
  }
}

EmbeddedBlock = Identifier

// TODO group multiline comments
// like this comment
Comment = '//' comment:(!Newline AnyCharacter)* {
  return {
    type: 'Comment',
    value: comment.map(c => c[1]).join('').trim()
  }
}

// Conditional = ExprTerm _ "?" _ ExprTerm _ ":" _ ExprTerm

// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#expression-terms
ExprTerm = term:(
    CollectionValue
    / LiteralValue
    // / HeredocTemplate
    / StringLit
    / FunctionCall
    / VariableExpr
    // / ForExpr
    / Index
    / GetAttr
    // / Splat
    / "(" Expression ")"
  ) {
    return term
  }

// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#operations
Operation = unaryOp / binaryOp;
unaryOp = ("-" / "!") ExprTerm;
binaryOp = ExprTerm _ binaryOperator _ ExprTerm;
binaryOperator = compareOperator / arithmeticOperator / logicOperator;
compareOperator = "==" / "!=" / "<" / ">" / "<=" / ">="
arithmeticOperator = "+" / "-" / "*" / "/" / "%"
logicOperator = "&&" / "||" / "!"

// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#template-expressions
// HeredocTemplate = (
//     ("<<" / "<<-") Identifier Newline
//     chars
//     Identifier Newline
// );

// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#variables-and-variable-expressions
VariableExpr = ident:(SelectorExpression / Identifier) {
  return {
    type: "Variable",
    value: ident
  }
}

// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#functions-and-function-calls
FunctionCall = ident:Identifier args:Arguments {
  return {
    type: "Function",
    ident: ident,
    arguments: args
  }
}

Arguments = '(' list:ArgumentList* ')' {
  return list
}

ArgumentList = head:Expression tail:(_ ',' _ Expression)* {
  return [head].concat(tail.map(t => t[3]))
}

// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#for-expressions
// ForExpr = forTupleExpr / forObjectExpr
// forTupleExpr = "[" forIntro Expression forCond? "]"
// forObjectExpr = "{" forIntro Expression "=>" Expression "..."? forCond? "}"
// forIntro = "for" Identifier ("," Identifier)? "in" Expression ":"
// forCond = "if" Expression

// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#index-operator
Index = "[" Expression "]"

// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#attribute-access-operator
GetAttr = "." Identifier;

// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#splat-operators
// Splat = attrSplat / fullSplat
// attrSplat = "." "*" GetAttr*
// fullSplat = "[" "*" "]" (GetAttr / Index)*

// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#literal-values
LiteralValue = term:(NumericLit / "true" / "false" / "null") {
  return term
}

// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#collection-values
CollectionValue = tuple / object
tuple = "[" _ exprs:((Expression ((_ "," _ / AtleastNewline) Expression)* (_ "," _ / AtleastNewline)?)? ) _ "]" {
  return {
    type: "Array",
    // TODO:
    exprs: []
  }
}
object = "{" _ (((objectelem / Comment) ((_ "," _ / AtleastNewline) (objectelem / Comment))* (_ "," _ / AtleastNewline)? )?) _ "}" {
  return {
    type: "Object",
    // TODO:
    expr: {}
  }
}
objectelem = (Identifier / Expression) _ "=" _ Expression Comment?

SelectorExpression = ident:Identifier "." expr:ExprTerm {
  return {
    type: "SelectorExpr",
    ident: ident,
    expr: expr
  }
}

// TODO not 100% compliant:
// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#identifiers
Identifier = ID_Start ID_Continue* {
  return {
    type: 'Identifier',
    value: text()
  }
}
ID_Start = id:[A-Za-z] { return id }
ID_Continue = id:[A-Za-z0-9_-] { return id }

// TODO might not be 100% compliant: I wasn't able to follow the spec here.
// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#template-expressions
StringLit = quotation_mark chars:chars quotation_mark {
  return {
    type: 'StringLit',
    value: chars,
  }
}

Newline = "\n"

// https://github.com/hashicorp/hcl2/blob/master/hcl/hclsyntax/spec.md#numeric-literals
// TODO: support floats and exponentials
NumericLit = int:decimal+ float:("." decimal+)? exp:(expmark decimal+)? {
  return parseInt(int, 10)
}
decimal    = [0-9]
expmark    = ('e' / 'E') ("+" / "-")?

chars = chars:char* {
  return text()
}

// Pulled from: https://github.com/pegjs/pegjs/blob/master/examples/json.pegjs
char
  = unescaped
  / escape
    sequence:(
        '"'
      / "\\"
      / "/"
      / "b" { return "\b"; }
      / "f" { return "\f"; }
      / "n" { return "\n"; }
      / "r" { return "\r"; }
      / "t" { return "\t"; }
      / "u" digits:$(HEXDIG HEXDIG HEXDIG HEXDIG) {
          return String.fromCharCode(parseInt(digits, 16));
        }
    )
    { return sequence; }
escape
  = "\\"
quotation_mark
  = '"'
unescaped
  = [^\0-\x1F\x22\x5C]
// See RFC 4234, Appendix B (http://tools.ietf.org/html/rfc4234).
DIGIT  = [0-9]
HEXDIG = [0-9a-f]i

// Whitespace alias
_ "whitespace" = ws*
__ "whitespace" = ws+
ws = [ \t\n\r]

// Require atleast a newline, we can have more whitespace
// but one newline is required
AtleastNewline = [ \t\r]* Newline [ \t\r]*

AnyCharacter = . { return text() }
```

**NOTE:** This was an MVP and is primarily used to quickly iterate and ensuring
our grammar didn't have ambiguities. It should only be used as a reference
implementation as it may be out of date with the final parser.
