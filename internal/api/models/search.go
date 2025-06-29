package models

// SearchPostsInput represents search posts request parameters
type SearchPostsInput struct {
	Query    string `query:"q" example:"golang concurrency" required:"true"`
	Subforum string `query:"subforum" example:"golang"`
	Author   string `query:"author" example:"user_display_name"`
	Sort     string `query:"sort" example:"relevance"` // "relevance", "hot", "top", "new", "comments"
	Time     string `query:"time" example:"all"`       // "hour", "day", "week", "month", "year", "all"
	Page     int    `query:"page" example:"1"`
	Limit    int    `query:"limit" example:"25"`
}

// SearchUsersInput represents search users request parameters
type SearchUsersInput struct {
	Query string `query:"q" example:"john" required:"true"`
	Page  int    `query:"page" example:"1"`
	Limit int    `query:"limit" example:"25"`
}

// SearchPost represents a search result post
type SearchPost struct {
	PostID       int          `json:"post_id" example:"123"`
	Title        string       `json:"title" example:"Understanding Golang Concurrency"`
	Content      string       `json:"content" example:"Post content about golang concurrency..."`
	Score        int          `json:"score" example:"1250"`
	CommentCount int          `json:"comment_count" example:"45"`
	CreatedAt    string       `json:"created_at" example:"2024-01-01T12:00:00Z"`
	Author       Author       `json:"author"`
	Subforum     SubforumInfo `json:"subforum"`
}

// SearchUser represents a search result user
type SearchUser struct {
	PseudonymID string `json:"pseudonym_id" example:"abc123def456..."`
	DisplayName string `json:"display_name" example:"john_doe"`
	KarmaScore  int    `json:"karma_score" example:"1250"`
	CreatedAt   string `json:"created_at" example:"2024-01-01T12:00:00Z"`
}

// SearchPostsResponseBody represents the body of search posts response
type SearchPostsResponseBody struct {
	Query      string       `json:"query" example:"golang concurrency"`
	Posts      []SearchPost `json:"posts"`
	Pagination Pagination   `json:"pagination"`
}

// SearchUsersResponseBody represents the body of search users response
type SearchUsersResponseBody struct {
	Query      string       `json:"query" example:"john"`
	Users      []SearchUser `json:"users"`
	Pagination Pagination   `json:"pagination"`
}

// SearchPostsResponse represents search posts response
type SearchPostsResponse struct {
	Status int                     `json:"-" example:"200"`
	Body   SearchPostsResponseBody `json:"body"`
}

// SearchUsersResponse represents search users response
type SearchUsersResponse struct {
	Status int                     `json:"-" example:"200"`
	Body   SearchUsersResponseBody `json:"body"`
}

// NewSearchPostsResponse creates a new search posts response
func NewSearchPostsResponse(query string, posts []SearchPost, page, limit, total int) *SearchPostsResponse {
	pages := (total + limit - 1) / limit // Ceiling division

	return &SearchPostsResponse{
		Status: 200,
		Body: SearchPostsResponseBody{
			Query: query,
			Posts: posts,
			Pagination: Pagination{
				Page:  page,
				Limit: limit,
				Total: total,
				Pages: pages,
			},
		},
	}
}

// NewSearchUsersResponse creates a new search users response
func NewSearchUsersResponse(query string, users []SearchUser, page, limit, total int) *SearchUsersResponse {
	pages := (total + limit - 1) / limit // Ceiling division

	return &SearchUsersResponse{
		Status: 200,
		Body: SearchUsersResponseBody{
			Query: query,
			Users: users,
			Pagination: Pagination{
				Page:  page,
				Limit: limit,
				Total: total,
				Pages: pages,
			},
		},
	}
}
