package repository

type ProductFilters struct{
	Page     int
	Limit    int
	Search   string
	Category string
	Sort     string
	Order    string
}