package repository

type InMemoryRepo struct {

}

func New() *InMemoryRepo {
	return &InMemoryRepo{}
}