package service

import (
	"context"
	"strings"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

// Column limits shared by every table that snapshots an author (auditor
// assignments, Project feedback, Work Item comments).
const (
	maxAuthorNameRunes  = 200
	maxAuthorEmailRunes = domain.MaxEmailRunes
)

// resolveAuthor returns the display name and email to snapshot for actor. It
// prefers the directory profile and falls back to the token email; a failed
// lookup is not an error because the token email is always a usable label.
func resolveAuthor(ctx context.Context, directory out.ProjectDirectory, actor domain.Actor) (name, email string) {
	name, email = actor.Email, actor.Email
	if directory != nil {
		if user, err := directory.FindUserByID(ctx, actor.UserID); err == nil {
			if n := strings.TrimSpace(user.Name); n != "" {
				name = n
			}
			if e := strings.TrimSpace(user.Email); e != "" {
				email = e
			}
		}
	}
	return truncateRunes(name, maxAuthorNameRunes), truncateRunes(email, maxAuthorEmailRunes)
}

func truncateRunes(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit])
}
