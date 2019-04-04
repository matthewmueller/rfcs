package prisma

// DB interface
type DB interface {
}

type users struct{}

// Users implementation
var Users = &users{}

// String fn
func String(s string) *string { return &s }

// Strings fn
func Strings(s ...string) *[]string { return &s }

// Int fn
func Int(s int) *int { return &s }

// Ints fn
func Ints(s ...string) *[]string { return &s }

// User struct
type User struct {
	ID        string
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
}

// UsersOrderBy type
type UsersOrderBy string

// OrderBy Enums
const (
	UsersEmailASC  UsersOrderBy = "email_ASC"
	UsersEmailDESC UsersOrderBy = "email_DESC"
	// ...
)

// UsersWhere struct
type UsersWhere struct {
	ID         *string
	IDContains *string
	IDIn       *[]string
	// ...

	Email         *string
	EmailContains *string
	EmailIn       *[]string
	// ...

	FirstName         *string
	FirstNameContains *string
	FirstNameIn       *[]string
	// ...

	LastName         *string
	LastNameContains *string
	LastNameIn       *[]string
	// ...

	PostsEvery *PostsWhere
	PostsSome  *PostsWhere
	PostsNone  *PostsWhere
}

// UsersFindMany struct
type UsersFindMany struct {
	After   *string
	Before  *string
	First   *int
	Last    *int
	Skip    *int
	OrderBy *UsersOrderBy
	Where   *UsersWhere
}

func (u *users) FindMany(db DB, condition *UsersFindMany) (users []*User, err error) {
	return users, err
}

func (u *users) FromMany(condition *UsersFindMany) *UsersFromMany {
	return &UsersFromMany{condition, Posts}
}

// UsersFromMany struct
type UsersFromMany struct {
	condition *UsersFindMany
	Posts     *posts
}

type posts struct{}

// Posts implementation
var Posts = &posts{}

// Post implementation
type Post struct {
}

// PostsOrderBy type
type PostsOrderBy string

// OrderBy Enums
const (
	PostsTitleASC  PostsOrderBy = "title_ASC"
	PostsTitleDESC PostsOrderBy = "title_DESC"
	// ...
)

// PostsFindMany struct
type PostsFindMany struct {
	After   *string
	Before  *string
	First   *int
	Last    *int
	Skip    *int
	OrderBy *PostsOrderBy
	Where   *PostsWhere
}

// PostsWhere struct
type PostsWhere struct {
	ID         *string
	IDContains *string
	IDIn       *[]string
	// ...

	Title         *string
	TitleContains *string
	TitleIn       *[]string
	// ...

	Body         *string
	BodyContains *string
	BodyIn       *[]string
	// ...

	CommentsSome  *CommentsWhere
	CommentsEvery *CommentsWhere
	CommentsNone  *CommentsWhere
	// ...
}

func (p *posts) FindMany(db DB, condition *PostsFindMany) (posts []*Post, err error) {
	return posts, err
}

func (p *posts) FromMany(condition *PostsFindMany) *PostsFromMany {
	return &PostsFromMany{condition, Comments}
}

// PostsFromMany struct
type PostsFromMany struct {
	condition *PostsFindMany
	Comments  *comments
}

type comments struct{}

// Comments implementation
var Comments = &comments{}

// Comment implementation
type Comment struct {
}

// CommentsOrderBy type
type CommentsOrderBy string

// OrderBy Enums
const (
	CommentsCommentASC  CommentsOrderBy = "comment_ASC"
	CommentsCommentDESC CommentsOrderBy = "comment_DESC"
	// ...
)

// CommentsFindMany struct
type CommentsFindMany struct {
	After   *string
	Before  *string
	First   *int
	Last    *int
	Skip    *int
	OrderBy *CommentsOrderBy
	Where   *CommentsWhere
}

// CommentsWhere struct
type CommentsWhere struct {
	ID         *string
	IDContains *string
	IDIn       *[]string
	// ...

	Comment         *string
	CommentContains *string
	CommentIn       *[]string
	// ...
}

func (c *comments) FindMany(db DB, condition *CommentsFindMany) (comments []*Comment, err error) {
	return comments, err
}
