package main

import (
	"bytes"
	"fmt"
	"math"
	"net/url"
	"strconv"
)

type Option struct {
	// DefaultPerPage is the default number of items per page.
	DefaultPerPage int

	// MaxPerPage is the maximum number of items per page.
	// If a client requests more than this number, it will be reduced to this number.
	MaxPerPage int

	// MaxNumPageNums is the maximum number of page numbers to show in the pagination.
	// e.g if numpagenums is 5, and current page is 10, the pagination will show (1 , 2 ,3 , 4 , 5, ... , 10)
	NumPageNums int

	// PageParam is the query parameter for the per page number.
	PerPageParam string

	// PageParam is the query parameter for the current page number.
	PageParam string

	// AllowAll allows the client to request all items without pagination.
	AllowAll bool

	// AllowAllParam is the query parameter to request all items without pagination.
	AllowAllParam string
}

// Paginator represents a paginator instance.
type Paginator struct {
	o Option
}

// Set represents pagination values for the query
type Set struct {
	// These value are json tagged in case they need to be embedded
	// in a struct that's sent to the outside world.
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	Total      int `json:"total"`

	// Computed values for queries.
	Offset int `json:"-"`
	Limit  int `json:"-"`

	// Fields for rendering page numbers.
	PinFirstPage bool  `json:"-"`
	PinLastPage  bool  `json:"-"`
	Pages        []int `json:"-"`
	pg           *Paginator
}

// Default returns a paginator.Opt with default values set.
func Default() Option {
	return Option{
		DefaultPerPage: 10,
		MaxPerPage:     50,
		NumPageNums:    10,
		PageParam:      "page",
		PerPageParam:   "per_page",
		AllowAll:       false,
		AllowAllParam:  "all",
	}
}

// New returns a new paginator instance.
func New(o Option) *Paginator {
	if o.AllowAllParam == "" {
		o.AllowAllParam = "all"
	}

	return &Paginator{
		o: o,
	}
}

func (p *Paginator) NewFromUrl(q url.Values) Set {
	var (
		perPage, _ = strconv.Atoi(q.Get("per_page"))
		page, _    = strconv.Atoi(q.Get("page"))
	)

	if q.Get("per_page") == p.o.AllowAllParam {
		perPage = -1
	}

	return p.New(page, perPage)
}

// New returns a new paginator set.
func (p *Paginator) New(page, perPage int) Set {
	if perPage < 0 && p.o.AllowAll {
		perPage = 0
	} else if perPage < 1 {
		perPage = p.o.DefaultPerPage
	} else if !p.o.AllowAll && perPage > p.o.MaxPerPage {
		perPage = p.o.MaxPerPage
	}

	if page < 1 {
		page = 1
	}

	return Set{
		Page:    page,
		PerPage: perPage,
		Offset:  (page - 1) * perPage,
		Limit:   perPage,
		pg:      p,
	}
}

//

func (s *Set) SetTotal(t int) {
	s.Total = t
	s.generateNumbers()
}

func (s *Set) generateNumbers() {
	if s.Total <= s.PerPage {
		return
	}

	numPages := int(math.Ceil(float64(s.Total) / float64(s.PerPage)))
	s.TotalPages = numPages
	half := s.pg.o.NumPageNums / 2

	var (
		first = s.Page - half
		last  = s.Page + half
	)

	if first < 1 {
		first = 1
	}

	if last > numPages {
		last = numPages
	}

	if numPages > s.pg.o.NumPageNums {
		if last < numPages && s.Page <= half {
			last = first + s.pg.o.NumPageNums - 1
		}
		if s.Page > numPages-half {
			first = last - s.pg.o.NumPageNums
		}
	}

	// If first in the page number series isn't 1, pin it.
	if first != 1 {
		s.PinFirstPage = true
	}

	// If last page in the page number series is not the actual last page,
	// pin it.
	if last != numPages {
		s.PinFirstPage = true
	}

	s.Pages = make([]int, 0, last-first+1)
	for i := first; i <= last; i++ {
		s.Pages = append(s.Pages, i)
	}
}

// HTML prints pagination as HTML.
func (s *Set) HTML(uri string) string {
	var b bytes.Buffer
	if s.PinFirstPage {
		b.WriteString(`<a class="pg-page-first" href="` + fmt.Sprintf(uri, 1) + `">`)
		b.WriteString("1")
		b.WriteString(`</a> `)
		b.WriteString(`<span class="pg-page-ellipsis-first">...</span> `)
	}
	for _, p := range s.Pages {
		c := ""
		if s.Page == p {
			c = " pg-selected"
		}
		b.WriteString(`<a class="pg-page` + c + `" href="` + fmt.Sprintf(uri, p) + `">`)
		b.WriteString(fmt.Sprintf("%d", p))
		b.WriteString(`</a> `)
	}
	if s.PinLastPage {
		b.WriteString(`<span class="pg-page-ellipsis-last">...</span> `)
		b.WriteString(`<a class="pg-page-last" href="` + fmt.Sprintf(uri, s.TotalPages) + `">`)
		b.WriteString(fmt.Sprintf("%d", s.TotalPages))
		b.WriteString(`</a> `)
	}
	return b.String()
}
