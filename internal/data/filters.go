package data

import "github.com/anukuljoshi/greenlight/internal/validator"

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
