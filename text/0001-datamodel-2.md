# Datamodel 2 (Experimental Syntax)

This RFC attempts to improve the syntax laid out in [current RFC on Data Model 2](https://github.com/prisma/rfcs/blob/datamodel/text/0000-datamodel.md#1-1), spec out a few additional features and unify some concepts.

This RFC is a potential answer to an open question posed in the previous RFC:

> "If we were a little more radical with the syntax, could we create something much better?"

- **Warning:** There is a lot missing from this RFC that is properly speced out in the previous RFC. If we like the direction of this syntax, I can start moving concepts over to that spec.

## Motivation

If we're changing the syntax to something new anyway, which others will (skeptically) have to learn, then we might as well try to make it a beautiful dialect for data modeling.

Getting this right will be extremely important. Whether you're introspecting or starting from scratch, the datamodel is the first window into the Prisma world ▲. We want to make a good impression.

My goal here is to explore the space a bit further and hopefully reach consensus on a syntax we can all get behind.

## Requirements

- Break from the existing GraphQL SDL syntax where it makes sense
- Clearly separate responsibilities into two categories: Core Prisma primitives and Connector specific primitives
- High-level relationships without ambiguities
- Easily parsable ([avoid symbol tables, ideally](https://golang.org/doc/faq#different_syntax))
- Abstraction over raw column names via field aliasing
- Can be rendered into JSON

## Nice to Have

- One configuration file for prisma (WIP)
- Strict Machine formatting (␀ bikeshedding)
- Multi-line support and optional single-line via commas

## Summary of adjustments from previous DM2 RFC

- Removed colon between name and type
- Moved model directives from the top of the model into the block
- Lowercase primitives, capitalized higher-level types
- Removed `ID` as a primitive type
- Merged prisma.yml configuration into the datamodel (WIP)
- Replaced relation metadata with `Model@field`
- Introduced `source` block for connectors
- Renamed `embedded` to `embed`
- Replaced `=` in favor of `default(...)`
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

type Numeric = postgres.Numeric(5, 2)

model User {
  meta = {
    // adjust name from "users" convention
    name = "people"
  }

  // model fields
  id             int              @primary @serial
  email          postgres.Citext  @unique @postgres.Like(".%.com")
  name           string?          @check(name > 2)
  role           Role
  profile        Profile?         @alias("my_profile")
  createdAt      datetime         @default(now())
  updatedAt      datetime         @onChange(now())

  weight         Numeric          @alias("my_weight")
  posts          Post[]

  // composite indexes
  unique(email, name)             @alias("email_name_index")
}

enum Role {
  USER   // unless explicit, defaults to "USER"
  ADMIN  @default("A")
}

model Profile {
  meta = {
    from = mongo
    name = "people_profiles"
  }

  // model fields
  id       int            @primary @serial
  author   User(id)
  bio      string

  // nullable array items and nullable photos field
  photos   Photo?[]?
}

// named embed (reusable)
embed Photo {
  id   mgo2.ObjectID
  url  string

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
  id                     int          @primary,
                                      @serial,
                                      // this is okay too...
                                      @default("some default") // default value
  // model attribute right after also fine
  unique(title, author)

  title      string
  author     User(id)
  reviewer   User(id)
  published  bool               @default(false)

  createdAt  datetime           @createdAt @default(now())
  updatedAt  datetime           @updatedAt @default(now())

  categories CategoriesPosts[]
}

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
  unique(post, category)
}
```

## Configuration

Prisma's configuration and DataModel share the same language. The configuration and datamodel can be in 1 file or spread across multiple files, environments and networks. See the "Should we store configuration alongside the Datamodel?" open question below for a potential implementation.

For syntax, we're essentially a superset of the HCL2, so the configuration of Terraform would apply here:

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
  name             string
  age              int
  someRandomField  string  @default(`${this.name} is ${this.age} years old`)
  ageInDays        int     @default(`${this.age * 365}`)
}
```

## Relations

### 1-1

#### Specifying Relation id side

```groovy
model User {
  id        int           @primary @serial
  customer  Customer@id?
  name      string
}

model Customer {
  id       int     @primary @serial
  user     User?
  address  string
}
```

The relationship can be made on either side, but the `@id` indicates where the data is stored. You can think of this as a pointer to the Customer id field.

- `@id` is required here because we need to know where to put the

### 1-M

```groovy
model Writer {
  id      int        @primary @serial
  blogs   Blog[]
}

model Blog {
  id      int        @primary @serial
  author  Writer@id
}
```

- `author Writer@id` points to the `id int` on `Writer` model, establishing the
  has-many relationship.
- We can make `@id` optional here as it can default to the primary key
- `blogs Blog[]` names the back-relation, but is entirely optional

### M-N

Blogs can have multiple writers

```groovy
model Blog {
  id       int        @primary @serial
  authors  Writer[]
}

model Writer {
  id      int      @primary @serial
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

### Self-Referential Models

```groovy
// @id could probably be implied here
model Employee {
  id         int          @primary @serial
  reportsTo  Employee@id
}
```

### Embedded Models

```groovy
model Human {
  id     int     @primary @serial
  name   string
  height int
}

model Employee {
  Human
  employer  string
  height    float
}
```

Models can be embedded inside of other models, resulting in an Employee that looks like this:

```groovy
model Employee {
  id      int     @primary @serial
  name    string
  height  float
}
```

### Multiple References

Models can have multiple references to the same model. Based on [this example](https://github.com/prisma/prisma/blob/50ba03f7248b59cb1dd3b1911b415de79b851cc4/cli/packages/prisma-db-introspection/src/__tests__/postgres/blackbox/withExistingSchema/ambiguousBackRelation.ts).

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

### Referencing composite indexes

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

  primary(projectID, revision)
}

model Block {
  id        int                            @primary
  document  Document(projectID, revision)
}
```

# Indexes

## Unique index on a field

This is the most common index. Adding a `@unique` attribute on a field will create an index on that field

```groovy
model Employee {
  id      int     @primary @serial
  name    string  @unique
  height  int
  height  float
}
```

## Composite indexes

Use model attributes to represent indexes across fields:

```groovy
model Employee {
  id          int     @primary @serial
  first_name  string
  email       string

  unique(first_name, email)
}
```

## Indexes for expressions

You can also create indexes for common expressions:

**Field Indexes**

```groovy
model Employee {
  id          int     @primary @serial
  first_name  string
  last_name   string
  email       string

  index(lower(first_name))
  index(first_name + " " + last_name)
}
```

> Based on: https://www.postgresql.org/docs/9.1/indexes-expressional.html

## Machine Formatting

Following the lead of [gofmt](https://golang.org/cmd/gofmt/) and [prettier](https://github.com/prettier/prettier). Our syntax ships with one way to format `.prisma` files.

Like gofmt and unlike prettier, we offer no options for configurability here. There is one way to format a prisma file. The goal of this is to end pointless bikeshedding. There's a saying in the Go community that, "Gofmt's style is nobody's favorite, but gofmt is everybody's favorite."

Here's an example of gofmt in action when I press save in VSCode:

![gofmt](https://cldup.com/ooQHBLtQtL.gif)

For our syntax, it would be nice to arrange the document into 3 columns:

```groovy
model User {
  id:             Int     @primary @postgres.serial()
  name:           String
  profile: {
    avatarUrl:    String?
    displayName:  String?
  }
  posts:          Post[]
}

model Post {
  id:             Int     @primary @postgres.serial()
  title:          String
  body:           String
}
```

It can produce some really bad pathological cases where a single long line introduces extraneous whitespace on thousands of lines in a large file.

## Triggers

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

1. `update trigger` that executes a procedure to update `update_at` whenever the row changes
2. `on delete` trigger deletes the `teammate` row when the referenced `team` is deleted
3. `on update` trigger updates the `teammates.team_id` whenever the `teams.id` is updated

Assuming that we cannot place custom procedures into our datamodel, we can bucket the above cases into field triggers & model triggers:

1. `updated_at` model trigger
2. `on delete` model trigger
3. `on update` field trigger

Given that, I propose:

```groovy
model Teammate {
  id          int       @primary @serial
  team_id     string    @onUpdate(cascade())
  slack_id    string
  created_at  datetime
  updated_at  datetime

  onUpdate(autoupdate(updated_at))
  onDelete(team_id, cascade())
}
```

We could also say that the `updated_at` model trigger is a special procedure that operates on fields and then we could do:

```groovy
model Teammate {
  id          int       @primary @serial
  team_id     string    @onUpdate(cascade())
  slack_id    string
  created_at  datetime
  updated_at  datetime  @onUpdate(autoupdate())

  onDelete(team_id, cascade())
}
```

# Resolved Questions

<details>
<summary>`string` or `text`?</summary>
<br>
🙃

- String is more familiar to programmers.
- Text is more familiar to English speakers
- Text is shorter.

### Answer

Consensus suggests that we stick with "developer terms". `string` it is!

</details>

<details>
<summary>Should we enforce link tables?</summary>
<br>

- Link tables are not usually needed right away, but are often good practice
  since you often want to attach metadata to that relation later on
  (e.g. `can_edit bool`). Some options:

1. We could make them optional at first, but create a table in the background (we'd need to do this anyway), but then when they specify the table and migrate, we'll be aware that this implicit join table became explicit in the datamodel

2. Enforce the link table at build-time when we run `prisma generate`. A bit simpler to implement and less magic, at the expense of cluttering up your datamodel file and forcing you to think more about your data layout earlier on (might not be a bad thing).

### Answer

Having a default behavior for implicit link tables for m:n relations is a good idea as it provides two benefits:

1. Simpler to get started
2. Portability between different databases that implement m:n relations in different ways.

</details>

<details>
<summary>Can back-relations have a different nullability than the forward relations?</summary>
<br>

e.g. Is this possible?

```groovy
model User {
  id        int           primary() serial()
  customer  Customer(id)?
  name      string
}

model Customer {
  id       int   primary() serial()
  user     User
  address  string
}
```

Or is it always:

```groovy
model User {
  id        int           primary() serial()
  customer  Customer(id)?
  name      string
}

model Customer {
  id       int    primary() serial()
  user     User?
  address  string
}
```

### Answer

Yes this is possible. I don't think this will change our syntax it all, it was more for my knowledge 😬

</details>

<details>
<summary>Should we support Mongo's many-to-many relationship in relational databases?</summary>
<br>
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

### Answer

Keep in mind, but punt on this use case for now.

</details>

<details>
<summary>Can we rename datamodel to schema?</summary>
<br>

I learned that the historical context for datamodel was that we needed something different than schema.graphql.

I think we should revisit this because `schema.prisma` is much shorter and sounds nicer.

### Answer

> Feedback: I still like datamodel. I think we will understand the role of the datamodel much better in a years time, and suggest we delay any renaming discussion til then.

> Feedback: Please no. We just decided to switch from the term schema to datamodel in the backend 😂

`datamodel.prisma` is fine. Stop complaining @matthewmueller.

</details>

## Lowercase primitives, uppercase indentifiers?

- If primitives are considered special and cannot be overridden
  then I think we should have special syntax for them. If they are
  simply types that are booted up at start (GraphQL), they should
  be treated like every other type.

### Answer

Primitives are special, so it most likely makes sense to separate their types from everything else. For model and embed blocks, we all seem to agree that they should be capitalized.

## Consider the low-level field and model names?

I'm starting to think that it might be best to use the low-level field names as the default for datamodel 2.

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

### Answer

Prisma doesn't want to expose all the low-level details of the underlying columns, but if you introspect aliases will map directly to the existing column names since there's no other reasonable default.

As far as aliases changing across languages, variations in case is not a big deal.

## Do we want back-relations to be optional?

- My typical stance is to enforce good practices (e.g. prettier),
  and provide one way to do it. We have some options with back-relations though:

  1. Implied back-relation when not provided (affects client API)
  2. No back-relation when not provided (affects client API)
  3. Build-time (`prisma generate`) error when no back-relation provided

### Answer

Historically Prisma did 3., but migrated to 1. for a better experience. We'll stick with 1. for now and plan to offer IDE features like green/blue/yellow edit squiggies to suggest changes.

## Should `id`, `created_at`, `updated_at` be special types that the database adds?

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
  id         id
  createdAt  created_at
  updatedAt  updated_at
}
```

One possible solution is to introduce a prisma namespace for high-level data types:

```groovy
model User {
  id         prisma.ID
  createdAt  prisma.CreatedAt
  updatedAt  prisma.UpdatedAt
}
```

### Answer

For cases like these we'd like to offer "type specifications" (better name? "type upgrade") as attributes. This way we can keep the fields types low-level and universal, but "upgrade" the type for databases that support it.

```groovy
model User {
  id         text  @as(postgres.UUID) @as(mongo.ObjectID)
}
```

## Can we make syntax more familiar?

Right now the syntax is a mismash of SQL, Terraform and Go. I think there are steps we can take to make it more familiar to GraphQL/Typescript users.

> Feedback: How to reduce the learning curve / make the datamodel syntax feel more familiar

### Ideas:

#### Use field: Type instead of field Type

This isn't a dealbreaker for me, but I find it to be an embellishment. From a parsing perspective, this syntax doesn't need to be here.

The reason I originally removed it was to make use of the colon for aliasing, but I didn't end up using it. We may want to have this piece of syntax in the future.

Also, it's a relatively new concept (C/C++ don't include it). I'd like to learn where it came from. My guess is that there was an ambiguity in going from Javascript to Typescript's syntax and they needed to add it and it has spread since then.

#### Use @attribute instead of attribute()

I like this idea because it would simplify `unique()` to `@unique`. I wish it wasn't an "at sign", because "at unique" doesn't make sense. Maybe we could just say it's the "attribute sign". 😬

### Answer

Drop the `:`, add the `@` attribute back in for both field attributes and model attributes.

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

The reason to bring the model attributes below is that it adds consistency to the attribute syntax and solves multi-line issues.

```
block-type block-name {
  field-name field-datatype  field-attributes
}
block-attributes
```

## Can we improve comma / multi-line support?

> Feedback: Comma/multi-line behavior seems error-prone

> Feedback: I appreciate that it gives us other freedoms, but it will be unexpected behavior to most developers. I we decide to use a special construct to support multi-line, we should try to find something more obvious or common.
>
> There are other ways to introduce multi-line support without requiring a special construct like this, but they introduce restrictions in the grammar.

This is a tricky one because we want to be able to support model attributes and multi-line field attributes. I chose the comma because I thought it the lesser of evils.

To give an example of where this is tricky:

```groovy
model Post {
  id  string  @attr1
              @attr2
  @attr3
  name string
}
```

Is `@attr3` an attribute on the `id` field or an attribute of the model? To disambiguate in this case I found this syntax to be acceptable:

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

I quite like the simplicity of field attributes and model attributes sharing the same syntax, but we can also look into merging metadata and model attributes. Based on feedback, something like this:

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

## Should we merge blocks with model attributes?

> Feedback: The current proposals requires discipline by the user to arrange it so that the fields are easy to read. Right now it would be possible to mix field and index declarations.
>
> Feedback: I like the idea of moving everything that is not field-specific into dedicated blocks. We should map out if we just want a single meta block or should rather have many purpose specific blocks:

I think the discipline required to arrange fields will largely be solved machine formatting. See the "Machine Formatting" section for more details.

Right now we have two supported syntaxes inside a definition block. In a psuedo-parsing language, it looks like this:

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

Context may help: I originally designed this syntax to only support block attributes, e.g.

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

But maybe it makes more sense to go the other way and bring the block attributes into the metadata. We could probably go [full terraform here](https://github.com/hashicorp/hcl2#information-model-and-syntax) and make it look like this:

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

If I'm honest, I don't think this looks as nice as the way SQL does it, but this would also resolve the named arguments & multi-line support open questions and may improve embedded embeds syntax so it might be worth it!

### Answer

Machine formatting should take care of most of the "discipline" issues and placing the model attributes below will force them into one consistent place.

## Should we have custom field type support or support primitives with attributes?

> Feedback: email postgres.Citext unique() postgres.Like(“.%.com”)
>
> Specifying a field like this makes it harder to read the Datamodel imo. When reading the Datamodel i am often just interested in the shape of data available for me in the client. In that case i need to map postgres.Citext to String in my head. I would therefore like to always see the common type such a special type is mapped to. E.g.:
>
> email String unique() postgres.Like(“.%.com”) postgres.Citext
>
> This also makes conversion to other databases much easier because you can simply remove all instances of postgres.xxx in your Datamodel.

This is a really great point and something I hadn't thought about. Who is the audience for our datamodel? Is it important for our datamodel to convey to developers the client's exact inputs and outputs? Or does the generated client's type-system resolve that?

I approached this as a Postgres user where you also don't really know what the underlying type is. _The required types only becomes apparent when you generate the client_.

```sql
create extension pgcrypto;
create extension citext;

create table users (
  id uuid primary key,
  email citext unique
);
```

One reason I like this approach is that it would give us an extensible architecture for adding custom types. It also fits a bit better with the proposed syntax (e.g. enums and type aliases):

```groovy
type Numeric = postgres.Numeric(5, 2)

model Customer {
  id       int     primary() serial()
  weight   Numeric
  gateway  Gateway
}

enum Gateway {
  PAYPAL,
  STRIPE
}
```

One more consideration: when I was writing my own generated Go clients the database inspection for the above SQL would generate something like this:

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

Where you could work with common higher-level types (in this case `google.UUID`) rather than `[16]byte` and the client itself would know how to serialize / deserialize.

I also think we could solve this with attributes and that we can also support this use case at the client layer, translating to simple types before sending to Rust, so I'm not too worried about this decision either way. Up to you!

### Answer

We're going to use "type specifications"/"type upgrades" to allow primitive types to be upgraded for databases that support custom features.

We'll want to support the UUID use-case above, so client generators will need to be able to understand these type specifications and what database they're generating for to build out these higher-level APIs.

## Should we store configuration alongside the Datamodel?

> Feedback: Storing the config along side with the Datamodel is not a good idea in my opinion. It looks somewhat neat like that at first glance but it will quickly turn into a nightmare if you have different configurations for 2 environments. Imagine a config for Mongo for example. Locally you have a super simple one that connects simply to localhost. On production you likely have something that involves a lot of settings around replica sets etc. In this case you want to omit some keys locally and this is where those config languages usually break down.
>
> Using the same Datamodel with different configuration files makes that a lot easier imo.

> Feedback: I don't like mixing database (endpoint) configuration with the datamodel. That should be two entirely different things. You might, for example, want to use the same datamodel for three databases (dev, staging, production).

I have a feeling my exploration into a single configuration will ultimately fail, but I'd like to try nonetheless.

My main motivation for a single configuration file is driven by anger: I can't stand when some random project or startup makes my application directory look like a disaster.

If you look at any modern Javascript project, most of it is metadata. For fun, count how many files are unrelated to the source code: https://github.com/yarnpkg/yarn. Couldn't this metadata be in the package.json? That's the point of package.json. Wouldn't this make readability and code contribution easier?

With that tangent aside 😅, I actually think it's possible to address your concerns **and** support a single configuration for those who want it.

I look to Terraform for answers. Terraform lets you break up configuration in 2 ways:

### 1. Multiple .tf in the same directory get concatenated

```sh
/app
  /infra
    ec2.tf
    iam.tf
    ses.tf
```

The above could address people's preference to separate configuration from the datamodel. For example:

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

The above could address people's preference to have different configuration per environment. Where you could do something like `prisma deploy infra/pro`.

**Caveat:** that that this should be a last resort. [Twelve-Factor App](https://12factor.net/config?S_TACT=105AGX28) encourages you to manage environment differences inside environment variables. Terraform supports this use case very well:

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

This proposal is not about forcing users to have all their configuration in 1 file. We should support 1 file and we should support many files locally or over a network.

The goal of this is more about sharing the same language between configuration and the datamodel (as opposed to SDL and YAML) and figuring out a way to join everything together into one final, consumable configuration.

### Answer

Yes lets unify the datamodel and configuration into one language(!)

I f we're a superset of HCL, we're piggybacking off of Terraform's battle-testesd configuration use cases.

### Should we reduce the syntax further, by eliminating/changing _multiple statements per line_ and _multi-line statements_?

I might be in the minority of people who really like the SQL syntax 😅

I think we could build on ANSI SQL's 33-year-old syntax with more structure, less punctuation and proper machine formatting. If we remove _multiple statements per line_ and _multi-line statements_ or add delimiters in theses cases (e.g. `,` or `\`). If so we could have a syntax like this:

```groovy
model User {
  meta {
    db = "people"
  }

  id          int       primary postgres.serial() start_at(100)
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
  GOOGLE
  TWITTER
  FACEBOOK
}

model Post {
  slug   string
  title  string

  primary(slug, created_at)
}
```

The difference being that we could add attributes that don't require empty parens `()`. There may be an ambiguity in the column identifiers and functions without `()`. I'd need to check that better if we push further down this road.

It's important to keep in mind that the current syntax highlighting is misleading. It would actually look more like this:

![actual syntax](https://cldup.com/VIxlQ084dV.png)

But wayyyy better 😅

#### Answer

- We're not going to support multiple fields per line.
- We're going to go with the @ attribute symbol and shift the model attributes below

### Apply a data-driven approach to finding the right syntax?

One question I keep asking myself is how will this syntax look across a wide spectrum of databases. We could apply a data-driven approach to finding this answer. By searching github for `language:sql`:

https://github.com/search?q=language%3Asql

Download a bunch of these. Spin up temporary databases with these schemas, introspect them, translate them to our evolving Datamodel AST, and then generate the Datamodel AST and compare results.

It would take a bit of time to go through and download these, but may give us the best results and also battle-test our introspection algorithms.

#### Answer

Done via [prisma-render](https://github.com/prisma/prisma-render) in [database-schema-examples](https://github.com/prisma/database-schema-examples).

# Open Questions

## Should enums be capitalize or lowercase

Generally the syntax suggests that all primitives are lowercase while all "block"s are uppercase. This breaks down with enum.

Instead of this:

```groovy
model User {
  id             int              @primary @serial
  role           Role
}

enum Role {
  USER   // unless explicit, defaults to "USER"
  ADMIN  @default("A")
}
```

We'd do this:

```groovy
model User {
  id             int              @primary @serial
  role           role
}

enum role {
  USER   // unless explicit, defaults to "USER"
  ADMIN  @default("A")
}
```

I didn't quite understand the implementation details of why enums would be lowercase, but it would break that mental model of all blocks being capitalized. We may want to change the syntax in that case to be consistent.

Maybe @marcus and @sorens want to clarify here?

## Should we have named arguments?

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

### Next Steps

This hasn't been decided yet but has been discussed. I think this will answer itself once we start generating full syntaxes

## Does `Model@field` make sense for relations?

> Feedback: Model@field doesn't feel right

> Feedback: I like the notation for relations and the fields they refer to: author User@id. How would a reference to a combined unique criteria look like? Something like this? author User@(email,name)?

> Feedback: Usage of @ to declare reference columns. Maybe we can just use a ., like when accessing the field of an object?

> Feedback: I really like this suggestion as a way to break out of default behavior. The default should still be to reference the @id (or whatever syntax we choose for it) field (Primary Key in relational databases and \_id in Mongo)

Marcus brought up a really great point about combined unique criteria. I like his suggestion for changing `@` to `(...)`

```groovy
model User {
  id        int                     primary() serial()
  customer  Customer(id, address)?
  name      string
}

model Customer {
  id       int                      primary() serial()
  email    string
  gateway  Gateway
  user     User?

  unique(email, gateway)
}

enum Gateway {
  PAYPAL,
  STRIPE
}
```

It'd also be familiar to SQL folks and would allow us to do `Customer(id, address)`. Alternatively we could maybe use aliases:

```groovy
model User {
  id        int                    primary() serial()
  customer  Customer(id_address)?
  name      string
}

model Customer {
  id       int                     primary() serial()
  address  string
  user     User?

  unique(id, address)  alias("id_address")
}
```

**Update:** I've updated the above spec to reflect this change.

### Next Steps

Generally we like the `Customer(id, address)` or `Customer(id_address)`, but @marcus and @sorens have a better idea of the edge cases so they will discuss foreign relations more and make a decision on a final syntax here.

## Is it okay to enforce Model@id for 1:1 relations?

For example, there's not enough data to determine where the reference should be `User.customer_id` or `Customer.user_id`.

```groovy
model User {
  id        int           primary() serial()
  customer  Customer?
  name      string
}

model Customer {
  id       int     primary() serial()
  user     User?
  address  string
}
```

If we can error out in this case, I think this is totally acceptable, but I'd like to hear what you think too.

### Next Steps

This feeds into the previous question about relations and will be decided by @sorens and @marcus.

---

More questions: https://github.com/prisma/rfcs/blob/datamodel/text/0000-datamodel.md#open-questions
