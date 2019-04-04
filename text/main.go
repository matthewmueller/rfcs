package main

import (
	"fmt"

	"github.com/matthewmueller/_hack/prisma/prisma"
)

func main() {
	var db prisma.DB

	// find many users based on nested resources
	emailAsc := prisma.UsersEmailASC
	users, err := prisma.Users.FindMany(db, &prisma.UsersFindMany{
		After:   prisma.String(""),
		Before:  prisma.String(""),
		First:   prisma.Int(1),
		Last:    prisma.Int(10),
		OrderBy: &emailAsc,
		Where: &prisma.UsersWhere{
			Email: prisma.String("alice@prisma.io"),
			PostsSome: &prisma.PostsWhere{
				TitleContains: prisma.String("my title"),
				CommentsEvery: &prisma.CommentsWhere{
					Comment: prisma.String("my comment"),
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(users)

	// find comments from posts from users
	comments, err := prisma.Users.
		FromMany(&prisma.UsersFindMany{
			Where: &prisma.UsersWhere{
				Email: prisma.String("alice@prisma.io"),
			},
		}).Posts.
		FromMany(&prisma.PostsFindMany{
			Where: &prisma.PostsWhere{
				TitleContains: prisma.String("my title"),
				CommentsEvery: &prisma.CommentsWhere{
					Comment: prisma.String("my comment"),
				},
			},
		}).Comments.
		FindMany(db, &prisma.CommentsFindMany{
			Where: &prisma.CommentsWhere{
				Comment: prisma.String("my comment"),
			},
		})
	if err != nil {
		panic(err)
	}

	fmt.Println(comments)
}
