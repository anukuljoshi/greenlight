package data

import (
	"math"
	"strings"

	"github.com/anukuljoshi/greenlight/internal/validator"
)

type Filters struct {
	Page int
	PageSize int
	Sort string
	// list of values that can be used for Sort
	SortSafeList []string
}

func ValidateFilters(v *validator.Validator, f Filters)  {
	// check if page is greater than 0
	v.Check(f.Page>0, "page", "must be greater than zero")
	// check if page is less than 10 mil
	v.Check(f.Page<=10_000_000, "page", "must be less than 10 million")
	// check if page_size is greater than 0
	v.Check(f.PageSize>0, "page_size", "must be greater than 0")
	// check if page_size is less than 100
	v.Check(f.PageSize<=100, "page_size", "must be less than 100")
	// check if sort value is in safe list
	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}

func (f Filters) SortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort==safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) SortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) GetOffset() int {
	return (f.Page-1)*f.PageSize
}

func (f Filters) GetLimit() int {
	return f.PageSize
}

// define a struct to hold metadata for pagination result
type Metadata struct {
	CurrentPage int `json:"current_page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
	FirstPage int `json:"first_page,omitempty"`
	LastPage int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// calculate metadata
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}
	return Metadata{
		CurrentPage: page,
		PageSize: pageSize,
		FirstPage: 1,
		LastPage: int(math.Ceil(float64(totalRecords)/float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
