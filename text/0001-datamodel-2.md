# Datamodel 2 (Experimental Syntax)

This RFC attempts to improve the syntax laid out in
[current RFC on Data Model 2](https://github.com/prisma/rfcs/blob/datamodel/text/0000-datamodel.md#1-1),
spec out a few additional features and unify some concepts.

This RFC is a potential answer to an open question posed in the previous RFC:

> "If we were a little more radical with the syntax, could we create something
> much better?"

## Motivation

If we're changing the syntax to something new anyway, which others will
(skeptically) have to learn, then we might as well try to make it a beautiful
dialect for data modeling.

Getting this right will be extremely important. Whether you're introspecting or
starting from scratch, the datamodel is the first window into the Prisma world
▲. We want to make a good impression.

My goal here is to explore the space a bit further and hopefully reach consensus
on a syntax we can all get behind.

## Requirements

- Break from the existing GraphQL SDL syntax where it makes sense
- Clearly separate responsibilities into two categories: Core Prisma primitives
  and Connector specific primitives
- High-level relationships without ambiguities
- Easily parsable
  ([avoid symbol tables, ideally](https://golang.org/doc/faq#different_syntax))
- Abstraction over raw column names via field aliasing
- Can be rendered into JSON

## Nice to Have

- One configuration file for prisma (WIP)
- Strict Machine formatting (␀ bikeshedding)
- Multi-line support and optional single-line via commas

## Summary of adjustments from previous DM2 RFC

- Removed colon between name and type
- Moved model Attributes from the top of the model into the block
- Lowercase primitives, capitalized higher-level types
- Removed `ID` as a primitive type
- Merged prisma.yml configuration into the datamodel
- introduced `source` block for connectors
- Renamed `embedded` to `embed`
- Replaced `=` in favor of `default(...)`
- Added "embedded embeds"
- Added metadata support to any block
- Added top-level configuration support
- Added connector constraints
- Revised connector-specific types
- Added model embedding (fragments in GraphQL)
- Adjust many-to-many convention `BlogToWriter` to `BlogsWriters`
- Prefix embedded embeds with `embed`
- introduce `@as` for type specifications
- rename `bool` to `bool`

## Basic Example

This example illustrate many aspects of the proposed syntax:

```groovy
postgres_url = env("POSTGRES_URL")
mongo_url    = env("MONGO_URL")

// postgres datasource
source postgres {
  type    = "postgres"
  url     = postgres_url
  default = true
}

// mongo datasource
source mongo {
  type = "mongo"
  url  = mongo_url
}

// connect to our secondary Mongo DB
source mgo2 {
  type     = "mongo"
  host     = "localhost"
  port     = 27017
  database = "neato2"

  query = {
    sslMode = "disable"
  }
}

// type definition
type Numeric Decimal = "2.0" @postgres.Numeric(5, 2)

enum Color {
  Red  = "RED"
  Teal = "TEAL"
}

model User {
  meta = {
    // adjust name from "users" convention
    name = "people"
  }

  // model fields
  id             int       @primary
  email          string    @unique  @postgres.Like(".%.com") @as(postgres.Citext)
  name           string?   @check(name > 2)
  role           Role
  profile        Profile?               @alias("my_profile")
  createdAt      datetime 
                                        @alias("ok")
  updatedAt      datetime               @onChange(now())

  weight         Numeric   @alias("my_weight")
  posts          Post[]


}
// composite indexes (last field optional)
@unique(email, name, "email_name_index")

enum Role {
  USER   enum  // unless explicit, defaults to "USER"
  ADMIN  enum  @default("A")
}

model Profile {
  meta = {
    from = mongo
    name = "people_profiles"
  }

  // model fields
  id       int            @primary
  author   User(id)
  bio      string

  // nullable array items and nullable photos field
  photos   Photo?[]?
}

// named embed (reusable)
embed Photo {
  id   string    @as(mgo2.ObjectID)
  url  string

  // anonymous embed (optional)
  embed size {
    height  int
    width   int
  }?

  // anonymous embed
  embed alternatives {
    height  int
    width   int
  }[]
}

model Post {
  // intentionally messy 😅
  // multi-line support, separated with commas
  id  int  @primary,
           @serial,
           // this is okay too...
           @default("some default") // default value

  title      string
  author     User(id)
  reviewer   User(id)
  published  bool      @default(false)

  createdAt  datetime  @default(now())
  updatedAt  datetime  @onChange(now())

  categories CategoriesPosts[]
}
@unique(title, author)

model MoviePost {
  // mixin Post's fields into MoviePost
  // Based on Go's struct embedding syntax
  Post

  stars   string
  review  string

  // duplicate title would replace the included model field
  title   string
}

// Comments in this datamodel have meaning. Look to godoc for inspiration:
// https://blog.golang.org/godoc-documenting-go-code

// Comments directly above a model are attached to the model
model Category {
  id     int                @primary
  name   string
  posts  CategoriesPosts[]
}

// Many-to-Many naming based on Rails conventions
model CategoriesPosts {
  post      Post(id)
  category  Category(id)
}
@unique(post, category)
```

## Motivation for core/connector split

> TODO: connector or data source?

Prisma core provides a set of data modeling primitives that are supported by
most connectors.

Connectors enable Prisma to work with a specific datastore. This can be a
database, a file format or even a web service.

- The primary job of a connector is to translate highher level Prisma concepts
  to lower level operations on the storage engine.
- Secondary, connectors enable the Prisma user to take full advantage of the
  underlying storage engine by exposing performance tuning options and extra
  functionality that can not be accessed through the core Prisma primitives.

**Core Prisma primitives**

Prisma provides a set of core primitives:

- Primitive Types (string, int, float, bool)
- Type Definitions (model, embed, enum, type)
- Relations
- Generators
- Constraints
- ID / Primary Key

**Connector specific primitives**

Connectors can additionally provide primitives specific to their underlying
datastore

- Type Specification
- Custom Primitive Types
- Custom Complex Types
- Connector Specific Constraints
- Indexes
- Connector specific generators

The Prisma Datamodel provides primitives that describe the structure of your
databases. Core Primitives are so fundamental that they map to most database
types. Some primitives are there only to express some special feature in a
single database. We call them Connector specific primitives. The following
sections describe why each primitive is either core or connector specific.

### Types (String, Int, Float etc)

Prisma specifies a common set of primitive types. Connectors have some
flexibility in how they implement the type, but there are minimum requirements
that must be satisfied.

This makes it possible to use diferent connectors in different environments.

### Type Declarations (model, embed, enum, type)

`model` declares a top level type that can be referenced. `embed` declares a
complex structure that is embed in a top level type. `enum` declares a set of
possible values. `type` allows you to extend primitive types with additional
constraints or database-specific types.

Prisma includes these 4 type declarations. It is not possible for a connector to
introduce a new type declaration.

### Relations

The concept of relations is provided by Prisma core. Relations between models is
fundamental to databases and web services dealing with data. Much of Prismas
value proposition is around making working with data relations easier (nested
mutations, relational filters, sub-selection), so requiring connectors to adhere
to a specific style of relations is worth it.

### Generators

Prisma core provides a set of connector agnostic generators. Additionally,
connectors can provide generators that take advantage of database specific
capabilities such as Sequences in Postgres.

### Constraints

Constraints such as uniqueness or value ranges are a data modeling concern. As
such, they are provided by Prisma core. Most connectors implement all the core
constraints, making it easy to use different connectors in different
environments.

Constraints that are bound to the data record (for example `age > 18`) are
implemented in Prisma Core. Constraints that must access other records (for
example uniqueness) must be implemented by the connectors as they can take
advantage of lower level primitives such as indexes and database triggers.

### ID / Primary Key

Prisma's relations rely on a primary key to uniquely identify a single record.
Therefore, primary keys are considered a data modeling concern and provided by
Prisma core.

### Generators

Prisma has a number of built in generators that produce one of the built in
primitive types without knowledge of the underlying datastore. Examples include
`now()`, `cuid()` and any literal value such as `42`, `"answer"` or `[1, 2, 3]`

Connectors can provide additional generators that take advantage of low level
storage engine primitives such as `SEQUENCE` in a relational database.

### Type Specification

Connectors can provide extra hooks to configure how the underlying storage
engine treats generic Prisma types. For example, the MySQL connector enables you
to specify that a `string` field is stored in a `varchar(100)` column.

Storage engines vary wildly, so it is not possible to have a generic interface
for deciding low level configuration.

### Custom Primitive Types and Custom Complex Types

Connectors can introduce primitive or complex types. These types can be used the
same way as a type extensions or complex type (model, embed or enum) declared
directly in the datamodel file.

### Indexes

Indexes are storage engine specific and mostly relevant for performance
configuration. Therefore, index configuration is provided by connectors.

There are two exceptions where indexes intersect with data modeling:

- Most storage engines implement the _unique constraint_ as an index. Unique
  constraint is provided by Prisma core, and the connector can choose to create
  only a single index if both a index and unique constraint is present on a
  single field in the datamodel.
- The concept of a _primary key_ is provided by Prisma Core (`@primary`). Many
  storage engines implement the primary key using a special index (sometimes
  called clustered index or primary index) that organises the data on disk by
  that field, even if no index is specified separately. These connectors will
  allow you to configure the index used for the primary key separately using the
  connector specific index configuration.

## Configuration

Prisma's configuration and datamodel share the same language. The configuration
and datamodel can be in 1 file, spread across multiple files, or spread across
the network.

For syntax, we're a superset of
[Terraform's HCL2](https://github.com/hashicorp/hcl2), so the configuration
format of Terraform would apply here:

```groovy
input postgres_url {
  description = "postgres connection string"
  default     = "postgres://localhost:5432/db"
  type        = "string" // optional
}

input mongo_url {
  description = "mongo connection string"
  default     = env("MONGO_URL")
}

// datasource named "pg"
source pg {
  type  = "postgres"
  url   = postgres_url
}

source mongo {
  url     = mongo_url
}

generate javascript {
  output = "photon/js"
  target = "es3"
}

generate typescript {
  output = "photon/ts"
  target = "es5"
}

generate go {
  output = "photon/go"
}
```

### Splitting Configuration

When your datamodels get sufficiently large, you'll look to split up the
configuration. Prisma offers 3 ways to do this, depending on the use case.

#### Multiple .prisma files in the same directory get concatenated

```sh
/app
  /prisma
    config.prisma
    comment.prisma
    post.prisma
    user.prisma
```

This will also address many people's preference to separate configuration from
the datamodel.

#### Connecting .prisma files across the network

Terraform's [module](https://www.terraform.io/docs/configuration/modules.html)
concept maps pretty well to our use case.

```groovy
import User {
  from = "git::https://github.com/my/app.git?ref=master//user"
}

// in github.com/my/app private repo
model User {
  id          int     @primary
  first_name  string
  last_name   string
}
@unique(first_name, last_name)
```

Terraform's module concept doesn't map well in 4 cases:

- `variable` it seems like a nice feature to support module scoping, but
  Terraform's configuration is fundamentally a directed-acyclic graph (DAG).
  They use variable and outputs to allow you to treat a remote DAG as a single
  node. We'll also have a DAG when it comes to the order in which tables are
  created and destroyed, but it's not clear to me whether we'd need to configure
  these schemas from the outside.
- `output` is an important concept for terraform modules as it allows you to
  treat very complex graph dependencies as a single node. This seems less
  important for Prisma's use case.
- I don't know how terraform handles non-resources inside modules. In practice
  this doesn't come up much because people know what they're doing, but I need
  to double-check. I think the appropriate answer is to only allow a subset of
  type definitions inside a module (e.g. model)

For the `from` parameter, we should definitely copy how terraform does it.
They've created a
[feature-rich fetching library](https://github.com/hashicorp/go-getter) that can
download dependencies from variety of protocols (local and remote).

**TODO** Finish thinking through what parts of Prisma's use cases should be
inspired by infrastructure as code (terraform / cloudformation / etc.) and what
makes our use cases fundamentally different.

#### Extending Models

Datamodel 2 also offers a way to break up large models.

```groovy
model Human {
  id     int     @primary
  name   string
  height int
}

model Employee {
  Human
  employer  string
  height    float
}
```

In this case Employee extends Human and would result in this:

```groovy
model Employee {
  id        int     @primary
  name      string
  employer  string
  height    float
}
```

#### Multiple directories for different Prisma environments

Finally, when your environments are wildly different, the datamodel supports
different Prisma environments in different directories:

```sh
/app
  /prisma
    /development
      config.prisma
      comment.prisma
      post.prisma
      user.prisma
    /staging
      config.prisma
      comment.prisma
      post.prisma
      user.prisma
    /production
      config.prisma
      comment.prisma
      post.prisma
      user.prisma
```

It's important to know that this should used as an escape hatch.
[Twelve-Factor App](https://12factor.net/config?S_TACT=105AGX28) encourages you
to manage environment differences inside environment variables.

## Types

### Primitive Types

Prisma 1.x lacks some primitive types. Other types are mapping to the wrong
storage engine type.

See https://github.com/prisma/prisma/issues/1753

#### Float

Our `float` primitive type is implemented as
`NUMERIC(precision = 65, scale = 30)`
([doc](https://github.com/prisma/prisma/issues/2934#issuecomment-451545681)).
There is no way to actually get a float.

We will change float to use the largest available floating point primitive type:

| GraphQL | MySQL             | Elastic Search                              | MongoDB    | PostgreS           |
| :------ | ----------------- | ------------------------------------------- | ---------- | ------------------ |
| Float   | FLOAT, **DOUBLE** | float, **double**, half_float, scaled_float | **Double** | float4, **float8** |

Most storage engines will provide a double/float8. It is possible to use type
specification to change to float/float4.

#### Decimal

The decimal type should map to a predefined configuration of the exact precision
type. We have used `NUMERIC(precision = 65, scale = 30)` since the beginning of
Graphcool and has never received a complaint, so we will use that as the default
on SQL storage engines. It is possible to use type specification to change this.

Elastic seems uninterested in supporting Decimal (BigDecimal):
https://github.com/elastic/elasticsearch/issues/17006

Mongo supports decimal (128-bit decimal floating point, not configurable)

> TODO: should we leave this to type specifications?

#### String

Prisma has only one `String` type that maps to the largest available text
representation supported by the storage engine. We make no effort to unify size
constraints across connectors.

Type specification can be used to specify a smaller storage type for
performance. On SQL it is common to use `varchar(128)`.

#### Binary

> DRAFT

> Task:
>
> Valid oprations and filters need to be specified

| GraphQL | MySQL                   | Elastic Search | MongoDB | PostgreS |
| :------ | ----------------------- | -------------- | ------- | -------- |
| Binary  | Binary, VarBinary, Blob | Binary         | binData | bytea    |

In practice, a binary type is a string type without collation.

#### JSON

> DRAFT

> Note: SQL connectors will implement Embedded types using JSON columns.
> Embedded types are different from JSON fields in that they have a schema that
> is enforced by Prisma on write

JSON is treated as a schema-less JSON value. Prisma validates that inserted
values are well-formed JSON.

> TASK:
>
> We should support generic JSON manipulation, ideally similar to the Mongo API.
> It should work the same across all connectors. We should support indexing
> Consider how explicit null vs not even in the document is handled

| GraphQL | MySQL | Elastic Search | MongoDB | PostgreS |
| :------ | ----- | -------------- | ------- | -------- |
| JSON    | JSON  | Object         | Object  | JSON     |

#### Datetime Types

Prisma date and time types are always following ISO 8601. DateTimes are always
stored with timezones.

Prisma support the 3 DateTime types: `DateTime`, `Date`, `Time`

| Prisma              | MySQL     | Elastic | Mongo | Postgres  |
| ------------------- | --------- | ------- | ----- | --------- |
| DateTime            | TIMESTAMP | -       | Date  | TIMESTAMP |
| Date                | DATE      | -       | -     | DATE      |
| Time                | TIME      | -       | -     | TIME      |
| Interval (from-to)  | -         | -       | -     | -         |
| Duration (timespan) | -         | -       | -     | INTERVAL  |

Elastic does not support `DataTime`. Instead DateTime is stored as a long number
representing milliseconds-since-the-epoch. DateTimes are returned by Elastic as
rendered strings. Prisma will convert them to be consistent with other
connectors.

Mongo and Elastic does not natively support `Date`. Prisma will simply map to
DateTime and set the time component to 0.

Mongo and Elastic does not support `Time`. Prisma will simply map to Int and
store a millisecond offset from midnight.

> Note: While Interval and Duration might be useful, Prisma does not specify
> these and individual connectors are free to implement this as needed

#### Spatial Types

> DRAFT

Sptial data types only make sense if they are augmented with proper operations,
like intersection tests or area calculation. PostGIS has some
[nice documentation](https://postgis.net/docs/manual-2.5/using_postgis_dbmanagement.html#PostGIS_GeographyVSGeometry)
which can serve as a starting point.

For spatial types, two conventions are meaningful:

- Geographic coordinates (lat, lon)
- Geometric coordinates (x, y)

| Prisma  | MySQL      | Elastic Search      | MongoDB    | PostgreS   |
| :------ | ---------- | ------------------- | ---------- | ---------- |
| Point   | POINT      | geo_shape/geo_point | Point      | POINT      |
| Line    | LINESTRING | geo_shape           | LineString | LINESTRING |
| Polygon | POLYGON    | geo_shape           | Polygon    | POLYGON    |

> Task:
>
> Decide what spatial primitive types Prisma should support
>
> Valid operations and filters need to be specified
>
> Can we leave this to type specifications? What's the underlying value here,
> float[]?

### Enums

declaring and using an enum:

```groovy
model Primitives {
  enum SomeEnum
}

enum SomeEnum {
  SomeEnumValue       int
  SomeOtherEnumValue  int
}
```

The following table specifies how connectors will implement enums. Note that
Prisma 1.x implements enums as a string, even when a dedicated ENUM type is
available.

| Prisma | MySQL | Elastic Search | MongoDB | PostgreS |
| :----- | ----- | -------------- | ------- | -------- |
| Enum   | ENUM  | text           | String  | ENUM     |

#### Ordered Enum values

Connectors for databases without native support for enums will store enums as
strings containing the name of the enum value. In the future we could add a
feature to specify an int representing the enum value similar to how protobuf
specifies the order of fields. This minimises space use and simplifies renaming
of values. This new feature will be backwards compatible:

```groovy
enum SomeEnum {
  SomeEnumValue      int @default(1)
  SomeOtherEnumValue int @default(2)
}
```

### Relations

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
- `1-m` The return value on one side is a nullable single value, on the other
  side a list that might be empty.
- `m-n` The return value on both sides is a list that might be empty. This is an
  improvement over the standard implementation in relational databases that
  require the application developer to deal with implementation details such as
  an intermediate table / join table. In Prisma, each connector will implement
  this concept in the way that is most efficient on the given storage engine and
  expose an API that hides the implementation details.

#### 1-1 (one-to-one)

```groovy
model User {
  id        int           @primary
  customer  Customer(id)?
  name      string
}

model Customer {
  id       int     @primary
  user     User?
  address  string
}
```

The relationship can be made on either side, but the `(id)` indicates where the
data is stored. You can think of this as a pointer to the Customer id field.
`customer_id` will be the same data type as `Customer.id`, in this case `int`.

Under the hood, the data looks like this:

| **User** |             |        |
| -------- | ----------- | ------ |
| id       | customer_id | name   |
| int      | int or null | string |

| **Customer** |         |
| ------------ | ------- |
| id           | address |
| int          | string  |

- `(id)` is required here because we need to know where to store the id

##### Rules for Nullability

- When `Customer(id)?` is nullable, the back-relation `User?`, must also be
  nullable.

#### 1-M (one-to-many)

A writer can have multiple blogs

```groovy
model Writer {
  id      int        @primary
  blogs   Blog[]
}

model Blog {
  id      int        @primary
  author  Writer(id)
}
```

- `author Writer(id)` points to the `id int` on `Writer` model, establishing the
  has-many relationship.
- We can make `(id)` optional here as it can default to the primary key
- `blogs Blog[]` names the back-relation, but is entirely optional

Connectors for relational databases will implement this as two tables with a
single relation column, exactly like the 1-1 relation:

| **Writer** |
| ---------- |
| id         |
| int        |

| **Blog** |           |
| -------- | --------- |
| id       | author_id |
| int      | int       |

##### Implicit (optional) relation field

It is possible to specify only one relation field:

```groovy
model Writer {
  id      int        @primary
  blogs   Blog[]
}

model Blog {
  id      int        @primary
}
```

This will be interpreted as if there was an implicit optional single item
relation field on `Blog` named after the `Writer` model:

```groovy
model Writer {
  id      int        @primary
  blogs   Blog[]
}

model Blog {
  id         int         @primary
  writer_id  Writer(id)?
}
```

| **Writer** |
| ---------- |
| id         |
| int        |

| **Blog** |           |
| -------- | --------- |
| id       | writer_id |
| int      | int       |

In this case `blogs` is either an empty list or a non-empty list. There is no
distinction between an optional and required many field.

If `Blog[]?` is provided, we should warn the user that the `?` has no affect.

Doing it the other way would not work as it would result in a 1-1 relation.

#### M-N (Implicit Many-to-Many)

Blogs can have multiple writers and a writer can write many blogs. Prisma
supports implicit join tables as a great for getting started.

```groovy
model Blog {
  id       int       @primary
  authors  Writer[]
}

model Writer {
  id      int      @primary
  blogs   Blog[]
}
```

Connectors for relational databases will implement this as two data tables and a
single join table. For data sources that support composite uniques, we'll use
`unique(blog_id, writer_id)` to ensure that there can't be more than one unique
association.

| **Blog** |
| -------- |
| id       |

| **Writer** |
| ---------- |
| id         |

| **\_BlogsWriters** |           |
| ------------------ | --------- |
| blog_id            | writer_id |

Many-to-Many's are always awkward to name. We'll follow
[ActiveRecord's conventions](https://guides.rubyonrails.org/association_basics.html#the-has-and-belongs-to-many-association)
until you decide to provide an explicit Join Table:

**Required relations**

Many-to-Many relations makes no distinction between required or optional

**One-sided Implicit relation field**

This would make the relationship a One-to-Many.

> TODO: Is there a technical reason for the underscore? Is it just to avoid name
> clashes?

#### M-N (Explicit Join Table)

As you get farther along, you may will likely want to attach metadata to the
association. To enable these workflows, Prisma supports a seamless upgrade to
Explicit Join Tables.

```groovy
model Blog {
  id       int        @primary
  authors  Writer[]
}

model Writer {
  id      int      @primary
  blogs   Blog[]
}

// many to many
model BlogsWriters {
  blog      Blog(id)
  author    Writer(id)
  is_owner  bool
}
// enforce a composite unique
@unique(author, blog)
```

| **Blog** |
| -------- |
| id       |

| **Writer** |
| ---------- |
| id         |

| **BlogsWriters** |           |
| ---------------- | --------- |
| blog_id          | author_id |
| int              | int       |

> TODO: This will be a very common path will be to start with an implicit
> many-to-many and eventually transition to an explicit join table. We'll need
> to make this flow as smooth as possible. I'm worried people will forget subtle
> things like the unique(blog, author)

#### Exotic Relations

##### Self-Referential Models

```groovy
// (id) could probably be implied here
model Employee {
  id         int          @primary
  reportsTo  Employee(id)
}
```

##### Multiple References

Models can have multiple references to the same model. Based on
[this example](https://github.com/prisma/prisma/blob/50ba03f7248b59cb1dd3b1911b415de79b851cc4/cli/packages/prisma-db-introspection/src/__tests__/postgres/blackbox/withExistingSchema/ambiguousBackRelation.ts).

Given the following SQL:

```sql
create table users (
  id serial not null primary key
);

create table questions (
  id serial not null primary key,
  asker_id integer not null references users(id),
  answerer_id integer references users(id)
);
```

Our model would look like this:

```groovy
model User {
  id        int         @primary
  asked     Question[]
  answered  Question[]
}

model Question {
  id        int       @primary
  asker     User(id)
  answerer  User(id)
}
```

##### Referencing composite indexes

You can have relationships to composite indexes. Example in SQL:

```sql
create table documents (
  project_id string not null default '',
  revision int not null DEFAULT '1',
  primary key(project_id, revision)
);

create table blocks (
  id serial not null primary key,
  document_id integer not null references documents(project_id, revision)
);
```

Our model would look like this:

```groovy
model Document {
  projectID  string  @default('')
  revision   int     @default(1)
  blocks     Block[]
}
@primary(projectID, revision)

model Block {
  id        int                            @primary
  document  Document(projectID, revision)
}
```

### Embedded Types

Embeds are supported natively by Prisma. There are 2 types of embeds: named
embeds (just called embeds) and inline embeds.

Unlike relations, embeds "come with the model". How the data is actually stored
(co-located or not) is not a concern of the data model. The data sources are
responsible for fetching and stitching the data together.

```groovy
model User {
  id        string
  customer  StripeCustomer?
}

embed StripeCustomer {
  id     string
  cards  Source[]
}

enum Card {
  Visa,
  Mastercard
}

embed Sources {
  type Card
}
```

#### Inline Embeds

When you know you don't need to reuse an embed, inline embeds are handy. Inlines
can be nested as deep as you want. Please don't go too deep though.

```groovy
model User {
  id        string
  customer  {
    id     string
    cards  {
      type Card
    }[]
  }?
}

enum Card {
  Visa,
  Mastercard
}
```

##### Note on formatting

Inline embeds are treated as their own block and will break up the formatting
above and below

```groovy
model User {
  id    string
  name  string
  customer  {
    id         string
    full_name  string
    cards  {
      type  Card
    }[]
  }?
  age   int
  email string
}

enum Card {
  Visa,
  Mastercard
}
```

### Optional fields

All primitive `types`, `enums`, `relations`, `lists` and `embeds` natively
support optional fields. By default, fields are required, but if you want to
make them optional, you add a `?` at the end.

```groovy
model User {
  names    string[]?
  ages     int?
  heights  float?
  card     Card?
}

enum Card {
  Visa,
  Mastercard
}
```

The default output for a nullable field is null.

### Lists

All primitive `types`, `enums`, `relations` and `embeds` natively support lists.
Lists are denoted with `[]` at the end of the type.

```groovy
model User {
  names    string[]
  ages     int[]
  heights  float[]
}
```

Lists can also be optional and will give the list a 3rd state, null:

```
Blog[]: empty list or non-empty list of blogs (default: [])
Blog[]?: empty list, non-empty list of blogs or null (default: null)
```

The default output for a required list is an empty list. The default output for
a nullable list is null.

### Custom Type Definitions

There are two distinct uses for custom primitive types. Users of Prisma can
create a custom type to encapsulate constraints or other configuration in a
reusable type. Implementers of connectors can declare primitive types that work
only in the context of that connector.

#### User-defined

If you have a certain field configuration that is used in multiple places, it
can be convenient to create a custom type instead of repeating the
configuration. This also ensures that all uses are in sync.

```groovy
type Email string @check.regexp(".*.com")

// Without custom type
model User {
  email  string  @check.regexp(".*.com")
}

// With custom type
model User {
  email  Email
}

// With additional field config
model User {
  email  Email  @as(postgres.varchar(250))
}
```

> A user-defined primitive type is a collection of field configurations that can
> be extended at place of use.

#### Connector-defined

When implementing a connector it might be necessary to augment Prisma with types
that are not already part of the built-in primitive types. For example, a legacy
database might have a special string type that support emoji, but does not
support indexing. The connector could introduce a new primitive type to expose
this type:

```groovy
type EmojiString string
```

Prisma users can then use it in their datamodel like this:

```groovy
model User {
  displayName EmojiString
}
```

Connector implementors can rely on Prisma for certain validations if they don't
need custom error messages. Any extra field configuration will work exactly the
same as if the Prisma user added it in their datamodel. In this example, if a
EmojiString is longer than 1000 characters, Prisma will reject it with a
standard error message without calling the connector:

```groovy
type EmojiString string @check.lengthLessThan(1000)
```

> TODO: try and map out check constraints, it's really weird in this case.

> TODO: this feels like it could be combined with type specification.

#### Connector-defined complex types

A conenctor might also want to introduce a complex type. For example a custom
connector for a legacy SOAP API could introduce a complex type that is
transparently mapped to a bitmap:

```groovy
embed UserSettingBitmap {
	sendEmail   bool
	showVideos  bool
}
```

> TODO: This needs to be mapped out in much greater detail

## Attributes

There are 2 types of attributes: model attributes & field attributes.

### Model Attributes

Model attributes apply to the whole model and appear at the end of the model.

```groovy
model User {
  to = "users"

  id          int     @primary
  first_name  string
  last_name   string
  salary      int
  bonus       int
}
@to("users")
@check(salary > bonus)
@postgres.unique(first_name, last_name)
```

> TODO: I don't have a good mental model of the difference between block
> configuration & block attributes. They feel like the same thing to me. Should
> we consolidate?

Prisma core supports the following model attributes:

- `@check`: Check creates a constraint across the model

Some examples from postgres (not feature complete):

- `@postgres.unique(fields...)`: Unique composites
- `@postgres.primary(fields...)`: Primary key composites
- `@postgres.gin(path)`: Support "data source"-specific indexes

### Field Attributes

Field attributes apply to a given field. Prisma core supports the following
field attributes:

- `@primary`: Defines the primary key
- `@as`: Defines a type specification
- `@alias`: Defines a type alias to be picked up by the generators
- `@check`: Check creates a single field constraint
- `@default`: Specifies a default value if null is provided

### Type specifications

Prisma's primitive types are implemented by all connectors. As such, they are
often too coarce to express the full power of a connectors type system. It is
possible to specify the exact type of a storage engine using type specification.

A type specification is always scoped to a specific connector. If the datamodel
is used with any other connector, it is ignored. It is possible to provide type
specification for multiple different connectors in a single datamodel.

```groovy
source pg {
  type = "postgres"
  url  = "postgres://localhost:5432/db"
}

model User {
  id           string   @as(pg.char(100))
  age          int      @as(pg.smallInt)
  name         string   @as(pg.varchar(128))
  height       float    @as(pg.float4)
  cashBalance  decimal  @as(pg.numeric(30, 60))
  props        json     @as(pg.mediumText)
}
```

Type specifications for multiple data sources:

```groovy
source pg {
  type: "postgres"
   url: "postgres://localhost:5432/db"
}

source ms {
  type = "mysql" // could be inferred from the protocol
   url = "mysql://localhost:3306/db"
}

model User {
  age int @as(pg.smallint) @as(ms.smallint)
}
```

> TODO: where does User persist in this case?

### Literals

Datamodel 2 supports the same literal values as HCL2.

| Primitive type | Literal generator |
| -------------- | ----------------- |
| int            | 1                 |
| float          | 1.1               |
| Decimal        | 1.1               |
| String         | "some text"       |
| bool           | true              |
| datetime       | "2018"            |
| enum           | SomeEnum          |
| json           | '{"a":3}'         |

> TODO:
>
> - spec out literals for remaining primitive types such as binary and spatial
> - Should we use colons instead of equals for better consistency with the JSON
>   primitive type?

### Functions

Datamodel 2 also supports functions as inputs.

#### Core

Prisma core provides a set of function expressions.

- `uuid()` - generates a fresh UUID
- `cuid()` - generates a fresh cuid
- `random(min, max)` - generates a random int in the specified range
- `now()` - current time

Default values using a dynamic generator can be specified as follows:

```groovy
model User {
  age        int       @default(random(1,5))
  height     float     @default(random(1,5))
  createdAt  datetime  @default(now())
}
```

#### Connectors

Connectors can provide additional functions that rely on specific database
primitives to work. For example, MySQL and Postgres connectors could provide the
`autoIncrement` generator that uses the underlying `AUTO_INCREMENT` and `SERIAL`
column modifiers respectively:

```groovy
model User {
  id  int  @as(pg.serial(100, 10))
           @as(ms.autoIncrement(100, 10))
           @as(maria.sequence(100, 10))
}
```

## Indexes

### Unique index on a field

This is the most common index. Adding a `@unique` attribute on a field will
create an index on that field

```groovy
model Employee {
  id      int     @primary
  name    string  @unique
  height  int
  height  float
}
```

## Composite indexes

Use model attributes to represent indexes across fields:

```groovy
model Employee {
  id          int     @primary
  first_name  string
  email       string

}
@unique(first_name, email)
```

## Indexes for expressions

You can also create indexes for common expressions:

**Field Indexes**

```groovy
model Employee {
  id          int     @primary
  first_name  string
  last_name   string
  email       string
}
@postgres.index(lower(first_name))
@postgres.index(first_name + " " + last_name)
```

## Machine Formatting

Following the lead of [gofmt](https://golang.org/cmd/gofmt/) and
[prettier](https://github.com/prettier/prettier). Our syntax ships with one way
to format `.prisma` files.

Like gofmt and unlike prettier, we offer no options for configurability here.
There is one way to format a prisma file. The goal of this is to end pointless
bikeshedding. There's a saying in the Go community that, "Gofmt's style is
nobody's favorite, but gofmt is everybody's favorite."

Here's an example of gofmt in action when I press save in VSCode:

![gofmt](https://cldup.com/ooQHBLtQtL.gif)

For our syntax, it would be nice to arrange the document into 3 columns:

```groovy
model User {
  id:             int     @primary @postgres.@serial
  name:           string
  profile: {
    avatarUrl:    string?
    displayName:  string?
  }
  posts:          Post[]
}

model Post {
  id:             int     @primary @postgres.@serial
  title:          string
  body:           string
}
```

It can produce some really bad pathological cases where a single long line
introduces extraneous whitespace on thousands of lines in a large file.

### Checks (Constraints)

> DRAFT

Checks restrict updating data when the update would violate a certain condition.
Checks can be defined with multiple scopes:

|                        | Single Field | Multi Field |
| ---------------------- | ------------ | ----------- |
| Only Updated Values    | x            | x           |
| All Values in Row      | x            | x           |
| All Values in Table    | x            | x           |
| All Values in Database | x            | x           |

#### Single field checks

The following is an excellent reading on
[single field checks](https://github.com/prisma/prisma/issues/728). Countless
extensions, or even using a simple expression language is thinkable.

Example:

```groovy
model Employee {
  salary     int
  bonus      int
  firstName  string
  lastName   string
  email      string @check.regexp("(^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$)")
}
```

#### Multi-field checks

For declaring multi-field checks, a similar structure as with indexes (on type
level) could be used.

Example:

```groovy
model Employee {
    salary      int
    bonus       int
    firstName   string
    lastName    string
  	email       string
}
@check(salary < bonus) // Custom check
```

#### checks that respect other values

**Caution**: checks that respect other values in tables or the database are
generally not supported natively by all databases.

Example:

```groovy
model Employee {
    salary      int     @check(salary < AVG(salary) * 1.5) // Cap salary
    bonus       int
    firstName   string
    lastName    string
  	email       string
}
```

> Task:
>
> Map out what it would look like to have reusable named constraints. Maybe
> custom scalars?

### Template Expressions

> DRAFT

Prisma could support expressions in generators:

```groovy
model User {
  name             string
  age              int
  someRandomField  string  @default(fmt("%s is %s years old", this.name, this.age))
  ageInDays        int     @default(this.age * 365)
}
```

> Based on: https://www.postgresql.org/docs/9.1/indexes-expressional.html

### Cascading Deletes

> DRAFT

> TODO: are cascading deletes a connector concern or something in core? For
> example, does Mongo support these?

With Cascading Deletes you can ensure that related data is automatically cleaned
up when a record is deleted. In general, when there is a parent-child
relationship, Cascading Deletes can be used to automatically delete the child.

There are 3 options:

- `CASCADE`: Delete all child records
- `RESTRICT`: Prevent deleting a parent record when there are child records
- `SET_NULL` (default): Break the relation, but leave child records alone

**1-1**

Cascading can be enabled on either side, but not both:

```groovy
model Blog {
  id     int @primary
  author Writer @onDelete(CASCADE)
}

model Writer {
  id   int @primary
  blog Blog?
}
```

```groovy
model Blog {
  id     int @primary
  author Writer
}

model Writer {
  id   int @primary
  blog Blog? @onDelete(CASCADE)
}
```

This would return an error:

```groovy
model Blog {
  id int @primary
  author Writer @onDelete: CASCADE
}

model Writer {
  id   int @primary
  blog Blog? @onDelete: CASCADE
}
```

**1-m**

Cascading can be enabled on the parent side only:

```groovy
model Blog {
  id     int @primary
  author Writer
}

model Writer {
  id   int @primary
  blog Blog? @onDelete(CASCADE)
}
```

This would return an error:

```groovy
model Blog {
  id     int @primary
  author Writer @onDelete(CASCADE)
}

model Writer {
  id   int @primary
  blog Blog?
}
```

**m-n**

In a `m-n` relation there is no natural parent-child relationship. Therefore,
cascading deletes are not supported.

> TODO: I could see a potential use case for this, where if you delete the join
> table, the entire relationship goes away.

### Edge Relation

> DRAFT

> Note: The Edge Relation part of the spec is additive and will not be part of
> the initial release of Prisma 2 Note: This is only a draft.

The edge relation concept comes from property graphs, implemented in systems
such as Neo4j and ArangoDB. An edge connects two nodes in the graph and can have
extra properties. The Edge Relation extends the Prisma API to handle this
concept without the use of a full model to represent the edge.

**Implementation in wire protocol**

```graphql
// Inserting relation data

mutation createWriter(data: {
  id: "a"
  blogs: { create: { id: "b" } relationData: { becameWriterOn: "${now()}" }}
})

// Reading relation data

writers {
  blogsConnection {
    node { id }
    relationData { becameWriterOn }
  }
}
```

## Triggers

> DRAFT

In Postgres you can have many kinds of triggers. For example:

```sql
create or replace function set_updated_at() returns trigger as $$
  begin
    new.updated_at := current_timestamp;
    return new;
  end;
$$ language plpgsql;

create table if not exists teams (
  id serial primary key,
  name text not null
);

create table if not exists teammates (
    id serial primary key,
    team_id text not null references teams(id) on delete cascade on update cascade,
    slack_id text not null,
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now()
);
create trigger updated_at before update on teammates for each row execute procedure set_updated_at();
```

In the above example we have a 3 triggers:

1. `update trigger` that executes a procedure to update `update_at` whenever the
   row changes
2. `on delete` trigger deletes the `teammate` row when the referenced `team` is
   deleted
3. `on update` trigger updates the `teammates.team_id` whenever the `teams.id`
   is updated

Assuming that we cannot place custom procedures into our datamodel, we can
bucket the above cases into field triggers & model triggers:

1. `updated_at` model trigger
2. `on delete` model trigger
3. `on update` field trigger

Given that, I propose:

```groovy
model Teammate {
  id          int       @primary
  team_id     string    @onUpdate(cascade())
  slack_id    string
  created_at  datetime
  updated_at  datetime

}
@onUpdate(autoupdate(updated_at))
@onDelete(team_id, cascade())
```

We could also say that the `updated_at` model trigger is a special procedure
that operates on fields and then we could do:

```groovy
model Teammate {
  id          int       @primary
  team_id     string    @onUpdate(cascade())
  slack_id    string
  created_at  datetime
  updated_at  datetime  @onUpdate(autoupdate())

}
@onDelete(team_id, cascade())
```

### Aggregations

> DRAFT

> NOTE: this section is an aside examining the applicability of the above API
> design to aggregations We will use the same datamodel

```ts
# Reading aggregate data

// aggregate record data
const writersWithBlogsRelationData: DynamicType[] = await prisma.writers
  .findAll()
  .blogsWithRelationData({select: {data: {$aggregate: { id: { avg: true } } }}})

// aggregate relation data
const writersWithBlogsRelationData: DynamicType[] = await prisma.writers
  .findAll()
  .blogsWithRelationData({select: {relation: {$aggregate: { becameWriterOn: { avg: true } } }}})

// or the same result relying on top level select
const writersWithBlogsRelationData: DynamicType[] = await prisma.writers
  .findAll({select: {id: true, blogs: {id: true, $aggregate: { id: { avg: true } }}}}) // note that we don't need the {data, relation} intermediate type

const writersWithBlogsRelationData: DynamicType[] = await prisma.writers
  .findAll({select: {id: true, blogs: {data: {id: true}, relation: { $aggregate: { becameWriterOn: { avg: true } }}}}}) // The {data, relation} intermediate type is the only way to access relation


// Filtering by aggregation data
const writers: Writer[] = await prisma.writers
  .findAll({ where: { blogs_all: { id_ne: "abba" }, blogs_aggregate: { id_avg_gt: "2018"} } })

const writers: Writer[] = await prisma.writers
  .findAll({ where: { blogs_all: { id_ne: "abba" }, blogs_relation_aggregate: { becameWriterOn_avg_gt: "2018"} } })
```

### Special Search Index

> DRAFT

Full text/phrase/spatial search in its simplest form can be reduced to providing
an index. For this, an optional `model` argument could be added to indices, to
represent this indices.

```groovy
model Post {
  id         int       @primary
  title      string
  published  datetime
  text       string
  author     user
}
@postgres.fuzzy(text, title)
```

An alterative would be a distinct `textIndex` directive which allows additional
tuning params.

```groovy
type Post {
  id: ID @id
  title: String
  published: DateTime
  text: String
  author: User
}
@postgres.fuzzy.weighted({
  text: 0.4
  title: 0.6
})
```

These indices would, when created, add extra filter fields to the schema.

### Proposed Special Indices

> DRAFT

| Description                                     | DirectiveName proposals                     | Filter fields                                           |
| ----------------------------------------------- | ------------------------------------------- | ------------------------------------------------------- |
| Spatial Geometry Contains                       | `@spatialIndex`, `@geoContainsIndex`        | `field_contains:Gemoetry`, `field_intersects: Geometry` |
| Full Text Trigram index, for contains queries   | `@fullTextIndex`, `@fullTextContainsIndex`  | `field_contains: String`                                |
| Full Text index with stemming for phrase search | `@fuzzyFullTextIndex`, `@phraseSearchIndex` | `field_matches: String`                                 |

For the fuzzy text index, it would be useful to also expose an order by field
which would allow to order by rank of the match.

> Task:
>
> Find further special inidices Map out all index tuning settings and create
> common capability groupings between DBs

## Inheritance

> DRAFT

Inheritance in this context describes the concept of sharing common fields
between types that are conceptually related. While inheritance is well discussed
and researched on a language level, we have to tie these concepts closely to
database models and to introspection.

> Polymorphic relations are a powerful concept and should be discussed here
> seperately.

This concept is not to be confused with polymorphic relations, which is
described in the
[datamodel v2 specification](https://github.com/prisma/prisma/issues/3407). The
polymorphic relation discussion is recommended reading for this topic as well.

Also, inheritance has to be distinguished from **interfaces**. The concept is
similar, but interfaces are not backed by the databases, and any model can
implement multiple interfaces.

> Tasks:
>
> What about union types?
>
> What are the precise implications on data layout when inheritance is used?
>
> Do we include discriminators or do we rely on native mechanisms, exposed by
> the client (like instanceof)?

### Inheritance in Prisma

> DRAFT

In the concept of prisma **inheritance** allows extending a type that is backed
by the database by **inheriting** from it.

Conventional **abstract** types behave like conventional types, but cannot be
created. We have to take care of existing data in the database correctly.

Inheritance in prisma respects **all properties** of base fields, including:

- Default Values
- Indices
- Types
- Field Names
- Constraints

Inheritance in the datamodel is declared by an `extends` clause:

```groovy
model LivingBeing {
  id int @primary
  dateOfBirth datetime @default(now())
}

model Human {
  LivingBeing(id)
  firstName string
  lastName string
}

model Pet {
  LivingBeing(id)
  nickname string
  owner Human
}

model Cat {
  Pet(LivingBeing(id))
  likesFish bool
}

model Dog {
  Pet(LivingBeing(id))
  likesFrisbee bool
}
```

In the example above, `Dog` would inherit all fields from `Pet` and
`LivingBeing` without explicitly declaring them.

When a prisma query for base type happens, all super types are taken into
consideration. In other words, when quering all Pets, all cats and dogs are
returned as well.

### Inheritance in Relational DBs

> DRAFT

Via **single table inheritance**: We simply have all base field and all fields
from superclasses in the same table.

Drawbacks: Impossible to enforce not null, field names collide. A `type` collumn
will be mandatory.

Via **concrete table inheritance**: We have a seperate table for each subtype,
copying base fields.

Drawbacks: No clear distinction between base and sub fields. It is hard to query
all for the base type. Auto incrementing PKs on the base type are hard to
achieve.

Via **class table inheritance**/**join table inhertiance**: We create a base
table for the base class and specific tables for subtypes, which are joined.

Drawback: Performance (Feedback from Marcus)

### Implementation in Prisma

> DRAFT

> This point should be discussion. Marcus mentioned that join table forms can
> lead to poor performance . Maybe single table is better for prisma - prisma
> could hide the not null shortcoming in the application layer.

Prisma always uses the **join table form** for relational DBs, as it poses the
least drawbacks. Optionally, prisma could offer support for the other
inheritance concepts to allow easier adoption of existing databases.

When introspecting, inheritance is never discovered, as there are no hints we
could salvage for detecting inheritance. However, a user can always declare an
inheritance in an existing datamodel to match the database.

> Marcus pointed out that it might make sense to limit or at least discourage to
> deep inheritance, since it can lead to poor performance.

> Task:
>
> Map out the MDL syntax for supporting all inheritance types

### Inheritance in Document DBs

> DRAFT

On Top Level, Document Databases can theoretically leverage the same approaches
as Relational Databases. For embedded types, only **single table inheritance**
is really feasible. In the context of document DBs, this means mixing all base
and sub types inside the same collection or array. This will require a type tag
on each object to function properly.

### Implementation in Prisma

> DRAFT

Prisma always stores super and sub types in the same collection, with a type
tag.

Introspection does not identify inheritance (in theory, it could with
heuristics), but allows a user to declare an existing inheritance relationship
in the datamodel. For that, a type tag needs to be added, which can be done
using provided tooling.

### Migration Considerations

> DRAFT

For any form of inheritance, migrating away from a super/subtype relationship
will move (and potentially duplicate!) a lot of data.

Migrating towards a class/subclass relationship is can be a difficult task if
it's allowed to create a base types for two types simultaneously because of
conflicts. Splitting a single type into super/subtype is less of a problem.

### Client considerations

> DRAFT

The prisma client needs to expose a way to distinguish between different
subclasses for superclass queries. This can be done in a language-native way or
with type tags.

> How does that work with querying?

### Impact on filters

> DRAFT

Filters on supertypes also include an is operator to check for a specific
subtype. This is needed for relations that point to a supertype.

When querying a specific subtype on top level, the appropriate sub type should
be queried directly.

## Interfaces

> DRAFT

Interfaces operate similar to inheritance, although interfaces ONLY transfer

- The field name
- The field type
- Field constraints

To a base type.

Other properties (indices, Attributes, default) cannot be declared on
interfaces.

Interfaces are not backed by the database and they do not change the API schema
per se. However, they are exposed in the generated client's type system and are
also included in the client API. A type can inherit multiple interfaces, as long
as single fields are not conflicting. The type still has to explicitly declare
all interface fields.

In other words, Interfaces offer a guarantee that a subset of a certain type
follows a certain schema.

Interfaces can be declared using the `Interface` keyword and used using the
`implements` clause:

```groovy
interface IDatabase {
  storageSize    int
}

interface IMessageQueue {
    capacity    int
}

type Kafka {
  IMessageQueue

  capacity    int
  serverName  string
}

type PostGres {
  IDatabase
  storageSize  int
  serverName   string
}

type Prisma {
  IDatabase
  IMessageQueue

  storageSize  int
  capacity     int
}
```

# Resolved Questions

<details>
<summary>`string` or `text`?</summary>
<br>

### Answer

Consensus suggests that we stick with "developer terms". `string` it is!

---

🙃

- String is more familiar to programmers.
- Text is more familiar to English speakers
- Text is shorter.

</details>

<details>
<summary>Should we enforce link tables?</summary>
<br>

### Answer

Having a default behavior for implicit link tables for m:n relations is a good
idea as it provides two benefits:

1. Simpler to get started
2. Portability between different databases that implement m:n relations in
   different ways.

---

- Link tables are not usually needed right away, but are often good practice
  since you often want to attach metadata to that relation later on (e.g.
  `can_edit bool`). Some options:

1. We could make them optional at first, but create a table in the background
   (we'd need to do this anyway), but then when they specify the table and
   migrate, we'll be aware that this implicit join table became explicit in the
   datamodel

2. Enforce the link table at build-time when we run `prisma generate`. A bit
   simpler to implement and less magic, at the expense of cluttering up your
   datamodel file and forcing you to think more about your data layout earlier
   on (might not be a bad thing).

</details>

<details>
<summary>Can back-relations have a different nullability than the forward relations?</summary>
<br>

### Answer

Yes this is possible. I don't think this will change our syntax it all, it was
more for my knowledge 😬

---

e.g. Is this possible?

```groovy
model User {
  id        int           @primary
  customer  Customer(id)?
  name      string
}

model Customer {
  id       int   @primary
  user     User
  address  string
}
```

Or is it always:

```groovy
model User {
  id        int           @primary
  customer  Customer(id)?
  name      string
}

model Customer {
  id       int    @primary
  user     User?
  address  string
}
```

</details>

<details>
<summary>Should we support Mongo's many-to-many relationship in relational databases?</summary>
<br>

### Answer

Keep in mind, but punt on this use case for now.

---

This is what facebook needed to do to scale MySQL. It looks like the
[original video from 2011](https://www.facebook.com/Engineering?sk=app_260691170608423)
was taken down, but the gist is that they successfully scaled MySQL by add the
foreign keys as an array in your table aand writing "integrity checkers" to
ensure that if a row gets deleted, the foreign keys that link to it will
eventually get deleted. In short, Facebook wrote a NoSQL implementation on a
MySQL database.

How this architecture plays out in your application is that instead of looking
up posts by a user's ID, you can lookup each post by ID within the user table.
This can tie into their memcache layer too so it's easily parallelizable.

In researching this question a bit better, it seems like they got the benefits
of NoSQL, without migrating all their data. Additionally, I came across Google
Spanner (considered [NewSQL](https://en.wikipedia.org/wiki/NewSQL) 😑) which
does seem to give you the benefits of this approach with better guarentees so
you don't need to write your own "integrity checkers".

The question is can we enforce relations at the application layer without
enforcement from the database layer? Furthermore, is there value in doing this
when something that's not using Prisma client can muck up the data?

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

How do we indicate that these are logical relationships without actually
enforcing those guarentees.

</details>

<details>
<summary>Can we rename datamodel to schema?</summary>
<br>

### Answer

`datamodel.prisma` is fine. Stop complaining @matthewmueller.

**Update:** Actually, @matthewmueller doesn't have to complain anymore because
if we do terraform's autodiscover `*.prisma` concatenation approach, I can name
the `.prisma` files whatever I want!

---

I learned that the historical context for datamodel was that we needed something
different than schema.graphql.

I think we should revisit this because `schema.prisma` is much shorter and
sounds nicer.

> Feedback: I still like datamodel. I think we will understand the role of the
> datamodel much better in a years time, and suggest we delay any renaming
> discussion til then.

> Feedback: Please no. We just decided to switch from the term schema to
> datamodel in the backend 😂

</details>

<details>
<summary>Lowercase primitives, uppercase indentifiers?</summary>
<br>

### Answer

Primitives are special, so we'll make them lowercase to separate them from
higher-level types. For model and embed blocks, we all seem to agree that they
should be PascalCase.

---

If primitives are considered special and cannot be overridden then I think we
should have special syntax for them. If they are simply types that are booted up
at start (GraphQL), they should be treated like every other type.

</details>

<details>
<summary>Consider the low-level field and model names?</summary>
<br>

### Answer

Prisma doesn't want to expose all the low-level details of the underlying
columns, but if you introspect aliases will map directly to the existing column
names since there's no other reasonable default.

As far as aliases changing across languages, variations in case is not a big
deal.

---

I'm starting to think that it might be best to use the low-level field names as
the default for datamodel 2.

```groovy
model users {
  id          int
  first_name  string
}

model posts {
  user_id  User
}
```

There are two reasons for this:

1. Less mental mapping between alias and actual column name.
2. Aliases won't be stable across languages, for example:

- In Javascript: `first_name => firstName`
- In Go: `first_name => FirstName`
- In Ruby/Python: `first_name => first_name`

</details>

<details>
<summary>Do we want back-relations to be optional?</summary>
<br>

### Answer

Historically Prisma did 3., but migrated to 1. for a better experience. We'll
stick with 1. for now and plan to offer IDE features like green/blue/yellow edit
squiggies to suggest changes.

---

- My typical stance is to enforce good practices (e.g. prettier), and provide
  one way to do it. We have some options with back-relations though:

  1. Implied back-relation when not provided (affects client API)
  2. No back-relation when not provided (affects client API)
  3. Build-time (`prisma generate`) error when no back-relation provided

</details>

<details>
<summary>Should `id`, `created_at`, `updated_at` be special types that the database adds?</summary>
<br>

### Answer

For cases like these we'd like to offer "type specifications" (better name?
"type upgrade"?) as attributes. This way we can keep the fields types low-level
and universal, but "upgrade" the type for databases that support it.

```groovy
model User {
  id  string  @as(postgres.UUID) @as(mongo.ObjectID)
}
```

---

The DM2 proposal proposes:

```groovy
model User {
  id int (id)
}
```

The reason I avoided this is because it's not always clear (at least in SQL
databases) whether you want your ID to be an `int`, a `text`, or a
`postgres.UUID`. I could definitely see value in higher-level types where Prisma
chooses the datatype based on the most common cases. We'd just need to have a
way to override them.

Same could go for `created_at` or `updated_at`. Where they bring a type and some
default functionality.

```groovy
model User {
  id         id
  createdAt  created_at
  updatedAt  updated_at
}
```

One possible solution is to introduce a prisma namespace for high-level data
types:

```groovy
model User {
  id         prisma.ID
  createdAt  prisma.CreatedAt
  updatedAt  prisma.UpdatedAt
}
```

</details>

<details>
<summary>Can we make syntax more familiar?</summary>
<br>

### Answer

Drop the `:`, add the `@` attribute back in for both field attributes and model
attributes.

The new syntax will look like this:

```groovy
model Post {
  id     int     @primary
  title  string
  author Author
}
// unique composite
@unique(title, author)
```

The reason to bring the model attributes below is that it adds consistency to
the attribute syntax and solves multi-line issues.

```
block-type block-name {
  field-name field-type  field-attributes
}
block-attributes
```

---

Right now the syntax is a mismash of SQL, Terraform and Go. I think there are
steps we can take to make it more familiar to GraphQL/Typescript users.

> Feedback: How to reduce the learning curve / make the datamodel syntax feel
> more familiar

### Ideas:

#### Use field: Type instead of field Type

This isn't a dealbreaker for me, but I find it to be an embellishment. From a
parsing perspective, this syntax doesn't need to be here.

The reason I originally removed it was to make use of the colon for aliasing,
but I didn't end up using it. We may want to have this piece of syntax in the
future.

Also, it's a relatively new concept (C/C++ don't include it). I'd like to learn
where it came from. My guess is that there was an ambiguity in going from
Javascript to Typescript's syntax and they needed to add it and it has spread
since then.

#### Use @attribute instead of attribute()

I like this idea because it would simplify `unique()` to `@unique`. I wish it
wasn't an "at sign", because "at unique" doesn't make sense. Maybe we could just
say it's the "attribute sign". 😬

</details>

<details>
<summary>Can we improve comma / multi-line support?</summary>
<br>

### Answer

The new syntax resolves multi-line issues without needing commas.

```groovy
model Post {
  id     int     @primary
  title  string
  author Author
}
// unique composite
@unique(title, author)
```

---

> Feedback: Comma/multi-line behavior seems error-prone

> Feedback: I appreciate that it gives us other freedoms, but it will be
> unexpected behavior to most developers. I we decide to use a special construct
> to support multi-line, we should try to find something more obvious or common.
>
> There are other ways to introduce multi-line support without requiring a
> special construct like this, but they introduce restrictions in the grammar.

This is a tricky one because we want to be able to support model attributes and
multi-line field attributes. I chose the comma because I thought it the lesser
of evils.

To give an example of where this is tricky:

```groovy
model Post {
  id  string  @attr1
              @attr2
  @attr3
  name string
}
```

Is `@attr3` an attribute on the `id` field or an attribute of the model? To
disambiguate in this case I found this syntax to be acceptable:

```groovy
model Post {
  id  string  @attr1,
              @attr2
  @attr3
  name string
}
```

We could also use different syntax for model attributes:

```groovy
model Post {
  id  string  @attr1
              @attr2
  attr3()
  name string
}
```

Or perhaps group the field attributes:

```groovy
model Post {
  id  string  [
                @attr1
                @attr2
              ]
  @attr3
  name string
}
```

I quite like the simplicity of field attributes and model attributes sharing the
same syntax, but we can also look into merging metadata and model attributes.
Based on feedback, something like this:

```groovy
model User {
  dataSource = {
    // use specific table name
    name = "tbl_user"
  }
  generator = {
    // adjust the plural form used by client generators
    plural = "people"
  }
  index = {
    unique(email, name)             alias("email_name_index")
  }
}
```

or

```groovy
model User {
  meta = {
    // adjust name from "users" convention
    name = "people"
    unique(email, name)             alias("email_name_index")
  }
}
```

</details>

<details>
<summary>Should we merge blocks with model attributes?</summary>
<br>

### Answer

Machine formatting should take care of most of the "discipline" issues and
placing the model attributes below will force them into one consistent place.

---

> Feedback: The current proposals requires discipline by the user to arrange it
> so that the fields are easy to read. Right now it would be possible to mix
> field and index declarations.
>
> Feedback: I like the idea of moving everything that is not field-specific into
> dedicated blocks. We should map out if we just want a single meta block or
> should rather have many purpose specific blocks:

I think the discipline required to arrange fields will largely be solved machine
formatting. See the "Machine Formatting" section for more details.

Right now we have two supported syntaxes inside a definition block. In a
psuedo-parsing language, it looks like this:

```
<block> <name> '{'

  <block-meta> '=' '{'
    (
      <ident> '=' <expression>
    )*
  '}'

  (
    <field-name> <data-type> <attributes>*
  )*

  (
    <block-attribute> '(' <expression>* ')'
  )*
'}'
```

Context may help: I originally designed this syntax to only support block
attributes, e.g.

```groovy
unique(name, email)
```

But metadata was pretty awkward looking:

```groovy
model Post {
  name("posts")
  generate("js")

  id     int     primary
  title  string
  author User

  unique(title, author)
}
```

So I introduced a meta block concept as syntactic sugar:

```groovy
model Post {
  meta = {
    name     = "posts"
    generate = "js"
  }
  unique(title, author)
}
```

Could be considered:

```groovy
model Post {
  meta.name("posts")
  meta.generate("js")
  unique(title, author)
}
```

But maybe it makes more sense to go the other way and bring the block attributes
into the metadata. We could probably go
[full terraform here](https://github.com/hashicorp/hcl2#information-model-and-syntax)
and make it look like this:

```groovy
model Post {
  meta {
    name     = "posts"
    generate = "js"
  }

  unique title_author_index = [title, author]

  // or
  unique title_author_index {
    fields = [title, author]
  }

  // or
  uniques {
    title_author_index = [title, author]
    first_last_index = [title, author]
  }

  // or
  indexes [
    unique(title, author)
  ]
}
```

If I'm honest, I don't think this looks as nice as the way SQL does it, but this
would also resolve the named arguments & multi-line support open questions and
may improve embedded embeds syntax so it might be worth it!

</details>

<details>
<summary>Should we have custom field type support or support primitives with attributes?</summary>
<br>

### Answer

We're going to use "type specifications"/"type upgrades" to allow primitive
types to be upgraded for databases that support custom features.

We'll want to support the UUID use-case above, so client generators will need to
be able to understand these type specifications and what database they're
generating for to build out these higher-level APIs.

---

> Feedback: email postgres.Citext unique() postgres.Like(“.%.com”)
>
> Specifying a field like this makes it harder to read the Datamodel imo. When
> reading the Datamodel i am often just interested in the shape of data
> available for me in the client. In that case i need to map postgres.Citext to
> String in my head. I would therefore like to always see the common type such a
> special type is mapped to. E.g.:
>
> email String unique() postgres.Like(“.%.com”) postgres.Citext
>
> This also makes conversion to other databases much easier because you can
> simply remove all instances of postgres.xxx in your Datamodel.

This is a really great point and something I hadn't thought about. Who is the
audience for our datamodel? Is it important for our datamodel to convey to
developers the client's exact inputs and outputs? Or does the generated client's
type-system resolve that?

I approached this as a Postgres user where you also don't really know what the
underlying type is. _The required types only becomes apparent when you generate
the client_.

```sql
create extension pgcrypto;
create extension citext;

create table users (
  id uuid primary key,
  email citext unique
);
```

One reason I like this approach is that it would give us an extensible
architecture for adding custom types. It also fits a bit better with the
proposed syntax (e.g. enums and type aliases):

```groovy
type Numeric = postgres.Numeric(5, 2)

model Customer {
  id       int     @primary
  weight   Numeric
  gateway  Gateway
}

enum Gateway {
  PAYPAL string,
  STRIPE string
}
```

One more consideration: when I was writing my own generated Go clients the
database inspection for the above SQL would generate something like this:

```go
package main

import (
  "github.com/google/uuid"
  "github.com/app/prisma"
  "github.com/app/prisma/users"
)

func main() {
  db := prisma.Load("some token")
  id := uuid.Must(uuid.NewRandom())
  user := users.New().ID(id).Email("blah@blah.com")
  // ^^^ users.New().ID(id) directly takes a google.UUID
  u, err := users.Create(db, user)
  if err != nil {
    panic(err)
  }
  u.ID.String()
  // ^^^ u.ID returns a google.UUID
}
```

Where you could work with common higher-level types (in this case `google.UUID`)
rather than `[16]byte` and the client itself would know how to serialize /
deserialize.

I also think we could solve this with attributes and that we can also support
this use case at the client layer, translating to simple types before sending to
Rust, so I'm not too worried about this decision either way. Up to you!

</details>

<details>
<summary>Should we store configuration alongside the Datamodel?</summary>
<br>

### Answer

Yes lets unify the datamodel and configuration into one language(!)

If we're a superset of HCL, we're piggybacking off of Terraform's battle-testesd
configuration use cases.

---

> Feedback: Storing the config along side with the Datamodel is not a good idea
> in my opinion. It looks somewhat neat like that at first glance but it will
> quickly turn into a nightmare if you have different configurations for 2
> environments. Imagine a config for Mongo for example. Locally you have a super
> simple one that connects simply to localhost. On production you likely have
> something that involves a lot of settings around replica sets etc. In this
> case you want to omit some keys locally and this is where those config
> languages usually break down.
>
> Using the same Datamodel with different configuration files makes that a lot
> easier imo.

> Feedback: I don't like mixing database (endpoint) configuration with the
> datamodel. That should be two entirely different things. You might, for
> example, want to use the same datamodel for three databases (dev, staging,
> production).

I have a feeling my exploration into a single configuration will ultimately
fail, but I'd like to try nonetheless.

My main motivation for a single configuration file is driven by anger: I can't
stand when some random project or startup makes my application directory look
like a disaster.

If you look at any modern Javascript project, most of it is metadata. For fun,
count how many files are unrelated to the source code:
https://github.com/yarnpkg/yarn. Couldn't this metadata be in the package.json?
That's the point of package.json. Wouldn't this make readability and code
contribution easier?

With that tangent aside 😅, I actually think it's possible to address your
concerns **and** support a single configuration for those who want it.

I look to Terraform for answers. Terraform lets you break up configuration in 2
ways:

### 1. Multiple .tf in the same directory get concatenated

```sh
/app
  /infra
    ec2.tf
    iam.tf
    ses.tf
```

The above could address people's preference to separate configuration from the
datamodel. For example:

```sh
/app
  /prisma
    datamodel.prisma
    config.prisma
```

### 2. Multiple directories for different terraform environments

```sh
/app
  /infra
    /pro
      ec2.tf
      iam.tf
      ses.tf
    /dev
      ec2.tf
      iam.tf
      ses.tf
```

The above could address people's preference to have different configuration per
environment. Where you could do something like `prisma deploy infra/pro`.

**Caveat:** that that this should be a last resort.
[Twelve-Factor App](https://12factor.net/config?S_TACT=105AGX28) encourages you
to manage environment differences inside environment variables. Terraform
supports this use case very well:

```tf
provider aws {
  version = "~> 1.60"
  region  = "ap-southeast-1"
}

variable "iam_user" {
  description = "name of the IAM user"
}

variable "env" {
  description = "name of the environment"
}

# Policy for the lambda role
# Write to cloudwatch and invoke lambdas within lambdas
data "aws_iam_policy_document" "policy" {
  statement {
    actions   = ["*"]
    resources = ["*"]
  }
}

# Create an IAM user
module "iam_user" {
  source = "git::https://github.com/matthewmueller/infra.git?ref=master//iam-user"
  name   = "${var.iam_user}-${var.env}"
  policy = "${data.aws_iam_policy_document.policy.json}"
}
```

Usage:

```sh
# for development, create prisma-dev user
terraform apply -var 'iam_user=prisma' -var 'env=dev'

# for production, create prisma-pro user
terraform apply -var 'iam_user=prisma' -var 'env=pro'
```

### In Summary

I wasn't clear enough in the original draft of this proposal.

This proposal is not about forcing users to have all their configuration in 1
file. We should support 1 file and we should support many files locally or over
a network.

The goal of this is more about sharing the same language between configuration
and the datamodel (as opposed to SDL and YAML) and figuring out a way to join
everything together into one final, consumable configuration.

</details>

<details>
<summary>Should we reduce the syntax further, by eliminating/changing _multiple statements per line_ and _multi-line statements_?</summary>
<br>

#### Answer

- We're not going to support multiple fields per line.
- We're going to go with the @ attribute symbol and shift the model attributes
  below

---

I might be in the minority of people who really like the SQL syntax 😅

I think we could build on ANSI SQL's 33-year-old syntax with more structure,
less punctuation and proper machine formatting. If we remove _multiple
statements per line_ and _multi-line statements_ or add delimiters in theses
cases (e.g. `,` or `\`). If so we could have a syntax like this:

```groovy
model User {
  meta {
    db = "people"
  }

  id          int       primary postgres.@serial start_at(100)
  first_name  string
  last_name   string
  email       string    unique
  posts       Posts[]
  accounts    Accounts
  created_at  datetime  default(now)
  updated_at  datetime  default(now)

  unique(first_name, last_name)
  before_update(updated_at, autoupdate())
}

embed Accounts {
  provider  AccountProvider
}

enum AccountProviders {
  GOOGLE    string
  TWITTER   string
  FACEBOOK  string
}

model Post {
  slug   string
  title  string

  primary(slug, created_at)
}
```

The difference being that we could add attributes that don't require empty
parens `()`. There may be an ambiguity in the column identifiers and functions
without `()`. I'd need to check that better if we push further down this road.

It's important to keep in mind that the current syntax highlighting is
misleading. It would actually look more like this:

![actual syntax](https://cldup.com/VIxlQ084dV.png)

But wayyyy better 😅

</details>

<details>
<summary>Apply a data-driven approach to finding the right syntax?</summary>
<br>

#### Answer

Done via [prisma-render](https://github.com/prisma/prisma-render) in
[database-schema-examples](https://github.com/prisma/database-schema-examples).

---

One question I keep asking myself is how will this syntax look across a wide
spectrum of databases. We could apply a data-driven approach to finding this
answer. By searching github for `language:sql`:

https://github.com/search?q=language%3Asql

Download a bunch of these. Spin up temporary databases with these schemas,
introspect them, translate them to our evolving Datamodel AST, and then generate
the Datamodel AST and compare results.

It would take a bit of time to go through and download these, but may give us
the best results and also battle-test our introspection algorithms.

</details>

<details>
<summary>Does `Model@field` make sense for relations?</summary>
<br>

> Feedback: Model@field doesn't feel right

> Feedback: I like the notation for relations and the fields they refer to:
> author User(id). How would a reference to a combined unique criteria look
> like? Something like this? author User@(email,name)?

> Feedback: Usage of @ to declare reference columns. Maybe we can just use a .,
> like when accessing the field of an object?

> Feedback: I really like this suggestion as a way to break out of default
> behavior. The default should still be to reference the (id) (or whatever
> syntax we choose for it) field (Primary Key in relational databases and \_id
> in Mongo)

Marcus brought up a really great point about combined unique criteria. I like
his suggestion for changing `@` to `(...)`

```groovy
model User {
  id        int                     @primary
  customer  Customer(id, address)?
  name      string
}

model Customer {
  id       int                      @primary
  email    string
  gateway  Gateway
  user     User?

  unique(email, gateway)
}

enum Gateway {
  PAYPAL string,
  STRIPE string
}
```

It'd also be familiar to SQL folks and would allow us to do
`Customer(id, address)`. Alternatively we could maybe use aliases:

```groovy
model User {
  id        int                    @primary
  customer  Customer(id_address)?
  name      string
}

model Customer {
  id       int                     @primary
  address  string
  user     User?

  unique(id, address)  alias("id_address")
}
```

**Update:** I've updated the above spec to reflect this change.

### Next Steps

Generally we like the `Customer(id, address)` or `Customer(id_address)`, but
@marcus and @sorens have a better idea of the edge cases so they will discuss
foreign relations more and make a decision on a final syntax here.

</details>

<details>
<summary> Should enums be capitalize or lowercase</summary>
<br>

Generally the syntax suggests that all primitives are lowercase while all
"block"s are uppercase. This breaks down with enum.

Instead of this:

```groovy
model User {
  id             int              @primary
  role           Role
}

enum Role {
  USER  string   // unless explicit, defaults to "USER"
  ADMIN string  @default("A")
}
```

We'd do this:

```groovy
model User {
  id             int              @primary
  role           role
}

enum role {
  USER  string // unless explicit, defaults to "USER"
  ADMIN string @default("A")
}
```

I didn't quite understand the implementation details of why enums would be
lowercase, but it would break that mental model of all blocks being capitalized.
We may want to change the syntax in that case to be consistent.

Maybe @marcus and @sorens want to clarify here?

### Answer

Keep as is to stay consistent.

</details>

<details>
<summary>What do you guys think of this crazy idea?</summary>
<br>

Given our new syntax format:

```
block-type block-name {
  field-name field-type  field-attributes
}
block-attributes
```

We could be even more consistent if we did

```
block-name block-type {
  field-name field-type  field-attributes
}
block-attributes
```

This would be pretty radical, but it actually reads well in English!

```groovy
Post model {
  id     string  @primary
  title  string
}

Role enum {
  PUBLISHED  string
  DRAFT      string
}
```

### Answer

Nah bro.

</details>

<details>
<summary>What do you guys think of this crazy idea?</summary>
<br>

</details>

# Open Questions

## Should we have named arguments?

I've been going back and forth on this one a lot. We have a couple options here:

1. All named arguments @alias(name: "my_weight") @sequence(start:1, step:100)
2. Optionally named: @alias("my_weight") @sequence(step:100, start:1). Naming
   allows you to specify the arguments in whatever order you want.
3. No named arguments, VSCode plugin will provide autocomplete signatures.
4. No named arguments, break up optional attributes when not obvious to be
   multiple 1 argument attributes @sequence.start(1) @sequence.step(100) or as a
   chain @sequence.start(1).step(100)

> Feedback from @marcus
>
> Option 3 is a no go for me because not everyone uses VSCode.
>
> Option 4 introduces more syntax complexity imo. I would not know when i am
> supposed to use the dot syntax versus the normal one.
>
> Option 1 is very consistent but would contradict the overall approach you took
> with the directives so far where they are most singly argument directives.
> This would necessitate the grouping of related settings. E.g. consider
> @relation(name: "MyRelation" onDelete: CASCADE) in DM 1.1 vs @relation(name:
> "MyRelation") onDelete(value: CASCADE). This would also require us to add lots
> of unnecessary argument names a la value.
>
> The direction of Option 2 seems to be the most promising to me. But we need to
> spec that out a bit further imo. How do we handle arguments in directives with
> multiple arguments? Your suggestion implies that @sequence(100,1) is possible
> which is not very readable. Therefore i suggest the following rule: Directives
> with 1 argument may omit the argument name. E.g. @alias("my_weight") and
> @alias(name:"my_weight") is allowed. Directives with multiple arguments must
> always specify all argument names.

I quite like the rule @marcus proposed and I can definitely be swayed. I think
my preference still leans slightly towards Option 3 (regardless of VSCode or
not) because common attributes like composite indexes would get messy:

```groovy
@unique(first_name, email)
// to...
@unique([ first_name, email ])
// ok, but what about named indexes?
@unique(fields: [ first_name, email ], name: "first_name_email_idx")
// perhaps...
@unique([ first_name, email ]) // and
@uniqueWithName(fields: [ first_name, email ], name: "first_name_email_idx")
```

With Option 3, we'd need to have the expectation that our attribute APIs will
[avoid Boolean Traps](https://ariya.io/2011/08/hall-of-api-shame-boolean-trap)
and other non-clear inputs.

This would mean attributes like `@sequence` would take an object:
`@sequence({ start: 100, step: 10 })` rather than take integers.

With this, I propose the same parameter-style as Typescript.

##

# Things I'm not happy with yet

As I've integrated Soren's proposal and my own, I've come across some areas
where I'm not quite happy with the grammar and could use some help.

## Definitions vs Instances

I'm probably overthinking this, but I haven't found a clear separation between
declarations and assignments. Essentially the difference between defining a
class and creating an instance of a class.

I'd consider:

```groovy
model Post {

}
```

As defining the model, while something like this:

```groovy
source pg {
  url = "..."
}
```

As more like an assignment, where `pg = { url }`. It's pretty unclear from the
syntax that this is the case. Even more confusing is Type Definitions. Should
the syntax be

**An Assignment**

```
type Numeric = int @as(postgres.Numeric(5,2))
```

**Or a Definition**

```
type Numeric int @as(postgres.Numeric(5,2))
```

Go uses the later as a Definition and the former as purely an alias to a Type
Definition. I'm using Go's interpretation in this spec.

## Block Configuration and Block Attributes are unclear

I don't have a good mental model of when we should be using one or the other.
This makes me think that we should pick one or the other but not both.

```groovy
model Post {
  db = "posts"
  slug        string
  created_at  datetime  @default(now())
}
@primary(slug, created_at)
```

Aestetically, I think prefer everything inside the block, but if we go that way,
we'd need to support functions too. If functions are too confusing, I think
outside looks better.

```groovy
model Post {
  slug        string
  created_at  datetime  @default(now())
}
@db("posts")
@primary(slug, created_at)
```

**Or...**

```groovy
model Post {
  db = "posts"
  slug        string
  created_at  datetime  @default(now())
  primary(slug, created_at)
}
```

## JSON objects as an expressions look weird with equals

If we support JSON as a primitive type, we'll also want to support a default for
JSON fields. Currently we place equal signs between key value pairs, but this
looks weird for json fields (not to mention is not json):

```groovy
source pg {
  url = env("POSTGRES_URL")
}

generate typescript {
  directory = "./photon"
  options = {
    some    = "more"
    options = "cool"
  }
}

model Post {
  meta = {
    db = "posts"
  }

  slug           string
  custom_fields  json      @default({
                             some    = "key"
                             another = "key"
                           })
}
```

I'm pretty indifferent to using `:` or `=` inside objects, but `json` as a
native type would be enough of a reason for me to switch over to `:`. I don't
see any reason why we should have 2 syntaxes.

```groovy
source pg {
  url: env("POSTGRES_URL")
}

generate typescript {
  directory: "./photon"
  options: {
       some: "more"
    options: "cool"
  }
}

model Post {
  meta: {
    db: "posts"
  }

  slug           string
  custom_fields  json      @default({
                                some: "key"
                             another: "key"
                           })
}
```

## Optional field on implicit 1-M relations are slightly inconsistent

This is very minor, but given this syntax:

```groovy
model Writer {
  id      int        @primary
  blogs   Blog[]?
}

model Blog {
  id      int        @primary
}
```

The previous spec suggests that `?` in `Blog[]?` has no affect. It gets turned
into:

```groovy
model Writer {
  id      int        @primary
  blogs   Blog[]
}

model Blog {
  id         int         @primary
  writer_id  Writer(id)?
}
```

It'd be more consistent if it did:

```groovy
model Writer {
  id      int        @primary
  blogs   Blog[]?
}

model Blog {
  id         int         @primary
  writer_id  Writer(id)?
}
```

And required implicit blogs:

```groovy
model Writer {
  id      int        @primary
  blogs   Blog[]
}

model Blog {
  id         int         @primary
  writer_id  Writer(id)?
}
```

Resulted in:

```groovy
model Writer {
  id      int        @primary
  blogs   Blog[]
}

model Blog {
  id         int         @primary
  writer_id  Writer(id)?
}
```

We probably still want an optional Writer by default.

Super minor, I'm not very opinionated on this. Just pointing it out.
