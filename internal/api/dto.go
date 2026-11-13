package api

import "github.com/william22913/clover-be/internal/store"

// movieDTO is the public shape. It carries no internal id and no filename.
type movieDTO struct {
	Slug             string  `json:"slug"`
	Title            string  `json:"title"`
	ShortDescription string  `json:"short_description"`
	ReleaseYear      int     `json:"release_year"`
	RuntimeSeconds   int     `json:"runtime_seconds"`
	PosterURL        string  `json:"poster_url"`
	StillURL         *string `json:"still_url"`
}

func toMovieDTO(m store.Movie) movieDTO {
	dto := movieDTO{
		Slug:             m.Slug,
		Title:            m.Title,
		ShortDescription: m.ShortDescription,
		ReleaseYear:      m.ReleaseYear,
		RuntimeSeconds:   m.RuntimeSeconds,
		PosterURL:        "/media/poster/" + m.PosterFile,
	}
	if m.StillFile != nil {
		still := "/media/still/" + *m.StillFile
		dto.StillURL = &still
	}
	return dto
}

type userDTO struct {
	Email string `json:"email"`
}

type userBody struct {
	User userDTO `json:"user"`
}

type movieListBody struct {
	Items []movieDTO `json:"items"`
}

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
