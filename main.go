package main

import (
	"context"
	"fmt"
	"html/template"
	"kube_a_day/sorting"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func getGoodFirstIssues(org string, cache map[string]sorting.IssueStub) []sorting.IssueStub {
	pat := os.Getenv("GITHUB_PAT")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	// Create a new GitHub client using the OAuth2 client
	client := github.NewClient(tc)

	// Set up the search criteria for repositories
	repoOpt := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 10},
	}

	// Get issues for each repository that have the label good first issue
	issueOpt := &github.IssueListByRepoOptions{
		State:       "open",
		Labels:      []string{"good first issue"},
		ListOptions: github.ListOptions{PerPage: 100},
	}

	// Initialize an empty slice to store the issues
	issues := []sorting.IssueStub{}
	// Set up a loop to retrieve all repositories in the organization
	for {
		// Perform the search for repositories
		repos, resp, err := client.Repositories.ListByOrg(context.Background(), org, repoOpt)
		if err != nil {
			log.Fatal(err)
		}
		repos = sorting.Quicksort(repos)
		// Iterate over the repositories
		for i := 0; i < len(repos); i++ {
			repo := repos[i]
			// Set up a loop to retrieve all issues for the repository
			for {
				// Perform the search for issues
				result, resp, err := client.Issues.ListByRepo(context.Background(), org, *repo.Name, issueOpt)
				if err != nil {
					log.Fatal(err)
				}

				for _, issue := range result {
					// Add the issues to the slice
					// check if this issue has pull requests associated with it
					// if it does, skip it
					// if it doesn't, add it to the List

					// check how many pull requests are associated with this issue
					if issue.PullRequestLinks == nil || *issue.PullRequestLinks.URL == "" && issue.Labels != nil && findTargetLabels(issue.Labels) {
						if val, isPresent := cache[issue.GetHTMLURL()]; isPresent {
							issues = append(issues, val)
							continue
						}
						validIssue := sorting.IssueStub{
							Title:     *issue.Title,
							Body:      issue.GetBody(),
							Url:       issue.GetHTMLURL(),
							Labels:    issue.Labels,
							CreatedAt: issue.GetCreatedAt(),
						}
						cache[validIssue.Url] = validIssue
						issues = append(issues, validIssue)
						fmt.Println(*issue.Title)
					}
				}
				// Check if there are more pages of results
				if resp.NextPage == 0 {
					break
				}
				issueOpt.Page = resp.NextPage
			}
		}
		// Check if there are more pages of results
		if resp.NextPage == 0 {
			break
		}
		repoOpt.Page = resp.NextPage
	}

	return sorting.MergeSort(issues)
}

func findTargetLabels(labels []github.Label) bool {
	for _, label := range labels {
		if label.Name != nil && *label.Name == "good first issue" {
			return true
		}
	}
	return false
}

func updateIssues(issues *[]sorting.IssueStub, cache map[string]sorting.IssueStub) {
	for {
		time.Sleep(24 * time.Hour)
		*issues = getGoodFirstIssues("kubernetes", cache)
		fmt.Println("issues updated")
	}
}

func main() {
	// k8sSigs := getGoodFirstIssues("kubernetes-sigs")
	cache := map[string]sorting.IssueStub{}
	k8s := getGoodFirstIssues("kubernetes", cache)
	// launch go routine to ocassionally update the values
	go updateIssues(&k8s, cache)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set up the data for the template
		data := struct {
			Issues []sorting.IssueStub
		}{
			Issues: k8s,
		}

		// Parse the template
		tmpl, err := template.ParseFiles("template.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Execute the template
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	fmt.Println("Hosting webserver on port :8080")
	http.ListenAndServe(":8080", nil)
}
