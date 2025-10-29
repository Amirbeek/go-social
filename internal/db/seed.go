package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/amirbeek/social/internal/store"
)

// Seed populates the database with test data: users, posts, and comments.
func Seed(store store.Storage, db *sql.DB) {
	ctx := context.Background()

	tx, _ := db.BeginTx(ctx, nil)

	users := generateUsers(100)
	for i := range users {
		if err := store.Users.Create(ctx, tx, &users[i]); err != nil {
			_ = tx.Rollback()
			log.Fatalf("Error creating user %s: %v", users[i].Username, err)
			return
		}
	}

	zeroUsers := 0
	for _, u := range users {
		if u.ID == 0 {
			zeroUsers++
		}
	}
	if zeroUsers > 0 {
		log.Fatalf(" %d users have zero ID — stopping seeding", zeroUsers)
	}
	log.Printf(" %d users created successfully", len(users))

	posts := generatePosts(200, users)
	for i := range posts {
		if err := store.Posts.Create(ctx, &posts[i]); err != nil {
			log.Fatalf(" Error creating post (user_id=%d, title=%q): %v",
				posts[i].UserId, posts[i].Title, err)
		}
	}

	zeroPosts := 0
	for _, p := range posts {
		if p.ID == 0 {
			zeroPosts++
		}
	}
	if zeroPosts > 0 {
		log.Fatalf(" %d posts have zero ID — stopping seeding", zeroPosts)
	}
	log.Printf(" %d posts created successfully", len(posts))

	comments := generateComments(500, users, posts)
	for i := range comments {
		if err := store.Comments.Create(ctx, comments[i]); err != nil {
			log.Fatalf(" Error creating comment (post_id=%d, user_id=%d): %v",
				comments[i].PostID, comments[i].UserID, err)
		}
	}
	log.Printf(" %d comments created successfully", len(comments))

	log.Println(" Database seeded successfully!")
}

func generateComments(count int, users []store.User, posts []store.Post) []*store.Comment {
	commentsList := make([]*store.Comment, count)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < count; i++ {
		commentsList[i] = &store.Comment{
			PostID:  posts[rng.Intn(len(posts))].ID,
			UserID:  int64(users[rng.Intn(len(users))].ID),
			Content: comments[rng.Intn(len(comments))],
		}
	}

	return commentsList
}

func generateUsers(i int) []store.User {
	users := make([]store.User, i)

	for j := 0; j < i; j++ {
		users[j] = store.User{
			Username: username[j%len(username)] + fmt.Sprintf("%d", j),
			Email:    email[j%len(email)] + fmt.Sprintf("%d", j),
			Role: store.Role{
				Name: "user",
			},
		}
	}
	return users
}

func generatePosts(count int, users []store.User) []store.Post {
	posts := make([]store.Post, 0, count)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < count; i++ {
		user := users[rng.Intn(len(users))]

		post := store.Post{
			UserId:    int64(user.ID),
			Title:     fmt.Sprintf("%s #%d", titles[rng.Intn(len(titles))], i+1),
			Content:   contents[rng.Intn(len(contents))],
			Version:   1,
			Tags:      tagPool[rng.Intn(len(tagPool))],
			CreatedAt: time.Now().Add(-time.Duration(rng.Intn(1000)) * time.Hour),
		}

		posts = append(posts, post)
	}

	return posts
}
