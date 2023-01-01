package sorting

// import all the packages used in this file
import (
	"github.com/google/go-github/github"
	"time"
)

type IssueStub struct {
	Title     string
	Body      string
	Url       string
	Labels    []github.Label
	CreatedAt time.Time
}

func Quicksort(repos []*github.Repository) []*github.Repository {
	if len(repos) < 2 {
		return repos
	}

	pivotIndex := len(repos) / 2
	pivot := repos[pivotIndex]
	repos = append(repos[:pivotIndex], repos[pivotIndex+1:]...)

	less := []*github.Repository{}
	greater := []*github.Repository{}

	for _, repo := range repos {
		if *repo.ForksCount > *pivot.ForksCount {
			less = append(less, repo)
		} else {
			greater = append(greater, repo)
		}
	}

	return append(append(Quicksort(less), pivot), Quicksort(greater)...)
}

func MergeSort(issues []IssueStub) []IssueStub {
	if len(issues) <= 1 {
		return issues
	}

	// Split the slice into two halves
	middle := len(issues) / 2
	left := MergeSort(issues[:middle])
	right := MergeSort(issues[middle:])

	// Merge the two halves
	return merge(left, right)
}

func merge(left, right []IssueStub) []IssueStub {
	result := make([]IssueStub, 0, len(left)+len(right))

	for len(left) > 0 || len(right) > 0 {
		if len(left) == 0 {
			return append(result, right...)
		}
		if len(right) == 0 {
			return append(result, left...)
		}
		if left[0].CreatedAt.After(right[0].CreatedAt) {
			result = append(result, left[0])
			left = left[1:]
		} else {
			result = append(result, right[0])
			right = right[1:]
		}
	}

	return result
}
